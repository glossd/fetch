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
		Id    string
		Name  string
		Saved bool
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/pets", ToHandlerFunc(func(in Request[Pet]) (*Pet, error) {
		if in.Body.Name != "Lola" {
			t.Errorf("request: name isn't Lola")
		}
		return &Pet{Name: "Lola", Id: "1", Saved: true}, nil
	}))
	server := &http.Server{Addr: ":7349", Handler: mux}
	go server.ListenAndServe()
	defer server.Shutdown(context.Background())
	time.Sleep(time.Millisecond)

	res, err := Post[Pet]("http://localhost:7349/pets", Pet{Name: "Lola"})
	assert(t, err, nil)
	assert(t, res.Id, "1")
	assert(t, res.Name, "Lola")
	assert(t, res.Saved, true)
}
