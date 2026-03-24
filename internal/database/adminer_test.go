package database

import (
	"net"
	"testing"
)

func TestFreePort(t *testing.T) {
	t.Run("returns port in expected range", func(t *testing.T) {
		startPort := 19000
		port, err := freePort(startPort)
		if err != nil {
			t.Fatalf("freePort(%d) error: %v", startPort, err)
		}
		if port < startPort || port >= startPort+20 {
			t.Errorf("freePort(%d) = %d, want in [%d, %d)", startPort, port, startPort, startPort+20)
		}
	})

	t.Run("skips already bound port", func(t *testing.T) {
		// Bind a port so freePort must skip it
		ln, err := net.Listen("tcp", "127.0.0.1:19100")
		if err != nil {
			t.Skip("cannot bind 127.0.0.1:19100 for test setup")
		}
		defer ln.Close()

		port, err := freePort(19100)
		if err != nil {
			t.Fatalf("freePort(19100) error: %v", err)
		}
		if port == 19100 {
			t.Error("freePort should skip the already-bound port 19100")
		}
		if port < 19100 || port >= 19120 {
			t.Errorf("freePort returned %d, want in range [19100, 19120)", port)
		}
	})

	t.Run("returned port is actually free", func(t *testing.T) {
		port, err := freePort(19200)
		if err != nil {
			t.Fatalf("freePort(19200) error: %v", err)
		}
		// Verify we can bind the returned port
		ln, err := net.Listen("tcp", "127.0.0.1:"+intToStr(port))
		if err != nil {
			t.Errorf("port %d returned by freePort is not actually free: %v", port, err)
		} else {
			ln.Close()
		}
	})
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
