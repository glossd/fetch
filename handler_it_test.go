package fetch

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestHandle(t *testing.T) {
	mock = false
	defer func() { mock = true }()
	type Pet struct {
		Id   int
		Name string
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/pets", ToHandlerFunc(func(in Pet) (Pet, error) {
		if in.Name != "Lola" {
			t.Errorf("request: name isn't Lola")
		}
		in.Id = 1
		return in, nil
	}))
	server := &http.Server{Addr: ":7349", Handler: mux}
	go server.ListenAndServe()
	defer server.Shutdown(context.Background())
	time.Sleep(time.Millisecond)

	res, err := Post[Pet]("http://localhost:7349/pets", Pet{Name: "Lola"})
	if err != nil {
		t.Fatalf("Got error: %s", err)
	}
	if res.Id != 1 {
		t.Errorf("response: expected id 1")
	}
	if res.Name != "Lola" {
		t.Errorf("response: expected name Lola")
	}
}
