package scf

import (
	"net/http"

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
	options *registryOptions
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
		options: options,
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

func safeMediaTypes(mediaTypes []string) []string {
	if len(mediaTypes) == 0 {
		return []string{"application/json"}
	}
	return mediaTypes
}
