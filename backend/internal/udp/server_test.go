package udp

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestHandleRegisterAddsClientAndConfirms(t *testing.T) {
	// Create UDP connections for server and client
	serverConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatalf("failed to create server UDP conn: %v", err)
	}
	defer serverConn.Close()

	clientConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatalf("failed to create client UDP conn: %v", err)
	}
	defer clientConn.Close()

	server := NewServer(serverConn.LocalAddr().String(), nil)
	server.conn = serverConn

	packet := &Packet{
		Type: PacketTypeRegister,
		Payload: RegisterRequest{
			UserID:   42,
			NovelIDs: []int64{1, 2},
			DeviceID: "device-123",
		},
	}

	data, err := SerializePacket(packet)
	if err != nil {
		t.Fatalf("failed to serialize packet: %v", err)
	}

	// Invoke handler directly to simulate a received packet
	server.handlePacket(context.Background(), data, clientConn.LocalAddr().(*net.UDPAddr))

	// Ensure client was registered
	if got := server.GetClientCount(); got != 1 {
		t.Fatalf("expected 1 registered client, got %d", got)
	}

	// Expect confirmation packet
	buf := make([]byte, 1024)
	if err := clientConn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("failed to set read deadline: %v", err)
	}

	n, _, err := clientConn.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("failed to read confirmation: %v", err)
	}

	resp, err := ParsePacket(buf[:n])
	if err != nil {
		t.Fatalf("failed to parse confirmation: %v", err)
	}

	if resp.Type != PacketTypeConfirm {
		t.Fatalf("expected confirm packet, got %s", resp.Type)
	}

	payloadMap, ok := resp.Payload.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map payload, got %T", resp.Payload)
	}

	if success, ok := payloadMap["success"].(bool); !ok || !success {
		t.Fatalf("expected success confirmation, got %#v", payloadMap["success"])
	}
}
