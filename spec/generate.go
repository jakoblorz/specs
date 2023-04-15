package spec

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/jakoblorz/scf"
)

// https://github.com/getkin/kin-openapi/blob/master/openapi3/operation_test.go
// https://github.com/getkin/kin-openapi/blob/master/openapi3gen/simple_test.go

func Generate[T interface{}](r scf.Registry[T]) (t *openapi3.T) {
	t = new(openapi3.T)
	g := openapi3gen.NewGenerator()
	for operationId, endpoint := range r {
		operation := openapi3.Operation{
			Tags:        endpoint.Tags,
			Summary:     endpoint.Title,
			Description: endpoint.Description,
			OperationID: operationId,
			Deprecated:  endpoint.Deprecated,
		}

		// parameterRef, err := g.NewSchemaRefForValue(endpoint.Parameters, nil)
		// if err != nil {
		// 	panic(err)
		// }

		// parameterRef.Value.Properties
		if endpoint.Method != http.MethodGet {
			content := make(map[string]*openapi3.MediaType)
			for _, requestBodyDeclaration := range endpoint.Consumes {
				requestBodyRef, err := g.NewSchemaRefForValue(endpoint.Consumes, nil)
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

		for status, response := range endpoint.Produces {
			responseRef, err := g.NewSchemaRefForValue(response.Value, nil)
			if err != nil {
				panic(err)
			}

			operation.Responses[fmt.Sprintf("%d", status)] = &openapi3.ResponseRef{
				Value: &openapi3.Response{
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