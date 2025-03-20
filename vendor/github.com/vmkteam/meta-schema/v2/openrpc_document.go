package openrpc_document

import (
	"encoding/json"
	"errors"

	"github.com/iancoleman/orderedmap"
)

type Openrpc string

const (
	OpenrpcEnum0  Openrpc = "1.2.6"
	OpenrpcEnum1  Openrpc = "1.2.5"
	OpenrpcEnum2  Openrpc = "1.2.4"
	OpenrpcEnum3  Openrpc = "1.2.3"
	OpenrpcEnum4  Openrpc = "1.2.2"
	OpenrpcEnum5  Openrpc = "1.2.1"
	OpenrpcEnum6  Openrpc = "1.2.0"
	OpenrpcEnum7  Openrpc = "1.1.12"
	OpenrpcEnum8  Openrpc = "1.1.11"
	OpenrpcEnum9  Openrpc = "1.1.10"
	OpenrpcEnum10 Openrpc = "1.1.9"
	OpenrpcEnum11 Openrpc = "1.1.8"
	OpenrpcEnum12 Openrpc = "1.1.7"
	OpenrpcEnum13 Openrpc = "1.1.6"
	OpenrpcEnum14 Openrpc = "1.1.5"
	OpenrpcEnum15 Openrpc = "1.1.4"
	OpenrpcEnum16 Openrpc = "1.1.3"
	OpenrpcEnum17 Openrpc = "1.1.2"
	OpenrpcEnum18 Openrpc = "1.1.1"
	OpenrpcEnum19 Openrpc = "1.1.0"
	OpenrpcEnum20 Openrpc = "1.0.0"
	OpenrpcEnum21 Openrpc = "1.0.0-rc1"
	OpenrpcEnum22 Openrpc = "1.0.0-rc0"
)

type ContactObject struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Url   string `json:"url,omitempty"`
}

type LicenseObject struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}
type InfoObject struct {
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	TermsOfService string `json:"termsOfService,omitempty"`
	Version        string `json:"version"`
	Contact        string `json:"contact,omitempty"`
	License        string `json:"license,omitempty"`
}

// information about external documentation
type ExternalDocumentationObject struct {
	Description string `json:"description,omitempty"`
	Url         string `json:"url"`
}

type ServerObjectVariable struct {
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type ServerObject struct {
	Url         string                          `json:"url"`
	Name        string                          `json:"name,omitempty"`
	Description string                          `json:"description,omitempty"`
	Summary     string                          `json:"summary,omitempty"`
	Variables   map[string]ServerObjectVariable `json:"variables,omitempty"`
}

type TagObject struct {
	Name         string                       `json:"name"`
	Description  string                       `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`
}

type ReferenceObject struct {
	Ref string `json:"$ref"`
}

type TagOrReference struct {
	*TagObject
	*ReferenceObject
}

// UnmarshalJSON implements the json Unmarshaler interface.
// This implementation DOES NOT assert that ONE AND ONLY ONE
// of the simple properties is satisfied; it lazily uses the first one that is satisfied.
// Ergo, it will not return an error if more than one property is valid.
func (o *TagOrReference) UnmarshalJSON(bytes []byte) error {
	var myTagObject TagObject
	if err := json.Unmarshal(bytes, &myTagObject); err == nil {
		o.TagObject = &myTagObject
		return nil
	}
	var myReferenceObject ReferenceObject
	if err := json.Unmarshal(bytes, &myReferenceObject); err == nil {
		o.ReferenceObject = &myReferenceObject
		return nil
	}
	return errors.New("failed to unmarshal one of the object properties")
}
func (o TagOrReference) MarshalJSON() ([]byte, error) {
	if o.TagObject != nil {
		return json.Marshal(o.TagObject)
	}
	if o.ReferenceObject != nil {
		return json.Marshal(o.ReferenceObject)
	}
	return nil, errors.New("failed to marshal any one of the object properties")
}

// Format the server expects the params. Defaults to 'either'.
type MethodObjectParamStructure string

const (
	MethodObjectParamStructureEnum0 MethodObjectParamStructure = "by-position"
	MethodObjectParamStructureEnum1 MethodObjectParamStructure = "by-name"
	MethodObjectParamStructureEnum2 MethodObjectParamStructure = "either"
)

type AlwaysTrue interface{}
type SchemaMap []JSONSchema

func (m SchemaMap) MarshalJSON() ([]byte, error) {
	omap := orderedmap.New()
	for _, schema := range m {
		if schema.JSONSchemaObject != nil && schema.JSONSchemaObject.Id != "" {
			object := *schema.JSONSchemaObject
			object.Id = ""

			omap.Set(schema.JSONSchemaObject.Id, object)
		}
	}

	return json.Marshal(omap)
}

func (m *SchemaMap) UnmarshalJSON(bytes []byte) error {
	omap := orderedmap.New()
	if err := json.Unmarshal(bytes, omap); err != nil {
		return err
	}

	var v map[string]JSONSchema
	if err := json.Unmarshal(bytes, &v); err != nil {
		return err
	}

	smap := SchemaMap{}
	for _, k := range omap.Keys() {
		if p, ok := v[k]; ok {
			p.JSONSchemaObject.Id = k
			smap = append(smap, p)
		}
	}
	*m = smap

	return nil
}

func (m SchemaMap) Get(key string) (JSONSchema, bool) {
	for _, schema := range m {
		if schema.JSONSchemaObject != nil && schema.JSONSchemaObject.Id == key {
			return schema, true
		}
	}

	return JSONSchema{}, false
}

func (m *SchemaMap) Add(key string, object JSONSchema) {
	if m == nil || object.JSONSchemaObject == nil {
		return
	}

	object.JSONSchemaObject.Id = key

	*m = append(*m, object)
}

type SchemaArray []JSONSchema
type Items struct {
	*JSONSchema
	*SchemaArray
}

func (a *Items) UnmarshalJSON(bytes []byte) error {
	var ok bool
	var myJSONSchema JSONSchema
	if err := json.Unmarshal(bytes, &myJSONSchema); err == nil {
		ok = true
		a.JSONSchema = &myJSONSchema
	}
	var mySchemaArray SchemaArray
	if err := json.Unmarshal(bytes, &mySchemaArray); err == nil {
		ok = true
		a.SchemaArray = &mySchemaArray
	}
	if ok {
		return nil
	}
	return errors.New("failed to unmarshal any of the object properties")
}
func (a Items) MarshalJSON() ([]byte, error) {
	out := []interface{}{}
	if a.JSONSchema != nil {
		out = append(out, a.JSONSchema)
	}
	if a.SchemaArray != nil {
		out = append(out, a.SchemaArray)
	}
	return json.Marshal(out)
}

type StringArray []string
type DependenciesSet struct {
	*JSONSchema
	*StringArray
}

func (a *DependenciesSet) UnmarshalJSON(bytes []byte) error {
	var ok bool
	var myJSONSchema JSONSchema
	if err := json.Unmarshal(bytes, &myJSONSchema); err == nil {
		ok = true
		a.JSONSchema = &myJSONSchema
	}
	var myStringArray StringArray
	if err := json.Unmarshal(bytes, &myStringArray); err == nil {
		ok = true
		a.StringArray = &myStringArray
	}
	if ok {
		return nil
	}
	return errors.New("failed to unmarshal any of the object properties")
}
func (o DependenciesSet) MarshalJSON() ([]byte, error) {
	out := []interface{}{}
	if o.JSONSchema != nil {
		out = append(out, o.JSONSchema)
	}
	if o.StringArray != nil {
		out = append(out, o.StringArray)
	}
	return json.Marshal(out)
}

type SimpleType string

const (
	SimpleTypeArray   = "array"
	SimpleTypeBoolean = "boolean"
	SimpleTypeInteger = "integer"
	SimpleTypeNull    = "null"
	SimpleTypeNumber  = "number"
	SimpleTypeObject  = "object"
	SimpleTypeString  = "string"
)

type ArrayOfSimpleTypes []SimpleType
type Type struct {
	SimpleType
	*ArrayOfSimpleTypes
}

func (a *Type) UnmarshalJSON(bytes []byte) error {
	var ok bool
	var mySimpleTypes SimpleType
	if err := json.Unmarshal(bytes, &mySimpleTypes); err == nil {
		ok = true
		a.SimpleType = mySimpleTypes
	}
	var myArrayOfSimpleTypes ArrayOfSimpleTypes
	if err := json.Unmarshal(bytes, &myArrayOfSimpleTypes); err == nil {
		ok = true
		a.ArrayOfSimpleTypes = &myArrayOfSimpleTypes
	}
	if ok {
		return nil
	}
	return errors.New("failed to unmarshal any of the object properties")
}
func (o Type) MarshalJSON() ([]byte, error) {
	if o.SimpleType != "" {
		return json.Marshal(o.SimpleType)
	}
	out := []interface{}{}
	if o.ArrayOfSimpleTypes != nil {
		out = append(out, o.ArrayOfSimpleTypes)
	}
	return json.Marshal(out)
}

type JSONSchemaObject struct {
	Id                   string                 `json:"$id,omitempty,identifier"`
	Schema               string                 `json:"$schema,omitempty"`
	Ref                  string                 `json:"$ref,omitempty"`
	Comment              string                 `json:"$comment,omitempty"`
	Title                string                 `json:"title,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Default              *AlwaysTrue            `json:"default,omitempty"`
	ReadOnly             bool                   `json:"readOnly,omitempty"`
	Examples             []interface{}          `json:"examples,omitempty"`
	MultipleOf           float64                `json:"multipleOf,omitempty"`
	Maximum              float64                `json:"maximum,omitempty"`
	ExclusiveMaximum     float64                `json:"exclusiveMaximum,omitempty"`
	Minimum              float64                `json:"minimum,omitempty"`
	ExclusiveMinimum     float64                `json:"exclusiveMinimum,omitempty"`
	MaxLength            int64                  `json:"maxLength,omitempty"`
	MinLength            int64                  `json:"minLength,omitempty"`
	Pattern              string                 `json:"pattern,omitempty"`
	AdditionalItems      *JSONSchema            `json:"additionalItems,omitempty"`
	Items                *Items                 `json:"items,omitempty"`
	MaxItems             int64                  `json:"maxItems,omitempty"`
	MinItems             int64                  `json:"minItems,omitempty"`
	UniqueItems          bool                   `json:"uniqueItems,omitempty"`
	Contains             *JSONSchema            `json:"contains,omitempty"`
	MaxProperties        int64                  `json:"maxProperties,omitempty"`
	MinProperties        int64                  `json:"minProperties,omitempty"`
	Required             []string               `json:"required,omitempty"`
	AdditionalProperties *JSONSchema            `json:"additionalProperties,omitempty"`
	Definitions          *SchemaMap             `json:"definitions,omitempty"`
	Properties           *SchemaMap             `json:"properties,omitempty"`
	PatternProperties    *SchemaMap             `json:"patternProperties,omitempty"`
	Dependencies         map[string]interface{} `json:"dependencies,omitempty"`
	PropertyNames        *JSONSchema            `json:"propertyNames,omitempty"`
	Const                *AlwaysTrue            `json:"const,omitempty"`
	Enum                 []interface{}          `json:"enum,omitempty"`
	Type                 *Type                  `json:"type,omitempty"`
	Format               string                 `json:"format,omitempty"`
	ContentMediaType     string                 `json:"contentMediaType,omitempty"`
	ContentEncoding      string                 `json:"contentEncoding,omitempty"`
	If                   *JSONSchema            `json:"if,omitempty"`
	Then                 *JSONSchema            `json:"then,omitempty"`
	Else                 *JSONSchema            `json:"else,omitempty"`
	AllOf                []JSONSchema           `json:"allOf,omitempty"`
	AnyOf                []JSONSchema           `json:"anyOf,omitempty"`
	OneOf                []JSONSchema           `json:"oneOf,omitempty"`
	Not                  *JSONSchema            `json:"not,omitempty"`
}

// Always valid if true. Never valid if false. Is constant.
type JSONSchemaBoolean bool

//
// --- Default ---
//
// {}
type JSONSchema struct {
	*JSONSchemaObject
	*JSONSchemaBoolean
}

// UnmarshalJSON implements the json Unmarshaler interface.
// This implementation DOES NOT assert that ONE AND ONLY ONE
// of the simple properties is satisfied; it lazily uses the first one that is satisfied.
// Ergo, it will not return an error if more than one property is valid.
func (o *JSONSchema) UnmarshalJSON(bytes []byte) error {
	var myJSONSchemaObject JSONSchemaObject
	if err := json.Unmarshal(bytes, &myJSONSchemaObject); err == nil {
		o.JSONSchemaObject = &myJSONSchemaObject
		return nil
	}
	var myJSONSchemaBoolean JSONSchemaBoolean
	if err := json.Unmarshal(bytes, &myJSONSchemaBoolean); err == nil {
		o.JSONSchemaBoolean = &myJSONSchemaBoolean
		return nil
	}
	return errors.New("failed to unmarshal one of the object properties")
}

func (o JSONSchema) MarshalJSON() ([]byte, error) {
	if o.JSONSchemaObject != nil {
		return json.Marshal(o.JSONSchemaObject)
	}
	if o.JSONSchemaBoolean != nil {
		return json.Marshal(o.JSONSchemaBoolean)
	}
	return nil, errors.New("failed to marshal any one of the object properties")
}

type ContentDescriptorObject struct {
	Name        string      `json:"name,identifier"`
	Description string      `json:"description,omitempty"`
	Summary     string      `json:"summary,omitempty"`
	Schema      *JSONSchema `json:"schema"`
	Required    bool        `json:"required,omitempty"`
	Deprecated  bool        `json:"deprecated,omitempty"`
}
type ContentDescriptorOrReference struct {
	*ContentDescriptorObject
	*ReferenceObject
}

// UnmarshalJSON implements the json Unmarshaler interface.
// This implementation DOES NOT assert that ONE AND ONLY ONE
// of the simple properties is satisfied; it lazily uses the first one that is satisfied.
// Ergo, it will not return an error if more than one property is valid.
func (o *ContentDescriptorOrReference) UnmarshalJSON(bytes []byte) error {
	var myContentDescriptorObject ContentDescriptorObject
	if err := json.Unmarshal(bytes, &myContentDescriptorObject); err == nil {
		o.ContentDescriptorObject = &myContentDescriptorObject
		return nil
	}
	var myReferenceObject ReferenceObject
	if err := json.Unmarshal(bytes, &myReferenceObject); err == nil {
		o.ReferenceObject = &myReferenceObject
		return nil
	}
	return errors.New("failed to unmarshal one of the object properties")
}
func (o ContentDescriptorOrReference) MarshalJSON() ([]byte, error) {
	if o.ContentDescriptorObject != nil {
		return json.Marshal(o.ContentDescriptorObject)
	}
	if o.ReferenceObject != nil {
		return json.Marshal(o.ReferenceObject)
	}
	return nil, errors.New("failed to marshal any one of the object properties")
}

type DescriptorsMap []ContentDescriptorObject

func (m DescriptorsMap) MarshalJSON() ([]byte, error) {
	omap := orderedmap.New()
	for _, schema := range m {
		if schema.Name != "" {
			omap.Set(schema.Name, schema)
		}
	}

	return json.Marshal(omap)
}

func (m *DescriptorsMap) UnmarshalJSON(bytes []byte) error {
	omap := orderedmap.New()
	if err := json.Unmarshal(bytes, omap); err != nil {
		return err
	}

	var v map[string]ContentDescriptorObject
	if err := json.Unmarshal(bytes, &v); err != nil {
		return err
	}

	dmap := DescriptorsMap{}
	for _, k := range omap.Keys() {
		if p, ok := v[k]; ok {
			dmap = append(dmap, p)
		}
	}
	*m = dmap

	return nil
}

func (m DescriptorsMap) Get(key string) (ContentDescriptorObject, bool) {
	for _, descriptor := range m {
		if descriptor.Name == key {
			return descriptor, true
		}
	}

	return ContentDescriptorObject{}, false
}

func (m *DescriptorsMap) Add(key string, object ContentDescriptorObject) {
	if m == nil {
		return
	}

	object.Name = key

	*m = append(*m, object)
}

type MethodObjectResult struct {
	*ContentDescriptorObject
	*ReferenceObject
}

// UnmarshalJSON implements the json Unmarshaler interface.
// This implementation DOES NOT assert that ONE AND ONLY ONE
// of the simple properties is satisfied; it lazily uses the first one that is satisfied.
// Ergo, it will not return an error if more than one property is valid.
func (o *MethodObjectResult) UnmarshalJSON(bytes []byte) error {
	var myContentDescriptorObject ContentDescriptorObject
	if err := json.Unmarshal(bytes, &myContentDescriptorObject); err == nil {
		o.ContentDescriptorObject = &myContentDescriptorObject
		return nil
	}
	var myReferenceObject ReferenceObject
	if err := json.Unmarshal(bytes, &myReferenceObject); err == nil {
		o.ReferenceObject = &myReferenceObject
		return nil
	}
	return errors.New("failed to unmarshal one of the object properties")
}
func (o MethodObjectResult) MarshalJSON() ([]byte, error) {
	if o.ContentDescriptorObject != nil {
		return json.Marshal(o.ContentDescriptorObject)
	}
	if o.ReferenceObject != nil {
		return json.Marshal(o.ReferenceObject)
	}
	return nil, errors.New("failed to marshal any one of the object properties")
}

// Defines an application level error.
type ErrorObject struct {
	Code    int64       `json:"code,identifier"` // A Number that indicates the error type that occurred. This MUST be an integer. The error codes from and including -32768 to -32000 are reserved for pre-defined errors. These pre-defined errors SHOULD be assumed to be returned from any JSON-RPC api.
	Message string      `json:"message"`         // A String providing a short description of the error. The message SHOULD be limited to a concise single sentence.
	Data    interface{} `json:"data,omitempty"`  // A Primitive or Structured value that contains additional information about the error. This may be omitted. The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.).
}
type ErrorOrReference struct {
	*ErrorObject
	*ReferenceObject
}

// UnmarshalJSON implements the json Unmarshaler interface.
// This implementation DOES NOT assert that ONE AND ONLY ONE
// of the simple properties is satisfied; it lazily uses the first one that is satisfied.
// Ergo, it will not return an error if more than one property is valid.
func (o *ErrorOrReference) UnmarshalJSON(bytes []byte) error {
	var myErrorObject ErrorObject
	if err := json.Unmarshal(bytes, &myErrorObject); err == nil {
		o.ErrorObject = &myErrorObject
		return nil
	}
	var myReferenceObject ReferenceObject
	if err := json.Unmarshal(bytes, &myReferenceObject); err == nil {
		o.ReferenceObject = &myReferenceObject
		return nil
	}
	return errors.New("failed to unmarshal one of the object properties")
}
func (o ErrorOrReference) MarshalJSON() ([]byte, error) {
	if o.ErrorObject != nil {
		return json.Marshal(o.ErrorObject)
	}
	if o.ReferenceObject != nil {
		return json.Marshal(o.ReferenceObject)
	}
	return nil, errors.New("failed to marshal any one of the object properties")
}

type LinkObjectServer struct {
	Url         string                          `json:"url"`
	Name        string                          `json:"name,omitempty"`
	Description string                          `json:"description,omitempty"`
	Summary     string                          `json:"summary,omitempty"`
	Variables   map[string]ServerObjectVariable `json:"variables,omitempty"`
}
type LinkObject struct {
	Name        string            `json:"name,omitempty"`
	Summary     string            `json:"summary,omitempty"`
	Method      string            `json:"method,omitempty"`
	Description string            `json:"description,omitempty"`
	Params      interface{}       `json:"params,omitempty"`
	Server      *LinkObjectServer `json:"server,omitempty"`
}
type LinkOrReference struct {
	*LinkObject
	*ReferenceObject
}

// UnmarshalJSON implements the json Unmarshaler interface.
// This implementation DOES NOT assert that ONE AND ONLY ONE
// of the simple properties is satisfied; it lazily uses the first one that is satisfied.
// Ergo, it will not return an error if more than one property is valid.
func (o *LinkOrReference) UnmarshalJSON(bytes []byte) error {
	var myLinkObject LinkObject
	if err := json.Unmarshal(bytes, &myLinkObject); err == nil {
		o.LinkObject = &myLinkObject
		return nil
	}
	var myReferenceObject ReferenceObject
	if err := json.Unmarshal(bytes, &myReferenceObject); err == nil {
		o.ReferenceObject = &myReferenceObject
		return nil
	}
	return errors.New("failed to unmarshal one of the object properties")
}
func (o LinkOrReference) MarshalJSON() ([]byte, error) {
	if o.LinkObject != nil {
		return json.Marshal(o.LinkObject)
	}
	if o.ReferenceObject != nil {
		return json.Marshal(o.ReferenceObject)
	}
	return nil, errors.New("failed to marshal any one of the object properties")
}

type MethodObjectLinks []LinkOrReference
type ExampleObjectValue interface{}
type ExampleObject struct {
	Summary     string      `json:"summary,omitempty"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
	Name        string      `json:"name"`
}
type ExampleOrReference struct {
	*ExampleObject
	*ReferenceObject
}

// UnmarshalJSON implements the json Unmarshaler interface.
// This implementation DOES NOT assert that ONE AND ONLY ONE
// of the simple properties is satisfied; it lazily uses the first one that is satisfied.
// Ergo, it will not return an error if more than one property is valid.
func (o *ExampleOrReference) UnmarshalJSON(bytes []byte) error {
	var myExampleObject ExampleObject
	if err := json.Unmarshal(bytes, &myExampleObject); err == nil {
		o.ExampleObject = &myExampleObject
		return nil
	}
	var myReferenceObject ReferenceObject
	if err := json.Unmarshal(bytes, &myReferenceObject); err == nil {
		o.ReferenceObject = &myReferenceObject
		return nil
	}
	return errors.New("failed to unmarshal one of the object properties")
}
func (o ExampleOrReference) MarshalJSON() ([]byte, error) {
	if o.ExampleObject != nil {
		return json.Marshal(o.ExampleObject)
	}
	if o.ReferenceObject != nil {
		return json.Marshal(o.ReferenceObject)
	}
	return nil, errors.New("failed to marshal any one of the object properties")
}

type ExamplePairingObjectResult struct {
	*ExampleObject
	*ReferenceObject
}

// UnmarshalJSON implements the json Unmarshaler interface.
// This implementation DOES NOT assert that ONE AND ONLY ONE
// of the simple properties is satisfied; it lazily uses the first one that is satisfied.
// Ergo, it will not return an error if more than one property is valid.
func (o *ExamplePairingObjectResult) UnmarshalJSON(bytes []byte) error {
	var myExampleObject ExampleObject
	if err := json.Unmarshal(bytes, &myExampleObject); err == nil {
		o.ExampleObject = &myExampleObject
		return nil
	}
	var myReferenceObject ReferenceObject
	if err := json.Unmarshal(bytes, &myReferenceObject); err == nil {
		o.ReferenceObject = &myReferenceObject
		return nil
	}
	return errors.New("failed to unmarshal one of the object properties")
}
func (o ExamplePairingObjectResult) MarshalJSON() ([]byte, error) {
	if o.ExampleObject != nil {
		return json.Marshal(o.ExampleObject)
	}
	if o.ReferenceObject != nil {
		return json.Marshal(o.ReferenceObject)
	}
	return nil, errors.New("failed to marshal any one of the object properties")
}

type ExamplePairingObject struct {
	Name        string                      `json:"name"`
	Description string                      `json:"description,omitempty"`
	Params      []ExamplePairingOrReference `json:"params"`
	Result      *ExamplePairingObjectResult `json:"result"`
}
type ExamplePairingOrReference struct {
	*ExamplePairingObject
	*ReferenceObject
}

// UnmarshalJSON implements the json Unmarshaler interface.
// This implementation DOES NOT assert that ONE AND ONLY ONE
// of the simple properties is satisfied; it lazily uses the first one that is satisfied.
// Ergo, it will not return an error if more than one property is valid.
func (o *ExamplePairingOrReference) UnmarshalJSON(bytes []byte) error {
	var myExamplePairingObject ExamplePairingObject
	if err := json.Unmarshal(bytes, &myExamplePairingObject); err == nil {
		o.ExamplePairingObject = &myExamplePairingObject
		return nil
	}
	var myReferenceObject ReferenceObject
	if err := json.Unmarshal(bytes, &myReferenceObject); err == nil {
		o.ReferenceObject = &myReferenceObject
		return nil
	}
	return errors.New("failed to unmarshal one of the object properties")
}
func (o ExamplePairingOrReference) MarshalJSON() ([]byte, error) {
	if o.ExamplePairingObject != nil {
		return json.Marshal(o.ExamplePairingObject)
	}
	if o.ReferenceObject != nil {
		return json.Marshal(o.ReferenceObject)
	}
	return nil, errors.New("failed to marshal any one of the object properties")
}

type MethodObject struct {
	Name           string                         `json:"name,identifier"`       // The cannonical name for the method. The name MUST be unique within the methods array.
	Description    string                         `json:"description,omitempty"` // A verbose explanation of the method behavior. GitHub Flavored Markdown syntax MAY be used for rich text representation.
	Summary        string                         `json:"summary,omitempty"`     // A short summary of what the method does.
	Servers        []ServerObject                 `json:"servers,omitempty"`
	Tags           []TagOrReference               `json:"tags,omitempty"`
	ParamStructure MethodObjectParamStructure     `json:"paramStructure,omitempty"`
	Params         []ContentDescriptorOrReference `json:"params"`
	Result         *MethodObjectResult            `json:"result"`
	Errors         []ErrorOrReference             `json:"errors,omitempty"`
	Links          []LinkOrReference              `json:"links,omitempty"`
	Examples       []ExamplePairingOrReference    `json:"examples,omitempty"`
	Deprecated     bool                           `json:"deprecated,omitempty"`
	ExternalDocs   *ExternalDocumentationObject   `json:"externalDocs,omitempty"`
}

func (o MethodObject) MarshalJSON() ([]byte, error) {
	type aux MethodObject
	if o.Params == nil {
		o.Params = []ContentDescriptorOrReference{}
	}

	return json.Marshal(aux(o))
}

type MethodOrReference struct {
	*MethodObject
	*ReferenceObject
}

// UnmarshalJSON implements the json Unmarshaler interface.
// This implementation DOES NOT assert that ONE AND ONLY ONE
// of the simple properties is satisfied; it lazily uses the first one that is satisfied.
// Ergo, it will not return an error if more than one property is valid.
func (o *MethodOrReference) UnmarshalJSON(bytes []byte) error {
	var myMethodObject MethodObject
	if err := json.Unmarshal(bytes, &myMethodObject); err == nil {
		o.MethodObject = &myMethodObject
		return nil
	}
	var myReferenceObject ReferenceObject
	if err := json.Unmarshal(bytes, &myReferenceObject); err == nil {
		o.ReferenceObject = &myReferenceObject
		return nil
	}
	return errors.New("failed to unmarshal one of the object properties")
}
func (o MethodOrReference) MarshalJSON() ([]byte, error) {
	if o.MethodObject != nil {
		return json.Marshal(o.MethodObject)
	}
	if o.ReferenceObject != nil {
		return json.Marshal(o.ReferenceObject)
	}
	return nil, errors.New("failed to marshal any one of the object properties")
}

type Components struct {
	Schemas            *SchemaMap                      `json:"schemas,omitempty"`
	Links              map[string]LinkObject           `json:"links,omitempty"`
	Errors             map[string]ErrorObject          `json:"errors,omitempty"`
	Examples           map[string]ExampleObject        `json:"examples,omitempty"`
	ExamplePairings    map[string]ExamplePairingObject `json:"examplePairings,omitempty"`
	ContentDescriptors *DescriptorsMap                 `json:"contentDescriptors,omitempty"`
	Tags               map[string]TagObject            `json:"tags,omitempty"`
}
type OpenrpcDocument struct {
	Openrpc      *Openrpc                     `json:"openrpc"`
	Info         *InfoObject                  `json:"info"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`
	Servers      []ServerObject               `json:"servers,omitempty"`
	Methods      []MethodOrReference          `json:"methods"`
	Components   *Components                  `json:"components,omitempty"`
}

const RawOpenrpc_document = "{\"$schema\":\"https://meta.json-schema.tools/\",\"$id\":\"https://meta.open-rpc.org/\",\"title\":\"openrpcDocument\",\"type\":\"object\",\"required\":[\"info\",\"methods\",\"openrpc\"],\"additionalProperties\":false,\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}},\"properties\":{\"openrpc\":{\"$ref\":\"#/definitions/openrpc\"},\"info\":{\"$ref\":\"#/definitions/infoObject\"},\"externalDocs\":{\"$ref\":\"#/definitions/externalDocumentationObject\"},\"servers\":{\"$ref\":\"#/definitions/servers\"},\"methods\":{\"$ref\":\"#/definitions/methods\"},\"components\":{\"$ref\":\"#/definitions/components\"}},\"definitions\":{\"openrpc\":{\"title\":\"openrpc\",\"type\":\"string\",\"enum\":[\"1.2.6\",\"1.2.5\",\"1.2.4\",\"1.2.3\",\"1.2.2\",\"1.2.1\",\"1.2.0\",\"1.1.12\",\"1.1.11\",\"1.1.10\",\"1.1.9\",\"1.1.8\",\"1.1.7\",\"1.1.6\",\"1.1.5\",\"1.1.4\",\"1.1.3\",\"1.1.2\",\"1.1.1\",\"1.1.0\",\"1.0.0\",\"1.0.0-rc1\",\"1.0.0-rc0\"]},\"infoObjectProperties\":{\"title\":\"infoObjectProperties\",\"type\":\"string\"},\"infoObjectDescription\":{\"title\":\"infoObjectDescription\",\"type\":\"string\"},\"infoObjectTermsOfService\":{\"title\":\"infoObjectTermsOfService\",\"type\":\"string\",\"format\":\"uri\"},\"infoObjectVersion\":{\"title\":\"infoObjectVersion\",\"type\":\"string\"},\"contactObjectName\":{\"title\":\"contactObjectName\",\"type\":\"string\"},\"contactObjectEmail\":{\"title\":\"contactObjectEmail\",\"type\":\"string\"},\"contactObjectUrl\":{\"title\":\"contactObjectUrl\",\"type\":\"string\"},\"specificationExtension\":{\"title\":\"specificationExtension\"},\"contactObject\":{\"title\":\"contactObject\",\"type\":\"object\",\"additionalProperties\":false,\"properties\":{\"name\":{\"$ref\":\"#/definitions/contactObjectName\"},\"email\":{\"$ref\":\"#/definitions/contactObjectEmail\"},\"url\":{\"$ref\":\"#/definitions/contactObjectUrl\"}},\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}}},\"licenseObjectName\":{\"title\":\"licenseObjectName\",\"type\":\"string\"},\"licenseObjectUrl\":{\"title\":\"licenseObjectUrl\",\"type\":\"string\"},\"licenseObject\":{\"title\":\"licenseObject\",\"type\":\"object\",\"additionalProperties\":false,\"properties\":{\"name\":{\"$ref\":\"#/definitions/licenseObjectName\"},\"url\":{\"$ref\":\"#/definitions/licenseObjectUrl\"}},\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}}},\"infoObject\":{\"title\":\"infoObject\",\"type\":\"object\",\"additionalProperties\":false,\"required\":[\"title\",\"version\"],\"properties\":{\"title\":{\"$ref\":\"#/definitions/infoObjectProperties\"},\"description\":{\"$ref\":\"#/definitions/infoObjectDescription\"},\"termsOfService\":{\"$ref\":\"#/definitions/infoObjectTermsOfService\"},\"version\":{\"$ref\":\"#/definitions/infoObjectVersion\"},\"contact\":{\"$ref\":\"#/definitions/contactObject\"},\"license\":{\"$ref\":\"#/definitions/licenseObject\"}},\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}}},\"externalDocumentationObjectDescription\":{\"title\":\"externalDocumentationObjectDescription\",\"type\":\"string\"},\"externalDocumentationObjectUrl\":{\"title\":\"externalDocumentationObjectUrl\",\"type\":\"string\",\"format\":\"uri\"},\"externalDocumentationObject\":{\"title\":\"externalDocumentationObject\",\"type\":\"object\",\"additionalProperties\":false,\"description\":\"information about external documentation\",\"required\":[\"url\"],\"properties\":{\"description\":{\"$ref\":\"#/definitions/externalDocumentationObjectDescription\"},\"url\":{\"$ref\":\"#/definitions/externalDocumentationObjectUrl\"}},\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}}},\"serverObjectUrl\":{\"title\":\"serverObjectUrl\",\"type\":\"string\",\"format\":\"uri\"},\"serverObjectName\":{\"title\":\"serverObjectName\",\"type\":\"string\"},\"serverObjectDescription\":{\"title\":\"serverObjectDescription\",\"type\":\"string\"},\"serverObjectSummary\":{\"title\":\"serverObjectSummary\",\"type\":\"string\"},\"serverObjectVariableDefault\":{\"title\":\"serverObjectVariableDefault\",\"type\":\"string\"},\"serverObjectVariableDescription\":{\"title\":\"serverObjectVariableDescription\",\"type\":\"string\"},\"serverObjectVariableEnumItem\":{\"title\":\"serverObjectVariableEnumItem\",\"type\":\"string\"},\"serverObjectVariableEnum\":{\"title\":\"serverObjectVariableEnum\",\"type\":\"array\",\"items\":{\"$ref\":\"#/definitions/serverObjectVariableEnumItem\"}},\"serverObjectVariable\":{\"title\":\"serverObjectVariable\",\"type\":\"object\",\"required\":[\"default\"],\"properties\":{\"default\":{\"$ref\":\"#/definitions/serverObjectVariableDefault\"},\"description\":{\"$ref\":\"#/definitions/serverObjectVariableDescription\"},\"enum\":{\"$ref\":\"#/definitions/serverObjectVariableEnum\"}}},\"serverObjectVariables\":{\"title\":\"serverObjectVariables\",\"type\":\"object\",\"patternProperties\":{\"[0-z]+\":{\"$ref\":\"#/definitions/serverObjectVariable\"}}},\"serverObject\":{\"title\":\"serverObject\",\"type\":\"object\",\"required\":[\"url\"],\"additionalProperties\":false,\"properties\":{\"url\":{\"$ref\":\"#/definitions/serverObjectUrl\"},\"name\":{\"$ref\":\"#/definitions/serverObjectName\"},\"description\":{\"$ref\":\"#/definitions/serverObjectDescription\"},\"summary\":{\"$ref\":\"#/definitions/serverObjectSummary\"},\"variables\":{\"$ref\":\"#/definitions/serverObjectVariables\"}},\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}}},\"servers\":{\"title\":\"servers\",\"type\":\"array\",\"additionalItems\":false,\"items\":{\"$ref\":\"#/definitions/serverObject\"}},\"methodObjectName\":{\"title\":\"methodObjectName\",\"description\":\"The cannonical name for the method. The name MUST be unique within the methods array.\",\"type\":\"string\",\"minLength\":1},\"methodObjectDescription\":{\"title\":\"methodObjectDescription\",\"description\":\"A verbose explanation of the method behavior. GitHub Flavored Markdown syntax MAY be used for rich text representation.\",\"type\":\"string\"},\"methodObjectSummary\":{\"title\":\"methodObjectSummary\",\"description\":\"A short summary of what the method does.\",\"type\":\"string\"},\"tagObjectName\":{\"title\":\"tagObjectName\",\"type\":\"string\",\"minLength\":1},\"tagObjectDescription\":{\"title\":\"tagObjectDescription\",\"type\":\"string\"},\"tagObject\":{\"title\":\"tagObject\",\"type\":\"object\",\"additionalProperties\":false,\"required\":[\"name\"],\"properties\":{\"name\":{\"$ref\":\"#/definitions/tagObjectName\"},\"description\":{\"$ref\":\"#/definitions/tagObjectDescription\"},\"externalDocs\":{\"$ref\":\"#/definitions/externalDocumentationObject\"}},\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}}},\"$ref\":{\"title\":\"$ref\",\"type\":\"string\",\"format\":\"uri-reference\"},\"referenceObject\":{\"title\":\"referenceObject\",\"type\":\"object\",\"additionalProperties\":false,\"required\":[\"$ref\"],\"properties\":{\"$ref\":{\"$ref\":\"#/definitions/$ref\"}}},\"tagOrReference\":{\"title\":\"tagOrReference\",\"oneOf\":[{\"$ref\":\"#/definitions/tagObject\"},{\"$ref\":\"#/definitions/referenceObject\"}]},\"methodObjectTags\":{\"title\":\"methodObjectTags\",\"type\":\"array\",\"items\":{\"$ref\":\"#/definitions/tagOrReference\"}},\"methodObjectParamStructure\":{\"title\":\"methodObjectParamStructure\",\"type\":\"string\",\"description\":\"Format the server expects the params. Defaults to 'either'.\",\"enum\":[\"by-position\",\"by-name\",\"either\"],\"default\":\"either\"},\"contentDescriptorObjectName\":{\"title\":\"contentDescriptorObjectName\",\"type\":\"string\",\"minLength\":1},\"contentDescriptorObjectDescription\":{\"title\":\"contentDescriptorObjectDescription\",\"type\":\"string\"},\"contentDescriptorObjectSummary\":{\"title\":\"contentDescriptorObjectSummary\",\"type\":\"string\"},\"$id\":{\"title\":\"$id\",\"type\":\"string\",\"format\":\"uri-reference\"},\"$schema\":{\"title\":\"$schema\",\"type\":\"string\",\"format\":\"uri\"},\"$comment\":{\"title\":\"$comment\",\"type\":\"string\"},\"title\":{\"title\":\"title\",\"type\":\"string\"},\"description\":{\"title\":\"description\",\"type\":\"string\"},\"AlwaysTrue\":true,\"readOnly\":{\"title\":\"readOnly\",\"type\":\"boolean\",\"default\":false},\"examples\":{\"title\":\"examples\",\"type\":\"array\",\"items\":true},\"multipleOf\":{\"title\":\"multipleOf\",\"type\":\"number\",\"exclusiveMinimum\":0},\"maximum\":{\"title\":\"maximum\",\"type\":\"number\"},\"exclusiveMaximum\":{\"title\":\"exclusiveMaximum\",\"type\":\"number\"},\"minimum\":{\"title\":\"minimum\",\"type\":\"number\"},\"exclusiveMinimum\":{\"title\":\"exclusiveMinimum\",\"type\":\"number\"},\"nonNegativeInteger\":{\"title\":\"nonNegativeInteger\",\"type\":\"integer\",\"minimum\":0},\"nonNegativeIntegerDefaultZero\":{\"title\":\"nonNegativeIntegerDefaultZero\",\"type\":\"integer\",\"minimum\":0,\"default\":0},\"pattern\":{\"title\":\"pattern\",\"type\":\"string\",\"format\":\"regex\"},\"schemaArray\":{\"title\":\"schemaArray\",\"type\":\"array\",\"minItems\":1,\"items\":{\"$ref\":\"#/definitions/JSONSchema\"}},\"items\":{\"title\":\"items\",\"anyOf\":[{\"$ref\":\"#/definitions/JSONSchema\"},{\"$ref\":\"#/definitions/schemaArray\"}],\"default\":true},\"uniqueItems\":{\"title\":\"uniqueItems\",\"type\":\"boolean\",\"default\":false},\"string_doaGddGA\":{\"type\":\"string\",\"title\":\"string_doaGddGA\"},\"stringArray\":{\"title\":\"stringArray\",\"type\":\"array\",\"items\":{\"$ref\":\"#/definitions/string_doaGddGA\"},\"uniqueItems\":true,\"default\":[]},\"definitions\":{\"title\":\"definitions\",\"type\":\"object\",\"additionalProperties\":{\"$ref\":\"#/definitions/JSONSchema\"},\"default\":{}},\"properties\":{\"title\":\"properties\",\"type\":\"object\",\"additionalProperties\":{\"$ref\":\"#/definitions/JSONSchema\"},\"default\":{}},\"patternProperties\":{\"title\":\"patternProperties\",\"type\":\"object\",\"additionalProperties\":{\"$ref\":\"#/definitions/JSONSchema\"},\"propertyNames\":{\"title\":\"propertyNames\",\"format\":\"regex\"},\"default\":{}},\"dependenciesSet\":{\"title\":\"dependenciesSet\",\"anyOf\":[{\"$ref\":\"#/definitions/JSONSchema\"},{\"$ref\":\"#/definitions/stringArray\"}]},\"dependencies\":{\"title\":\"dependencies\",\"type\":\"object\",\"additionalProperties\":{\"$ref\":\"#/definitions/dependenciesSet\"}},\"enum\":{\"title\":\"enum\",\"type\":\"array\",\"items\":true,\"minItems\":1,\"uniqueItems\":true},\"simpleTypes\":{\"title\":\"simpleTypes\",\"enum\":[\"array\",\"boolean\",\"integer\",\"null\",\"number\",\"object\",\"string\"]},\"arrayOfSimpleTypes\":{\"title\":\"arrayOfSimpleTypes\",\"type\":\"array\",\"items\":{\"$ref\":\"#/definitions/simpleTypes\"},\"minItems\":1,\"uniqueItems\":true},\"type\":{\"title\":\"type\",\"anyOf\":[{\"$ref\":\"#/definitions/simpleTypes\"},{\"$ref\":\"#/definitions/arrayOfSimpleTypes\"}]},\"format\":{\"title\":\"format\",\"type\":\"string\"},\"contentMediaType\":{\"title\":\"contentMediaType\",\"type\":\"string\"},\"contentEncoding\":{\"title\":\"contentEncoding\",\"type\":\"string\"},\"JSONSchemaObject\":{\"title\":\"JSONSchemaObject\",\"type\":\"object\",\"properties\":{\"$id\":{\"$ref\":\"#/definitions/$id\"},\"$schema\":{\"$ref\":\"#/definitions/$schema\"},\"$ref\":{\"$ref\":\"#/definitions/$ref\"},\"$comment\":{\"$ref\":\"#/definitions/$comment\"},\"title\":{\"$ref\":\"#/definitions/title\"},\"description\":{\"$ref\":\"#/definitions/description\"},\"default\":true,\"readOnly\":{\"$ref\":\"#/definitions/readOnly\"},\"examples\":{\"$ref\":\"#/definitions/examples\"},\"multipleOf\":{\"$ref\":\"#/definitions/multipleOf\"},\"maximum\":{\"$ref\":\"#/definitions/maximum\"},\"exclusiveMaximum\":{\"$ref\":\"#/definitions/exclusiveMaximum\"},\"minimum\":{\"$ref\":\"#/definitions/minimum\"},\"exclusiveMinimum\":{\"$ref\":\"#/definitions/exclusiveMinimum\"},\"maxLength\":{\"$ref\":\"#/definitions/nonNegativeInteger\"},\"minLength\":{\"$ref\":\"#/definitions/nonNegativeIntegerDefaultZero\"},\"pattern\":{\"$ref\":\"#/definitions/pattern\"},\"additionalItems\":{\"$ref\":\"#/definitions/JSONSchema\"},\"items\":{\"$ref\":\"#/definitions/items\"},\"maxItems\":{\"$ref\":\"#/definitions/nonNegativeInteger\"},\"minItems\":{\"$ref\":\"#/definitions/nonNegativeIntegerDefaultZero\"},\"uniqueItems\":{\"$ref\":\"#/definitions/uniqueItems\"},\"contains\":{\"$ref\":\"#/definitions/JSONSchema\"},\"maxProperties\":{\"$ref\":\"#/definitions/nonNegativeInteger\"},\"minProperties\":{\"$ref\":\"#/definitions/nonNegativeIntegerDefaultZero\"},\"required\":{\"$ref\":\"#/definitions/stringArray\"},\"additionalProperties\":{\"$ref\":\"#/definitions/JSONSchema\"},\"definitions\":{\"$ref\":\"#/definitions/definitions\"},\"properties\":{\"$ref\":\"#/definitions/properties\"},\"patternProperties\":{\"$ref\":\"#/definitions/patternProperties\"},\"dependencies\":{\"$ref\":\"#/definitions/dependencies\"},\"propertyNames\":{\"$ref\":\"#/definitions/JSONSchema\"},\"const\":true,\"enum\":{\"$ref\":\"#/definitions/enum\"},\"type\":{\"$ref\":\"#/definitions/type\"},\"format\":{\"$ref\":\"#/definitions/format\"},\"contentMediaType\":{\"$ref\":\"#/definitions/contentMediaType\"},\"contentEncoding\":{\"$ref\":\"#/definitions/contentEncoding\"},\"if\":{\"$ref\":\"#/definitions/JSONSchema\"},\"then\":{\"$ref\":\"#/definitions/JSONSchema\"},\"else\":{\"$ref\":\"#/definitions/JSONSchema\"},\"allOf\":{\"$ref\":\"#/definitions/schemaArray\"},\"anyOf\":{\"$ref\":\"#/definitions/schemaArray\"},\"oneOf\":{\"$ref\":\"#/definitions/schemaArray\"},\"not\":{\"$ref\":\"#/definitions/JSONSchema\"}}},\"JSONSchemaBoolean\":{\"title\":\"JSONSchemaBoolean\",\"description\":\"Always valid if true. Never valid if false. Is constant.\",\"type\":\"boolean\"},\"JSONSchema\":{\"$schema\":\"https://meta.json-schema.tools/\",\"$id\":\"https://meta.json-schema.tools/\",\"title\":\"JSONSchema\",\"default\":{},\"oneOf\":[{\"$ref\":\"#/definitions/JSONSchemaObject\"},{\"$ref\":\"#/definitions/JSONSchemaBoolean\"}]},\"contentDescriptorObjectRequired\":{\"title\":\"contentDescriptorObjectRequired\",\"type\":\"boolean\",\"default\":false},\"contentDescriptorObjectDeprecated\":{\"title\":\"contentDescriptorObjectDeprecated\",\"type\":\"boolean\",\"default\":false},\"contentDescriptorObject\":{\"title\":\"contentDescriptorObject\",\"type\":\"object\",\"additionalProperties\":false,\"required\":[\"name\",\"schema\"],\"properties\":{\"name\":{\"$ref\":\"#/definitions/contentDescriptorObjectName\"},\"description\":{\"$ref\":\"#/definitions/contentDescriptorObjectDescription\"},\"summary\":{\"$ref\":\"#/definitions/contentDescriptorObjectSummary\"},\"schema\":{\"$ref\":\"#/definitions/JSONSchema\"},\"required\":{\"$ref\":\"#/definitions/contentDescriptorObjectRequired\"},\"deprecated\":{\"$ref\":\"#/definitions/contentDescriptorObjectDeprecated\"}},\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}}},\"contentDescriptorOrReference\":{\"title\":\"contentDescriptorOrReference\",\"oneOf\":[{\"$ref\":\"#/definitions/contentDescriptorObject\"},{\"$ref\":\"#/definitions/referenceObject\"}]},\"methodObjectParams\":{\"title\":\"methodObjectParams\",\"type\":\"array\",\"items\":{\"$ref\":\"#/definitions/contentDescriptorOrReference\"}},\"methodObjectResult\":{\"title\":\"methodObjectResult\",\"oneOf\":[{\"$ref\":\"#/definitions/contentDescriptorObject\"},{\"$ref\":\"#/definitions/referenceObject\"}]},\"errorObjectCode\":{\"title\":\"errorObjectCode\",\"description\":\"A Number that indicates the error type that occurred. This MUST be an integer. The error codes from and including -32768 to -32000 are reserved for pre-defined errors. These pre-defined errors SHOULD be assumed to be returned from any JSON-RPC api.\",\"type\":\"integer\"},\"errorObjectMessage\":{\"title\":\"errorObjectMessage\",\"description\":\"A String providing a short description of the error. The message SHOULD be limited to a concise single sentence.\",\"type\":\"string\"},\"errorObjectData\":{\"title\":\"errorObjectData\",\"description\":\"A Primitive or Structured value that contains additional information about the error. This may be omitted. The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.).\"},\"errorObject\":{\"title\":\"errorObject\",\"type\":\"object\",\"description\":\"Defines an application level error.\",\"additionalProperties\":false,\"required\":[\"code\",\"message\"],\"properties\":{\"code\":{\"$ref\":\"#/definitions/errorObjectCode\"},\"message\":{\"$ref\":\"#/definitions/errorObjectMessage\"},\"data\":{\"$ref\":\"#/definitions/errorObjectData\"}}},\"errorOrReference\":{\"title\":\"errorOrReference\",\"oneOf\":[{\"$ref\":\"#/definitions/errorObject\"},{\"$ref\":\"#/definitions/referenceObject\"}]},\"methodObjectErrors\":{\"title\":\"methodObjectErrors\",\"description\":\"Defines an application level error.\",\"type\":\"array\",\"items\":{\"$ref\":\"#/definitions/errorOrReference\"}},\"linkObjectName\":{\"title\":\"linkObjectName\",\"type\":\"string\",\"minLength\":1},\"linkObjectSummary\":{\"title\":\"linkObjectSummary\",\"type\":\"string\"},\"linkObjectMethod\":{\"title\":\"linkObjectMethod\",\"type\":\"string\"},\"linkObjectDescription\":{\"title\":\"linkObjectDescription\",\"type\":\"string\"},\"linkObjectParams\":{\"title\":\"linkObjectParams\"},\"linkObjectServer\":{\"title\":\"linkObjectServer\",\"type\":\"object\",\"required\":[\"url\"],\"additionalProperties\":false,\"properties\":{\"url\":{\"$ref\":\"#/definitions/serverObjectUrl\"},\"name\":{\"$ref\":\"#/definitions/serverObjectName\"},\"description\":{\"$ref\":\"#/definitions/serverObjectDescription\"},\"summary\":{\"$ref\":\"#/definitions/serverObjectSummary\"},\"variables\":{\"$ref\":\"#/definitions/serverObjectVariables\"}},\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}}},\"linkObject\":{\"title\":\"linkObject\",\"type\":\"object\",\"additionalProperties\":false,\"properties\":{\"name\":{\"$ref\":\"#/definitions/linkObjectName\"},\"summary\":{\"$ref\":\"#/definitions/linkObjectSummary\"},\"method\":{\"$ref\":\"#/definitions/linkObjectMethod\"},\"description\":{\"$ref\":\"#/definitions/linkObjectDescription\"},\"params\":{\"$ref\":\"#/definitions/linkObjectParams\"},\"server\":{\"$ref\":\"#/definitions/linkObjectServer\"}},\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}}},\"linkOrReference\":{\"title\":\"linkOrReference\",\"oneOf\":[{\"$ref\":\"#/definitions/linkObject\"},{\"$ref\":\"#/definitions/referenceObject\"}]},\"methodObjectLinks\":{\"title\":\"methodObjectLinks\",\"type\":\"array\",\"items\":{\"$ref\":\"#/definitions/linkOrReference\"}},\"examplePairingObjectName\":{\"title\":\"examplePairingObjectName\",\"type\":\"string\",\"minLength\":1},\"examplePairingObjectDescription\":{\"title\":\"examplePairingObjectDescription\",\"type\":\"string\"},\"exampleObjectSummary\":{\"title\":\"exampleObjectSummary\",\"type\":\"string\"},\"exampleObjectValue\":{\"title\":\"exampleObjectValue\"},\"exampleObjectDescription\":{\"title\":\"exampleObjectDescription\",\"type\":\"string\"},\"exampleObjectName\":{\"title\":\"exampleObjectName\",\"type\":\"string\",\"minLength\":1},\"exampleObject\":{\"title\":\"exampleObject\",\"type\":\"object\",\"required\":[\"name\",\"value\"],\"properties\":{\"summary\":{\"$ref\":\"#/definitions/exampleObjectSummary\"},\"value\":{\"$ref\":\"#/definitions/exampleObjectValue\"},\"description\":{\"$ref\":\"#/definitions/exampleObjectDescription\"},\"name\":{\"$ref\":\"#/definitions/exampleObjectName\"}},\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}}},\"exampleOrReference\":{\"title\":\"exampleOrReference\",\"oneOf\":[{\"$ref\":\"#/definitions/exampleObject\"},{\"$ref\":\"#/definitions/referenceObject\"}]},\"examplePairingObjectParams\":{\"title\":\"examplePairingObjectParams\",\"type\":\"array\",\"items\":{\"$ref\":\"#/definitions/exampleOrReference\"}},\"examplePairingObjectResult\":{\"title\":\"examplePairingObjectResult\",\"oneOf\":[{\"$ref\":\"#/definitions/exampleObject\"},{\"$ref\":\"#/definitions/referenceObject\"}]},\"examplePairingObject\":{\"title\":\"examplePairingObject\",\"type\":\"object\",\"required\":[\"name\",\"params\",\"result\"],\"properties\":{\"name\":{\"$ref\":\"#/definitions/examplePairingObjectName\"},\"description\":{\"$ref\":\"#/definitions/examplePairingObjectDescription\"},\"params\":{\"$ref\":\"#/definitions/examplePairingObjectParams\"},\"result\":{\"$ref\":\"#/definitions/examplePairingObjectResult\"}}},\"examplePairingOrReference\":{\"title\":\"examplePairingOrReference\",\"oneOf\":[{\"$ref\":\"#/definitions/examplePairingObject\"},{\"$ref\":\"#/definitions/referenceObject\"}]},\"methodObjectExamples\":{\"title\":\"methodObjectExamples\",\"type\":\"array\",\"items\":{\"$ref\":\"#/definitions/examplePairingOrReference\"}},\"methodObjectDeprecated\":{\"title\":\"methodObjectDeprecated\",\"type\":\"boolean\",\"default\":false},\"methodObject\":{\"title\":\"methodObject\",\"type\":\"object\",\"required\":[\"name\",\"result\",\"params\"],\"additionalProperties\":false,\"properties\":{\"name\":{\"$ref\":\"#/definitions/methodObjectName\"},\"description\":{\"$ref\":\"#/definitions/methodObjectDescription\"},\"summary\":{\"$ref\":\"#/definitions/methodObjectSummary\"},\"servers\":{\"$ref\":\"#/definitions/servers\"},\"tags\":{\"$ref\":\"#/definitions/methodObjectTags\"},\"paramStructure\":{\"$ref\":\"#/definitions/methodObjectParamStructure\"},\"params\":{\"$ref\":\"#/definitions/methodObjectParams\"},\"result\":{\"$ref\":\"#/definitions/methodObjectResult\"},\"errors\":{\"$ref\":\"#/definitions/methodObjectErrors\"},\"links\":{\"$ref\":\"#/definitions/methodObjectLinks\"},\"examples\":{\"$ref\":\"#/definitions/methodObjectExamples\"},\"deprecated\":{\"$ref\":\"#/definitions/methodObjectDeprecated\"},\"externalDocs\":{\"$ref\":\"#/definitions/externalDocumentationObject\"}},\"patternProperties\":{\"^x-\":{\"$ref\":\"#/definitions/specificationExtension\"}}},\"methodOrReference\":{\"title\":\"methodOrReference\",\"oneOf\":[{\"$ref\":\"#/definitions/methodObject\"},{\"$ref\":\"#/definitions/referenceObject\"}]},\"methods\":{\"title\":\"methods\",\"type\":\"array\",\"additionalItems\":false,\"items\":{\"$ref\":\"#/definitions/methodOrReference\"}},\"schemaComponents\":{\"title\":\"schemaComponents\",\"type\":\"object\",\"patternProperties\":{\"[0-z]+\":{\"$ref\":\"#/definitions/JSONSchema\"}}},\"linkComponents\":{\"title\":\"linkComponents\",\"type\":\"object\",\"patternProperties\":{\"[0-z]+\":{\"$ref\":\"#/definitions/linkObject\"}}},\"errorComponents\":{\"title\":\"errorComponents\",\"type\":\"object\",\"patternProperties\":{\"[0-z]+\":{\"$ref\":\"#/definitions/errorObject\"}}},\"exampleComponents\":{\"title\":\"exampleComponents\",\"type\":\"object\",\"patternProperties\":{\"[0-z]+\":{\"$ref\":\"#/definitions/exampleObject\"}}},\"examplePairingComponents\":{\"title\":\"examplePairingComponents\",\"type\":\"object\",\"patternProperties\":{\"[0-z]+\":{\"$ref\":\"#/definitions/examplePairingObject\"}}},\"contentDescriptorComponents\":{\"title\":\"contentDescriptorComponents\",\"type\":\"object\",\"patternProperties\":{\"[0-z]+\":{\"$ref\":\"#/definitions/contentDescriptorObject\"}}},\"tagComponents\":{\"title\":\"tagComponents\",\"type\":\"object\",\"patternProperties\":{\"[0-z]+\":{\"$ref\":\"#/definitions/tagObject\"}}},\"components\":{\"title\":\"components\",\"type\":\"object\",\"properties\":{\"schemas\":{\"$ref\":\"#/definitions/schemaComponents\"},\"links\":{\"$ref\":\"#/definitions/linkComponents\"},\"errors\":{\"$ref\":\"#/definitions/errorComponents\"},\"examples\":{\"$ref\":\"#/definitions/exampleComponents\"},\"examplePairings\":{\"$ref\":\"#/definitions/examplePairingComponents\"},\"contentDescriptors\":{\"$ref\":\"#/definitions/contentDescriptorComponents\"},\"tags\":{\"$ref\":\"#/definitions/tagComponents\"}}}}}"
