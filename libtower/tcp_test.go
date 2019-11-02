package libtower

import (
	"testing"
	"time"
)

func TestTCPPortCheck(t *testing.T) {
	tr := TCPResult{Host: "google.com", Port: "80", Timeout: time.Second * 2}
	_, err := tr.TCPPortCheck()
	if err != nil {
		t.Errorf("test failed %v", err)
	}
}
