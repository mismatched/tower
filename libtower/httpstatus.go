package libtower

import (
	"net/http"
	"time"
)

// HTTPStatus check
func (hsr *HTTP) HTTPStatus() error {
	// setup client
	client := &http.Client{}
	req, err := http.NewRequest(hsr.Method, hsr.URL, nil)
	if err != nil {
		return err
	}
	hsr.Start = time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	hsr.End = time.Now()

	hsr.Duration = hsr.End.Sub(hsr.Start)
	hsr.StatusCode = resp.StatusCode
	// TODO : add response body
	hsr.Status, hsr.Proto, hsr.ProtoMajor, hsr.ProtoMinor, hsr.Header, hsr.ContentLength, hsr.TransferEncoding, hsr.Close, hsr.Uncompressed, hsr.Trailer =
		resp.Status, resp.Proto, resp.ProtoMajor, resp.ProtoMinor, resp.Header, resp.ContentLength, resp.TransferEncoding, resp.Close, resp.Uncompressed, resp.Trailer

	return err
}
