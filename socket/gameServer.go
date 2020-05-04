package socket

import (
	"Doudizhu-server-go/game"
	"container/list"
	"encoding/json"
	"log"
	"strconv"
)

type MatchItem struct {
	ws *Client
	name string
	seq int
}

type GameServer struct {
	// 匹配队列
	queue *list.List

	// 房间map
	rooms map[int]*RoomServer

	// 当前最小可用房间编号
	roomIndex int
}

func NewGameServer() *GameServer {
	return &GameServer{
		queue:     list.New(),
		rooms:     make(map[int]*RoomServer),
		roomIndex: 1,
	}
}

// 处理实际的信息
func (gameServer *GameServer) handleMsg(client *Client, message *Message) {
	switch message.Command {
	case MATCH_PLAYER: // 匹配

		matchCommand := game.MatchCommand{}
		err := json.Unmarshal(message.Content, &matchCommand)

		//解析失败会报错。
		if err != nil {
			break
		}
		gameServer.matchPlayer(client, strconv.Itoa(matchCommand.Name), message.Seq)
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
	log.Printf("user %s try match", name)
	// TODO 将client与name进行绑定
	newMatchItem := MatchItem{
		ws: client,
		name: name,
		seq: seq,
	}
	
	gameServer.queue.PushBack(newMatchItem)

	// 每三个进行处理
	if gameServer.queue.Len() >= 3 {
		var players []MatchItem = make([]MatchItem, 3)
		var names []string = make([]string, 3)
		var roomPlayers []RoomPlayItem = make([]RoomPlayItem, 3)

		for i := 0; i < 3; i++ {
			tempItem := gameServer.queue.Front()
			gameServer.queue.Remove(tempItem)
			p, ok := tempItem.Value.(MatchItem);
			if ! ok {
				return
			}

			players[i] = p
			names[i] = p.name
			playItem := RoomPlayItem{
				ws: p.ws,
				name: p.name,
			}
			roomPlayers[i] = playItem
		}

		// 发送匹配成功消息
		resultContent := game.MatchResultCommand{}
		resultContent.Players = names
		resultContent.RoomId = gameServer.roomIndex
		for i := 0; i < 3; i++ {
			resp := Message{}
			resp.Code = 0
			resp.Command = MATCH_PLAYER
			resp.Seq = players[i].seq
			resp.Content, _ = json.Marshal(resultContent)

			tempMsg, _ := json.Marshal(resp)
			players[i].ws.send  <- tempMsg
		}

		room := NewRoomServer()
		room.roomId = gameServer.roomIndex;
		room.players = roomPlayers
		gameServer.roomIndex++;
		gameServer.rooms[room.roomId] = room
		room.InitGame()
	}
}

// 单局游戏内消息
func (gameServer *GameServer) playGame(client *Client, message *Message) {
	commonRoomCommand := game.CommonRoomCommand{}
	err := json.Unmarshal(message.Content, &commonRoomCommand)

	//解析失败会报错。
	if err != nil {
		return
	}

	room := gameServer.rooms[commonRoomCommand.RoomId]
	room.handlePlayCard(message)
}

// 抢地主消息
func (gameServer *GameServer) wantDizhu(client *Client, message *Message) {
	commonRoomCommand := game.CommonRoomCommand{}
	err := json.Unmarshal(message.Content, &commonRoomCommand)

	//解析失败会报错。
	if err != nil {
		return
	}

	room := gameServer.rooms[commonRoomCommand.RoomId]
	room.handleWantDizhu(message)
}
