// Package address holds the one rule about where a live connection may go: it
// stays on this computer (ADR 0016). There is no authentication, so it is a
// development tool and not a channel to manage an application from far away.
package address

import (
	"fmt"
	"net"
)

// Loopback reports an error unless addr (host:port) is on this computer.
func Loopback(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("address %q: %w", addr, err)
	}
	if host == "localhost" {
		return nil
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return nil
	}
	return fmt.Errorf("address %q is not on this computer: a live connection uses 127.0.0.1, ::1 or localhost", addr)
}
