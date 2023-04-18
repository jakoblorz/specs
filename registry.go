package specs

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	httpRegistry = registry[http.Handler]{}
)

func GET(path string, handler http.Handler) *builder[http.Handler] {
	return httpRegistry.GET(path, handler)
}

func POST(path string, handler http.Handler) *builder[http.Handler] {
	return httpRegistry.POST(path, handler)
}

func PUT(path string, handler http.Handler) *builder[http.Handler] {
	return httpRegistry.PUT(path, handler)
}

func PATCH(path string, handler http.Handler) *builder[http.Handler] {
	return httpRegistry.PATCH(path, handler)
}

func DELETE(path string, handler http.Handler) *builder[http.Handler] {
	return httpRegistry.DELETE(path, handler)
}

type OperationIDGeneratorFunc func(method string, path string) string

func OperationIDGenerator(f OperationIDGeneratorFunc) RegistryOption {
	return func(o *registryOptions) {
		o.OperationIDGenerator = f
	}
}

func DefaultOperationIDGenerator(method string, path string) string {
	id, err := gonanoid.New()
	if err != nil {
		panic(err)
	}
	return id
}

type registryOptions struct {
	OperationIDGenerator OperationIDGeneratorFunc
}

type RegistryOption func(*registryOptions)

type Registry[T interface{}] map[string]*Endpoint[T]

type registry[T interface{}] struct {
	options registryOptions
	routes  map[string]*Endpoint[T]
}

func NewRegistry[T interface{}](opts ...RegistryOption) *registry[T] {
	options := &registryOptions{
		OperationIDGenerator: DefaultOperationIDGenerator,
	}
	for _, applyOption := range opts {
		applyOption(options)
	}

	return &registry[T]{
		options: *options,
		routes:  map[string]*Endpoint[T]{},
	}
}

func (r *registry[T]) generateOperationID(method string, path string) string {
	return r.options.OperationIDGenerator(method, path)
}

func (r *registry[T]) Eject() Registry[T] {
	return r.routes
}

func (r *registry[T]) Build(method string, path string, handler T) *builder[T] {
	operationID := r.generateOperationID(method, path)
	e := Endpoint[T]{
		OperationID: operationID,
		Method:      method,
		Path:        path,
		Handler:     handler,
	}
	r.Add(&e)
	return &builder[T]{&e}
}

func (r *registry[T]) Add(e *Endpoint[T]) {
	r.routes[e.OperationID] = e
}

func (r *registry[T]) GET(path string, handler T) *builder[T] {
	return r.Build(http.MethodGet, path, handler)
}

func (r *registry[T]) POST(path string, handler T) *builder[T] {
	return r.Build(http.MethodPost, path, handler)
}

func (r *registry[T]) PUT(path string, handler T) *builder[T] {
	return r.Build(http.MethodPut, path, handler)
}

func (r *registry[T]) PATCH(path string, handler T) *builder[T] {
	return r.Build(http.MethodPatch, path, handler)
}

func (r *registry[T]) DELETE(path string, handler T) *builder[T] {
	return r.Build(http.MethodDelete, path, handler)
}

func (r *registry[T]) HEAD(path string, handler T) *builder[T] {
	return r.Build(http.MethodHead, path, handler)
}

func (r *registry[T]) OPTIONS(path string, handler T) *builder[T] {
	return r.Build(http.MethodOptions, path, handler)
}

func (r *registry[T]) TRACE(path string, handler T) *builder[T] {
	return r.Build(http.MethodTrace, path, handler)
}

func (r *registry[T]) Annotate(t *openapi3.T) {
	schemas := make(openapi3.Schemas)

	typeInfoCache := NewTypeInfoCache()
	schemaGenerator := NewSchemaRefGenerator(WithTypeInfoCache(typeInfoCache))

	for operationId, endpoint := range r.routes {
		operation := openapi3.Operation{
			Tags:        endpoint.Tags,
			Summary:     endpoint.Title,
			Description: endpoint.Description,
			OperationID: operationId,
			Deprecated:  endpoint.Deprecated,
		}

		if endpoint.Parameters != nil {
			parameterRef, err := schemaGenerator.GenerateSchemaRef(endpoint.Parameters, schemas)
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

		if endpoint.Query != nil {
			queryRef, err := schemaGenerator.GenerateSchemaRef(endpoint.Query, schemas)
			if err != nil {
				panic(err)
			}

			if operation.Parameters == nil {
				operation.Parameters = make(openapi3.Parameters, 0)
			}
			for name, property := range queryRef.Value.Properties {
				operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:   name,
						In:     "query",
						Schema: property,
					},
				})
			}
		}

		if endpoint.Method != http.MethodGet {
			content := make(map[string]*openapi3.MediaType)
			for _, requestBodyDeclaration := range endpoint.Payload {
				requestBodyRef, err := schemaGenerator.GenerateSchemaRef(requestBodyDeclaration.Value, schemas)
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
			responseRef, err := schemaGenerator.GenerateSchemaRef(response.Value, schemas)
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

	t.Components = &openapi3.Components{
		Schemas: schemas,
	}

}

func safeMediaTypes(mediaTypes []string) []string {
	if len(mediaTypes) == 0 {
		return []string{"application/json"}
	}
	return mediaTypes
}
