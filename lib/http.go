package libtower

import (
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"time"
)

// HTTPResult type
type HTTPResult struct {
	URL                  string
	Method               string
	DNS                  time.Duration
	TLSHandshake         time.Duration
	Connect              time.Duration
	GotFirstResponseByte time.Duration
	Total                time.Duration
}

// HTTP check
func HTTP(url, method string) (HTTPResult, error) {
	res := HTTPResult{URL: url, Method: method}

	req, err := http.NewRequest(res.Method, res.URL, nil)
	if err != nil {
		return res, err
	}

	var start, connect, dns, tlsHandshake time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			res.DNS = time.Since(dns)
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			res.TLSHandshake = time.Since(tlsHandshake)
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			res.Connect = time.Since(connect)
		},

		GotFirstResponseByte: func() {
			res.Connect = time.Since(start)
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
		return res, err
	}
	res.Total = time.Since(start)

	return res, nil
}
