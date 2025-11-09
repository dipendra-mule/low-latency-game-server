// protocol defines the protocol for the game
// the protocol is a simple protocol that is used to send input and update packets
// the packets are serialized using big endian
package protocol

import (
	"encoding/binary"
	"math"
)

const (
	MaxPacketSize = 1024
	ServerPort    = 8080
)

// packet types
const (
	PacketJoin   = 1
	PacketLeave  = 2
	PacketInput  = 3
	PacketUpdate = 4
)

type PacketHeader struct {
	Type      uint8
	PlayerID  uint32
	Timestamp uint64
}

type JoinPacket struct {
	Header PacketHeader
}

type InputPacket struct {
	Header   PacketHeader
	Sequence uint32
	InputX   float32
	InputY   float32
}

type PlayerState struct {
	PlayerID  uint32
	X         float32
	Y         float32
	Timestamp uint64
}

type GameUpdatePacket struct {
	Header       PacketHeader
	PlayerCount  uint8
	PlayerStates []PlayerState
}

func SerializeUpdatePacket(packet *GameUpdatePacket) []byte {
	buf := make([]byte, MaxPacketSize)
	offset := 0
	buf[offset] = packet.Header.Type
	offset += 1

	binary.BigEndian.PutUint32(buf[offset:], packet.Header.PlayerID)
	offset += 4

	binary.BigEndian.PutUint64(buf[offset:], packet.Header.Timestamp)
	offset += 8

	buf[offset] = packet.PlayerCount
	offset += 1

	// TODO: serialize player states
	for i := 0; i < len(packet.PlayerStates); i++ {
		binary.BigEndian.PutUint32(buf[offset:], packet.PlayerStates[i].PlayerID)
		offset += 4
		binary.BigEndian.PutUint32(buf[offset:], math.Float32bits(packet.PlayerStates[i].X))
		offset += 4
		binary.BigEndian.PutUint32(buf[offset:], math.Float32bits(packet.PlayerStates[i].Y))
		offset += 4
		binary.BigEndian.PutUint64(buf[offset:], packet.PlayerStates[i].Timestamp)
		offset += 4
	}
	return buf[:offset]
}

func DeserializeUpdatePacket(buf []byte) *InputPacket {
	if len(buf) < 21 { // header(13) + sequence(4) + input(4) + inputY(4)
		return nil
	}

	return &InputPacket{
		Header: PacketHeader{
			Type:      buf[0],
			PlayerID:  binary.BigEndian.Uint32(buf[1:5]),
			Timestamp: binary.BigEndian.Uint64(buf[5:13]),
		},
		Sequence: binary.BigEndian.Uint32(buf[13:17]),
		InputX:   math.Float32frombits(binary.BigEndian.Uint32(buf[17:21])),
		InputY:   math.Float32frombits(binary.BigEndian.Uint32(buf[21:25])),
	}
}
