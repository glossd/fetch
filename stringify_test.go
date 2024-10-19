package fetch

import (
	"testing"
)

type Pet struct {
	Name string
}

type PetLowerCaseTag struct {
	Name string `json:"name"`
}

type PetUpperCaseTag struct {
	Name string `json:"Name"`
}

func TestMarshalStruct(t *testing.T) {
	type testCase struct {
		V any
		E string
	}
	var cases = []testCase{
		{V: &Pet{Name: "1"}, E: `{"name":"1"}`},
		{V: &PetLowerCaseTag{Name: "2"}, E: `{"name":"2"}`},
		{V: &PetUpperCaseTag{Name: "3"}, E: `{"Name":"3"}`},
	}
	for _, c := range cases {
		r, err := Marshal(c.V)
		if err != nil {
			t.Fatalf("Marshal error: %s", err)
		}
		if c.E != r {
			t.Errorf("Marshal result mismatch, got=%s, expected=%s", r, c.E)
		}
	}
}
