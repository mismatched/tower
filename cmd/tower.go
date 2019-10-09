// tower
package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "tower"
	app.Usage = "network uptime and status checker"
	app.Version = "0.0.1"
	app.Action = func(c *cli.Context) error {
		tower(c.Args().Get(0), c.Args().Get(1))
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

func tower(url, method string) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(0)
	}

	var start, connect, dns, tlsHandshake time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			fmt.Printf("DNS Done: %v\n", time.Since(dns))
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			fmt.Printf("TLS Handshake: %v\n", time.Since(tlsHandshake))
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			fmt.Printf("Connect time: %v\n", time.Since(connect))
		},

		GotFirstResponseByte: func() {
			fmt.Printf("Time from start to first byte: %v\n", time.Since(start))
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	if _, err := http.DefaultTransport.RoundTrip(req); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total time: %v\n", time.Since(start))
}

func getMethod(arg string) string {
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
	arg = strings.ToUpper(arg)
	for _, m := range methods {
		if m == arg {
			return arg
		}
	}
	return "GET"
}
