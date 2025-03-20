package openrpc

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"

	openrpc "github.com/vmkteam/meta-schema/v2"
	"github.com/vmkteam/zenrpc/v2/smd"
)

func NewSchema(schema smd.Schema, title, baseurl string) openrpc.OpenrpcDocument {
	orpc := openrpc.OpenrpcEnum0

	bs, _ := json.Marshal(schema)

	doc := openrpc.OpenrpcDocument{
		Openrpc: &orpc,
		Info: &openrpc.InfoObject{
			Title:       title,
			Description: schema.Description,
			Version:     fmt.Sprintf("v0.0.0-%x", md5.Sum(bs)),
		},
		Servers: []openrpc.ServerObject{{
			Name: sanitizeHost(baseurl),
			Url:  fmt.Sprintf("%s%s", baseurl, schema.Target),
		}},
		Methods:    newMethods(schema),
		Components: newComponents(schema),
	}

	return doc
}

func newMethods(schema smd.Schema) []openrpc.MethodOrReference {
	var methods []openrpc.MethodOrReference

	for serviceName, service := range schema.Services {
		parts := strings.Split(serviceName, ".")

		method := openrpc.MethodObject{
			Name:           serviceName,
			Tags:           []openrpc.TagOrReference{{TagObject: &openrpc.TagObject{Name: parts[0]}}},
			ParamStructure: newParamStruct(service),
			Params:         newParams(serviceName, service),
			Result:         newResult(serviceName, service),
			Errors:         newErrors(service),
			Summary:        service.Description,
		}

		methods = append(methods, openrpc.MethodOrReference{MethodObject: &method})
	}

	sort.Slice(methods, func(i, j int) bool {
		return methods[i].MethodObject.Name < methods[j].MethodObject.Name
	})

	return methods
}

func newParamStruct(service smd.Service) openrpc.MethodObjectParamStructure {
	for _, param := range service.Parameters {
		if param.Name == "" {
			return openrpc.MethodObjectParamStructureEnum0
		} else {
			return openrpc.MethodObjectParamStructureEnum1
		}
	}

	return ""
}

func newParams(serviceName string, service smd.Service) []openrpc.ContentDescriptorOrReference {
	var conts []openrpc.ContentDescriptorOrReference
	for i, param := range service.Parameters {
		name := param.Name
		if param.Name == "" {
			name = fmt.Sprintf("%s%d", serviceName, i)
		}

		cont := openrpc.ContentDescriptorObject{
			Name:     name,
			Summary:  param.Description,
			Schema:   newJSONSchema(serviceName, param),
			Required: !param.Optional,
		}

		conts = append(conts, openrpc.ContentDescriptorOrReference{ContentDescriptorObject: &cont})
	}

	return conts
}

func newResult(serviceName string, service smd.Service) *openrpc.MethodObjectResult {
	switch service.Returns.Type {
	case "":
		return &openrpc.MethodObjectResult{ReferenceObject: &openrpc.ReferenceObject{Ref: cdRefName(respName("null"))}}
	case smd.Array:
		if itemType, ok := service.Returns.Items["type"]; ok {
			return &openrpc.MethodObjectResult{ReferenceObject: &openrpc.ReferenceObject{Ref: cdRefName(arrayRespName(itemType))}}
		}
		fallthrough
	case smd.Object:

		return &openrpc.MethodObjectResult{ContentDescriptorObject: &openrpc.ContentDescriptorObject{
			Name:    objName(respName(serviceName)),
			Summary: service.Returns.Description,
			Schema:  newJSONSchema(respName(serviceName), service.Returns),
		}}
	default:
		return &openrpc.MethodObjectResult{ReferenceObject: &openrpc.ReferenceObject{Ref: cdRefName(respName(service.Returns.Type))}}
	}
}

func newErrors(service smd.Service) []openrpc.ErrorOrReference {
	var errors []openrpc.ErrorOrReference
	for c, m := range service.Errors {
		errors = append(errors, openrpc.ErrorOrReference{
			ErrorObject: &openrpc.ErrorObject{
				Code:    int64(c),
				Message: m,
			},
		})
	}

	sort.Slice(errors, func(i, j int) bool {
		return errors[i].ErrorObject.Code < errors[j].ErrorObject.Code
	})

	return errors
}

func newComponents(schema smd.Schema) *openrpc.Components {
	components := openrpc.SchemaMap{}
	descriptors := openrpc.DescriptorsMap{}

	for serviceName, service := range schema.Services {
		for _, param := range service.Parameters {
			parseComponentsFromSchema(serviceName, param, &components)
		}

		parseDescriptorsFromSchema(respName(serviceName), service.Returns, &components, &descriptors)
	}

	sort.Slice(components, func(i, j int) bool {
		if components[i].JSONSchemaObject != nil && components[j].JSONSchemaObject != nil {
			return components[i].JSONSchemaObject.Id < components[j].JSONSchemaObject.Id
		}

		return false
	})

	sort.Slice(descriptors, func(i, j int) bool {
		return descriptors[i].Name < descriptors[j].Name
	})

	return &openrpc.Components{Schemas: &components, ContentDescriptors: &descriptors}
}

// recursive hell
func parseComponentsFromSchema(serviceName string, schema smd.JSONSchema, components *openrpc.SchemaMap) {
	sch := newJSONSchema(serviceName, schema)
	if sch.JSONSchemaObject.Ref != "" && len(schema.Properties) > 0 {
		base := refBase(sch.JSONSchemaObject.Ref)

		if _, ok := components.Get(base); !ok {
			components.Add(base, *newPropertiesFromList(schema.Properties, components))
		}
	}

	if len(schema.Definitions) > 0 {
		parseComponentsFromDefinitions(schema.Definitions, components)
	}
}

func parseDescriptorsFromSchema(serviceName string, schema smd.JSONSchema, components *openrpc.SchemaMap, descriptors *openrpc.DescriptorsMap) {
	sch := newJSONSchema(serviceName, schema)

	var (
		descriptor *openrpc.ContentDescriptorObject
		component  *openrpc.JSONSchema
		base       string
	)

	switch schema.Type {
	case "":
		descriptor = newSimpleResponse("null")
		base = descriptor.Name
	case smd.Array:
		if itemType, ok := schema.Items["type"]; ok {
			descriptor = newSimpleArrayResponse(itemType)
			base = descriptor.Name
		}
	case smd.Object:
		if sch.JSONSchemaObject.Ref != "" && len(schema.Properties) > 0 {
			component = newPropertiesFromList(schema.Properties, components)
			base = refBase(sch.JSONSchemaObject.Ref)
		}
	default:
		descriptor = newSimpleResponse(schema.Type)
		base = descriptor.Name
	}

	if descriptor != nil {
		if _, ok := descriptors.Get(base); !ok {
			descriptors.Add(base, *descriptor)
		}
	}

	if component != nil {
		if _, ok := components.Get(base); !ok {
			components.Add(base, *component)
		}
	}

	if len(schema.Definitions) > 0 {
		parseComponentsFromDefinitions(schema.Definitions, components)
	}
}

func parseComponentsFromDefinitions(definitions map[string]smd.Definition, components *openrpc.SchemaMap) {
	for n, definition := range definitions {
		name := objName(n)
		if _, ok := components.Get(name); !ok {
			components.Add(name, *newPropertiesFromList(definition.Properties, components))
		}
	}
}

func newPropertiesFromList(props smd.PropertyList, components *openrpc.SchemaMap) *openrpc.JSONSchema {
	var required []string
	result := openrpc.SchemaMap{}

	for _, prop := range props {
		if len(prop.Definitions) > 0 {
			parseComponentsFromDefinitions(prop.Definitions, components)
		}

		if !prop.Optional {
			required = append(required, prop.Name)
		}

		switch prop.Type {
		case smd.Object:
			result.Add(prop.Name, openrpc.JSONSchema{JSONSchemaObject: &openrpc.JSONSchemaObject{
				Ref:         refName(prop.Ref),
				Description: prop.Description,
			}})
		case smd.Array:
			items := &openrpc.JSONSchemaObject{}
			if prop.Items["$ref"] != "" {
				items = &openrpc.JSONSchemaObject{
					Ref: refName(prop.Items["$ref"]),
				}
			} else {
				items = &openrpc.JSONSchemaObject{
					Type: &openrpc.Type{SimpleType: openrpc.SimpleType(prop.Items["type"])},
				}
			}

			result.Add(prop.Name, openrpc.JSONSchema{JSONSchemaObject: &openrpc.JSONSchemaObject{
				Description: prop.Description,
				Type:        &openrpc.Type{SimpleType: openrpc.SimpleType(prop.Type)},
				Items:       &openrpc.Items{JSONSchema: &openrpc.JSONSchema{JSONSchemaObject: items}},
			}})

		default:
			result.Add(prop.Name, openrpc.JSONSchema{JSONSchemaObject: &openrpc.JSONSchemaObject{
				Description: prop.Description,
				Type:        &openrpc.Type{SimpleType: openrpc.SimpleType(prop.Type)},
			}})
		}
	}

	return &openrpc.JSONSchema{JSONSchemaObject: &openrpc.JSONSchemaObject{
		Properties: &result,
		Required:   required,
	}}
}

func newJSONSchema(serviceName string, schema smd.JSONSchema) *openrpc.JSONSchema {

	switch schema.Type {
	case smd.Object:
		var ref string
		if schema.TypeName != "" {
			ref = refName(schema.TypeName)
		} else if isObjName(schema.Description) {
			ref = refName(schema.Description)
		} else {
			ref = refName(objName(serviceName, schema.Name))
		}

		return &openrpc.JSONSchema{JSONSchemaObject: &openrpc.JSONSchemaObject{
			Ref:         ref,
			Description: schema.Description,
		}}
	case smd.Array:
		items := &openrpc.JSONSchemaObject{}
		if schema.Items["$ref"] != "" {
			items = &openrpc.JSONSchemaObject{
				Ref: refName(schema.Items["$ref"]),
			}
		} else {
			items = &openrpc.JSONSchemaObject{
				Type: &openrpc.Type{SimpleType: openrpc.SimpleType(schema.Items["type"])},
			}
		}

		return &openrpc.JSONSchema{JSONSchemaObject: &openrpc.JSONSchemaObject{
			Type:        &openrpc.Type{SimpleType: openrpc.SimpleType(schema.Type)},
			Items:       &openrpc.Items{JSONSchema: &openrpc.JSONSchema{JSONSchemaObject: items}},
			Description: schema.Description,
		}}

	default:
		return &openrpc.JSONSchema{JSONSchemaObject: &openrpc.JSONSchemaObject{
			Type:        &openrpc.Type{SimpleType: openrpc.SimpleType(schema.Type)},
			Description: schema.Description,
		}}
	}
}

func refName(name string) string {
	if strings.HasPrefix(name, "#/definitions/") {
		name = strings.Replace(name, "#/definitions/", "", 1)
	}

	return fmt.Sprintf("#/components/schemas/%s", objName(name))
}

func cdRefName(name string) string {
	return fmt.Sprintf("#/components/contentDescriptors/%s", objName(name))
}

func refBase(ref string) string {
	return path.Base(ref)
}

func isObjName(name string) bool {
	return name != "" && objName(name) == name
}

func objName(names ...string) string {
	buf := strings.Builder{}
	for i, name := range names {
		if i == 0 {
			buf.WriteString(name)
		} else {
			buf.WriteString(strings.Title(name))
		}
	}

	return strings.Title(regexp.MustCompile(`[^a-zA-z1-9_]`).ReplaceAllString(buf.String(), ""))
}

func respName(name string) string {
	return strings.Title(fmt.Sprintf("%sResponse", name))
}

func arrayRespName(name string) string {
	return strings.Title(fmt.Sprintf("%sArrayResponse", name))
}

func newSimpleArrayResponse(typeName string) *openrpc.ContentDescriptorObject {
	return &openrpc.ContentDescriptorObject{
		Name:    arrayRespName(typeName),
		Summary: fmt.Sprintf("%s array response", typeName),
		Schema: &openrpc.JSONSchema{JSONSchemaObject: &openrpc.JSONSchemaObject{
			Items: &openrpc.Items{JSONSchema: &openrpc.JSONSchema{JSONSchemaObject: &openrpc.JSONSchemaObject{
				Type: &openrpc.Type{SimpleType: openrpc.SimpleType(typeName)},
			}}},
		}},
	}
}

func newSimpleResponse(typeName string) *openrpc.ContentDescriptorObject {
	return &openrpc.ContentDescriptorObject{
		Name:    respName(typeName),
		Summary: fmt.Sprintf("%s response", typeName),
		Schema: &openrpc.JSONSchema{JSONSchemaObject: &openrpc.JSONSchemaObject{
			Type: &openrpc.Type{SimpleType: openrpc.SimpleType(typeName)},
		}},
	}
}

func sanitizeHost(baseurl string) string {
	u, err := url.Parse(baseurl)
	if err != nil {
		return objName(baseurl)
	}

	return u.Host
}
