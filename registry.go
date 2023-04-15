package scf

import (
	"fmt"
	"net/http"
)

type Endpoint[T interface{}] struct {
	Title       string
	Description string

	Handler T

	Protocol string
	Method   string
	Path     string
	Status   int

	Parameters interface{}
	Query      interface{}

	Consumes interface{}
	Produces map[int]interface{}
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

func (b *builder[T]) Consumes(consumes interface{}) *builder[T] {
	b.e.Consumes = consumes
	return b
}

func (b *builder[T]) Produces(status int, data interface{}) *builder[T] {
	if b.e.Produces == nil {
		b.e.Produces = map[int]interface{}{}
	}
	b.e.Produces[status] = data
	return b
}

func (b *builder[T]) Build() *Endpoint[T] {
	return b.e
}

var (
	httpRegistry = Registry[http.Handler]{}
)

func Register(method string, path string, handler http.Handler) *builder[http.Handler] {
	return httpRegistry.Register(method, path, handler)
}

type Registry[T interface{}] map[string]Endpoint[T]

func (r *Registry[T]) Register(method string, path string, handler T) *builder[T] {
	e := Endpoint[T]{Method: method, Path: path, Handler: handler}
	(*r)[fmt.Sprintf("%s-%s", e.Method, e.Path)] = e
	return &builder[T]{
		e: &e,
	}
}
