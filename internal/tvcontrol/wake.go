package tvcontrol

import (
	"fmt"
	"net"
	"strings"
)

func Wake(options WakeOptions) (WakeResult, error) {
	mac, err := net.ParseMAC(strings.TrimSpace(options.MAC))
	if err != nil {
		return WakeResult{}, fmt.Errorf("parse mac address: %w", err)
	}

	target := strings.TrimSpace(options.Broadcast)
	if target == "" {
		target = "255.255.255.255:9"
	}

	conn, err := net.Dial("udp", target)
	if err != nil {
		return WakeResult{}, fmt.Errorf("open udp socket: %w", err)
	}
	defer conn.Close()

	packet := wakePacket(mac)
	n, err := conn.Write(packet)
	if err != nil {
		return WakeResult{}, fmt.Errorf("send magic packet: %w", err)
	}

	return WakeResult{
		Target:    target,
		MAC:       strings.ToLower(mac.String()),
		SentBytes: n,
		OK:        n == len(packet),
	}, nil
}

func wakePacket(mac net.HardwareAddr) []byte {
	packet := make([]byte, 6+16*len(mac))
	for i := 0; i < 6; i++ {
		packet[i] = 0xff
	}
	for i := 6; i < len(packet); i += len(mac) {
		copy(packet[i:i+len(mac)], mac)
	}
	return packet
}
