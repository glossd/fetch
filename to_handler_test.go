package fetch

import (
	"bytes"
	"context"
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

func TestToHandlerFunc_MultiplePathValue(t *testing.T) {
	type Pet struct {
		Category string `pathval:"category"`
		Id       string `pathval:"id"`
		Name     string
	}
	f := ToHandlerFunc(func(in Pet) (Empty, error) {
		assert(t, in.Category, "cats")
		assert(t, in.Id, "1")
		assert(t, in.Name, "Charles")
		return Empty{}, nil
	})
	mw := newMockWriter()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /categories/{category}/ids/{id}", f)
	r, err := http.NewRequest("POST", "/categories/cats/ids/1", bytes.NewBuffer([]byte(`{"name":"Charles"}`)))
	assert(t, err, nil)
	mux.ServeHTTP(mw, r)
	assert(t, mw.status, 200)
}

func TestToHandlerFunc_PathvalParseInt(t *testing.T) {
	type Pet struct {
		Id   int `pathval:"id"`
		Name string
	}
	f := ToHandlerFunc(func(in Pet) (Empty, error) {
		assert(t, in.Id, 1)
		assert(t, in.Name, "Charles")
		return Empty{}, nil
	})
	mw := newMockWriter()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /ids/{id}", f)
	r, err := http.NewRequest("POST", "/ids/1", bytes.NewBuffer([]byte(`{"name":"Charles"}`)))
	assert(t, err, nil)
	mux.ServeHTTP(mw, r)
	assert(t, mw.status, 200)
}

func TestToHandlerFunc_GetWithPathvalAndNothingToUnmarshal(t *testing.T) {
	type Pet struct {
		Id int `pathval:"id"`
	}
	f := ToHandlerFunc(func(in Pet) (Empty, error) {
		assert(t, in.Id, 1)
		return Empty{}, nil
	})
	mw := newMockWriter()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ids/{id}", f)
	r, err := http.NewRequest("GET", "/ids/1", bytes.NewBuffer([]byte(``)))
	assert(t, err, nil)
	mux.ServeHTTP(mw, r)
	assert(t, mw.status, 200)
}

func TestToHandlerFunc_J(t *testing.T) {
	f := ToHandlerFunc(func(in J) (J, error) {
		assert(t, in.Q("name").String(), "Lola")
		return M{"status": "ok"}, nil
	})
	mw := newMockWriter()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /j", f)
	r, err := http.NewRequest("POST", "/j", bytes.NewBuffer([]byte(`{"name":"Lola"}`)))
	assert(t, err, nil)
	mux.ServeHTTP(mw, r)
	assert(t, mw.status, 200)
	assert(t, string(mw.body), `{"status":"ok"}`)
}

func TestToHandlerFunc_Header(t *testing.T) {
	type Pet struct {
		Content string `header:"Content"`
	}
	f := ToHandlerFunc(func(in Pet) (Empty, error) {
		assert(t, in.Content, "mycontent")
		return Empty{}, nil
	})
	mw := newMockWriter()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /pets", f)
	r, err := http.NewRequest("POST", "/pets", bytes.NewBuffer([]byte(`{}`)))
	r.Header.Set("Content", "mycontent")
	assert(t, err, nil)
	mux.ServeHTTP(mw, r)
	assert(t, mw.status, 200)
}

func TestToHandlerFunc_Context(t *testing.T) {
	type Pet struct {
		Context context.Context
		Name    string
	}
	f := ToHandlerFunc(func(in Pet) (Empty, error) {
		assert(t, in.Context.Err(), nil)
		assert(t, in.Name, "Lola")
		return Empty{}, nil
	})
	mw := newMockWriter()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /pets", f)
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
	mux.HandleFunc("POST /pets", f)
	r, err := http.NewRequest("POST", "/pets", bytes.NewBuffer([]byte(`{}`)))
	assert(t, err, nil)
	mux.ServeHTTP(mw, r)
	assert(t, mw.status, 422)
}
