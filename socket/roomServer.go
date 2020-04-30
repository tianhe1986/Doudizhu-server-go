package socket

import (
	"Doudizhu-server-go/game"
	"encoding/json"
	"math/rand"
	"time"
)

type RoomPlayItem struct {
	ws   *Client
	name string
}

type RoomServer struct {
	// 房间id
	roomId int

	// 玩家列表
	players []RoomPlayItem

	// 三位玩家手中的牌
	playerCards [3][20]int

	// 底牌
	dzCards [3]int

	// 当前抢地主/出牌玩家
	nowPlayer int

	// 当前最大牌情况，座位，牌型，牌型中的头牌，具体牌
	nowMaxPlayer int
	nowMaxType   int
	nowMaxHeader int
	nowMaxCards  [20]int

	// 地主座位
	dizhu int

	// 抢地主到几分了
	dizhuScore int

	// 当前已经抢地主次数
	wantDizhuTimes int

	// 当前由于炸弹而翻倍的次数
	bombTimes int
}

func NewRoomServer() *RoomServer {
	// TODO: 做一些实际初始化的事
	return &RoomServer{
		players: make([]RoomPlayItem, 3),
	}
}

func (roomServer *RoomServer) InitGame() {
	// 发牌
	cards := roomServer.getNewCards54()
	for i := 0; i < 17; i++ {
		roomServer.playerCards[0][i] = cards[i]
		roomServer.playerCards[1][i] = cards[i + 17]
		roomServer.playerCards[2][i] = cards[i + 34]
	}

	for i := 0; i < 3; i++ {
		roomServer.dzCards[i] = cards[i + 51]
	}

	// 将牌消息发送给客户端
	for i := 0; i < 3; i++ {
		giveCardCommand := game.GiveCardCommand{}
		giveCardCommand.State = 0;
		giveCardCommand.RoomdId = roomServer.roomId
		giveCardCommand.Cards = roomServer.playerCards[i][0:17]

		msg := Message{}
		msg.Code = 0
		msg.Command = PLAY_GAME
		msg.Seq = 0
		msg.Content, _ = json.Marshal(giveCardCommand)

		tempMsg, _ := json.Marshal(msg)
		roomServer.players[i].ws.send  <- tempMsg
	}
}

func (roomServer *RoomServer) sendToOnepLayer(index int, data Message) {
	jsonData, _ := json.Marshal(data)
	roomServer.players[index - 1].ws.send <- jsonData
}

func (roomServer *RoomServer) getNewCards54() []int {
	cards := make([]int, 54)
	for i := 0; i < 54; i++ {
		cards[i] = i + 1
	}

	rand.Seed(time.Now().UnixNano())

	// 洗牌算法
	for i := 53; i >= 0; i-- {
		j := rand.Intn(i + 1)
		if i != j {
			temp := cards[i]
			cards[i] = cards[j]
			cards[j] = temp
		}
	}

	return cards
}
