package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"time"
)

const (
	ServerHost = "localhost"
	ServerPort = 8080
)

type GameClient struct {
	conn     *net.UDPConn
	playerID uint32
	sequence uint32
}

func NewGameClient() *GameClient {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", ServerHost, ServerPort))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatal(err)
	}

	return &GameClient{
		conn:     conn,
		sequence: 1,
	}
}

func (gc *GameClient) Join() {
	// Send join packet
	joinPacket := make([]byte, 13)
	joinPacket[0] = 1                              // PacketJoin
	binary.BigEndian.PutUint32(joinPacket[1:5], 0) // PlayerID (0 for new player)
	binary.BigEndian.PutUint64(joinPacket[5:13], uint64(time.Now().UnixNano()))

	gc.conn.Write(joinPacket)
	fmt.Println("Joined game server")
}

func (gc *GameClient) SendInput(inputX, inputY float32) {
	packet := make([]byte, 25)
	packet[0] = 3 // PacketInput
	binary.BigEndian.PutUint32(packet[1:5], gc.playerID)
	binary.BigEndian.PutUint64(packet[5:13], uint64(time.Now().UnixNano()))
	binary.BigEndian.PutUint32(packet[13:17], gc.sequence)
	binary.BigEndian.PutUint32(packet[17:21], math.Float32bits(inputX))
	binary.BigEndian.PutUint32(packet[21:25], math.Float32bits(inputY))

	gc.conn.Write(packet)
	gc.sequence++
}

func (gc *GameClient) ReceiveUpdates() {
	buffer := make([]byte, 1024)

	for {
		n, err := gc.conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading from server: %v", err)
			continue
		}

		if n >= 14 { // minimum update packet size
			packetType := buffer[0]
			if packetType == 4 { // PacketUpdate
				gc.handleGameUpdate(buffer[:n])
			}
		}
	}
}

func (gc *GameClient) handleGameUpdate(data []byte) {
	playerCount := data[13]
	fmt.Printf("Game update - Players: %d\n", playerCount)

	// Parse player states
	offset := 14
	for i := 0; i < int(playerCount); i++ {
		if offset+24 <= len(data) {
			playerID := binary.BigEndian.Uint32(data[offset : offset+4])
			x := math.Float32frombits(binary.BigEndian.Uint32(data[offset+4 : offset+8]))
			y := math.Float32frombits(binary.BigEndian.Uint32(data[offset+8 : offset+12]))

			fmt.Printf("  Player %d: (%.1f, %.1f)\n", playerID, x, y)
			offset += 24
		}
	}
}

func main() {
	client := NewGameClient()
	defer client.conn.Close()

	// Join game
	client.Join()

	// Start receiving updates
	go client.ReceiveUpdates()

	// Simulate player input
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	counter := 0
	for range ticker.C {
		// Simulate circular movement
		angle := float32(counter) * 0.1
		inputX := float32(math.Cos(float64(angle))) * 0.5
		inputY := float32(math.Sin(float64(angle))) * 0.5

		client.SendInput(inputX, inputY)
		counter++

		if counter >= 100 {
			break
		}
	}
}
