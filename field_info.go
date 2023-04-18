package specs

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

type fieldInfo_BSON struct {
	BSON_TypeIsMarshaler   bool
	BSON_TypeIsUnmarshaler bool
	BSON_OmitEmpty         bool
	BSON_Inline            bool
}

func (inFieldInfo *fieldInfo_BSON) Resolve(f reflect.StructField) (name string, fieldInfo *fieldInfo_BSON) {
	bsonTag := f.Tag.Get("bson")
	if bsonTag == "-" {
		return
	}
	if bsonTag == "" {
		return
	}

	fieldInfo = inFieldInfo
	_, fieldInfo.BSON_TypeIsMarshaler = f.Type.MethodByName("MarshalBSON")
	_, fieldInfo.BSON_TypeIsUnmarshaler = f.Type.MethodByName("UnmarshalBSON")

	for i, part := range strings.Split(bsonTag, ",") {
		if i == 0 {
			if part != "" {
				name = part
			}
		} else {
			switch part {
			case "omitempty":
				fieldInfo.BSON_OmitEmpty = true
			case "inline":
				fieldInfo.BSON_Inline = true
			}
		}
	}

	return
}

type fieldInfo_JSON struct {
	JSON_TypeIsMarshaler   bool
	JSON_TypeIsUnmarshaler bool
	JSON_OmitEmpty         bool
	JSON_String            bool
}

func (inFieldInfo *fieldInfo_JSON) Resolve(f reflect.StructField) (name string, fieldInfo *fieldInfo_JSON) {
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

	*fieldInfo_JSON
	*fieldInfo_BSON
	*fieldInfo_Validator
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

		jsonTag, hasJSONTag := f.Tag.Lookup("json")
		if hasJSONTag && jsonTag == "-" {
			continue
		}

		index := make([]int, 0, len(parentIndex)+1)
		index = append(index, parentIndex...)
		index = append(index, i)

		if f.Anonymous && !hasJSONTag {
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

		_, field.fieldInfo_BSON = new(fieldInfo_BSON).Resolve(f)
		_, field.fieldInfo_Validator = new(fieldInfo_Validator).Resolve(f)

		var jsonName string
		jsonName, field.fieldInfo_JSON = new(fieldInfo_JSON).Resolve(f)
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
