package libtower

import "time"

// HTTPStatusResult type
type HTTPStatusResult struct {
	URL    string
	Method string

	Status           string // e.g. "200 OK"
	StatusCode       int
	Proto            string // e.g. "HTTP/1.0"
	ProtoMajor       int    // e.g. 1
	ProtoMinor       int    // e.g. 0
	Header           map[string][]string
	Body             []byte
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Uncompressed     bool
	Trailer          map[string][]string

	// TODO: add tls fields

	Start    time.Time
	End      time.Time
	Duration time.Duration
}

// HTTPStatus check
func HTTPStatus(url, method string) (HTTPStatusResult, error) {
	return HTTPStatusResult{}, nil
}
