package tvcontrol

import (
	"net"
	"testing"
)

func TestWakePacket(t *testing.T) {
	mac, err := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("parse mac: %v", err)
	}

	packet := wakePacket(mac)
	if len(packet) != 102 {
		t.Fatalf("expected wake packet length 102, got %d", len(packet))
	}
	for i := 0; i < 6; i++ {
		if packet[i] != 0xff {
			t.Fatalf("expected packet prefix at %d", i)
		}
	}
}
