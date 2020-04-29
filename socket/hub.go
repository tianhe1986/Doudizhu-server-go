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
			switch data.message.Code {
			case SYSTEM_MSG: // TODO: 系统类消息
			case REGISTER:
			case LOGIN:
				break
			case MATCH_PLAYER:
			case PLAY_GAME:
			case PLAYER_PLAYCARD:
			case PLAYER_WANTDIZHU:
				break
			default:
				break
			}
		}
	}
}
