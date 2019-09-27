package libtower

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	// ProtocolICMP DSCP
	ProtocolICMP = 1
)

// Ping an address
func Ping(addr string, seq int) (*net.IPAddr, time.Duration, error) {
	// listening for icmp replies
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, 0, err
	}
	defer c.Close()

	// Resolve DNS and get the real IP of the it
	dst, err := DNSLookup(addr)

	// Make a new ICMP message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  seq,
			Data: []byte(""),
		},
	}
	bmsg, err := msg.Marshal(nil)
	if err != nil {
		return dst, 0, err
	}

	// Send ICMP message
	start := time.Now()
	n, err := c.WriteTo(bmsg, dst)
	if err != nil {
		return dst, 0, err
	} else if n != len(bmsg) {
		return dst, 0, fmt.Errorf("got %v; want %v", n, len(bmsg))
	}

	// Wait for an ICMP reply
	reply := make([]byte, 1500)
	err = c.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return dst, 0, err
	}
	n, peer, err := c.ReadFrom(reply)
	if err != nil {
		return dst, 0, err
	}
	duration := time.Since(start)

	rm, err := icmp.ParseMessage(ProtocolICMP, reply[:n])
	if err != nil {
		return dst, 0, err
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		return dst, duration, nil
	default:
		return dst, 0, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
	}
}
