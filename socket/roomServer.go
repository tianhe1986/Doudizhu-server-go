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
	curPlayerIndex int

	// 当前最大牌情况，座位，牌型，牌型中的头牌，具体牌
	curCard game.CurrentCard

	// 地主座位
	dizhu int

	// 抢地主到几分了
	dizhuScore int

	// 当前已经抢地主次数
	wantDizhuTimes int

	// 当前由于炸弹而翻倍的次数
	bombTimes int

	// 得分情况
	scores map[int]int
}

func NewRoomServer() *RoomServer {
	// TODO: 做一些实际初始化的事
	return &RoomServer{
		curPlayerIndex: 1,
		players: make([]RoomPlayItem, 3),
		curCard: game.CurrentCard{
			Type: game.NO_CARDS,
			Header: 0,
			Cards: nil,
		},
		scores: make(map[int]int),
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

	roomServer.changeState(3)
}

// 状态变更  2是结算，1是游戏中, 3是抢地主
func (roomServer *RoomServer) changeState(state int) {
	stateChangeCommand := game.StateChangeCommand{
		State: state,
	}
	msg := Message{}
	switch (state) {
	case 1:
		stateChangeCommand.CurPlayerIndex = roomServer.curPlayerIndex
		stateChangeCommand.CurCard = roomServer.curCard
		msg.Command = PLAY_GAME
		break;
	case 2:
		stateChangeCommand.Scores = roomServer.scores
		msg.Command = PLAY_GAME
		break;
	case 3:
		roomServer.curPlayerIndex = rand.Intn(3) + 1
		stateChangeCommand.CurPlayerIndex = roomServer.curPlayerIndex
		stateChangeCommand.NowScore = 0
		msg.Command = PLAYER_WANTDIZHU
		break;
	default:
		return;
	}

	msg.Content, _ = json.Marshal(stateChangeCommand)
	roomServer.sendToRoomPlayers(msg)
}

// 处理抢地主消息
func (roomServer *RoomServer) handleWantDizhu(message *Message) {
	seq := message.Seq

	wantDizhuCommand := game.WantDizhuInputCommand{}

	// 这里解析肯定是成功的，不然传不进来
	json.Unmarshal(message.Content, &wantDizhuCommand)

	score := wantDizhuCommand.Score
	index := wantDizhuCommand.Index
	roomServer.wantDizhuTimes++;

	// 发送ack
	ackMsg := Message{}
	ackMsg.Command = PLAYER_WANTDIZHU
	ackMsg.Seq = seq
	ackMsg.Code = 0
	roomServer.sendToOnePlayer(index, ackMsg)

	if (score == 3) { // 直接喊3分，成为地主
		roomServer.dizhuScore = score

		roomServer.giveDzCards(index - 1)
		roomServer.dizhu = index
		roomServer.curPlayerIndex = index
		// 告知地主结果
		roomServer.notifyDizhuResult()

		// 状态变更
		roomServer.changeState(1)
		return
	} else if (score > roomServer.dizhuScore) { // 叫了个更高分，更新
		roomServer.dizhuScore = score
		roomServer.dizhu = index
	}

	// 如果是第三次，表示每个人都表过态了
	if (roomServer.wantDizhuTimes == 3) {
		roomServer.giveDzCards(roomServer.dizhu - 1)
		roomServer.curPlayerIndex = roomServer.dizhu
		// 告知地主结果
		roomServer.notifyDizhuResult()

		// 状态变更
		roomServer.changeState(1)
	} else { // 继续问下一个人
		roomServer.addCurIndex()

		command := game.WantDizhuOutputCommand{
			State: 3,
			CurPlayerIndex: roomServer.curPlayerIndex,
			NowScore: roomServer.dizhuScore,
		}
		msg := Message{}

		msg.Command = PLAYER_WANTDIZHU
		msg.Content, _ = json.Marshal(command)
		roomServer.sendToRoomPlayers(msg)
	}
}

// 下一个座位
func (roomServer *RoomServer) addCurIndex() {
	roomServer.curPlayerIndex++;
	if (roomServer.curPlayerIndex > 3) {
		roomServer.curPlayerIndex = 1; //每次到4就变回1
	}
}

// 将底牌给地主
func (roomServer *RoomServer) giveDzCards(index int) {
	for i := 0; i < 3; i++ {
		roomServer.playerCards[index][17 + i] = roomServer.dzCards[i]
	}
}

// 通知抢地主结果
func (roomServer *RoomServer) notifyDizhuResult() {
	dizhuResultCommand := game.DizhuResultCommand{
		State: 3,
		Dizhu: roomServer.curPlayerIndex,
		DizhuCards: roomServer.dzCards[0:],
		NowScore: roomServer.dizhuScore,
	}

	msg := Message{}
	msg.Command = PLAYER_WANTDIZHU
	msg.Content, _ = json.Marshal(dizhuResultCommand)
	roomServer.sendToRoomPlayers(msg)
}

// 给房间内单个用户发送消息
func (roomServer *RoomServer) sendToOnePlayer(index int, data Message) {
	jsonData, _ := json.Marshal(data)
	roomServer.players[index - 1].ws.send <- jsonData
}

// 给房间内所有用户发送消息
func (roomServer *RoomServer) sendToRoomPlayers(data Message) {
	jsonData, _ := json.Marshal(data)
	for i := 0; i < len(roomServer.players); i++ {
		roomServer.players[i].ws.send <- jsonData
	}
}

// 拿到一副新好的牌
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
