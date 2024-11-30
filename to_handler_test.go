package fetch

import (
	"bytes"
	"net/http"
	"testing"
)

func TestToHandlerFunc_EmptyIn(t *testing.T) {
	f := ToHandlerFunc(func(in Empty) (J, error) {
		return M{"name": "Lola"}, nil
	})
	mw := newMockWriter()
	r, err := http.NewRequest("", "", bytes.NewBuffer(nil))
	assert(t, err, nil)
	f(mw, r)
	assert(t, mw.status, 200)
	assert(t, string(mw.body), `{"name":"Lola"}`)
}

func TestToHandlerFunc_EmptyOut(t *testing.T) {
	f := ToHandlerFunc(func(in J) (Empty, error) {
		assert(t, in.Q("name").String(), "Charles")
		return Empty{}, nil
	})
	mw := newMockWriter()
	r, err := http.NewRequest("POST", "http:/localhost:7543/pets", bytes.NewBuffer([]byte(`{"name":"Charles"}`)))
	assert(t, err, nil)
	f(mw, r)
	assert(t, mw.status, 200)
	assert(t, string(mw.body), ``)
}
