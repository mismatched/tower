package libtower

import "testing"

func TestTCPPortCheck(t *testing.T) {
	_, err := TCPPortCheck("google.com", "80")
	if err != nil {
		t.Errorf("test failed %v", err)
	}
}
