package internal

import (
	"bytes"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	// 允许客户端发送的最大消息大小。
	maxMessageSize = 512

	// 允许客户端读取下一条响应消息的时间。
	pongWait = 60 * time.Second

	// 在此期间向客户端发送 ping。时长必须小于 pongWait。
	pingPeriod = (pongWait * 9) / 10

	// 允许向客户端写入的时长。
	writeWait = 10 * time.Second

	// 发送缓冲区的大小
	bufSize = 256
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client 是 websocket 连接和 hub 之间的中间人。
type Client struct {
	hub *Hub

	// websocket 连接。
	conn *websocket.Conn

	// 出站消息的缓冲通道。
	send chan []byte
}

// ServeWs 处理来自客户端的 websocket 请求。
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, bufSize),
	}
	client.hub.register <- client

	// 允许通过在新的 goroutine 中完成所有工作来收集调用者引用的内存。
	go client.writePump()
	go client.readPump()
}

// readPump 将来自 websocket 连接的消息泵送到集线器。
//
// 为每个连接启动一个运行 readPump 的协程。
// 该程序通过执行该协程的所有读取来确保一个连接至多有一个读取器。
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("错误：%v", err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- message
	}
}

// writePump 将消息从 hub 泵送到 websocket 连接。
//
// 为每个连接启动一个运行 writePump 的协程。
// 该程序通过执行该协程的所有写入来确保一个连接至多有一个写入器。
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// 集线器关闭了通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 将排队中的聊天消息添加到当前 websocket 消息中。
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
