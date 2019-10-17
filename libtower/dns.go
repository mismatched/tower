package libtower

import (
	"context"
	"errors"
	"net"
	"net/http/httptrace"
	"time"
)

// DNSLookup func
func DNSLookup(addr string) (*net.IPAddr, time.Duration, error) {
	var dns time.Time
	var DNS time.Duration
	traceDNS := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) {
			dns = time.Now()
		},
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			DNS = time.Since(dns)
		},
	}
	ctx := httptrace.WithClientTrace(context.Background(), traceDNS)
	ips, err := (&net.Resolver{}).LookupIPAddr(ctx, addr)
	if err != nil {
		return new(net.IPAddr), DNS, err
	}
	if len(ips) == 0 {
		return new(net.IPAddr), DNS, errors.New("ips len is zero")
	}
	return &ips[0], DNS, nil
}
