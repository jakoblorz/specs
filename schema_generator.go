package scf

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// This code is heavily inspired by github.com/getkin/kin-openapi/openapi3gen, though a few things needed to be adjusted
// and enough relevant parts were added to warrant a "rewrite".

var (
	ErrCycleDetected  = errors.New("cycle detected")
	ErrSchemaExcluded = errors.New("schema excluded")
)

var (
	timeType       = reflect.TypeOf(time.Time{})
	rawMessageType = reflect.TypeOf(json.RawMessage{})

	zeroInt   = float64(0)
	maxInt8   = float64(math.MaxInt8)
	minInt8   = float64(math.MinInt8)
	maxInt16  = float64(math.MaxInt16)
	minInt16  = float64(math.MinInt16)
	maxUint8  = float64(math.MaxUint8)
	maxUint16 = float64(math.MaxUint16)
	maxUint32 = float64(math.MaxUint32)
	maxUint64 = float64(math.MaxUint64)
)

var (
	refSchemaRef = openapi3.NewSchemaRef("Ref", openapi3.NewObjectSchema().WithProperty("$ref", openapi3.NewStringSchema().WithMinLength(1)))
)

type SchemaRefGeneratorOption func(*schemaRefGeneratorOption)

type schemaRefGeneratorOption struct {
	throwErrorOnCycle bool
	typeInfoCache     *TypeInfoCache
}

func ThrowErrorOnCycle() SchemaRefGeneratorOption {
	return func(opt *schemaRefGeneratorOption) {
		opt.throwErrorOnCycle = true
	}
}

func WithTypeInfoCache(cache *TypeInfoCache) SchemaRefGeneratorOption {
	return func(opt *schemaRefGeneratorOption) {
		opt.typeInfoCache = cache
	}
}

type SchemaRefGenerator struct {
	options schemaRefGeneratorOption

	Types map[reflect.Type]*openapi3.SchemaRef

	// SchemaRefs contains all references and their counts.
	// If count is 1, it's not ne
	// An OpenAPI identifier has been assigned to each.
	SchemaRefs map[*openapi3.SchemaRef]int

	// componentSchemaRefs is a set of schemas that must be defined in the components to avoid cycles
	componentSchemaRefs map[string]struct{}
}

func NewSchemaRefGenerator(opts ...SchemaRefGeneratorOption) *SchemaRefGenerator {
	options := &schemaRefGeneratorOption{
		typeInfoCache: DefaultTypeInfoCache,
	}
	for _, applyOption := range opts {
		applyOption(options)
	}

	return &SchemaRefGenerator{
		Types:      make(map[reflect.Type]*openapi3.SchemaRef),
		SchemaRefs: make(map[*openapi3.SchemaRef]int),
		options:    *options,
	}
}

func (g *SchemaRefGenerator) GenerateSchemaRef(v interface{}, schemas openapi3.Schemas) (*openapi3.SchemaRef, error) {
	t := reflect.TypeOf(v)
	if ref := g.Types[t]; ref != nil {
		g.SchemaRefs[ref]++
		return ref, nil
	}

	ref, err := g.generateSchemaRef(nil, t, "_root", nil)
	if errors.Is(err, ErrSchemaExcluded) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if ref != nil {
		g.Types[t] = ref
		g.SchemaRefs[ref]++
	}
	for ref := range g.SchemaRefs {
		if _, ok := g.componentSchemaRefs[ref.Ref]; ok && schemas != nil {
			schemas[ref.Ref] = &openapi3.SchemaRef{
				Value: ref.Value,
			}
		}
		if strings.HasPrefix(ref.Ref, "#/components/schemas/") {
			ref.Value = nil
		} else {
			ref.Ref = ""
		}
	}
	return ref, nil
}

func (g *SchemaRefGenerator) generateSchemaRef(parents []*TypeInfo, t reflect.Type, name string, parentField *Field) (*openapi3.SchemaRef, error) {
	typeInfo := g.options.typeInfoCache.GetTypeInfo(t)
	for _, parent := range parents {
		if parent == typeInfo {
			return nil, ErrCycleDetected
		}
	}
	if cap(parents) == 0 {
		parents = make([]*TypeInfo, 0, 4)
	}
	parents = append(parents, typeInfo)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if strings.HasSuffix(t.Name(), "Ref") {
		_, a := t.FieldByName("Ref")
		v, b := t.FieldByName("Value")
		if a && b {
			vs, err := g.generateSchemaRef(parents, v.Type, name, nil)
			if err != nil {
				if errors.Is(err, ErrCycleDetected) && !g.options.throwErrorOnCycle {
					g.SchemaRefs[vs]++
					return vs, nil
				}

				return nil, err
			}
			g.SchemaRefs[refSchemaRef]++
			ref := openapi3.NewSchemaRef(t.Name(), &openapi3.Schema{
				OneOf: []*openapi3.SchemaRef{
					refSchemaRef,
					vs,
				},
			})
			g.SchemaRefs[ref]++
			return ref, nil
		}
	}

	schema := &openapi3.Schema{}

	switch t.Kind() {
	case reflect.Func, reflect.Chan:
		return nil, nil
	case reflect.Bool:
		schema.Type = "boolean"

	case reflect.Int:
		schema.Type = "integer"
	case reflect.Int8:
		schema.Type = "integer"
		schema.Min = &minInt8
		schema.Max = &maxInt8
	case reflect.Int16:
		schema.Type = "integer"
		schema.Min = &minInt16
		schema.Max = &maxInt16
	case reflect.Int32:
		schema.Type = "integer"
		schema.Format = "int32"
	case reflect.Int64:
		schema.Type = "integer"
		schema.Format = "int64"
	case reflect.Uint:
		schema.Type = "integer"
		schema.Min = &zeroInt
	case reflect.Uint8:
		schema.Type = "integer"
		schema.Min = &zeroInt
		schema.Max = &maxUint8
	case reflect.Uint16:
		schema.Type = "integer"
		schema.Min = &zeroInt
		schema.Max = &maxUint16
	case reflect.Uint32:
		schema.Type = "integer"
		schema.Min = &zeroInt
		schema.Max = &maxUint32
	case reflect.Uint64:
		schema.Type = "integer"
		schema.Min = &zeroInt
		schema.Max = &maxUint64
	case reflect.Float32:
		schema.Type = "number"
		schema.Format = "float"
	case reflect.Float64:
		schema.Type = "number"
		schema.Format = "double"
	case reflect.String:
		schema.Type = "string"

	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			if t == rawMessageType {
				return &openapi3.SchemaRef{Value: schema}, nil
			}
			schema.Type = "string"
			schema.Format = "byte"
		} else {
			schema.Type = "array"
			items, err := g.generateSchemaRef(parents, t.Elem(), name, nil)
			if err != nil {
				if errors.Is(err, ErrCycleDetected) && !g.options.throwErrorOnCycle {
					items = g.generateCycleSchemaRef(t.Elem(), schema)
				} else {
					return nil, err
				}
			}
			if items != nil {
				g.SchemaRefs[items]++
				schema.Items = items
			}
		}

	case reflect.Map:
		schema.Type = "object"
		additionalProperties, err := g.generateSchemaRef(parents, t.Elem(), name, nil)
		if err != nil {
			if errors.Is(err, ErrCycleDetected) && !g.options.throwErrorOnCycle {
				additionalProperties = g.generateCycleSchemaRef(t.Elem(), schema)
			} else {
				return nil, err
			}
		}
		if additionalProperties != nil {
			g.SchemaRefs[additionalProperties]++
			schema.AdditionalProperties = openapi3.AdditionalProperties{Schema: additionalProperties}
		}

	case reflect.Struct:
		if t == timeType {
			schema.Type = "string"
			schema.Format = "date-time"
		} else {
			for _, fieldInfo := range typeInfo.Fields {
				fieldName, fType := fieldInfo.Name, fieldInfo.Type
				ref, err := g.generateSchemaRef(parents, fType, fieldName, &fieldInfo)
				if err != nil {
					if errors.Is(err, ErrCycleDetected) && !g.options.throwErrorOnCycle {
						ref = g.generateCycleSchemaRef(fType, schema)
					} else {
						return nil, err
					}
				}
				if ref != nil {
					g.SchemaRefs[ref]++
					schema.WithPropertyRef(fieldName, ref)
				}
			}

			// Object only if it has properties
			if schema.Properties != nil {
				schema.Type = "object"
			}
		}
	}

	return openapi3.NewSchemaRef(t.Name(), schema), nil
}

func (g *SchemaRefGenerator) generateCycleSchemaRef(t reflect.Type, schema *openapi3.Schema) *openapi3.SchemaRef {
	var typeName string
	switch t.Kind() {
	case reflect.Ptr:
		return g.generateCycleSchemaRef(t.Elem(), schema)
	case reflect.Slice:
		ref := g.generateCycleSchemaRef(t.Elem(), schema)
		sliceSchema := openapi3.NewSchema()
		sliceSchema.Type = "array"
		sliceSchema.Items = ref
		return openapi3.NewSchemaRef("", sliceSchema)
	case reflect.Map:
		ref := g.generateCycleSchemaRef(t.Elem(), schema)
		mapSchema := openapi3.NewSchema()
		mapSchema.Type = "object"
		mapSchema.AdditionalProperties = openapi3.AdditionalProperties{Schema: ref}
		return openapi3.NewSchemaRef("", mapSchema)
	default:
		typeName = t.Name()
	}

	g.componentSchemaRefs[typeName] = struct{}{}
	return openapi3.NewSchemaRef(fmt.Sprintf("#/components/schemas/%s", typeName), schema)
}
