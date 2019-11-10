package libtower

import (
	"net"
	"net/url"
	"time"
)

// TCP type
type TCP struct {
	URL     *url.URL
	Timeout time.Duration

	Start    time.Time
	End      time.Time
	Duration time.Duration
}

// TCPPortCheck checks if a tcp port is open
func (tr *TCP) TCPPortCheck() (bool, error) {
	tr.Start = time.Now()
	conn, err := net.DialTimeout("tcp", tr.URL.Host, tr.Timeout)
	tr.End = time.Now()
	tr.Duration = tr.End.Sub(tr.Start)
	if err != nil {
		return false, err
	}
	if conn != nil {
		defer conn.Close()
		return true, nil
	}
	return false, nil
}
