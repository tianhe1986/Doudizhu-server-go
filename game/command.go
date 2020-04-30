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
