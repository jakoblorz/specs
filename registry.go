package scf

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrQueryAnnotationFailed      = errors.New("query annotation failed")
	ErrParametersAnnotationFailed = errors.New("parameters annotation failed")
	ErrResponseAnnotationFailed   = errors.New("response annotation failed")
	ErrPayloadAnnotationFailed    = errors.New("payload annotation failed")
)

type Response struct {
	Description string
	MediaType   string
	Value       interface{}
}

type Body struct {
	MediaType string
	Value     interface{}
}

type Endpoint[T interface{}] struct {
	OperationID string
	Title       string
	Description string

	Deprecated bool
	Tags       []string

	Handler T

	Protocol string
	Method   string
	Path     string
	Status   int

	Parameters interface{}
	Query      interface{}

	Payload  []Body
	Response map[int]Response
}

type builder[T interface{}] struct {
	e *Endpoint[T]
}

func (b *builder[T]) panic(err error) {
	panic(fmt.Errorf("failed to build endpoint (%s %s): %w", b.e.Method, b.e.Path, err))
}

func (b *builder[T]) Title(title string) *builder[T] {
	b.e.Title = title
	return b
}

func (b *builder[T]) Description(description string) *builder[T] {
	b.e.Description = description
	return b
}

func (b *builder[T]) Deprecated() *builder[T] {
	b.e.Deprecated = true
	return b
}

func (b *builder[T]) Tags(tags ...string) *builder[T] {
	b.e.Tags = tags
	return b
}

func (b *builder[T]) Protocol(protocol string) *builder[T] {
	b.e.Protocol = protocol
	return b
}

func (b *builder[T]) Status(status int) *builder[T] {
	b.e.Status = status
	return b
}

func (b *builder[T]) Parameters(parameters interface{}) *builder[T] {
	if b.e.Parameters != nil {
		b.panic(fmt.Errorf("parameters already defined: %w", ErrParametersAnnotationFailed))
	}
	b.e.Parameters = parameters
	return b
}

func (b *builder[T]) Query(query interface{}) *builder[T] {
	if b.e.Query != nil {
		b.panic(fmt.Errorf("query already defined: %w", ErrQueryAnnotationFailed))
	}
	b.e.Query = query
	return b
}

func (b *builder[T]) Payload(data interface{}, mediaTypes ...string) *builder[T] {
	if b.e.Payload == nil {
		b.e.Payload = []Body{}
	}
	for _, mediaType := range safeMediaTypes(mediaTypes) {
		for _, payload := range b.e.Payload {
			if payload.MediaType == mediaType {
				b.panic(fmt.Errorf("payload with media type %s already defined: %w", mediaType, ErrPayloadAnnotationFailed))
			}
		}
		b.e.Payload = append(b.e.Payload, Body{
			MediaType: mediaType,
			Value:     data,
		})
	}
	return b
}

func (b *builder[T]) Response(status int, data interface{}, description string, mediaTypes ...string) *builder[T] {
	if b.e.Response == nil {
		b.e.Response = map[int]Response{}
	}
	if _, hasStatusDefined := b.e.Response[status]; hasStatusDefined {
		b.panic(fmt.Errorf("response with status code %d already defined: %w", status, ErrResponseAnnotationFailed))
	}
	for _, mediaType := range safeMediaTypes(mediaTypes) {
		b.e.Response[status] = Response{
			Description: description,
			MediaType:   mediaType,
			Value:       data,
		}
	}
	return b
}

func (b *builder[T]) Build() *Endpoint[T] {
	return b.e
}

var (
	httpRegistry = Registry[http.Handler]{}
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

type Registry[T interface{}] map[string]*Endpoint[T]

func (r *Registry[T]) Bind(method string, path string, handler T) *builder[T] {
	e := Endpoint[T]{OperationID: fmt.Sprintf("%s-%s", method, path), Method: method, Path: path, Handler: handler}
	(*r)[e.OperationID] = &e
	return &builder[T]{
		e: &e,
	}
}

func (r *Registry[T]) GET(path string, handler T) *builder[T] {
	return r.Bind(http.MethodGet, path, handler)
}

func (r *Registry[T]) POST(path string, handler T) *builder[T] {
	return r.Bind(http.MethodPost, path, handler)
}

func (r *Registry[T]) PUT(path string, handler T) *builder[T] {
	return r.Bind(http.MethodPut, path, handler)
}

func (r *Registry[T]) PATCH(path string, handler T) *builder[T] {
	return r.Bind(http.MethodPatch, path, handler)
}

func (r *Registry[T]) DELETE(path string, handler T) *builder[T] {
	return r.Bind(http.MethodDelete, path, handler)
}

func safeMediaTypes(mediaTypes []string) []string {
	if len(mediaTypes) == 0 {
		return []string{"application/json"}
	}
	return mediaTypes
}
