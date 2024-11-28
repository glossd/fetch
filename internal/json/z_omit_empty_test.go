package json

import (
	"encoding/json"
	"testing"
)

func TestOmitEmptyByDefault(t *testing.T) {
	type TestStruct struct {
		Name string
		Tag  string
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

	var ts TestStruct
	err = json.Unmarshal(marsh, &ts)
	if err != nil {
		t.Fatal(err)
	}
	if ts.Name != "hello" {
		t.Errorf("expected to have name ummarshaled, got=%s", ts.Name)
	}
}

func TestDontOmitEmptyIfTagIsSpecified(t *testing.T) {
	type TestStruct struct {
		Name string `json:"Name"`
	}
	marsh, err := Marshal(TestStruct{})
	if err != nil {
		t.Fatal(err)
	}
	got := string(marsh)
	expected := `{"Name":""}`
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
