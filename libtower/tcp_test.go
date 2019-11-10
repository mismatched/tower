package libtower

import (
	"net/url"
	"testing"
	"time"
)

func TestTCPPortCheck(t *testing.T) {
	URL, err := url.Parse("tcp://google.com:80")
	if err != nil {
		t.Errorf("test failed %v", err)
	}
	tr := TCP{URL: URL, Timeout: time.Second * 2}
	_, err = tr.TCPPortCheck()
	if err != nil {
		t.Errorf("test failed %v", err)
	}
}
