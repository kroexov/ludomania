package golang

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/vmkteam/rpcgen/v2/gen"
	"github.com/vmkteam/zenrpc/v2/smd"
)

// NewSchema convert smd.Schema to Schema
func NewSchema(schema smd.Schema) Schema {
	var res Schema

	for _, namespaceName := range getNamspaceNames(schema) {
		res.Namespaces = append(res.Namespaces, Namespace{
			Name:    namespaceName,
			Methods: getNamespaceMethods(schema, namespaceName),
		})
	}

	sort.Slice(res.Namespaces, func(i, j int) bool {
		return res.Namespaces[i].Name < res.Namespaces[j].Name
	})

	return res
}

type Schema struct {
	gen.GeneratorData
	Package    string
	Namespaces []Namespace
}

// NamespaceNames return all namespace names.
func (s Schema) NamespaceNames() (res []string) {
	for _, ns := range s.Namespaces {
		res = append(res, ns.Name)
	}

	return res
}

// NamespaceMethodNames return all methodNames by namespace name.
func (s Schema) NamespaceMethodNames(namespace string) (res []string) {
	ns := s.NamespaceByName(namespace)

	for _, method := range ns.Methods {
		res = append(res, method.Name)
	}

	sort.Strings(res)

	return
}

// Models return all models from all methods
func (s Schema) Models() (res []Model) {
	for _, ns := range s.Namespaces {
		for _, method := range ns.Methods {
			res = append(res, method.Models...)
		}
	}

	res = cleanModelList(res, "", "")

	// sort model fields
	for k := range res {
		k := k
		sort.Slice(res[k].Fields, func(i, j int) bool { return res[k].Fields[i].Name < res[k].Fields[j].Name })
	}

	// sort models
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})

	return res
}

func (s Schema) NamespaceByName(namespace string) Namespace {
	for _, ns := range s.Namespaces {
		if ns.Name == namespace {
			return ns
		}
	}

	return Namespace{}
}

func (s Schema) MethodByName(namespace, methodName string) Method {
	ns := s.NamespaceByName(namespace)

	for _, method := range ns.Methods {
		if method.Name == methodName {
			return method
		}
	}

	return Method{}
}

type Namespace struct {
	Name    string
	Methods []Method
}

type Method struct {
	Name string

	// Params contains all method arguments
	Params []Value

	// RPC requires single one return value
	Returns *Value

	// Models contains all models which linked from Params or Return values by ModelName
	Models []Model

	// Docs fields

	// Description describes what the method does
	Description string

	// Typed errors
	Errors []Error
}

func (m Method) HasResult() bool {
	return m.Returns != nil
}

func (m Method) HasErrors() bool {
	return len(m.Errors) > 0
}

// CommentDescription add to head of all lines two slashes
func (m Method) CommentDescription() string {
	return commentText(m.Description)
}

type Error struct {
	Code    int
	Message string
}

func (e Error) StringCode() string {
	if e.Code < 0 {
		return "_" + strconv.Itoa(e.Code*-1)
	}
	return strconv.Itoa(e.Code)
}

type Model struct {
	Name        string
	Description string

	// IsParamModel indicate model is top-level method model
	IsParamModel bool
	// ParamName from smd method param name
	ParamName string
	// IsReturnModel indicate is return model
	IsReturnModel bool

	Fields []Value
}

type Value struct {
	Name        string
	Type        string
	Optional    bool
	Description string

	ArrayItemType string

	ModelName string
}

// LocalModelName rename external packages model names: pkg.Model -> PkgModel
func (v Value) LocalModelName() string {
	return localModelName(v.ModelName)
}

// CommentDescription add to head of all lines two slashes
func (v Value) CommentDescription() string {
	return commentText(v.Description)
}

func localModelName(name string) string {
	return strings.ReplaceAll(titleFirstLetter(name), ".", "")
}

// GoType convert Value jsType to golang type
func (v *Value) GoType() string {
	if v == nil {
		return ""
	}

	if v.Type == smd.Array {
		if v.ModelName != "" {
			return fmt.Sprintf("[]%s", v.LocalModelName())
		}

		return fmt.Sprintf("[]%s", simpleGoType(v.ArrayItemType))
	} else if v.Type == smd.Object {
		return v.LocalModelName()
	}

	return simpleGoType(v.Type)
}

// simpleGoType convert js type to go type
func simpleGoType(jsType string) string {
	switch jsType {
	case "boolean":
		return "bool"
	case "number":
		return "float64"
	case "integer":
		return "int"
	default:
		return jsType
	}
}

// getNamspaceNames return all namspace names from schema.
func getNamspaceNames(schema smd.Schema) (res []string) {
	m := map[string]struct{}{}
	for name := range schema.Services {
		m[getNamespace(name)] = struct{}{}
	}

	for name := range m {
		res = append(res, name)
	}

	sort.Strings(res)

	return res
}

// getNamespaceMethods return all namespace methods
func getNamespaceMethods(schema smd.Schema, namespace string) (res []Method) {
	for name, service := range schema.Services {
		if getNamespace(name) == namespace {
			res = append(res, newMethod(service, getNamespace(name), getMethodName(name)))
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})

	return res
}

// getNamespace extract namespace from string like "namespace.method"
func getNamespace(methodName string) string {
	return strings.Split(methodName, ".")[0]
}

// newMethod convert smd.Service method to Method
func newMethod(service smd.Service, namespace, methodName string) Method {
	var args []Value

	for _, arg := range service.Parameters {
		args = append(args, newValue(arg, namespace, methodName, true, false))
	}

	var methodErrors []Error

	for code, message := range service.Errors {
		methodErrors = append(methodErrors, Error{
			Code:    code,
			Message: message,
		})
	}

	// sort for prevent smd's errors map shuffle
	sort.SliceStable(methodErrors, func(i, j int) bool {
		return methodErrors[i].Code < methodErrors[j].Code
	})

	method := Method{
		Name:        methodName,
		Description: service.Description,
		Params:      args,
		Models:      newMethodModels(service, namespace, methodName),
		Errors:      methodErrors,
	}

	if service.Returns.Type != "" {
		method.Returns = newValuePointer(service.Returns, namespace, methodName)

		// fix return model name
		if method.Returns.Type == smd.Object && method.Returns.ModelName == "" {
			method.Returns.ModelName = fmt.Sprintf("%s%sResponse", titleFirstLetter(namespace), titleFirstLetter(methodName))
		}
	}

	return method
}

// newValue convert smd.JSONSchema to value
func newValue(in smd.JSONSchema, namespace, methodName string, isParam, isReturn bool) Value {
	value := Value{
		Name:     in.Name,
		Type:     in.Type,
		Optional: in.Optional,
	}

	// in is array
	if in.Type == smd.Array {
		if in.Items["type"] != "" { // simple type
			value.ArrayItemType = in.Items["type"]
		} else { // complex type
			value.ArrayItemType = "object"
			value.ModelName = strings.TrimPrefix(in.Items["$ref"], gen.DefinitionsPrefix)
		}
	}

	// in is object
	if in.Type == smd.Object {
		if in.TypeName != "" {
			value.ModelName = in.TypeName
		} else if in.Description != "" && smd.IsSMDTypeName(in.Description, in.Type) {
			value.ModelName = in.Description
		} else if isParam {
			value.ModelName = fmt.Sprintf("%s%s%sParam", titleFirstLetter(namespace), titleFirstLetter(methodName), titleFirstLetter(in.Name))
		} else if isReturn {
			value.ModelName = fmt.Sprintf("%s%s%sResponse", titleFirstLetter(namespace), titleFirstLetter(methodName), titleFirstLetter(in.Name))
		}
	} else {
		value.Description = in.Description
	}

	return value
}

func newValuePointer(in smd.JSONSchema, namespace, methodName string) *Value {
	v := newValue(in, namespace, methodName, false, true)

	return &v
}

func newValueFromProp(in smd.Property) Value {
	value := Value{
		Name:        in.Name,
		Type:        in.Type,
		Optional:    in.Optional,
		Description: in.Description,
	}

	// in is an object
	if in.Type == smd.Object && in.Ref != "" {
		value.ModelName = strings.TrimPrefix(in.Ref, gen.DefinitionsPrefix)
	}

	// in is an array
	if in.Type == smd.Array {
		if in.Items["type"] != "" { // simple type
			value.ArrayItemType = in.Items["type"]
		} else { // complex type
			value.ArrayItemType = "object"
			value.ModelName = strings.TrimPrefix(in.Items["$ref"], gen.DefinitionsPrefix)
		}
	}

	return value
}

// newMethodModels scrape all method models
//
// Model will be created from:
// - method has argument or return with type `array` with $ref to definition. Name by definition name.
// - method has argument or return with type `object`. Name if model gather from Description field.
// - method return object has property with type `object` and $ref to definition.
// - definition has property with type `object` and $ref to definition.
//
// NOTE: definitions limited to smd.JSONSchema level, but in our Schema
// the definition will be raised to a higher level in the scope of the method.
func newMethodModels(method smd.Service, namespace, methodName string) (res []Model) {
	// get models from params
	for _, arg := range method.Parameters {
		res = append(res, newParamModels(arg, fmt.Sprintf("%s%s", titleFirstLetter(namespace), titleFirstLetter(methodName)))...)
	}

	// get models from return
	res = append(res, newReturnsModels(method.Returns)...)

	return cleanModelList(res, namespace, methodName)
}

// cleanModelList remove duplicates, add NamespaceMethod prefix for each model.
func cleanModelList(models []Model, namespace, methodName string) (res []Model) {
	var dup = map[string]struct{}{}

	for _, model := range models {
		// fix empty model names
		if model.Name == "" {
			if model.IsParamModel {
				model.Name = fmt.Sprintf("%s%s%sParam", titleFirstLetter(namespace), titleFirstLetter(methodName), titleFirstLetter(model.ParamName))
			} else if model.IsReturnModel {
				model.Name = fmt.Sprintf("%s%sResponse", titleFirstLetter(namespace), titleFirstLetter(methodName))
			}
		}

		// filter out duplicates
		if _, has := dup[model.Name]; has {
			continue
		}

		// filter out time.Time model - is not model
		if model.Name == "time.Time" {
			continue
		}

		dup[model.Name] = struct{}{}
		res = append(res, model)
	}

	return res
}

// newParamModels wrapper set isParam flag for top-level model
func newParamModels(in smd.JSONSchema, modelNamePrefix string) (res []Model) {
	return newModels(in, true, false, modelNamePrefix)
}

// newReturnsModels wrapper set isReturn flag for top-level model
func newReturnsModels(in smd.JSONSchema) (res []Model) {
	return newModels(in, false, true, "")
}

// newModels convert
func newModels(in smd.JSONSchema, isParam, isReturn bool, modelNamePrefix string) (res []Model) {
	// definitions
	for name, def := range in.Definitions {
		res = append(res, convertDefinitionToModel(def, name))
	}

	// object args and return
	if in.Type == smd.Object {
		var values []Value

		for _, prop := range in.Properties {
			values = append(values, newValueFromProp(prop))
		}

		var name string
		if isParam {
			name = modelNamePrefix + titleFirstLetter(in.Name) + "Param"
		} else if isReturn {
			name = modelNamePrefix + titleFirstLetter(in.Name) + "Response"
		}

		if in.TypeName != "" {
			name = in.TypeName
		} else if smd.IsSMDTypeName(in.Description, in.Type) {
			name = in.Description
		}

		model := Model{
			Name:          localModelName(name),
			Fields:        values,
			IsParamModel:  isParam,
			IsReturnModel: isReturn,
		}

		if isParam {
			model.ParamName = in.Name
		}

		res = append(res, model)
	}

	return res
}

// convertDefinitionToModel convert smd definition to Object.
// Definition is always an object.
func convertDefinitionToModel(def smd.Definition, name string) Model {
	model := Model{
		Name: localModelName(name),
	}

	for _, property := range def.Properties {
		model.Fields = append(model.Fields, newValueFromProp(property))
	}

	return model
}

// commentText add to head of all lines two slashes
func commentText(text string) string {
	if text == "" {
		return ""
	}

	return "// " + strings.ReplaceAll(text, "\n", "\n// ")
}
