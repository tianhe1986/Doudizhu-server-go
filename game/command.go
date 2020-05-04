package game

// 发起匹配content
type MatchCommand struct {
	Name int `json:"name"`
}

// 匹配结果content
type MatchResultCommand struct {
	Players []string `json:"players"`
	RoomId int `json:"roomId"`
}

// 发牌content
type GiveCardCommand struct {
	State int `json:"state"`
	RoomdId int `json:"roomId"`
	Cards []int `json:"cards"`
}

type CurrentCard struct {
	Type CardType `json:"type"`
	Header int `json:"header"`
	Cards []int `json:"cards"`
}

// 状态变化command
type StateChangeCommand struct {
	State int `json:"state"`
	CurPlayerIndex int `json:"curPlayerIndex"`
	CurCard CurrentCard `json:"curCard"`
	Scores map[int]int `json:"scores"`
	NowScore int `json:"nowScore"`
}

// 房间内基本信息
type CommonRoomCommand struct {
	RoomId int `json:"roomId"`
	Index int `json:"index"`
}

// 接收到的抢地主消息
type WantDizhuInputCommand struct {
	RoomId int `json:"roomId"`
	Index int `json:"index"`
	Score int `json:"score"`
}

type WantDizhuOutputCommand struct {
	State int `json:"state"`
	CurPlayerIndex int `json:"curPlayerIndex"`
	NowScore int `json:"nowScore"`
}

// 抢地主结果消息
type DizhuResultCommand struct {
	State int `json:"state"`
	Dizhu int `json:"dizhu"`
	DizhuCards []int `json:"dizhuCards"`
	NowScore int `json:"nowScore"`
}