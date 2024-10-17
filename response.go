package fetch

type Response[T any] struct {
	Status           int
	DuplicateHeaders map[string][]string
	Body             T
	BodyBytes        []byte
}

func (r Response[T]) Headers() map[string]string {
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
