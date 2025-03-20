package rpcgen

import (
	smd1 "github.com/vmkteam/zenrpc/smd"
	"github.com/vmkteam/zenrpc/v2/smd"
)

// smdv1ToSMD convert smd v1 to smd v2
func smdv1ToSMD(schema smd1.Schema) smd.Schema {
	return smd.Schema{
		Transport:   schema.Transport,
		Envelope:    schema.Envelope,
		ContentType: schema.ContentType,
		SMDVersion:  schema.SMDVersion,
		Target:      schema.Target,
		Description: schema.Description,
		Services:    newSMDServicesMap(schema.Services),
	}
}

// newSMDServicesMap convert smd v1 services map to smd v2
func newSMDServicesMap(servicesMap map[string]smd1.Service) map[string]smd.Service {
	m := map[string]smd.Service{}
	for k, srv := range servicesMap {
		m[k] = newSMDService(srv)
	}

	return m
}

// newSMDService convert smd1 Service to smd2 Service
func newSMDService(service smd1.Service) smd.Service {
	return smd.Service{
		Description: service.Description,
		Parameters:  newJSONSchemas(service.Parameters),
		Returns:     newJSONSchema(service.Returns, false),
		Errors:      service.Errors,
	}
}

// newJSONSchema convert smd1 JSONSchema to smd2 JSONSchema
func newJSONSchema(smd1Schema smd1.JSONSchema, parseTypeName bool) smd.JSONSchema {
	schema := smd.JSONSchema{
		Name:        smd1Schema.Name,
		Type:        smd1Schema.Type,
		Optional:    smd1Schema.Optional,
		Default:     smd1Schema.Default,
		Description: smd1Schema.Description,
		Properties:  newPropertiesMap(smd1Schema.Properties),
		Definitions: newSmdDefinitionsMap(smd1Schema.Definitions),
		Items:       smd1Schema.Items,
	}

	if parseTypeName && smd.IsSMDTypeName(schema.Description, schema.Type) {
		schema.TypeName = smd.TypeName(schema.Description, schema.Type)
	}

	return schema
}

// newJSONSchemas convert slice of smd1 JSONSchema to slice of smd2 JSONSchema
func newJSONSchemas(schemas []smd1.JSONSchema) (res []smd.JSONSchema) {
	for _, s := range schemas {
		res = append(res, newJSONSchema(s, false))
	}

	return res
}

// newPropertiesMap convert smd1 to smd2 properties map
func newPropertiesMap(propertiesMap map[string]smd1.Property) smd.PropertyList {
	var m smd.PropertyList
	for k, prop := range propertiesMap {
		m = append(m, newSmdProperty(k, prop))
	}

	return m
}

// newSmdProperty convert smd1 property to smd2
func newSmdProperty(key string, prop smd1.Property) smd.Property {
	return smd.Property{
		Name:        key,
		Type:        prop.Type,
		Description: prop.Description,
		Items:       prop.Items,
		Definitions: newSmdDefinitionsMap(prop.Definitions),
		Ref:         prop.Ref,
	}
}

// newSmdDefinitionsMap convert smd1 to smd2 definitions map
func newSmdDefinitionsMap(definitionsMap map[string]smd1.Definition) map[string]smd.Definition {
	m := map[string]smd.Definition{}

	for k, def := range definitionsMap {
		m[k] = newSmdDefinition(def)
	}

	return m
}

// newSmdDefinition convert smd1 to smd2 definition
func newSmdDefinition(def smd1.Definition) smd.Definition {
	return smd.Definition{
		Type:       def.Type,
		Properties: newPropertiesMap(def.Properties),
	}
}
