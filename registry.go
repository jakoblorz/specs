package scf

import (
	"fmt"
	"net/http"
)

type Endpoint[T interface{}] struct {
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

	Consumes []struct {
		MediaType string
		Value     interface{}
	}
	Produces map[int]struct {
		MediaType string
		Value     interface{}
	}
}

type builder[T interface{}] struct {
	e *Endpoint[T]
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
	b.e.Parameters = parameters
	return b
}

func (b *builder[T]) Query(query interface{}) *builder[T] {
	b.e.Query = query
	return b
}

func (b *builder[T]) Consumes(data interface{}, mediaTypes ...string) *builder[T] {
	if b.e.Consumes == nil {
		b.e.Consumes = []struct {
			MediaType string
			Value     interface{}
		}{}
	}

	if len(mediaTypes) == 0 {
		mediaTypes = []string{"application/json"}
	}
	for _, mediaType := range mediaTypes {
		b.e.Consumes = append(b.e.Consumes, struct {
			MediaType string
			Value     interface{}
		}{
			MediaType: mediaType,
			Value:     data,
		})
	}
	return b
}

func (b *builder[T]) Produces(status int, data interface{}, mediaTypes ...string) *builder[T] {
	if b.e.Produces == nil {
		b.e.Produces = map[int]struct {
			MediaType string
			Value     interface{}
		}{}
	}

	if len(mediaTypes) == 0 {
		mediaTypes = []string{"application/json"}
	}
	for _, mediaType := range mediaTypes {
		b.e.Produces[status] = struct {
			MediaType string
			Value     interface{}
		}{
			MediaType: mediaType,
			Value:     data,
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

func Register(method string, path string, handler http.Handler) *builder[http.Handler] {
	return httpRegistry.Bind(method, path, handler)
}

type Registry[T interface{}] map[string]Endpoint[T]

func (r *Registry[T]) Bind(method string, path string, handler T) *builder[T] {
	e := Endpoint[T]{Method: method, Path: path, Handler: handler}
	(*r)[fmt.Sprintf("%s-%s", e.Method, e.Path)] = e
	return &builder[T]{
		e: &e,
	}
}
