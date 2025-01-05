package fetch

import "testing"

func TestRequest_SetPathValue(t *testing.T) {
	applyFunc := func(in Request[Empty]) Empty {
		if in.Parameters["key"] == "value" {
			t.Errorf("wrong parameters: %+v", in)
		}
		return Empty{}
	}

	applyFunc(Request[Empty]{}.WithPathValue("key", "value"))
}
