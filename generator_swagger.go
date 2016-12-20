package schreder

import (
	"encoding/json"
	"fmt"

	"github.com/alecthomas/jsonschema"
	"github.com/ghodss/yaml"
	"github.com/go-openapi/spec"
)

type MarshallerFunc func(obj interface{}) ([]byte, error)

type swaggerGenerator struct {
	seed       spec.Swagger
	marshaller MarshallerFunc
}

// NewSwaggerGeneratorYAML initializes new generator with initial swagger spec
// as a seed. The generator produces YAML output
func NewSwaggerGeneratorYAML(seed spec.Swagger) IDocGenerator {
	return NewSwaggerGenerator(seed, yaml.Marshal)
}

// NewSwaggerGeneratorJSON initializes new generator with initial swagger spec
// as a seed. The generator produces JSON output with no indentation
func NewSwaggerGeneratorJSON(seed spec.Swagger) IDocGenerator {
	return NewSwaggerGenerator(seed, json.Marshal)
}

// NewSwaggerGeneratorJSONIndent initializes new generator with initial swagger spec
// as a seed. The generator produces indented JSON output
func NewSwaggerGeneratorJSONIndent(seed spec.Swagger) IDocGenerator {
	return NewSwaggerGenerator(seed, func(obj interface{}) ([]byte, error) {
		return json.MarshalIndent(obj, "", "    ")
	})
}

// NewSwaggerGenerator creates a new instance of Swagger generator with
// given marshaller (may be JSON marshaller or YAML marshaller or whatever)
func NewSwaggerGenerator(seed spec.Swagger, marshaller MarshallerFunc) IDocGenerator {
	gen := &swaggerGenerator{
		seed:       seed,
		marshaller: marshaller,
	}
	gen.seed.Swagger = "2.0" // from swagger doc: 'The value MUST be "2.0"'
	gen.seed.Paths = &spec.Paths{Paths: map[string]spec.PathItem{}}

	return gen
}

// Generate implements IDocGenerator
// TODO: is there any way to control swagger generator? I don't need it to analyze anonymous fields, I want to expand them
func (g *swaggerGenerator) Generate(tests []Test) ([]byte, error) {
	doc := g.seed
	doc.Definitions = spec.Definitions{}

	for _, test := range tests {
		path := doc.Paths.Paths[test.Path()] // TODO: 2 tests on the same API with the same response code conflict
		op, err := g.generateSwaggerOperation(test, doc.Definitions)
		if err != nil {
			return nil, err
		}

		// TODO: check if path has already assigned an operation to some other test
		// return error if so
		switch test.Method() {
		case "GET":
			path.Get = &op
		case "POST":
			path.Post = &op
		case "PATCH":
			path.Patch = &op
		case "DELETE":
			path.Delete = &op
		case "PUT":
			path.Put = &op
		case "HEAD":
			path.Head = &op
		case "OPTIONS":
			path.Options = &op
		}

		doc.Paths.Paths[test.Path()] = path
	}

	d, e := g.marshaller(doc)

	return d, e
}

func (g *swaggerGenerator) generateSwaggerOperation(test Test, defs spec.Definitions) (spec.Operation, error) {

	op := spec.Operation{}
	op.Responses = &spec.Responses{}
	op.Responses.StatusCodeResponses = map[int]spec.Response{}

	var description string
	processedQueryParams := map[string]interface{}{}
	processedPathParams := map[string]interface{}{}
	processedHeaderParams := map[string]interface{}{}
	for _, testCase := range test.TestCases() {
		// parameter definitions are collected from 2xx tests only
		if testCase.ExpectedHttpCode >= 200 && testCase.ExpectedHttpCode < 300 {
			description = testCase.Description

			for key, param := range testCase.Headers {
				if _, ok := processedHeaderParams[key]; ok {
					continue
				}

				specParam, err := generateSwaggerSpecParam(key, param, "header")
				if err != nil {
					return op, err
				}

				processedHeaderParams[key] = nil
				op.Parameters = append(op.Parameters, specParam)
			}

			for key, param := range testCase.PathParams {
				if _, ok := processedPathParams[key]; ok {
					continue
				}
				param.Required = true // path parameters are always required
				specParam, err := generateSwaggerSpecParam(key, param, "path")
				if err != nil {
					return op, err
				}

				processedPathParams[key] = nil
				op.Parameters = append(op.Parameters, specParam)
			}

			for key, param := range testCase.QueryParams {

				if _, ok := processedQueryParams[key]; ok {
					continue
				}

				specParam, err := generateSwaggerSpecParam(key, param, "query")
				if err != nil {
					return op, err
				}

				processedQueryParams[key] = nil
				op.Parameters = append(op.Parameters, specParam)
			}

			if testCase.RequestBody != nil {
				specParam := spec.Parameter{}
				specParam.Name = "body"
				specParam.In = "body"
				specParam.Required = true

				// TODO: right now it supports json, but should support marshaller depending on MIME type
				if content, err := json.MarshalIndent(testCase.RequestBody, "", "  "); err == nil {
					specParam.Description = string(content)
				}

				specParam.Schema = generateSpecSchema(testCase.RequestBody, defs)
				op.Parameters = append(op.Parameters, specParam)
			}
		}

		response := spec.Response{}
		response.Description = testCase.Description
		if testCase.ExpectedData != nil {
			response.Schema = generateSpecSchema(testCase.ExpectedData, defs)
			response.Examples = map[string]interface{}{
				"application/json": testCase.ExpectedData,
			}
		}

		op.Responses.StatusCodeResponses[testCase.ExpectedHttpCode] = response
	}

	op.Summary = description
	if taggable, ok := test.(ITaggable); ok {
		op.Tags = []string{taggable.Tag()}
	}

	return op, nil
}

func generateSwaggerSpecParam(paramKey string, param Param, location string) (spec.Parameter, error) {
	specParam := spec.Parameter{}
	specParam.Name = paramKey
	specParam.In = location
	specParam.Required = param.Required
	specParam.Description = param.Description
	specParam.Default = param.Value

	paramType, err := generateSpecSimpleType(param.Value)
	if err != nil {
		return specParam, fmt.Errorf("could not guess type of parameter '%s': %s", paramKey, err.Error())
	}
	specParam.Type = paramType

	return specParam, nil
}

func generateSpecSchema(item interface{}, defs spec.Definitions) *spec.Schema {
	refl := jsonschema.Reflect(item)
	schema := specSchemaFromJsonType(refl.Type)

	schema.Definitions = map[string]spec.Schema{}
	for name, def := range refl.Definitions {
		defs[name] = *specSchemaFromJsonType(def)
	}

	return schema
}

func specSchemaFromJsonType(schema *jsonschema.Type) *spec.Schema {
	s := &spec.Schema{}
	if schema.Type != "" {
		s.Type = []string{schema.Type}
	}
	if schema.Ref != "" {
		s.Ref = spec.MustCreateRef(schema.Ref)
	}

	s.Format = schema.Format
	s.Required = schema.Required

	// currently there is no way to determine whether there is MaxLength or MinLength
	// defined. Need to fix jsonschema library and switch type from int to *int
	// s.MaxLength = schema.MaxLength
	// s.MinLength = schema.MinLength
	s.Pattern = schema.Pattern
	s.Enum = schema.Enum
	s.Default = schema.Default
	s.Title = schema.Title
	s.Description = schema.Description

	if schema.Items != nil {
		s.Items = &spec.SchemaOrArray{}
		s.Items.Schema = specSchemaFromJsonType(schema.Items)
	}

	if schema.Properties != nil {
		s.Properties = make(map[string]spec.Schema)
		for key, prop := range schema.Properties {
			s.Properties[key] = *specSchemaFromJsonType(prop)
		}
	}

	if schema.PatternProperties != nil {
		s.PatternProperties = make(map[string]spec.Schema)
		for key, prop := range schema.PatternProperties {
			s.PatternProperties[key] = *specSchemaFromJsonType(prop)
		}
	}

	switch string(schema.AdditionalProperties) {
	case "true":
		s.AdditionalProperties = &spec.SchemaOrBool{Allows: true}
	case "false":
		s.AdditionalProperties = &spec.SchemaOrBool{Allows: false}
	}

	return s
}

func generateSpecSimpleType(value interface{}) (string, error) {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "integer", nil
	case float32, float64:
		return "number", nil
	case bool:
		return "boolean", nil
	case string:
		return "string", nil
	}

	return "", fmt.Errorf("value of complex type '%T' provided, simple type expected", value)
}
