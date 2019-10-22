package libtower

// TCPResult type
type TCPResult struct {
	Host    string
	Port    int
	Timeout time.Duration

	Start    time.Duration
	End      time.Duration
	Duration time.Duration
}

// TCPPortCheck checks if a tcp port is open
func TCPPortCheck(host string, port string) (bool, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return flase, err
	}
	if conn != nil {
		defer conn.Close()
		return true, nil
	}
	return false, nil
}
