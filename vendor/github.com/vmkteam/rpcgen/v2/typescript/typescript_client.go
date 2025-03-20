package typescript

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/vmkteam/rpcgen/v2/gen"
	"github.com/vmkteam/zenrpc/v2/smd"
)

const (
	interfacePrefix = "I"
	voidResponse    = "void"
	viewOps         = "viewOps"
	numberType      = "number"
)

type Generator struct {
	schema smd.Schema

	settings Settings
}

type Settings struct {
	TypeMapper        TypeMapper
	ExcludedNamespace []string
	WithClasses       bool
}

func NewClient(schema smd.Schema, settings Settings) *Generator {
	return &Generator{schema: schema, settings: settings}
}

type TypeMapper func(in smd.JSONSchema, tsType Type) Type

// Generate returns generate TypeScript client
func (g *Generator) Generate() ([]byte, error) {
	tsModels := g.tsModels()
	tsModels.GeneratorData = gen.DefaultGeneratorData()

	var fns = template.FuncMap{
		"len": func(a interface{}) int {
			return reflect.ValueOf(a).Len() - 1
		},
	}

	tmpl, err := template.New("test").Funcs(fns).Parse(client)
	if err != nil {
		return nil, err
	}

	// compile template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tsModels); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type tsInterface struct {
	Name       string
	Parameters []Type
}

type Type struct {
	Name       string
	Comment    string
	Type       string
	Optional   bool
	HasDefault bool
	Default    *string
}

type tsServiceNamespace struct {
	Name     string
	Services []tsService
}

type tsService struct {
	Namespace string
	Name      string
	NameLCF   string
	HasParams bool
	Params    string
	Response  string
}

type tsModels struct {
	gen.GeneratorData
	WithClasses bool
	Interfaces  []tsInterface
	Namespaces  []tsServiceNamespace
}

func (ts tsInterface) ModelName() string {
	return strings.TrimPrefix(ts.Name, interfacePrefix)
}

func (ts tsInterface) EntityName() string {
	return strings.TrimSuffix(nameLCF(ts.ModelName()), "Summary")
}

func (ts tsInterface) EntityNameTmpl() string {
	return strings.ToLower(ts.EntityName())
}

func (ts tsInterface) IsSearch() bool {
	entityName := ts.EntityName()

	if entityName == viewOps || strings.HasSuffix(entityName, "Search") {
		return true
	}

	return false
}

func (t Type) DefaultTmpl() string {
	result := "null"
	if t.Default != nil {
		result = *t.Default
	}

	if t.HasDefault && result == "null" {
		switch t.Type {
		case numberType:
			result = "0"
		case "string":
			result = `""`
		case "boolean":
			result = "false"
		case "Array<number>":
			result = "[0]"
		case "Array<string>":
			result = `[""]`
		case "Array<boolean>":
			result = "[false]"
		}
	}

	return result
}

// tsModels return converted schema to TypeScript.
func (g *Generator) tsModels() tsModels {
	var (
		models          tsModels
		interfacesCache = map[string]interface{}{}
	)

	// iterate over all services
skipNS:
	for serviceName, service := range g.schema.Services {
		serviceNameParts := strings.Split(serviceName, ".")
		if len(serviceNameParts) != 2 {
			continue
		}
		namespace := serviceNameParts[0]
		method := serviceNameParts[1]
		interfaceName := fmt.Sprintf("%s%s%sParams", interfacePrefix, strings.Title(namespace), strings.Title(method))

		for _, ns := range g.settings.ExcludedNamespace {
			if namespace == ns {
				continue skipNS
			}
		}

		// add service params as TypeScript interfaces
		if len(service.Parameters) > 0 {
			tsTypes := make([]Type, len(service.Parameters))
			for i := range service.Parameters {
				tsTypes[i] = convertTSType(&models, interfacesCache, service.Parameters[i], "", g.settings.TypeMapper)
			}
			addTSInterface(&models, interfacesCache, tsInterface{
				Name:       interfaceName,
				Parameters: tsTypes,
			})
		}

		// add service "returns" as TypeScript interface
		respType := convertTSType(&models, interfacesCache, service.Returns, "", g.settings.TypeMapper)

		// add namespace to TypeScript services
		nIdx := -1
		for i := range models.Namespaces {
			if models.Namespaces[i].Name == namespace {
				nIdx = i
			}
		}
		if nIdx == -1 {
			models.Namespaces = append(models.Namespaces, tsServiceNamespace{
				Name:     namespace,
				Services: nil,
			})
			nIdx = len(models.Namespaces) - 1
		}

		// add service to TypeScript services
		respService := tsService{
			Namespace: namespace,
			Name:      method,
			NameLCF:   nameLCF(method),
			HasParams: false,
			Params:    "",
			Response:  respType.Type,
		}
		if len(service.Parameters) > 0 {
			respService.HasParams = true
			respService.Params = interfaceName
		}
		if respService.Response == "" {
			respService.Response = voidResponse
		}

		models.Namespaces[nIdx].Services = append(models.Namespaces[nIdx].Services, respService)
	}

	// sort interfaces
	sort.Slice(models.Interfaces, func(i, j int) bool {
		return models.Interfaces[i].Name < models.Interfaces[j].Name
	})

	// sort namespaces
	sort.Slice(models.Namespaces, func(i, j int) bool {
		return models.Namespaces[i].Name < models.Namespaces[j].Name
	})

	// sort methods
	for idx := range models.Namespaces {
		sort.Slice(models.Namespaces[idx].Services, func(i, j int) bool {
			return models.Namespaces[idx].Services[i].Name < models.Namespaces[idx].Services[j].Name
		})
	}

	models.WithClasses = g.settings.WithClasses

	return models
}

// addTSInterface adds TypeScript interface to client.
func addTSInterface(models *tsModels, interfaces map[string]interface{}, ti tsInterface) {
	if len(ti.Parameters) == 0 {
		return
	}

	// set "HasDefault" for each parameter
	isSearch := ti.IsSearch()
	for i := range ti.Parameters {
		if isSearch || ti.Parameters[i].Type == "boolean" || ti.Parameters[i].Type == numberType {
			ti.Parameters[i].HasDefault = true
		}
	}

	// default 25 pageSize for viewOps
	if ti.EntityName() == viewOps {
		for i := range ti.Parameters {
			if ti.Parameters[i].Name == "pageSize" {
				pageSize := "25"
				ti.Parameters[i].Default = &pageSize
			}
			if ti.Parameters[i].Name == "page" {
				page := "1"
				ti.Parameters[i].Default = &page
			}
		}
	}

	if _, ok := interfaces[ti.Name]; !ok {
		models.Interfaces = append(models.Interfaces, ti)
		interfaces[ti.Name] = struct{}{}
	}
}

// convertTSScalar converts TypeScript scalars.
func convertTSScalar(t string) string {
	switch t {
	case "integer", "int":
		return "number"
	default:
		return t
	}
}

// convertTSType converts smd.JSONSchema to Type.
func convertTSType(models *tsModels, interfacesCache map[string]interface{}, in smd.JSONSchema, comment string, typeMapper TypeMapper) Type {
	result := Type{
		Name:     in.Name,
		Comment:  comment,
		Type:     convertTSScalar(in.Type),
		Optional: in.Optional,
	}

	// detect array sub type
	if in.Type == "array" {
		var subType string
		if scalar, ok := in.Items["type"]; ok {
			subType = convertTSScalar(scalar)
		}
		if ref, ok := in.Items["$ref"]; ok {
			subType = interfacePrefix + strings.TrimPrefix(ref, gen.DefinitionsPrefix)
		}

		result.Type = fmt.Sprintf("Array<%s>", subType)
	}

	// add object as complex type
	if in.Type == "object" && (in.TypeName != "" || in.Description != "") {
		if in.TypeName == "" {
			in.TypeName = in.Description
		}

		addTSComplexInterface(models, interfacesCache, in, typeMapper)
		result.Type = interfacePrefix + in.TypeName

		if in.TypeName == "time.Time" {
			result.Type = "string"
		}
	}

	// add definitions as complex types
	for name, d := range in.Definitions {
		addTSComplexInterface(models, interfacesCache, smd.JSONSchema{
			Name:        name,
			TypeName:    name,
			Description: name,
			Type:        d.Type,
			Properties:  d.Properties,
		}, typeMapper)
	}

	// apply hook
	if typeMapper != nil {
		result = typeMapper(in, result)
	}

	return result
}

// addTSComplexInterface converts complex type stored in smd1.JSONSchema to tsInterface and adds it to client.
func addTSComplexInterface(models *tsModels, interfacesCache map[string]interface{}, in smd.JSONSchema, typeMapper func(in smd.JSONSchema, tsType Type) Type) {
	var tsTypes []Type

	for _, p := range in.Properties {
		tsTypes = append(tsTypes, convertTSType(models, interfacesCache, smd.JSONSchema{
			Name:        p.Name,
			Optional:    p.Optional,
			Description: strings.TrimPrefix(p.Ref, gen.DefinitionsPrefix),
			Type:        p.Type,
			Items:       p.Items,
		}, p.Description, typeMapper))
	}

	addTSInterface(models, interfacesCache, tsInterface{
		Name:       interfacePrefix + in.TypeName,
		Parameters: tsTypes,
	})
}

// nameLCF converts "GetKits" to "getKits", "FAQ" to "FAQ"
func nameLCF(str string) string {
	if strings.ToUpper(str) == str {
		return str
	}

	for _, v := range str {
		u := string(unicode.ToLower(v))
		return u + str[len(u):]
	}
	return ""
}
