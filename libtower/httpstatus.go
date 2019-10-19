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
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return hsr, err
	}
	hsr.Start = time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return hsr, err
	}
	hsr.End = time.Now()
	
	hsr.Duration = hsr.Start.Sub(hsr.End)
	hsr.StatusCode = resp.StatusCode
	// TODO : add response body
	hsr.Status, hsr.Proto, hsr.ProtoMajor, hsr.ProtoMinor, hsr.Header, hsr.ContentLength, hsr.TransferEncoding, hsr.Close, hsr.Uncompressed, hsr.Trailer = 
		resp.Status, resp.Proto, resp.ProtoMajor, resp.ProtoMinor, resp.Header, resp.ContentLength, resp.TransferEncoding, resp.Close, resp.Uncompressed, resp.Trailer

	return hsr, err
}
