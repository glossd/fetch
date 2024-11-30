package fetch

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
)

var defaultHandlerConfig = HandlerConfig{
	ErrorHook: func(err error) {
		fmt.Printf("fetch.Handle failed to respond: %s\n", err)
	},
}

// SetDefaultHandleConfig sets default HandlerConfig globally.
func SetDefaultHandlerConfig(hc HandlerConfig) {
	defaultHandlerConfig = hc
}

type HandlerConfig struct {
	// ErrorHook is called if an error happens while sending an HTTP response
	ErrorHook func(err error)
}

// ApplyFunc represents a simple function to be converted to http.Handler with
// In type as a request body and Out type as a response body.
type ApplyFunc[In any, Out any] func(in In) (Out, error)

// ToHandlerFunc converts ApplyFunc into http.HandlerFunc,
// which can be used later in http.ServeMux#HandleFunc.
// It unmarshals the HTTP request body into the ApplyFunc argument and
// then marshals the returned value into the HTTP response body.
func ToHandlerFunc[In any, Out any](apply ApplyFunc[In, Out], config ...HandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := defaultHandlerConfig
		if len(config) > 0 {
			cfg = config[0]
		}
		var in In
		if typeFor[In]() != reflect.TypeOf(Empty{}) {
			reqBody, err := io.ReadAll(r.Body)
			if err != nil {
				cfg.ErrorHook(err)
				err = RespondError(w, 400, err)
				if err != nil {
					cfg.ErrorHook(err)
				}
				return
			}
			in, err = Unmarshal[In](string(reqBody))
			if err != nil {
				cfg.ErrorHook(err)
				err = RespondError(w, 400, err)
				if err != nil {
					cfg.ErrorHook(err)
				}
				return
			}
		}
		out, err := apply(in)
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
