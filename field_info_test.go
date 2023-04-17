package scf

import (
	"reflect"
	"testing"
)

func ptr[T interface{}](v T) *T {
	return &v
}

func TestFields_Append(t *testing.T) {
	type embeddedStruct struct {
		Test11 string
		Test12 string
	}

	type field struct {
		Name  *string
		Type  *reflect.Type
		Index *[]int
	}

	tests := []struct {
		name string
		args reflect.Type
		want []field
	}{
		{
			name: "should include json-named field",
			args: reflect.TypeOf(struct {
				Test string `json:"test"`
			}{}),
			want: []field{{
				Name:  ptr("test"),
				Type:  ptr(reflect.TypeOf("")),
				Index: ptr([]int{0}),
			}},
		},
		{
			name: "should include unnamed field",
			args: reflect.TypeOf(struct {
				Test string
			}{}),
			want: []field{{
				Name:  ptr("Test"),
				Type:  ptr(reflect.TypeOf("")),
				Index: ptr([]int{0}),
			}},
		},
		{
			name: "should not include lowercase unnamed field",
			args: reflect.TypeOf(struct {
				test string
			}{}),
			want: []field{},
		},
		{
			name: "should include json-named and unnamed fields",
			args: reflect.TypeOf(struct {
				Test1 string
				Test2 string `json:"test2"`
			}{}),
			want: []field{
				{
					Name:  ptr("Test1"),
					Type:  ptr(reflect.TypeOf("")),
					Index: ptr([]int{0}),
				},
				{
					Name:  ptr("test2"),
					Type:  ptr(reflect.TypeOf("")),
					Index: ptr([]int{1}),
				},
			},
		},
		{
			name: "should include json-named and omit excluded fields",
			args: reflect.TypeOf(struct {
				Test1 string
				Test2 string `json:"-"`
			}{}),
			want: []field{
				{
					Name:  ptr("Test1"),
					Type:  ptr(reflect.TypeOf("")),
					Index: ptr([]int{0}),
				},
			},
		},
		{
			name: "should include json-unnamed embedded struct's fields",
			args: reflect.TypeOf(struct {
				embeddedStruct
			}{}),
			want: []field{
				{
					Name:  ptr("Test11"),
					Type:  ptr(reflect.TypeOf("")),
					Index: ptr([]int{0, 0}),
				},
				{
					Name:  ptr("Test12"),
					Type:  ptr(reflect.TypeOf("")),
					Index: ptr([]int{0, 1}),
				},
			},
		},
		{
			name: "should not include unnamed embedded struct's fields",
			args: reflect.TypeOf(struct {
				Test1 embeddedStruct
			}{}),
			want: []field{
				{
					Name:  ptr("Test1"),
					Type:  ptr(reflect.TypeOf(embeddedStruct{})),
					Index: ptr([]int{0}),
				},
			},
		},
		{
			name: "should not include unnamed embedded struct's fields",
			args: reflect.TypeOf(struct {
				Test1 embeddedStruct `json:"test1"`
			}{}),
			want: []field{
				{
					Name:  ptr("test1"),
					Type:  ptr(reflect.TypeOf(embeddedStruct{})),
					Index: ptr([]int{0}),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArr := make(Fields, 0).Append([]int{}, tt.args)
			if len(gotArr) != len(tt.want) {
				t.Errorf("len(Fields.Append()) = %v, want %v", len(gotArr), len(tt.want))
				return
			}
			for i, wantField := range tt.want {
				got := gotArr[i]
				if wantField.Name != nil && !reflect.DeepEqual(*wantField.Name, got.Name) {
					t.Errorf("Fields.Append()[0].Name = %v, want %v", got.Name, *wantField.Name)
				}
				if wantField.Type != nil && !reflect.DeepEqual(*wantField.Type, got.Type) {
					t.Errorf("Fields.Append()[0].Type = %v, want %v", got.Type, *wantField.Type)
				}
				if wantField.Index != nil && !reflect.DeepEqual(*wantField.Index, got.Index) {
					t.Errorf("Fields.Append()[0].Index = %v, want %v", got.Index, *wantField.Index)
				}
			}
		})
	}
}
