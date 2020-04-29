package socket

import (
	"encoding/json"
)

type MessageCode int

const (
	SYSTEM_MSG       MessageCode = 1
	REGISTER         MessageCode = 2
	LOGIN            MessageCode = 3
	MATCH_PLAYER     MessageCode = 4
	PLAY_GAME        MessageCode = 5
	ROOM_NOTIFY      MessageCode = 6
	PLAYER_PLAYCARD  MessageCode = 7
	PLAYER_WANTDIZHU MessageCode = 8
	WS_CLOSE         MessageCode = 9
)

type Message struct {
	Seq     int             `json:"seq"`
	Code    MessageCode     `json:"code"`
	Command int             `json:"command"`
	Content json.RawMessage `json:"content"`
}
