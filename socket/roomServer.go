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

	// 扑克逻辑处理
	pokerLogic *game.PokerLogic

	// 玩家列表
	players []RoomPlayItem

	// 三位玩家手中的牌
	playerCards [3][20]int

	// 底牌
	dzCards [3]int

	// 当前抢地主/出牌玩家
	curPlayerIndex int

	// 当前最大牌情况，牌型，牌型中的头牌，具体牌
	curCard game.CardSet

	// 地主座位
	dizhu int

	// 抢地主到几分了
	dizhuScore int

	// 当前已经抢地主次数
	wantDizhuTimes int

	// 当前由于炸弹而翻倍的次数
	bombTimes int

	// 当前牌最大的位置
	nowBigger int

	// 当前有几个人过牌
	passNum int

	// 得分情况
	scores map[int]int
}

func NewRoomServer() *RoomServer {
	// TODO: 做一些实际初始化的事
	return &RoomServer{
		curPlayerIndex: 1,
		pokerLogic: &game.PokerLogic{},
		players: make([]RoomPlayItem, 3),
		curCard: game.CardSet{
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

// 处理出牌消息
func (roomServer *RoomServer) handlePlayCard(message *Message) {
	seq := message.Seq

	playCardCommand := game.PlayCardInCommand{}

	// Todo: 发送错误返回？
	err := json.Unmarshal(message.Content, &playCardCommand)
	if err != nil {
		return
	}

	index := playCardCommand.Index

	curCard := playCardCommand.CurCard;
	if (len(curCard.Cards) > 0) {
		// 判断是否符合出牌规则

		// 牌型检查
		cardType := roomServer.pokerLogic.CalcuPokerType(curCard.Cards)
		if cardType != curCard.Type { // 计算出的牌型与传过来的不匹配
			//log.Printf("计算出的类型为 %d,传过来的为 %d", cardType, curCard.Type)
			// 告知玩家出牌失败
			roomServer.sendAck(PLAYER_PLAYCARD, index, seq, -2)
			return
		}

		// 头牌检查
		cardHeader := roomServer.pokerLogic.CalcuPokerHeader(curCard.Cards, curCard.Type)
		if cardHeader != curCard.Header { // 计算出的头牌与传过来的不匹配
			//log.Printf("计算出的头牌为 %d,传过来的为 %d", cardHeader, curCard.Header)
			// 告知玩家出牌失败
			roomServer.sendAck(PLAYER_PLAYCARD, index, seq, -2)
			return
		}

		// 是否可出牌检查
		if ! roomServer.pokerLogic.CanOut(&curCard, &roomServer.curCard) {
			//log.Printf("当前牌型为 %d, 头牌为 %d", roomServer.curCard.Type, roomServer.curCard.Header)
			//log.Printf("新来的牌型为 %d, 头牌为 %d", curCard.Type, curCard.Header)
			roomServer.sendAck(PLAYER_PLAYCARD, index, seq, -3)
			return
		}

		// 移除手中的牌
		if ! roomServer.removeCards(index - 1, curCard.Cards) {
			// 告知玩家出牌失败
			roomServer.sendAck(PLAYER_PLAYCARD, index, seq, -1)
			return
		}
	}

	// 告知玩家出牌成功
	roomServer.sendAck(PLAYER_PLAYCARD, index, seq, 0)

	if curCard.Type != game.PASS_CARDS { // 如果不是过牌，处理新的最大牌
		roomServer.curCard = curCard
		roomServer.passNum = 0

		// 通知本轮出牌，以及下一个应出牌的玩家
		roomServer.addCurIndex()
		roomServer.sendNextCardOut()
	} else {
		roomServer.passNum++
		if (1 == roomServer.passNum) {
			roomServer.nowBigger = roomServer.curPlayerIndex - 1;
			if (roomServer.nowBigger == 0) {
				roomServer.nowBigger = 3
			}
			roomServer.addCurIndex()
			roomServer.sendPassMsg()
		} else { // 不是1就是2
			roomServer.passNum = 0
			roomServer.curPlayerIndex = roomServer.nowBigger
			roomServer.curCard = game.CardSet{
				Type: game.NO_CARDS,
				Header: 0,
				Cards: nil,
			}
			roomServer.nowBigger = 0
			roomServer.sendNextCardOut()
		}
	}
}

func (roomServer *RoomServer) sendAck(command MessageCode, index int, seq int, code int) {
	ackMsg := Message{}
	ackMsg.Command = command
	ackMsg.Seq = seq
	ackMsg.Code = code
	roomServer.sendToOnePlayer(index, ackMsg)
	return
}

func (roomServer *RoomServer) sendPassMsg() {
	passCurCard := game.CardSet{
		Type: game.PASS_CARDS,
		Header: 0,
		Cards: nil,
	}

	cardOutCommand := game.CardOutCommand{
		State: 1,
		CurPlayerIndex: roomServer.curPlayerIndex,
		CurCard: passCurCard,
	}

	msg := Message{}
	msg.Command = PLAY_GAME
	msg.Content, _ = json.Marshal(cardOutCommand)
	roomServer.sendToRoomPlayers(msg)
}

func (roomServer *RoomServer) sendNextCardOut() {
	cardOutCommand := game.CardOutCommand{
		State: 1,
		CurPlayerIndex: roomServer.curPlayerIndex,
		CurCard: roomServer.curCard,
	}

	msg := Message{}
	msg.Command = PLAY_GAME
	msg.Content, _ = json.Marshal(cardOutCommand)
	roomServer.sendToRoomPlayers(msg)
}

// 移除牌
func (roomServer *RoomServer) removeCards(index int, cards []int) bool {
	cardGroup := &roomServer.playerCards[index]

	var haveHard bool
	for i, length := 0, len(cards); i < length; i++ {
		haveHard = false

		for j := 0; j < 20; j++ {
			if (cards[i] == cardGroup[j]) { // 清除对应的牌
				haveHard = true
				cardGroup[j] = 0
				break
			}
		}

		if ( ! haveHard) {
			return false
		}
	}

	// 检查是否出完牌了
	hasOut := true
	for j := 0; j < 20; j++ {
		if (0 != cardGroup[j]) {
			hasOut = false
			break
		}
	}

	if (hasOut) {
		roomServer.countScore(index + 1)
		roomServer.changeState(2)
		roomServer.exit()
	}

	return true
}

// 算分
func (roomServer *RoomServer) countScore(winIndex int) {
	// 喊地主分
	score := roomServer.dizhuScore

	// 翻倍
	for i := 0; i < roomServer.bombTimes; i++ {
		score = score * 2
	}

	other1 := winIndex - 1
	if (other1 == 0) {
		other1 = 3
	}

	other2 := winIndex + 1
	if (other2 == 4) {
		other2 = 1
	}

	dizhu := roomServer.dizhu
	if (winIndex == dizhu) { // 胜者是地主
		roomServer.scores[other1] = -score
		roomServer.scores[other2] = -score
		roomServer.scores[dizhu] = 2 * score
	} else {
		roomServer.scores[other1] = score
		roomServer.scores[other2] = score
		roomServer.scores[winIndex] = score
		roomServer.scores[dizhu] = -2 * score
	}
}

// 处理抢地主消息
func (roomServer *RoomServer) handleWantDizhu(message *Message) {
	seq := message.Seq

	wantDizhuCommand := game.WantDizhuInputCommand{}

	// Todo: 发送错误返回？
	err := json.Unmarshal(message.Content, &wantDizhuCommand)
	if err != nil {
		return
	}

	score := wantDizhuCommand.Score
	index := wantDizhuCommand.Index
	roomServer.wantDizhuTimes++;

	// 发送ack
	roomServer.sendAck(PLAYER_WANTDIZHU, index, seq, 0)

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

// 退出房间
func (roomServer *RoomServer) exit() {
	for i := 0; i < len(roomServer.players); i++ {
		roomServer.players[i].ws.roomId = 0
	}

	msg := Message{}
	msg.Command = ROOM_EXIT
	roomServer.sendToRoomPlayers(msg)
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
