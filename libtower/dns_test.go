package libtower

import (
	"net"
	"reflect"
	"testing"
)

func TestDNSLookup(t *testing.T) {
	addr, _, err := DNSLookup("google.com")
	if err != nil {
		t.Errorf("test failed %v", err)
	}
	var b *net.IPAddr
	if reflect.TypeOf(addr) != reflect.TypeOf(b) {
		t.Errorf("its not a valid ip address")
	}
}

func TestDNSLookupFrom(t *testing.T) {
	addr, _, err := DNSLookupFrom("google.com", "8.8.8.8")
	if err != nil {
		t.Errorf("test failed %v", err)
	}
	var b *net.IPAddr
	if reflect.TypeOf(addr) != reflect.TypeOf(b) {
		t.Errorf("its not a valid ip address")
	}
}
