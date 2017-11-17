package structcopy_test

import (
	"reflect"
	"testing"
	"time"

	. "github.com/kataras/pkg/structcopy"
)

type emb struct {
	Name string
}

func TestCopy(t *testing.T) {
	var destination = struct {
		ID           int    // should not be setted
		AnotherField string // should not be setted
		IP           string
		Name         string
		SetInitial   string
		CreatedAt    time.Time // should not be setted
	}{SetInitial: "should stay"}

	var source = struct {
		IP string
		emb
		Country string // should be ignored
	}{
		"192.168.1.1",
		emb{"Gerasimos Maropoulos"},
		"Greece",
	}

	Copy(&destination, source)

	// unchanged
	if expected, got := 0, destination.ID; !reflect.DeepEqual(expected, got) {
		t.Fatalf("ID should be %v as source doesn't contain this field but got: %v", expected, got)
	}

	if expected, got := "", destination.AnotherField; !reflect.DeepEqual(expected, got) {
		t.Fatalf("AnotherField should be %v as source doesn't contain this field but got: %v", expected, got)
	}

	zeroTime := time.Time{}
	if expected, got := zeroTime, destination.CreatedAt; !reflect.DeepEqual(expected, got) {
		t.Fatalf("CreatedAt should be %v as source doesn't contain this field but got: %v", expected, got)
	}

	if expected, got := "should stay", destination.SetInitial; !reflect.DeepEqual(expected, got) {
		t.Fatalf("SetInitial should stay as it was, '%v' as source doesn't contain this field but got: %v", expected, got)
	}

	// changed
	if expected, got := "192.168.1.1", destination.IP; !reflect.DeepEqual(expected, got) {
		t.Fatalf("IP should be %v but got: %v", expected, got)
	}
	if expected, got := "Gerasimos Maropoulos", destination.Name; !reflect.DeepEqual(expected, got) {
		t.Fatalf("Name should be %v but got: %v", expected, got)
	}

	// and for any case, maybe the lib clears a value of the "source" (this will be a huge disaster if this test fails.)
	if expected, got := "Greece", source.Country; !reflect.DeepEqual(expected, got) {
		t.Fatalf("Source's Country should not be changed. Should be %v but got: %v", expected, got)
	}
}
