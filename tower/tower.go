// tower
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dariubs/fmt2"
	"github.com/mismatched/tower/libtower"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "tower"
	app.Usage = "network uptime and status checker"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Dariush Abbasi",
			Email: "poshtehani@gmail.com",
		},
		cli.Author{
			Name:  "Hasan Aminfar",
			Email: "aminfar69@gmail.com",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "ping, p",
			Usage: "ping a url. you must run it as root",
		},
		cli.StringFlag{
			Name:  "dns",
			Usage: "dns resolve time of an address",
		},
		cli.StringFlag{
			Name:  "trace",
			Usage: "http trace time",
		},
		cli.StringFlag{
			Name:  "http-status, s",
			Usage: "http status time",
		},
	}
	app.Action = ActionHandler

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

// ActionHandler handle cli actions
func ActionHandler(c *cli.Context) error {
	if c.String("ping") != "" {
		r, d, err := libtower.Ping(c.String("ping"), 1)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}

		fmt.Printf("Ping %s in %v ms\n", r, d)
		return nil
	} else if c.String("dns") != "" {
		r, d, err := libtower.DNSLookup(c.String("dns"))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}

		fmt.Printf("DNS of %s with %s ip resolves in %v ms\n", c.String("dns"), r, d)
		return nil
	} else if c.String("trace") != "" {
		// TODO: get http method from user
		// TODO: use http:// schema if for urls with no schemas
		r := libtower.HTTPTrace{URL: c.String("trace"), Method: "GET"}
		err := r.Trace()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}
		if r.DNS != 0 {
			fmt.Printf("DNS Done: %v\n", r.DNS)
		}
		if r.TLSHandshake != 0 {
			fmt.Printf("TLS Handshake: %v\n", r.TLSHandshake)
		}
		if r.Connect != 0 {
			fmt.Printf("Connect time: %v\n", r.Connect)
		}
		if r.GotFirstResponseByte != 0 {
			fmt.Printf("Time from start to first byte: %v\n", r.GotFirstResponseByte)
		}
		if r.Total != 0 {
			fmt.Printf("Total time: %v\n", r.Total)
		}
		return nil
	} else if c.String("http-status") != "" {
		client := libtower.HTTP{URL: c.String("http-status"), Method: "GET"}
		err := client.HTTPStatus()
		if err != nil {
			fmt2.Printlnf("Error: %v\n", err)
			return err
		}
		fmt2.Printlnf("%s with status %s code %d in %v", c.String("http-status"), client.Status, client.StatusCode, client.Duration)
		return nil
	} else {
		fmt.Println("Command not found in Tower")
	}

	return nil
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
