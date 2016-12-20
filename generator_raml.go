package schreder

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/alecthomas/jsonschema"
	"github.com/go-raml/raml"
	"gopkg.in/yaml.v2"
)

type ramlGenerator struct {
	seed raml.APIDefinition
}

// NewRamlGenerator creates an instance of RAML generator
// seed as used as a source of initial data for resulting doc
func NewRamlGenerator(seed raml.APIDefinition) IDocGenerator {
	generator := &ramlGenerator{
		seed: seed,
	}

	generator.seed.RAMLVersion = "0.8"
	generator.seed.Resources = map[string]raml.Resource{}

	return generator
}

func (g *ramlGenerator) Generate(tests []Test) ([]byte, error) {
	doc := g.seed // copy seed

	for _, test := range tests {
		// path MUST begin with '/'
		path := test.Path()
		if path[0] != '/' {
			path = "/" + path
		}

		resource, ok := doc.Resources[path]
		if !ok { // new resource created by default
			resource.UriParameters = map[string]raml.NamedParameter{}
		}

		m := raml.Method{
			Responses:       map[raml.HTTPCode]raml.Response{},
			Headers:         map[raml.HTTPHeader]raml.Header{},
			QueryParameters: map[string]raml.NamedParameter{},
			Description:     test.Description(),
			Name:            test.Method(),
		}

		processedHeaderParams := map[string]interface{}{}
		processedPathParams := map[string]interface{}{}
		processedQueryParams := map[string]interface{}{}

		for _, testCase := range test.TestCases() {
			m.Description = testCase.Description
			for key, param := range testCase.PathParams {
				if _, ok := processedPathParams[key]; ok {
					continue
				}

				uriParam := generateRamlNamedParameter(key, param)

				processedPathParams[key] = nil
				resource.UriParameters[key] = uriParam
			}

			for key, param := range testCase.Headers {
				if _, ok := processedHeaderParams[key]; ok {
					continue
				}

				h := generateRamlNamedParameter(key, param)

				m.Headers[raml.HTTPHeader(key)] = raml.Header(h)
				processedHeaderParams[key] = nil
			}

			for key, param := range testCase.QueryParams {
				if _, ok := processedQueryParams[key]; ok {
					continue
				}

				queryParam := generateRamlNamedParameter(key, param)

				processedQueryParams[key] = nil
				m.QueryParameters[key] = queryParam
			}

			response := raml.Response{}
			response.Description = testCase.Description
			response.HTTPCode = raml.HTTPCode(testCase.ExpectedHttpCode)
			if testCase.ExpectedData != nil {
				schema := jsonschema.Reflect(testCase.ExpectedData)

				// TODO: marshal data according to MIME type, coming soon with RAML 1.0
				schemaBytes, _ := json.MarshalIndent(schema, "", "  ")
				response.Bodies.DefaultSchema = string(schemaBytes)

				// TODO: marshal data according to MIME type, coming soon with RAML 1.0
				exampleBytes, _ := json.MarshalIndent(testCase.ExpectedData, "", "  ")
				response.Bodies.DefaultExample = string(exampleBytes)
			}

			m.Responses[raml.HTTPCode(testCase.ExpectedHttpCode)] = response

		}

		// TODO: check if path has already assigned an method to some other test
		// return error if so
		switch test.Method() {
		case "GET":
			resource.Get = &m
		case "POST":
			resource.Post = &m
		case "PATCH":
			resource.Patch = &m
		case "DELETE":
			resource.Delete = &m
		case "PUT":
			resource.Put = &m
		case "HEAD":
			resource.Head = &m
		}

		doc.Resources[path] = resource
	}

	generatedDoc, err := yaml.Marshal(doc)
	if err == nil {
		header := []byte(fmt.Sprintf("#%%RAML %s\n", doc.RAMLVersion))
		generatedDoc = append(header, generatedDoc...)

	}

	return generatedDoc, err
}

func generateRamlNamedParameter(paramKey string, param Param) raml.NamedParameter {
	return raml.NamedParameter{
		Name:        paramKey,
		Description: param.Description,
		Required:    param.Required,
		Default:     param.Value,
		Type:        resolveRamlType(param.Value),
	}
}

func resolveRamlType(data interface{}) string {
	switch data.(type) {
	case []byte:
		return "string"
	case time.Time, *time.Time:
		return "date"
	default:
		val := reflect.ValueOf(data)
		tpe := val.Type()
		switch tpe.Kind() {
		case reflect.Bool:
			return "boolean"
		case reflect.String:
			return "string"
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint8, reflect.Uint16, reflect.Uint32:
			fallthrough
		case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
			return "integer"
		case reflect.Float32, reflect.Float64:
			return "number"
		case reflect.Ptr:
			return resolveRamlType(reflect.Indirect(val).Interface())
		}
	}
	return ""
}
