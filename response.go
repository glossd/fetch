package fetch

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
	Status int
	// HTTP headers are not unique.
	// In the majority of the cases Headers is enough.
	// Headers are filled with the last value from DuplicateHeaders.
	DuplicateHeaders map[string][]string
	Headers          map[string]string
	Body             T
	BodyBytes        []byte
}

// ResponseEmpty is a special ResponseType that completely ignores the HTTP body.
// Can be used as the (generic) ReturnType for any HTTP method.
type ResponseEmpty struct {
	Status           int
	Headers          map[string]string
	DuplicateHeaders map[string][]string
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

// Empty represents an empty response or request body, skipping JSON handling.
// Can be used in any HTTP method or to fit the signature of ApplyFunc.
type Empty struct{}
