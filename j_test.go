package fetch

import (
	"fmt"
	"testing"
)

func TestJ_Q(t *testing.T) {
	type testCase struct {
		In string
		P  string
		E  any
	}

	cases := []testCase{
		{In: `{"name": "Lola"}`, P: ".name", E: "Lola"},
		{In: `{"name": "Lola"}`, P: "name", E: "Lola"},
		{In: `{"dog": {"name": "Lola"}}`, P: ".dog.name", E: "Lola"},
		{In: `{"pet": {"dog": {"name": "Lola"}}}`, P: ".pet.dog.name", E: "Lola"},
		{In: `{"num": 1}`, P: ".num", E: 1.0},
		{In: `1`, P: ".", E: 1.0},
		{In: `true`, P: ".", E: true},
		{In: `"hello"`, P: ".", E: "hello"},
		{In: `{"sold": false}`, P: ".sold", E: false},
		{In: `[1, 2, 3]`, P: ".[0]", E: 1.0},
		{In: `[{"name":"Lola"}, {"name":"Buster"}]`, P: ".[1].name", E: "Buster"},
		{In: `{"nums":[1, 2, 57]}`, P: ".nums[2]", E: 57.0},
		{In: `{"pets":[{"name":"Lola"}]}`, P: ".pets[0].name", E: "Lola"},
		{In: `[{}, {"pets":[{},{}, {"name":"Lola", "tags":[{"id": 12}]}]}]`, P: ".[1].pets[2].tags[0].id", E: 12.0},
		{In: `{"":"hello"}`, P: "..", E: "hello"},
		{In: `{"":{"name":"Lola"}}`, P: "..name", E: "Lola"},
		{In: `[[1, 2, 3]]`, P: ".[0][2]", E: 3.0},
		{In: `[[1, 2, 3]]`, P: ".[0].[2]", E: 3.0},
		{In: `{}`, P: ".name", E: nil},
		{In: `{}`, P: ".name.category", E: nil},
		{In: `{}`, P: ".name.tags[0]", E: nil},
		{In: `{}`, P: ".[0]", E: nil},
		{In: `[]`, P: ".name", E: nil},
		{In: `{}`, P: "..", E: nil},
		{In: `{"name":"Lola"}`, P: ".name.category", E: nil},
		{In: `[1, 2]`, P: ".[3]", E: nil},
		{In: `{"tags":[{"id":12}]}`, P: ".tags.id", E: nil},
		{In: `{"category":{"name":"dog"}}`, P: ".category[0]", E: nil},
		{In: `[1, 2, 3]`, P: ".[1", E: nil},
		{In: `[1, 2, 3]`, P: ".]1", E: nil},
		{In: `[1, 2, 3]`, P: ".[]", E: nil},
		{In: `[1, 2, 3]`, P: ".[hello]", E: nil},
		{In: `[1, 2, 3]`, P: ".[0]name", E: nil},
	}
	for i, c := range cases {
		j, err := Unmarshal[J](c.In)
		if err != nil {
			t.Errorf("Unmarshal error: %s", err)
			continue
		}

		//// debugging
		//if i == 14 {
		//	fmt.Print()
		//}

		got := j.Q(c.P).Elem()
		if got != c.E {
			t.Errorf("case #%d: wrong value, expected=%v, got=%v", i, c.E, got)
		}
	}
}

func TestJ_String(t *testing.T) {
	type testCase struct {
		I J
		O string
	}

	var cases = []testCase{
		{I: M{"name": "Lola"}, O: `{"name":"Lola"}`},
		{I: A{1, 2, 3}, O: `[1,2,3]`},
		{I: F(1), O: `1`},
		{I: S("hello"), O: `hello`},
		{I: A{M{"name": "Lola"}}, O: `[{"name":"Lola"}]`},
		{I: M{"tags": A{1, 2, 3}}, O: `{"tags":[1,2,3]}`},
	}

	for _, c := range cases {
		got := c.I.String()
		if got != c.O {
			t.Errorf("J.String() expected=%s, got=%s", c.O, got)
		}
	}

	if (M{"name": "Lola"}).String() != `{"name":"Lola"}` {
		t.Errorf("")
	}

	var j J = M{"name": "Lola"}
	if fmt.Sprint(j) != `{"name":"Lola"}` {
		t.Errorf("J.String j inteface wrong value")
	}
}

func TestJ_Nil(t *testing.T) {
	j, err := Unmarshal[J](`{"name":"Lola"}`)
	if err != nil {
		t.Fatal(err)
	}

	if j.Q(".id") == nil {
		t.Errorf("didn't expect J value to be nil")
	}
	if !j.Q("id").IsNil() {
		t.Errorf("expected J.IsNil to be true")
	}
	if j.Q(".id").Elem() != nil {
		t.Errorf("expected id to be nil")
	}
	if j.Q(".id").String() != "nil" {
		t.Errorf("expected id to print nil")
	}
	if j.Q(".id").Q(".yaid").Elem() != nil {
		t.Errorf("expected id to be nil")
	}
}

func TestJ_AsFirstValue(t *testing.T) {
	m, _ := mustUnmarshal(`{"key":"value"}`).AsObject()
	if m["key"] != "value" {
		t.Errorf("object value is wrong")
	}

	a, _ := mustUnmarshal(`[1, 2, 3]`).AsArray()
	if len(a) != 3 {
		t.Errorf("array value is wrong")
	}

	n, _ := mustUnmarshal(`1`).AsNumber()
	if n != 1 {
		t.Errorf("number value is wrong")
	}

	s, _ := mustUnmarshal(`{"key":"value"}`).Q("key").AsString()
	if s != "value" {
		t.Errorf("string value is wrong")
	}

	b, _ := mustUnmarshal(`true`).AsBoolean()
	if !b {
		t.Errorf("boolean value is wrong")
	}
}

type AsCheck struct {
	Boolean bool
	Number  bool
	String  bool
	Array   bool
	Object  bool
	IsNil   bool
}

func TestJ_AsSecondValue(t *testing.T) {

	type testCase struct {
		I J
		O AsCheck
	}

	var cases = []testCase{
		{I: mustUnmarshal(`{"key":"value"}`), O: AsCheck{Object: true}},
		{I: mustUnmarshal(`{"outer":{"key":"value"}}`).Q("outer"), O: AsCheck{Object: true}},
		{I: mustUnmarshal(`{"key":[1, 2, 3]}`).Q("key"), O: AsCheck{Array: true}},
		{I: mustUnmarshal(`[1, 2]`), O: AsCheck{Array: true}},
		{I: mustUnmarshal(`0`), O: AsCheck{Number: true}},
		{I: mustUnmarshal(`1`), O: AsCheck{Number: true}},
		{I: mustUnmarshal(`false`), O: AsCheck{Boolean: true}},
		{I: mustUnmarshal(`true`), O: AsCheck{Boolean: true}},
		{I: mustUnmarshal(`{"key":"value"}`).Q("key"), O: AsCheck{String: true}},
		{I: mustUnmarshal(`{}`).Q("key"), O: AsCheck{IsNil: true}},
	}
	for i, c := range cases {
		if c.I.IsNil() != c.O.IsNil {
			t.Errorf("%d nil mismatch", i)
		}
		if _, ok := c.I.AsBoolean(); ok != c.O.Boolean {
			t.Errorf("%d boolean mismatch", i)
		}
		if _, ok := c.I.AsNumber(); ok != c.O.Number {
			t.Errorf("%d number mismatch", i)
		}
		if _, ok := c.I.AsString(); ok != c.O.String {
			t.Errorf("%d string mismatch", i)
		}
		if _, ok := c.I.AsArray(); ok != c.O.Array {
			t.Errorf("%d array mismatch", i)
		}
		if _, ok := c.I.AsObject(); ok != c.O.Object {
			t.Errorf("%d object mismatch", i)
		}
	}
}

func mustUnmarshal(s string) J {
	j, err := Unmarshal[J](s)
	if err != nil {
		panic(err)
	}
	return j
}
