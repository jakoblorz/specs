package scf

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrInvalidOperator = errors.New("invalid operator")
	ErrKeysTagRequired = errors.New("'" + endKeysTag + "' Operator encountered without a corresponding '" + keysTag + "' Operator")
	ErrDiveTagRequired = errors.New("'" + keysTag + "' Operator must be immediately preceded by the '" + diveTag + "' Operator")
)

type TagType uint8

const (
	TagTypeDefault TagType = iota
	TagTypeOmitEmpty
	TagTypeIsDefault
	TagTypeNoStructLevel
	TagTypeStructOnly
	TagTypeDive
	TagTypeOr
	TagTypeKeys
	TagTypeEndKeys
)

const (
	defaultValidateTagName = "validate"
	utf8HexComma           = "0x2C"
	utf8Pipe               = "0x7C"
	tagSeparator           = ","
	orSeparator            = "|"
	tagKeySeparator        = "="
	structOnlyTag          = "structonly"
	noStructLevelTag       = "nostructlevel"
	omitempty              = "omitempty"
	isdefault              = "isdefault"
	diveTag                = "dive"
	keysTag                = "Keys"
	endKeysTag             = "endkeys"
)

type FieldTag struct {
	Operator    string
	Param       string
	Keys        *FieldTag // only populated when using Operator's 'Keys' and 'endkeys' for map key validation
	Next        *FieldTag
	Type        TagType
	HasOperator bool
	HasParam    bool // true if parameter used eg. eq= where the equal sign has been set
	IsBlockEnd  bool // indicates the current Operator represents the last validation in the block
}

func parseFieldTags(tag string) (firstFieldTag *FieldTag, current *FieldTag) {
	var t string
	tags := strings.Split(tag, tagSeparator)

	for i := 0; i < len(tags); i++ {
		t = tags[i]

		var prevTagType TagType

		if i == 0 {
			current = &FieldTag{HasOperator: true, Type: TagTypeDefault}
			firstFieldTag = current
		} else {
			prevTagType = current.Type
			current.Next = &FieldTag{HasOperator: true}
			current = current.Next
		}

		switch t {
		case diveTag:
			current.Type = TagTypeDive
			continue

		case keysTag:
			current.Type = TagTypeKeys

			if i == 0 || prevTagType != TagTypeDive {
				panic(ErrDiveTagRequired)
			}

			current.Type = TagTypeKeys

			// need to pass along only Keys Operator
			// need to increment i to skip over the Keys tags
			b := make([]byte, 0, 64)

			i++

			for ; i < len(tags); i++ {

				b = append(b, tags[i]...)
				b = append(b, ',')

				if tags[i] == endKeysTag {
					break
				}
			}

			current.Keys, _ = parseFieldTags(string(b[:len(b)-1]))
			continue

		case endKeysTag:
			current.Type = TagTypeEndKeys

			// if there are more in tags then there was no keysTag defined
			// and an error should be thrown
			if i != len(tags)-1 {
				panic(ErrKeysTagRequired)
			}
			return

		case omitempty:
			current.Type = TagTypeOmitEmpty
			continue

		case structOnlyTag:
			current.Type = TagTypeStructOnly
			continue

		case noStructLevelTag:
			current.Type = TagTypeNoStructLevel
			continue

		default:
			if t == isdefault {
				current.Type = TagTypeIsDefault
			}
			// if a pipe character is needed within the Param you must use the utf8Pipe representation "0x7C"
			orVals := strings.Split(t, orSeparator)

			for j := 0; j < len(orVals); j++ {
				vals := strings.SplitN(orVals[j], tagKeySeparator, 2)

				if j > 0 {
					current.Next = &FieldTag{HasOperator: true}
					current = current.Next
				}
				current.HasParam = len(vals) > 1

				current.Operator = vals[0]
				if len(current.Operator) == 0 {
					panic(fmt.Errorf("len(operator) == 0: %w", ErrInvalidOperator))
				}
				if len(orVals) > 1 {
					current.Type = TagTypeOr
				}
				if len(vals) > 1 {
					current.Param = vals[1]
					current.Param = strings.Replace(current.Param, utf8HexComma, ",", -1)
					current.Param = strings.Replace(current.Param, utf8Pipe, "|", -1)
				}
			}
			current.IsBlockEnd = true
		}
	}
	return
}

type fieldInfo_Validator struct {
	rootFieldTag *FieldTag
}

func (inFieldInfo *fieldInfo_Validator) Resolve(f reflect.StructField) (name string, fieldInfo *fieldInfo_Validator) {
	validateTag := f.Tag.Get(defaultValidateTagName)
	if validateTag == "" {
		return
	}

	fieldInfo = inFieldInfo
	fieldInfo.rootFieldTag, _ = parseFieldTags(validateTag)
	return
}

type fieldTagWalker struct {
	rootFieldTag *FieldTag
}

func createFieldTagWalker(v *fieldInfo_Validator) *fieldTagWalker {
	r := new(fieldTagWalker)
	if v != nil {
		r.rootFieldTag = v.rootFieldTag
	}
	return r
}

func (v *fieldTagWalker) Walk(walkerFunc func(fieldTag *FieldTag)) {
	for current := v.rootFieldTag; current != nil; current = current.Next {
		walkerFunc(current)
	}
}
