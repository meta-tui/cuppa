package address

import "testing"

func TestOnlyThisComputerIsAllowed(t *testing.T) {
	for addr, ok := range map[string]bool{
		"127.0.0.1:7331": true, "localhost:7331": true, "[::1]:7331": true, "127.0.0.2:1": true,
		":7331": false, "0.0.0.0:7331": false, "192.168.1.5:7331": false, "example.com:80": false, "nonsense": false,
	} {
		if err := Loopback(addr); (err == nil) != ok {
			t.Errorf("%s: allowed=%v, error %v", addr, ok, err)
		}
	}
}
