package server

import (
	"time"
)

type GameState struct {
	playerManager *PlayerManager
	tickRate      time.Duration
	running       bool
}

func NewGameState(playerManager *PlayerManager) *GameState {
	return &GameState{
		playerManager: playerManager,
		tickRate:      time.Second / 60, // 60 Hz
		running:       true,
	}
}

func (gs *GameState) Start() {
	ticker := time.NewTicker(gs.tickRate)
	defer ticker.Stop()

	for gs.running {
		select {
		case <-ticker.C:
			gs.Update()
		}
	}
}

func (gs *GameState) Update() {
	// cleanup inactive players
	gs.playerManager.CleanupInactivePlayers(time.Second * 5)

	// game state updates would go here
	// collision detection, etc
}

func (gs *GameState) Stop() {
	gs.running = false
}
