package smd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	String  = "string"
	Integer = "integer"
	Array   = "array"
	Boolean = "boolean"
	Float   = "number"
	Object  = "object"
)

// Schema is struct for http://dojotoolkit.org/reference-guide/1.10/dojox/rpc/smd.html
// This struct doesn't implement complete specification.
type Schema struct {
	// Transport property defines the transport mechanism to be used to deliver service calls to servers.
	Transport string `json:"transport,omitempty"`

	// Envelope defines how a service message string is created from the provided parameters.
	Envelope string `json:"envelope,omitempty"`

	// ContentType is content type of the content returned by a service. Any valid MIME type is acceptable.
	// This property defaults to application/json.
	ContentType string `json:"contentType,omitempty"`

	// SMDVersion is a string that indicates the version level of the SMD being used.
	// This specification is at version level "2.0". This property SHOULD be included.
	SMDVersion string `json:"SMDVersion,omitempty"`
	// This should indicate what URL (or IP address in the case of TCP/IP transport) to use for the method call requests.
	//  A URL may be an absolute URL or a relative URL
	Target string `json:"target,omitempty"`

	// Description of the service. This property SHOULD be included.
	Description string `json:"description,omitempty"`

	// Services should be an Object value where each property in the Object represents one of the available services.
	// The property name represents the name of the service, and the value is the service description.
	// This property MUST be included.
	Services map[string]Service `json:"services"`
}

// Service is a web endpoint that can perform an action and/or return
// specific information in response to a defined network request.
type Service struct {
	Description string `json:"description"`

	// Parameters for the service calls. A parameters value MUST be an Array.
	// Each value in the parameters Array should describe a parameter and follow the JSON Schema property definition.
	// Each of parameters that are defined at the root level are inherited by each of service definition's parameters.
	Parameters []JSONSchema `json:"parameters"`

	// Returns indicates the expected type of value returned from the method call.
	// This value of this property should follow JSON Schema type definition.
	Returns JSONSchema `json:"returns"`

	// Errors describes error codes from JSON-RPC 2.0 Specification
	Errors map[int]string `json:"errors,omitempty"`
}

type JSONSchema struct {
	// Name of the parameter. If names are not provided for all the parameters,
	// this indicates positional/ordered parameter calls MUST be used.
	// If names are provided in the parameters this indicates that named parameters SHOULD be issued by
	// the client making the service call, and the server MUST support named parameters,
	// but positional parameters MAY be issued by the client and servers SHOULD support positional parameters.
	Name        string                `json:"name,omitempty"`
	Type        string                `json:"type,omitempty"`
	TypeName    string                `json:"typeName,omitempty"`
	Optional    bool                  `json:"optional,omitempty"`
	Default     *json.RawMessage      `json:"default,omitempty"`
	Description string                `json:"description,omitempty"`
	Properties  PropertyList          `json:"properties,omitempty"`
	Definitions map[string]Definition `json:"definitions,omitempty"`
	Items       map[string]string     `json:"items,omitempty"`
}

type Property struct {
	Name        string                `json:"-"`
	Type        string                `json:"type,omitempty"`
	Optional    bool                  `json:"optional,omitempty"`
	Description string                `json:"description,omitempty"`
	Items       map[string]string     `json:"items,omitempty"`
	Definitions map[string]Definition `json:"definitions,omitempty"`
	Ref         string                `json:"$ref,omitempty"`
}

type Definition struct {
	Type       string       `json:"type,omitempty"`
	Properties PropertyList `json:"properties,omitempty"`
}

type ServiceInfo struct {
	Description string
	Methods     map[string]Service
}

type PropertyList []Property

// RawMessageString returns string as *json.RawMessage.
func RawMessageString(m string) *json.RawMessage {
	r := json.RawMessage(m)
	return &r
}

// MarshalJSON implements the json.Marshaler interface for preserving map key sorting in SMD scheme.
func (pl PropertyList) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("{")
	for i, kv := range pl {
		if i != 0 {
			buf.WriteString(",")
		}
		// marshal key
		key, err := json.Marshal(kv.Name)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		// marshal value
		val, err := json.Marshal(kv)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

func (pl *PropertyList) UnmarshalJSON(data []byte) error {
	var v map[string]Property
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	var plist PropertyList
	for k, vl := range v {
		p := vl
		p.Name = k
		plist = append(plist, p)
	}
	*pl = plist
	return nil
}

var typeNameRegex = regexp.MustCompile(`[^a-zA-z1-9_]`)

func TypeName(n, t string) string {
	c := cases.Title(language.Und, cases.NoLower)
	name := c.String(typeNameRegex.ReplaceAllString(n, ""))

	if strings.EqualFold(t, Array) {
		return fmt.Sprintf("[]%s", name)
	}

	return name
}

func IsSMDTypeName(n, t string) bool {
	return n == TypeName(n, t)
}
