package scf

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
)

type SchemaGeneratorOption func(*schemaGeneratorOption)

type schemaGeneratorOption struct {
	useAllExportedFields bool
	throwErrorOnCycle    bool
}

func UseAllExportedFields() SchemaGeneratorOption {
	return func(opt *schemaGeneratorOption) {
		opt.useAllExportedFields = true
	}
}

func ThrowErrorOnCycle() SchemaGeneratorOption {
	return func(opt *schemaGeneratorOption) {
		opt.throwErrorOnCycle = true
	}
}

type SchemaGenerator struct {
	options schemaGeneratorOption

	Types map[reflect.Type]*openapi3.SchemaRef

	// SchemaRefs contains all references and their counts.
	// If count is 1, it's not ne
	// An OpenAPI identifier has been assigned to each.
	SchemaRefs map[*openapi3.SchemaRef]int
}

func NewSchemaGenerator(opts ...SchemaGeneratorOption) *SchemaGenerator {
	options := new(schemaGeneratorOption)
	for _, applyOption := range opts {
		applyOption(options)
	}

	return &SchemaGenerator{
		Types:      make(map[reflect.Type]*openapi3.SchemaRef),
		SchemaRefs: make(map[*openapi3.SchemaRef]int),
		options:    *options,
	}

}
