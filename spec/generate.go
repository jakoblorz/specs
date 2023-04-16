package spec

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/jakoblorz/scf"
)

func Generate[T interface{}](r scf.Registry[T]) (t *openapi3.T) {
	t = new(openapi3.T)
	t.OpenAPI = "3.0.0"

	g := openapi3gen.NewGenerator()
	for operationId, endpoint := range r {
		operation := openapi3.Operation{
			Tags:        endpoint.Tags,
			Summary:     endpoint.Title,
			Description: endpoint.Description,
			OperationID: operationId,
			Deprecated:  endpoint.Deprecated,
		}

		if endpoint.Parameters != nil {
			parameterRef, err := g.NewSchemaRefForValue(endpoint.Parameters, nil)
			if err != nil {
				panic(err)
			}

			if operation.Parameters == nil {
				operation.Parameters = make(openapi3.Parameters, 0)
			}
			for name, property := range parameterRef.Value.Properties {
				operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:     name,
						In:       "path",
						Required: true,
						Schema:   property,
					},
				})
			}
		}

		if endpoint.Method != http.MethodGet {
			content := make(map[string]*openapi3.MediaType)
			for _, requestBodyDeclaration := range endpoint.Payload {
				requestBodyRef, err := g.NewSchemaRefForValue(requestBodyDeclaration.Value, nil)
				if err != nil {
					panic(err)
				}

				content[requestBodyDeclaration.MediaType] = &openapi3.MediaType{
					Schema: requestBodyRef,
				}
			}

			operation.RequestBody = &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Required: true,
					Content:  content,
				},
			}
		}

		for status, response := range endpoint.Response {
			responseRef, err := g.NewSchemaRefForValue(response.Value, nil)
			if err != nil {
				panic(err)
			}

			if operation.Responses == nil {
				operation.Responses = make(map[string]*openapi3.ResponseRef)
			}
			operation.Responses[fmt.Sprintf("%d", status)] = &openapi3.ResponseRef{
				Value: &openapi3.Response{
					Description: &response.Description,
					Content: map[string]*openapi3.MediaType{
						response.MediaType: &openapi3.MediaType{
							Schema: responseRef,
						},
					},
				},
			}
		}

		t.AddOperation(endpoint.Path, endpoint.Method, &operation)
	}

	return
}
