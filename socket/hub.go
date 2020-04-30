package socket

// 游戏中消息的结构体
type MessageChanItem struct {
	// 对应的客户端
	client  *Client
	message Message
}

// 从websocket的demo中拿来的Hub类，用来做中控
type Hub struct {
	// 客户端map
	clients map[*Client]bool

	// 数据消息队列，带缓冲区
	dataList chan MessageChanItem

	// 新增连接队列
	register chan *Client

	// 移除连接队列
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		dataList:   make(chan MessageChanItem, maxDataListSize),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	gameServer := NewGameServer()
	for {
		select {
		case client := <-h.register: // 新增连接
			h.clients[client] = true
		case client := <-h.unregister: // 关闭连接
			if _, ok := h.clients[client]; ok { // TODO: 从排队队列中移除
				delete(h.clients, client)
				close(client.send)
			}
		case data := <-h.dataList: // 实际的数据处理
			switch data.message.Command {
			// TODO: 系统类消息
			case SYSTEM_MSG, REGISTER, LOGIN:
				break
			// 游戏内消息
			case MATCH_PLAYER, PLAY_GAME, PLAYER_PLAYCARD, PLAYER_WANTDIZHU:
				gameServer.handleMsg(data.client, &data.message)
				break
			default:
				break
			}
		}
	}
}
