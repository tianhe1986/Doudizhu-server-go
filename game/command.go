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

type CurrentCards struct {
	Type CardType `json:"type"`
	Header int `json:"header"`
	Cards []int `json:"cards"`
}

// 状态变化command
type StateChangeCommand struct {
	State int `json:"state"`
	CurPlayerIndex int `json:"curPlayerIndex"`
	CurCards CurrentCards `json:"curCards"`
	Scores map[int]int `json:"scores"`
	NowScore int `json:"nowScore"`
}
