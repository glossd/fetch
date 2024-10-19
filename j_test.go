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

		got := j.Q(c.P).Raw()
		if got != c.E {
			t.Errorf("case #%d: wrong value, expected=%v, got=%v", i, c.E, got)
		}
	}

	errorCases := []testCase{
		{In: `[1, 2, 3]`, P: ".[1", E: "expected ] for array index"},
		{In: `[1, 2, 3]`, P: ".[hello]", E: "expected a number for array index, got: 'hello'"},
		{In: `[1, 2, 3]`, P: ".[0]name", E: "expected . or [, got: 'name'"},
	}

	for i, c := range errorCases {
		j, err := Unmarshal[J](c.In)
		if err != nil {
			t.Errorf("Unmarshal error: %s", err)
			continue
		}

		got := j.Q(c.P)
		errStr, ok := c.E.(string)
		if !ok {
			panic("E should be string")
		}
		jqErr, ok := got.(*JQError)
		if !ok {
			t.Errorf("error case #%d: expected error, got=%v", i, got)
			continue
		}
		if jqErr.s != errStr {
			t.Errorf("error case #%d: wrong value, expected=%v, got=%v", i, errStr, jqErr.s)
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
	fmt.Println(j.Q(".id"))
	if j.Q(".id").Raw() != nil {
		t.Errorf("expected id to be nil")
	}
	if j.Q(".id").String() != "nil" {
		t.Errorf("expected id to print nil")
	}
	if j.Q(".id").Q(".yaid").Raw() != nil {
		t.Errorf("expected id to be nil")
	}
}
