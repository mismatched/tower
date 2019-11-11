package libtower

import "time"

// HTTP type
type HTTP struct {
	URL     string
	Method  string
	Timeout Timeout

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

	Time
}

// HTTPTrace type
type HTTPTrace struct {
	URL                  string
	Method               string
	DNS                  time.Duration
	TLSHandshake         time.Duration
	Connect              time.Duration
	GotFirstResponseByte time.Duration
	Total                time.Duration

	Time
}
