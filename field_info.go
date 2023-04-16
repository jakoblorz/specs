package scf

import (
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

func removeIndirect(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

type FieldInfo_JSON struct {
	JSON_TypeIsMarshaler   bool
	JSON_TypeIsUnmarshaler bool
	JSON_OmitEmpty         bool
	JSON_String            bool
}

func (inFieldInfo *FieldInfo_JSON) Resolve(f reflect.StructField) (name string, fieldInfo *FieldInfo_JSON) {
	jsonTag := f.Tag.Get("json")
	if jsonTag == "-" {
		return
	}
	if jsonTag == "" {
		return
	}

	fieldInfo = inFieldInfo
	_, fieldInfo.JSON_TypeIsMarshaler = f.Type.MethodByName("MarshalJSON")
	_, fieldInfo.JSON_TypeIsUnmarshaler = f.Type.MethodByName("UnmarshalJSON")

	for i, part := range strings.Split(jsonTag, ",") {
		if i == 0 {
			if part != "" {
				name = part
			}
		} else {
			switch part {
			case "omitempty":
				fieldInfo.JSON_OmitEmpty = true
			case "string":
				fieldInfo.JSON_String = true
			}
		}
	}

	return
}

type Field struct {
	Name  string
	Type  reflect.Type
	Index []int

	*FieldInfo_JSON
}

type Fields []Field

func (fields Fields) Append(parentIndex []int, t reflect.Type) Fields {
	t = removeIndirect(t)

	numField := t.NumField()
	for i := 0; i < numField; i++ {
		f := t.Field(i)

		switch f.Type.Kind() {
		case reflect.Func, reflect.Chan:
			continue
		}

		jsonTag := f.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		index := make([]int, 0, len(parentIndex)+1)
		index = append(index, parentIndex...)
		index = append(index, i)

		if f.Anonymous && jsonTag == "" {
			fields = fields.Append(index, f.Type)
			continue
		}

		firstRune, _ := utf8.DecodeRuneInString(f.Name)
		if unicode.IsLower(firstRune) {
			continue
		}

		field := Field{
			Index: index,
			Type:  f.Type,
			Name:  f.Name,
		}

		var jsonName string
		jsonName, field.FieldInfo_JSON = new(FieldInfo_JSON).Resolve(f)
		if jsonName != "" {
			field.Name = jsonName
		}

		fields = append(fields, field)
	}

	return fields
}

func (f Fields) Len() int           { return len(f) }
func (f Fields) Less(i, j int) bool { return f[i].Name < f[j].Name }
func (f Fields) Swap(i, j int)      { a, b := f[i], f[j]; f[i], f[j] = b, a }
