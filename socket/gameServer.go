package socket

import (
	"container/list"
)

type GameServer struct {
	// 匹配队列
	queue *list.List

	// 房间map
	rooms map[int]RoomServer

	// 当前最小可用房间编号
	roomIndex int
}

func NewGameServer() *GameServer {
	return &GameServer{
		queue:     list.New(),
		rooms:     make(map[int]RoomServer),
		roomIndex: 1,
	}
}

// 处理实际的信息
func (gameServer *GameServer) handleMsg(client *Client, message *Message) {
	switch message.Command {
	case MATCH_PLAYER: // 匹配
		gameServer.matchPlayer(client, "先占个位", message.Seq)
		break
	case PLAYER_PLAYCARD: // 一局游戏内消息，每一项单独处理
		gameServer.playGame(client, message)
		break
	case PLAYER_WANTDIZHU: // 抢地主消息
		gameServer.wantDizhu(client, message)
		break
	}
}

// 进入匹配队列，尝试匹配
func (gameServer *GameServer) matchPlayer(client *Client, name string, seq int) {

}

// 单局游戏内消息
func (gameServer *GameServer) playGame(client *Client, message *Message) {

}

// 抢地主消息
func (gameServer *GameServer) wantDizhu(client *Client, message *Message) {

}
