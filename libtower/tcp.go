package libtower

import (
	"net"
	"time"
)

// TCPResult type
type TCPResult struct {
	Host    string
	Port    string
	Timeout time.Duration

	Start    time.Time
	End      time.Time
	Duration time.Duration
}

// TCPPortCheck checks if a tcp port is open
func (tr *TCPResult) TCPPortCheck() (bool, error) {
	tr.Start = time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(tr.Host, tr.Port), tr.Timeout)
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
