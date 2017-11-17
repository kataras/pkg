package zerocheck_test

import (
	"reflect"
	"testing"

	. "github.com/kataras/pkg/zerocheck"
)

type testStruct struct {
	Field1 string
	Field2 bool
}

type testStructUnexportedFields struct {
	field1 string // this will always return non zero
}

func Test(t *testing.T) {
	tt := testStruct{Field1: "no zero"}

	if IsZero(reflect.ValueOf(tt)) {
		t.Fatalf("%T should be not zero because it's filled", tt)
	}

	tt2 := testStruct{}
	if !IsZero(reflect.ValueOf(tt2)) {
		t.Fatalf("%T should zero because non of the fields are filled", tt2)
	}

	tt2.Field2 = true
	if IsZero(reflect.ValueOf(tt2)) {
		t.Fatalf("%T should be not zero because it's filled", tt2)
	}

	tt3 := testStructUnexportedFields{field1: "this is filled but it's unexported"}
	if IsZero(reflect.ValueOf(tt2)) {
		t.Fatalf("%T should be not zero because its only one field is unexported, so it cannot be checked", tt3)
	}
}

type testStructAlwaysZero struct {
	Field1 string
}

func (ts testStructAlwaysZero) IsZero() bool {
	return true
}

func TestUserdefinedIsZero(t *testing.T) {
	tt := testStructAlwaysZero{
		Field1: "filled",
	}

	if !IsZero(reflect.ValueOf(tt)) {
		t.Fatalf("%T should be always zero because it's user-defined IsZero is always return true", tt)
	}
}
