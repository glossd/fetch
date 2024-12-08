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

// This test should fail to compile on go1.21 and successfully run on go1.22.
// Don't forget to update go.mod to 1.22 before running.
//func TestToHandlerFunc_MultiplePathValue(t *testing.T) {
//	type Pet struct {
//		Category string
//		Id       string
//		Name     string
//	}
//	f := ToHandlerFunc(func(in Request[Pet]) (Empty, error) {
//		if in.PathValue("category") != "cats" || in.PathValue("id") != "1" {
//			t.Errorf("wrong request, got %v", in)
//		}
//		return Empty{}, nil
//	})
//	mw := newMockWriter()
//	mux := http.NewServeMux()
//	mux.HandleFunc("/categories/{category}/ids/{id}", f)
//	r, err := http.NewRequest("POST", "/categories/cats/ids/1", bytes.NewBuffer([]byte(`{"name":"Charles"}`)))
//	assert(t, err, nil)
//	mux.ServeHTTP(mw, r)
//	assert(t, mw.status, 200)
//}

func TestToHandlerFunc_J(t *testing.T) {
	f := ToHandlerFunc(func(in J) (J, error) {
		assert(t, in.Q("name").String(), "Lola")
		return M{"status": "ok"}, nil
	})
	mw := newMockWriter()
	mux := http.NewServeMux()
	mux.HandleFunc("/j", f)
	r, err := http.NewRequest("POST", "/j", bytes.NewBuffer([]byte(`{"name":"Lola"}`)))
	assert(t, err, nil)
	mux.ServeHTTP(mw, r)
	assert(t, mw.status, 200)
	assert(t, string(mw.body), `{"status":"ok"}`)
}

func TestToHandlerFunc_Header(t *testing.T) {
	f := ToHandlerFunc(func(in Request[Empty]) (Empty, error) {
		if in.Headers["Content"] != "mycontent" {
			t.Errorf("wrong in %v", in)
		}
		return Empty{}, nil
	})
	mw := newMockWriter()
	mux := http.NewServeMux()
	mux.HandleFunc("/pets", f)
	r, err := http.NewRequest("POST", "/pets", bytes.NewBuffer([]byte(`{}`)))
	r.Header.Set("Content", "mycontent")
	assert(t, err, nil)
	mux.ServeHTTP(mw, r)
	assert(t, mw.status, 200)
}

func TestToHandlerFunc_Context(t *testing.T) {
	f := ToHandlerFunc(func(in Request[Empty]) (Empty, error) {
		assert(t, in.Context().Err(), nil)
		return Empty{}, nil
	})
	mw := newMockWriter()
	mux := http.NewServeMux()
	mux.HandleFunc("/pets", f)
	r, err := http.NewRequest("POST", "/pets", bytes.NewBuffer([]byte(`{"name":"Lola"}`)))
	assert(t, err, nil)
	mux.ServeHTTP(mw, r)
	assert(t, mw.status, 200)
}

func TestToHandlerFunc_Middleware(t *testing.T) {
	SetDefaultHandlerConfig(HandlerConfig{
		Middleware: func(w http.ResponseWriter, r *http.Request) bool {
			w.WriteHeader(422)
			return true
		},
	})
	defer SetDefaultHandlerConfig(HandlerConfig{Middleware: func(w http.ResponseWriter, r *http.Request) bool {
		return false
	}})

	f := ToHandlerFunc(func(in Empty) (Empty, error) {
		return Empty{}, nil
	})
	mw := newMockWriter()
	mux := http.NewServeMux()
	mux.HandleFunc("/pets", f)
	r, err := http.NewRequest("POST", "/pets", bytes.NewBuffer([]byte(`{}`)))
	assert(t, err, nil)
	mux.ServeHTTP(mw, r)
	assert(t, mw.status, 422)
}
