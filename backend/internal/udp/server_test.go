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

func TestHandleRegisterRespectsCapacity(t *testing.T) {
	serverConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatalf("failed to create server UDP conn: %v", err)
	}
	defer serverConn.Close()

	clientConn1, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatalf("failed to create first client UDP conn: %v", err)
	}
	defer clientConn1.Close()

	clientConn2, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatalf("failed to create second client UDP conn: %v", err)
	}
	defer clientConn2.Close()

	server := NewServer(serverConn.LocalAddr().String(), nil)
	server.conn = serverConn
	server.SetMaxClients(1)

	registerPacket := &Packet{
		Type: PacketTypeRegister,
		Payload: RegisterRequest{
			UserID:   1,
			NovelIDs: []int64{42},
		},
	}

	firstPayload, err := SerializePacket(registerPacket)
	if err != nil {
		t.Fatalf("failed to serialize first packet: %v", err)
	}

	server.handlePacket(context.Background(), firstPayload, clientConn1.LocalAddr().(*net.UDPAddr))

	if server.GetClientCount() != 1 {
		t.Fatalf("expected 1 registered client after first registration")
	}

	secondPacket := &Packet{
		Type: PacketTypeRegister,
		Payload: RegisterRequest{
			UserID:   2,
			NovelIDs: []int64{43},
		},
	}

	secondPayload, err := SerializePacket(secondPacket)
	if err != nil {
		t.Fatalf("failed to serialize second packet: %v", err)
	}

	server.handlePacket(context.Background(), secondPayload, clientConn2.LocalAddr().(*net.UDPAddr))

	// Should not accept additional clients beyond capacity
	if server.GetClientCount() != 1 {
		t.Fatalf("expected capacity limit to prevent additional registrations")
	}

	buf := make([]byte, 1024)
	if err := clientConn2.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatalf("failed to set read deadline for second client: %v", err)
	}

	n, _, err := clientConn2.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("expected error response for capacity limit: %v", err)
	}

	resp, err := ParsePacket(buf[:n])
	if err != nil {
		t.Fatalf("failed to parse capacity response: %v", err)
	}

	if resp.Type != PacketTypeError {
		t.Fatalf("expected error packet, got %s", resp.Type)
	}

	payload, ok := resp.Payload.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map payload for error packet")
	}

	if code, ok := payload["code"].(string); !ok || code != "server_capacity" {
		t.Fatalf("expected server_capacity code, got %#v", payload["code"])
	}
}
