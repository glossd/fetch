package fetch

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestToHandlerFunc(t *testing.T) {
	mock = false
	defer func() { mock = true }()
	type Pet struct {
		Id    string `pathval:"id"`
		Name  string
		Saved bool
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/pets/{id}", ToHandlerFunc(func(in *Pet) (*Pet, error) {
		assert(t, in.Id, "1")
		if in.Name != "Lola" {
			t.Errorf("request: name isn't Lola")
		}
		in.Saved = true
		return in, nil
	}))
	server := &http.Server{Addr: ":7349", Handler: mux}
	go server.ListenAndServe()
	defer server.Shutdown(context.Background())
	time.Sleep(time.Millisecond)

	res, err := Post[Pet]("http://localhost:7349/pets/1", Pet{Name: "Lola"})
	assert(t, err, nil)
	assert(t, res.Id, "1")
	assert(t, res.Name, "Lola")
	assert(t, res.Saved, true)
}
