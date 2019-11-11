package libtower

import (
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"time"
)

// Trace http
func (ht *HTTPTrace) Trace(url, method string) error {
	// res := HTTPTrace{URL: url, Method: method}

	req, err := http.NewRequest(ht.Method, ht.URL, nil)
	if err != nil {
		return err
	}

	var start, connect, dns, tlsHandshake time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			ht.DNS = time.Since(dns)
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			ht.TLSHandshake = time.Since(tlsHandshake)
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			ht.Connect = time.Since(connect)
		},

		GotFirstResponseByte: func() {
			ht.Connect = time.Since(start)
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
		return err
	}
	ht.Total = time.Since(start)

	return nil
}
