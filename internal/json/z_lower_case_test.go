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
