package json

import (
	"testing"
)

func TestSecondToLower(t *testing.T) {
	type testCase struct {
		In  string
		Out string
	}
	cases := []testCase{
		{In: `"Name"`, Out: `"name"`},
		{In: `"NameName"`, Out: `"nameName"`},
		{In: `"name"`, Out: `"name"`},
		{In: `"NN"`, Out: `"nN"`},
		{In: `"Имя"`, Out: `"имя"`},
		{In: `"姓名"`, Out: `"姓名"`},
	}
	for _, c := range cases {
		got := secondToLower(c.In)
		if got != c.Out {
			t.Errorf("failed to transform to lower case, got=%s, expected=%s", got, c.Out)
		}
	}
}

func TestMarshalToLowerCase(t *testing.T) {
	type TestStruct struct {
		Name string
	}
	marsh, err := Marshal(TestStruct{Name: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	got := string(marsh)
	expected := `{"name":"hello"}`
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestMarshalToUpperCaseWithTag(t *testing.T) {
	type TestStruct struct {
		Name string `json:"Name"`
	}
	marsh, err := Marshal(TestStruct{Name: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	got := string(marsh)
	expected := `{"Name":"hello"}`
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
