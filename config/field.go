package config

import (
	"reflect"
	"strings"
)

// Tag is the key of the field Tag that is used to match certain things and properties,
// currently only "required", i.e myField `config:"required"`.
//
// Can be changed to a custom one if needed.
var Tag = "config"

func lookupTag(f reflect.StructField) (string, bool) {
	tag := f.Tag.Get(Tag)
	if tag == "" {
		return "", false
	}
	return tag, true
}

func containsTagValue(f reflect.StructField, searchFor string) bool {
	tag, found := lookupTag(f)
	if !found {
		return false
	}

	if tag == searchFor {
		return true
	}

	values := strings.Split(tag, ",")
	for _, v := range values {
		if v == searchFor {
			return true
		}
	}

	return false
}

type field struct {
	// the actual struct's indexes of the field.
	Index []int
	// the actual name, the yaml(or other file decoder's tag name) one or the field name.
	Name string
	// if marked as required, by tag.
	// And as always; a bool false value is zero,
	// so if required then it will ask for it so make sure that you setup your configuration fields correctly,
	// look the "options" structure's guidelines on comments to see what I'm talking about.
	Required bool

	// true if it's password/secret, tag value contains "password" or "secret", it's being used
	// on survey to show a special password prompt.
	Secret bool
}

func structFieldIgnored(f reflect.StructField) bool {
	return containsTagValue(f, "-") // if the tag contains the "-" then ignore it.
}

func isRequired(f reflect.StructField) bool {
	if f.Anonymous || f.PkgPath != "" || f.Type.Kind() == reflect.Struct {
		// skip unexported, anonymous(embedded) and structs from this check, they are always false,
		// because the check should be happen on the fields of these structs and not on these as structures.
		return false
	}

	return !structFieldIgnored(f)
}

func isSecret(f reflect.StructField) bool {
	return containsTagValue(f, "password") || containsTagValue(f, "secret")
}

func lookupFields(typ reflect.Type, parentIndex int) (fields []field) {
	for i, n := 0, typ.NumField(); i < n; i++ {
		f := typ.Field(i)

		if f.Type.Kind() == reflect.Ptr {
			continue // skip pointers.
		}

		// embedded.
		if f.Type.Kind() == reflect.Struct && !structFieldIgnored(f) {
			fields = append(fields, lookupFields(f.Type, i)...)
			continue
		}

		required := isRequired(f)
		index := []int{i}
		if parentIndex >= 0 {
			index = append([]int{parentIndex}, index...)
		}

		field := field{
			Name:     f.Name,
			Index:    index,
			Required: required,
			Secret:   isSecret(f),
		}

		fields = append(fields, field)
	}

	return
}
