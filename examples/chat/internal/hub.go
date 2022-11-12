package internal

// Hub 维护活跃客户端集合及广播给客户端的消息。
type Hub struct {
	// 已注册客户端。
	clients map[*Client]bool

	// 来自客户端的入站消息。
	broadcast chan []byte

	// 注册来自客户端的请求。
	register chan *Client

	// 注销来自客户端的请求。
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
