package libtower

import (
	"net/http"
	"time"
)

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
	// setup client
	hsr := HTTPStatusResult{}
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		return hsr, err
	}

	req.Header.Add("If-None-Match", `W/"wyzzy"`)
	resp, err := client.Do(req)
	if err != nil {
		return hsr, err
	}

	hsr.Start = time.Now()
	hsr.StatusCode = resp.StatusCode
	hsr.End = time.Now()
	hsr.Duration = hsr.Start.Sub(hsr.End)

	return hsr, nil
}
