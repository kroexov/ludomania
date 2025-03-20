package swift

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/vmkteam/rpcgen/v2/gen"
	"github.com/vmkteam/zenrpc/v2/smd"
)

const (
	defaultClass = "RPCAPI"

	Bool   = "Bool"
	Int    = "Int"
	Double = "Double"
	String = "String"

	DefaultBoolFalse = "False"
	DefaultBoolTrue  = "True"
	DefaultString    = "EmptyString"
	DefaultArray     = "EmptyList"
	DefaultInteger   = "IntegerZero"
	DefaultDouble    = "DoubleZero"
	DefaultMap       = "EmptyMap"
)

var (
	reservedKeywords = []string{"associatedtype", "class", "deinit", "enum", "extension", "fileprivate", "func", "import", "init", "inout", "internal", "let", "open", "operator", "private", "precedencegroup", "protocol", "public", "rethrows", "static", "struct", "subscript", "typealias", "var"}
	linebreakRegex   = regexp.MustCompile("[\r\n]+")
)

type Method struct {
	Name        string
	SafeName    string
	Description []string
	Parameters  []Parameter
	Returns     Parameter
}

type Model struct {
	Name        string
	Description string
	Fields      []Parameter
}

type Parameter struct {
	Name             string
	Description      string
	Type             string
	BaseType         string
	ReturnType       string
	Optional         bool
	DefaultValue     string
	DecodableDefault string
	IsArray          bool
	IsObject         bool
	NeedEscaping     bool
	Properties       []Parameter
}

func (p Parameter) SafeName() string {
	return fmt.Sprintf("`%s`", p.Name)
}

type TypeMapper func(typeName string, in smd.Property, swiftType Parameter) Parameter

type templateData struct {
	gen.GeneratorData
	Class   string
	Methods []Method
	Models  []Model
}

type Generator struct {
	schema smd.Schema

	settings Settings
}

type Settings struct {
	Class      string
	TypeMapper TypeMapper
}

func NewClient(schema smd.Schema, settings Settings) *Generator {
	return &Generator{schema: schema, settings: settings}
}

// Generate returns generated Swift client
func (g *Generator) Generate() ([]byte, error) {
	data := templateData{Class: defaultClass, GeneratorData: gen.DefaultGeneratorData()}
	if g.settings.Class != "" {
		data.Class = g.settings.Class
	}

	modelsMap := make(map[string]Model, 0)
	// iterate over all services
	for serviceName, service := range g.schema.Services {
		var (
			params []Parameter
		)
		desc := linebreakRegex.ReplaceAllString(service.Description, "\n")
		method := Method{
			Name:        serviceName,
			SafeName:    strings.ReplaceAll(serviceName, ".", ""),
			Description: strings.Split(desc, "\n"),
			Returns:     g.prepareParameter(service.Returns),
		}

		g.defToModelMap(modelsMap, service.Returns.Definitions)
		paramToModelMap(modelsMap, method.Returns)
		for _, param := range service.Parameters {
			p := g.prepareParameter(param)
			g.defToModelMap(modelsMap, param.Definitions)
			paramToModelMap(modelsMap, p)
			params = append(params, p)
		}
		method.Parameters = params
		data.Methods = append(data.Methods, method)
	}

	for _, v := range modelsMap {
		data.Models = append(data.Models, v)
	}

	// sort methods
	sort.Slice(data.Methods, func(i, j int) bool {
		return data.Methods[i].Name < data.Methods[j].Name
	})

	// sort models
	sort.Slice(data.Models, func(i, j int) bool {
		return data.Models[i].Name < data.Models[j].Name
	})

	// sort models fields
	for idx := range data.Models {
		sort.Slice(data.Models[idx].Fields, func(i, j int) bool {
			return data.Models[idx].Fields[i].Name < data.Models[idx].Fields[j].Name
		})
	}

	funcMap := template.FuncMap{
		"notLast": func(index int, len int) bool {
			return index+1 != len
		},
	}
	tmpl, err := template.New("swift_client").Funcs(funcMap).Parse(client)
	if err != nil {
		return nil, err
	}

	// compile template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}

// propertiesToParams convert smd.PropertyList to []Parameter
func (g *Generator) propertiesToParams(typeName string, list smd.PropertyList) []Parameter {
	var parameters []Parameter
	for _, prop := range list {
		p := Parameter{
			Name:         prop.Name,
			Optional:     prop.Optional,
			Description:  prop.Description,
			NeedEscaping: needEscaping(prop.Name),
		}

		pType := swiftType(prop.Type)
		if prop.Type == smd.Object && prop.Ref != "" {
			pType = strings.TrimPrefix(prop.Ref, gen.DefinitionsPrefix)
			p.IsObject = true
		}

		if prop.Type == smd.Array {
			pType = arrayType(prop.Items)
			p.IsArray = true
		}

		p.Type = pType

		if !prop.Optional {
			p.DecodableDefault = swiftDefault(prop.Type)
			if prop.Type == smd.Array {
				p.DecodableDefault = DefaultArray
			}
		}

		if g.settings.TypeMapper != nil {
			p = g.settings.TypeMapper(typeName, prop, p)
		}
		parameters = append(parameters, p)
	}
	return parameters
}

// prepareParameter create Parameter from smd.JSONSchema
func (g *Generator) prepareParameter(in smd.JSONSchema) Parameter {
	out := Parameter{
		Name:         in.Name,
		NeedEscaping: needEscaping(in.Name),
		Description:  in.Description,
		BaseType:     swiftType(in.Type),
		Optional:     in.Optional,
		Properties:   g.propertiesToParams(in.TypeName, in.Properties),
	}

	pType := swiftType(in.Type)
	out.ReturnType = pType
	if in.Type == smd.Object {
		typeName := in.TypeName
		if typeName == "" && in.Description != "" && smd.IsSMDTypeName(in.Description, in.Type) {
			typeName = in.Description
		}

		if typeName != "" {
			pType = typeName
			out.ReturnType = pType
		}
		out.IsObject = true
	}
	out.Type = pType

	if in.Type == smd.Array {
		out.Type = arrayType(in.Items)
		out.ReturnType = arrayType(in.Items)
		out.IsArray = true
	}

	defaultValue := ""
	if in.Default != nil {
		defaultValue = string(*in.Default)
	}
	out.DefaultValue = defaultValue

	if !in.Optional {
		out.DecodableDefault = swiftDefault(in.Type)
		if in.Type == smd.Array {
			out.DecodableDefault = DefaultArray
		}
		if in.Type == smd.Boolean && out.DefaultValue == "true" {
			out.DecodableDefault = DefaultBoolTrue
		}
	}

	return out
}

// defToModelMap convert smd.Definition to Model and add to models map
func (g *Generator) defToModelMap(modelMap map[string]Model, definitions map[string]smd.Definition) {
	for name, def := range definitions {
		modelMap[name] = Model{Name: name, Fields: g.propertiesToParams(name, def.Properties)}
	}
}

// paramToModelMap add Parameter to model map if parameter is object
func paramToModelMap(modelsMap map[string]Model, p Parameter) {
	if p.IsObject && p.BaseType != p.Type {
		modelsMap[p.Type] = Model{
			Name:   p.Type,
			Fields: p.Properties,
		}
	}
}

// swiftType convert smd types to swift types
func swiftType(smdType string) string {
	switch smdType {
	case smd.String:
		return String
	case smd.Boolean:
		return Bool
	case smd.Float:
		return Double
	case smd.Integer:
		return Int
	}
	return smdType
}

// swiftDefault return default value for swift type
func swiftDefault(smdType string) string {
	switch smdType {
	case smd.String:
		return DefaultString
	case smd.Boolean:
		return DefaultBoolFalse
	case smd.Float:
		return DefaultDouble
	case smd.Integer:
		return DefaultInteger
	}
	return ""
}

func arrayType(items map[string]string) string {
	var subType string
	if scalar, ok := items["type"]; ok {
		subType = swiftType(scalar)
	}
	if ref, ok := items["$ref"]; ok {
		subType = strings.TrimPrefix(ref, gen.DefinitionsPrefix)
	}
	return fmt.Sprintf("[%s]", subType)
}

func needEscaping(name string) bool {
	for _, keyword := range reservedKeywords {
		if name == keyword {
			return true
		}
	}
	return false
}
