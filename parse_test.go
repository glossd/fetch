package fetch

import (
	"reflect"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	type testCase struct {
		V string
		E any
	}

	var cases = []testCase{
		{V: `{"name":"1"}`, E: Pet{Name: "1"}},
		{V: `{"name":"2"}`, E: PetLowerCaseTag{Name: "2"}},
		{V: `{"Name":"3"}`, E: PetUpperCaseTag{Name: "3"}},
	}

	for _, c := range cases {
		p := reflect.New(reflect.TypeOf(c.E)).Interface()
		err := UnmarshalInto(c.V, p)
		if err != nil {
			t.Fatalf("Unmarshal error: %s", err)
		}
		got := reflect.ValueOf(p).Elem().Interface()
		if got != c.E {
			t.Errorf("Marshal result mismatch, got=%v, expected=%v", got, c.E)
		}
	}
}

func TestUnmarshalString(t *testing.T) {
	_, err := Unmarshal[string]("hello")
	if err == nil {
		t.Fatal(err)
	}
}
