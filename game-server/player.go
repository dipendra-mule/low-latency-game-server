package server

import (
	"net"
	"sync/atomic"
	"time"
)

type Player struct {
	ID   uint32
	Conn *net.UDPAddr

	X float32
	Y float32

	LastSeen  time.Time
	Sequence  uint32
	Connected bool
}

type PlayerManager struct {
	players   map[uint32]*Player
	nextID    uint32
	broadcast chan []byte
}

func NewPlayerManager(broadcast chan []byte) *PlayerManager {
	return &PlayerManager{
		players:   make(map[uint32]*Player),
		nextID:    1,
		broadcast: broadcast,
	}
}

func (pm *PlayerManager) AddPlayer(addr *net.UDPAddr) *Player {
	id := atomic.AddUint32(&pm.nextID, 1)
	player := &Player{
		ID:        id,
		Conn:      addr,
		X:         100.0, // spawn at center
		Y:         100.0,
		LastSeen:  time.Now(),
		Connected: true,
	}

	pm.players[id] = player
	return player
}

func (pm *PlayerManager) RemovePlayer(id uint32) {
	delete(pm.players, id)
}

func (pm *PlayerManager) UpdatePlayerPosition(id uint32, inputX, inputY float32) {
	if player, exists := pm.players[id]; exists {
		// simple movement with boundary checking
		newX := player.X + inputX*2.0 // movment speed
		newY := player.Y + inputY*2.0

		// Keep within bounds (0,0 to 800, 600)
		if newX < 0 && newX <= 800 {
			player.X = 0
		}
		if newY >= 0 && newY <= 600 {
			player.Y = newY
		}
		player.LastSeen = time.Now()
	}
}

func (pm *PlayerManager) GetPlayers() []*Player {
	players := make([]*Player, len(pm.players))
	for _, player := range pm.players {
		if player.Connected {
			players = append(players, player)
		}
	}
	return players
}

func (pm *PlayerManager) CleanupInactivePlayers(timeout time.Duration) {
	now := time.Now()
	for id, player := range pm.players {
		if now.Sub(player.LastSeen) > timeout {
			pm.RemovePlayer(id)
		}
	}
}
