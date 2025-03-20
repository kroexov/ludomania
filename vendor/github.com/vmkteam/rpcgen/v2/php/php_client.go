package php

import (
	"bytes"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/vmkteam/rpcgen/v2/gen"
	"github.com/vmkteam/zenrpc/v2/smd"
)

const (
	defaultPhpNamespace = "JsonRpcClient"

	phpBoolean = "bool"
	phpInt     = "int"
	phpFloat   = "float"
	phpString  = "string"
	phpObject  = "object"
	phpArray   = "array"
)

var linebreakRegex = regexp.MustCompile("[\r\n]+")

type Generator struct {
	schema       smd.Schema
	phpNamespace string
}

func NewClient(schema smd.Schema, phpNamespace string) *Generator {
	ns := phpNamespace
	if ns == "" {
		ns = defaultPhpNamespace
	}
	return &Generator{schema: schema, phpNamespace: ns}
}

// Generate returns generate TypeScript client
func (g *Generator) Generate() ([]byte, error) {
	m := g.PHPModels()
	m.GeneratorData = gen.DefaultGeneratorData()

	funcMap := template.FuncMap{
		"now": time.Now,
	}
	tmpl, err := template.New("test").Funcs(funcMap).Parse(phpTpl)
	if err != nil {
		return nil, err
	}

	// compile template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, m); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type phpMethod struct {
	Name        string
	SafeName    string
	Description []string
	Parameters  []Parameter
	Returns     Parameter
}

type phpClass struct {
	Name        string
	Description string
	Fields      []Parameter
}

type Parameter struct {
	Name         string
	Description  string
	Type         string
	BaseType     string
	ReturnType   string
	Optional     bool
	DefaultValue string
	Properties   []Parameter
}

type phpModels struct {
	gen.GeneratorData
	Namespace string
	Methods   []phpMethod
	Classes   []phpClass
}

// PHPModels return converted schema to PHP.
func (g *Generator) PHPModels() phpModels {
	var pModels phpModels

	classesMap := make(map[string]phpClass, 0)

	// iterate over all services
	for serviceName, service := range g.schema.Services {
		var (
			params []Parameter
		)
		desc := linebreakRegex.ReplaceAllString(service.Description, "\n")
		pMethod := phpMethod{
			Name:        serviceName,
			SafeName:    strings.ReplaceAll(serviceName, ".", "_"),
			Description: strings.Split(desc, "\n"),
			Returns:     prepareParameter(service.Returns),
		}

		defToClassMap(classesMap, service.Returns.Definitions)
		paramToClassMap(classesMap, pMethod.Returns)

		for _, param := range service.Parameters {
			p := prepareParameter(param)
			defToClassMap(classesMap, param.Definitions)
			paramToClassMap(classesMap, p)
			params = append(params, p)
		}
		pMethod.Parameters = params
		pModels.Methods = append(pModels.Methods, pMethod)
	}

	for _, v := range classesMap {
		pModels.Classes = append(pModels.Classes, v)
	}

	// sort methods
	sort.Slice(pModels.Methods, func(i, j int) bool {
		return pModels.Methods[i].Name < pModels.Methods[j].Name
	})

	// sort classes
	sort.Slice(pModels.Classes, func(i, j int) bool {
		return pModels.Classes[i].Name < pModels.Classes[j].Name
	})

	// sort classes fields
	for idx := range pModels.Classes {
		sort.Slice(pModels.Classes[idx].Fields, func(i, j int) bool {
			return pModels.Classes[idx].Fields[i].Name < pModels.Classes[idx].Fields[j].Name
		})
	}
	pModels.Namespace = g.phpNamespace

	return pModels
}

// prepareParameter create Parameter from smd.JSONSchema
func prepareParameter(param smd.JSONSchema) Parameter {
	p := Parameter{
		Name:        param.Name,
		Description: param.Description,
		BaseType:    phpType(param.Type),
		Optional:    param.Optional,
		Properties:  propertiesToParams(param.Properties),
	}

	pType := phpType(param.Type)
	p.ReturnType = pType
	if param.Type == smd.Object {
		typeName := param.TypeName
		if typeName == "" && param.Description != "" && smd.IsSMDTypeName(param.Description, param.Type) {
			typeName = param.Description
		}

		if typeName != "" {
			pType = typeName
			p.ReturnType = pType
		}
	}
	if param.Type == smd.Array {
		pType = arrayType(param.Items)
		p.ReturnType = phpArray
	}
	p.Type = pType

	defaultValue := ""
	if param.Default != nil {
		defaultValue = string(*param.Default)
	}
	p.DefaultValue = defaultValue

	return p
}

// propertiesToParams convert smd.PropertyList to []Parameter
func propertiesToParams(list smd.PropertyList) []Parameter {
	var parameters []Parameter
	for _, prop := range list {
		p := Parameter{
			Name:        prop.Name,
			Optional:    prop.Optional,
			Description: prop.Description,
		}
		pType := phpType(prop.Type)
		if prop.Type == smd.Object && prop.Ref != "" {
			pType = objectType(prop.Ref)
		}
		if prop.Type == smd.Array {
			pType = arrayType(prop.Items)
		}

		p.Type = pType
		parameters = append(parameters, p)

	}
	return parameters
}

// defToClassMap convert smd.Definition to phpClass and add to classes map
func defToClassMap(classesMap map[string]phpClass, definitions map[string]smd.Definition) {
	for name, def := range definitions {
		classesMap[name] = phpClass{Name: name, Fields: propertiesToParams(def.Properties)}
	}
}

// paramToClassMap add Parameter to class map if parameter is object
func paramToClassMap(classesMap map[string]phpClass, p Parameter) {
	if p.BaseType == phpObject && p.BaseType != p.Type {
		classesMap[p.Type] = phpClass{
			Name:   p.Type,
			Fields: p.Properties,
		}
	}
}

// objectType return object type from $ref
func objectType(ref string) string {
	if ref == "" {
		return phpObject
	}
	return strings.TrimPrefix(ref, gen.DefinitionsPrefix)
}

// arrayType return array type from $ref
func arrayType(ref map[string]string) string {
	if r, ok := ref["$ref"]; ok {
		return strings.TrimPrefix(r, gen.DefinitionsPrefix) + "[]"
	}
	return phpArray
}

// phpType convert smd types to php types
func phpType(smdType string) string {
	switch smdType {
	case smd.String:
		return phpString
	case smd.Array:
		return phpArray
	case smd.Boolean:
		return phpBoolean
	case smd.Float:
		return phpFloat
	case smd.Integer:
		return phpInt
	case smd.Object:
		return phpObject
	}
	return "mixed"
}
