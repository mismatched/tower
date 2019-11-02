package libtower

import (
	"context"
	"errors"
	"net"
	"net/http/httptrace"
	"time"

	"github.com/miekg/dns"
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

//DNSLookupFrom func
func DNSLookupFrom(addr string, server string) (*net.IPAddr, time.Duration, error) {
	severIP := net.ParseIP(server)
	if severIP == nil {
		return new(net.IPAddr), time.Duration(0), errors.New("failed to parse server ip address")
	}
	serverAddress := server + ":53"

	msg := dns.Msg{}
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	msg.Question = []dns.Question{dns.Question{Name: dns.Fqdn(addr), Qtype: dns.TypeA, Qclass: dns.ClassINET}}

	client := dns.Client{Net: "udp"}
	resp, rtt, err := client.Exchange(&msg, serverAddress)

	if err != nil {
		return new(net.IPAddr), rtt, errors.New("dns exchange error: " + err.Error())
	}
	if resp == nil {
		return new(net.IPAddr), rtt, errors.New("response is nil")
	}
	if resp != nil && resp.Rcode != dns.RcodeSuccess {
		return new(net.IPAddr), rtt, errors.New(dns.RcodeToString[resp.Rcode])
	}
	for _, record := range resp.Answer {
		if t, ok := record.(*dns.A); ok {
			ipAddress := net.IPAddr{IP: t.A}
			return &ipAddress, rtt, nil
		}
	}
	return new(net.IPAddr), rtt, errors.New("record a not find in response")
}
