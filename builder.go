package specs

import (
	"errors"
	"fmt"
)

var (
	ErrQueryAnnotationFailed      = errors.New("query annotation failed")
	ErrParametersAnnotationFailed = errors.New("parameters annotation failed")
	ErrResponseAnnotationFailed   = errors.New("response annotation failed")
	ErrPayloadAnnotationFailed    = errors.New("payload annotation failed")
)

type Builder[T interface{}] interface {
	Title(title string) Builder[T]
	Description(description string) Builder[T]
	Deprecated() Builder[T]
	Tags(tags ...string) Builder[T]
	Protocol(protocol string) Builder[T]
	Status(status int) Builder[T]
	Parameters(parameters interface{}) Builder[T]
	Query(query interface{}) Builder[T]
	Payload(data interface{}, mediaTypes ...string) Builder[T]
	Response(status int, data interface{}, description string, mediaTypes ...string) Builder[T]
	Build() *Endpoint[T]
}

type builder[T interface{}] struct {
	e *Endpoint[T]
}

var (
	_ Builder[interface{}] = (*builder[interface{}])(nil)
)

func (b *builder[T]) panic(err error) {
	panic(fmt.Errorf("failed to build endpoint (%s %s): %w", b.e.Method, b.e.Path, err))
}

func (b *builder[T]) Title(title string) Builder[T] {
	b.e.Title = title
	return b
}

func (b *builder[T]) Description(description string) Builder[T] {
	b.e.Description = description
	return b
}

func (b *builder[T]) Deprecated() Builder[T] {
	b.e.Deprecated = true
	return b
}

func (b *builder[T]) Tags(tags ...string) Builder[T] {
	b.e.Tags = tags
	return b
}

func (b *builder[T]) Protocol(protocol string) Builder[T] {
	b.e.Protocol = protocol
	return b
}

func (b *builder[T]) Status(status int) Builder[T] {
	b.e.Status = status
	return b
}

func (b *builder[T]) Parameters(parameters interface{}) Builder[T] {
	if b.e.Parameters != nil {
		b.panic(fmt.Errorf("parameters already defined: %w", ErrParametersAnnotationFailed))
	}
	b.e.Parameters = parameters
	return b
}

func (b *builder[T]) Query(query interface{}) Builder[T] {
	if b.e.Query != nil {
		b.panic(fmt.Errorf("query already defined: %w", ErrQueryAnnotationFailed))
	}
	b.e.Query = query
	return b
}

func (b *builder[T]) Payload(data interface{}, mediaTypes ...string) Builder[T] {
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

func (b *builder[T]) Response(status int, data interface{}, description string, mediaTypes ...string) Builder[T] {
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
