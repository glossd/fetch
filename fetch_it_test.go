package fetch

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestRequestIntegration(t *testing.T) {
	mock = false
	defer func() { mock = true }()
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "noone.com")
		w.WriteHeader(303)
		_, err := w.Write([]byte("wrong neighborhood"))
		if err != nil {
			panic(err)
		}
	})

	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(200)
			w.Write([]byte(`{"name":"Lola"}`))
		} else {
			w.WriteHeader(405)
			w.Write([]byte(`{"message":"get out"}`))
		}
	})

	server := &http.Server{Addr: ":7349", Handler: mux}
	go server.ListenAndServe()
	defer server.Shutdown(context.Background())
	time.Sleep(time.Millisecond)

	_, err := Get[string]("localhost:7349/hello")
	if err == nil {
		t.Fatalf("expected 303 status error")
	}

	if err.(*Error).Status != 303 {
		t.Fatalf("expected 303 status, got=%d", err.(*Error).Status)
	}
	if err.(*Error).Headers["Access-Control-Allow-Origin"] != "noone.com" {
		t.Fatalf("expected header, got=%s", err.(*Error).Headers["Access-Control-Allow-Origin"])
	}
	if err.Error() != "http status=303, body=wrong neighborhood" {
		t.Fatalf("wrong error message, got=%s", err)
	}

	type Pet struct {
		Name string
	}
	p, err := Get[Pet]("localhost:7349/get")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Lola" {
		t.Errorf("unexpected name, got=%s", p.Name)
	}

	_, err = Post[Pet]("localhost:7349/get", "i'm post")
	if err.Error() != `http status=405, body={"message":"get out"}` {
		t.Errorf("expected 405 status error, got=%s", err)
	}
}

func TestIssue1(t *testing.T) {
	mock = false
	defer func() { mock = true }()
	mux := http.NewServeMux()
	mux.HandleFunc("/v3/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, err := w.Write([]byte("all good"))
		if err != nil {
			panic(err)
		}
	})
	server := &http.Server{Addr: ":7349", Handler: mux}
	go server.ListenAndServe()
	defer server.Shutdown(context.Background())
	time.Sleep(time.Millisecond)

	SetBaseURL("http://localhost:7349/v3")
	defer SetBaseURL("")

	res, err := Post[string]("/sessions", M{"key": "value"})
	if err != nil {
		t.Fatalf("Got error: %s", err)
	}

	if res != "all good" {
		t.Errorf("wrong response")
	}
}
