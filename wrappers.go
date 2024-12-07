package fetch

import "net/http"

/*
Response is a wrapper type for (generic) ReturnType to be used in
the HTTP methods. It allows you to access HTTP attributes
of the HTTP response and unmarshal the HTTP body.
e.g.

	type User struct {
		FirstName string
	}
	res, err := Get[Response[User]]("/users/1")
	if err != nil {panic(err)}
	if res.Status != 202 {
	   panic("unexpected status")
	}
	// Body is User type
	fmt.Println(res.Body.FirstName)
*/
type Response[T any] struct {
	Status  int
	Headers map[string]string
	Body    T
}

func uniqueHeaders(headers map[string][]string) map[string]string {
	h := make(map[string]string, len(headers))
	for key, val := range headers {
		if len(val) > 0 {
			// it takes the last element intentionally.
			h[key] = val[len(val)-1]
		}
	}
	return h
}

/*
Request can be used in ApplyFunc as a wrapper
for the input entity to access http attributes.
e.g.

	type Pet struct {
		Name string
	}
	http.HandleFunc("POST /pets/{id}", fetch.ToHandlerFunc(func(in fetch.Request[Pet]) (fetch.Empty, error) {
		in.Context()
		return fetch.Empty{}, nil
	}))
*/
type Request[T any] struct {
	*http.Request
	Headers map[string]string
	Body    T
}

// Empty represents an empty response or request body, skipping JSON handling.
// Can be used with the wrappers Response and Request or to fit the signature of ApplyFunc.
type Empty struct{}
