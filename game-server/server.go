package server

import (
	"log"
	"net"
	"time"

	"github.com/dipendra-mule/low-latency-game-server/protocol"
)

type GameServer struct {
	conn          *net.UDPConn
	playerManager *PlayerManager
	gameState     *GameState
	broadcastChan chan []byte
}

func NewGameServer() *GameServer {
	broadcastChan := make(chan []byte, 100)
	playerManager := NewPlayerManager(broadcastChan)

	return &GameServer{
		playerManager: playerManager,
		gameState:     NewGameState(playerManager),
		broadcastChan: broadcastChan,
	}
}

func (gs *GameServer) Start(port int) error {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return err
	}

	gs.conn = conn

	log.Printf("Game server started on port %d", port)

	// Start game loop
	go gs.gameState.Start()

	// Start broadcast system
	go gs.broadcastLoop()

	// Start receiving packets
	gs.receiveLoop()

	return nil
}

func (gs *GameServer) receiveLoop() {
	buffer := make([]byte, protocol.MaxPacketSize)

	for {
		n, addr, err := gs.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		go gs.handlePacket(buffer[:n], addr)
	}
}

func (gs *GameServer) handlePacket(data []byte, addr *net.UDPAddr) {
	if len(data) < 1 {
		return
	}

	packetType := data[0]

	switch packetType {
	case protocol.PacketJoin:
		gs.handleJoin(addr)
	case protocol.PacketInput:
		gs.handleInput(data, addr)
	}
}

func (gs *GameServer) handleJoin(addr *net.UDPAddr) {
	player := gs.playerManager.AddPlayer(addr)
	log.Printf("Player %d joined from %s", player.ID, addr.String())

	// Send welcome packet or initial state
	gs.sendGameUpdate()
}

func (gs *GameServer) handleInput(data []byte, addr *net.UDPAddr) {
	inputPacket := protocol.DeserializeUpdatePacket(data)
	if inputPacket == nil {
		return
	}

	// Update player position
	gs.playerManager.UpdatePlayerPosition(inputPacket.Header.PlayerID,
		inputPacket.InputX, inputPacket.InputY)

	// Broadcast updated game state to all players
	gs.sendGameUpdate()
}

func (gs *GameServer) sendGameUpdate() {
	players := gs.playerManager.GetPlayers()

	playerStates := make([]protocol.PlayerState, len(players))
	for i, player := range players {
		playerStates[i] = protocol.PlayerState{
			PlayerID:  player.ID,
			X:         player.X,
			Y:         player.Y,
			Timestamp: uint64(time.Now().UnixNano()),
		}
	}

	updatePacket := &protocol.GameUpdatePacket{
		Header: protocol.PacketHeader{
			Type:      protocol.PacketUpdate,
			Timestamp: uint64(time.Now().UnixNano()),
		},
		PlayerCount:  uint8(len(players)),
		PlayerStates: playerStates,
	}

	packetData := protocol.SerializeUpdatePacket(updatePacket)
	gs.broadcastChan <- packetData
}

func (gs *GameServer) broadcastLoop() {
	for packetData := range gs.broadcastChan {
		players := gs.playerManager.GetPlayers()

		for _, player := range players {
			if _, err := gs.conn.WriteToUDP(packetData, player.Conn); err != nil {
				log.Printf("Error broadcasting to player %d: %v", player.ID, err)
			}
		}
	}
}

func (gs *GameServer) Stop() {
	gs.gameState.Stop()
	close(gs.broadcastChan)
	if gs.conn != nil {
		gs.conn.Close()
	}
}
