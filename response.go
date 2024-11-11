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
	Status           int
	DuplicateHeaders map[string][]string
	Body             T
	BodyBytes        []byte
}

func (r Response[T]) Headers() map[string]string {
	return uniqueHeaders(r.DuplicateHeaders)
}

// ResponseEmpty is a special ResponseType that completely ignores the HTTP body.
// Can be used as the (generic) ReturnType for any HTTP method.
type ResponseEmpty struct {
	Status           int
	DuplicateHeaders map[string][]string
}

func (r ResponseEmpty) Headers() map[string]string {
	return uniqueHeaders(r.DuplicateHeaders)
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
