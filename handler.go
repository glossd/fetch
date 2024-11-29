package fetch

import (
	"fmt"
	"io"
	"net/http"
)

var defaultHandleConfig = HandleConfig{
	ErrorHook: func(err error) {
		fmt.Printf("fetch.Handle failed to respond: %s\n", err)
	},
}

// SetDefaultHandleConfig sets default HandleConfig globally.
func SetDefaultHandleConfig(hc HandleConfig) {
	defaultHandleConfig = hc
}

type HandleConfig struct {
	// ErrorHook is called if an error happens while sending an HTTP response
	ErrorHook func(err error)
}

// HandlerFunc represents a simple function with
// In type as a request body and Out type as a response body.
type HandlerFunc[In any, Out any] func(in In) (Out, error)

// ToHandlerFunc converts HandlerFunc into http.HandlerFunc,
// which can be used later in http.ServeMux#HandleFunc.
// It unmarshals the HTTP request body into the HandlerFunc argument and
// then marshals the returned value into the HTTP response body.
func ToHandlerFunc[In any, Out any](h HandlerFunc[In, Out], config ...HandleConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := defaultHandleConfig
		if len(config) > 0 {
			cfg = config[0]
		}
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			err = RespondError(w, 400, err)
			if err != nil {
				cfg.ErrorHook(err)
			}
			return
		}
		in, err := Unmarshal[In](string(reqBody))
		if err != nil {
			err = RespondError(w, 400, err)
			if err != nil {
				cfg.ErrorHook(err)
			}
			return
		}
		out, err := h(in)
		if err != nil {
			err = RespondError(w, 500, err)
			if err != nil {
				cfg.ErrorHook(err)
			}
			return
		}
		err = Respond(w, out)
		if err != nil {
			cfg.ErrorHook(err)
		}
	}
}
