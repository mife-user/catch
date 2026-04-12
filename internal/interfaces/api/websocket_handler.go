package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub *ProgressHub
}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{hub: GetProgressHub()}
}

func (h *WebSocketHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/ws", h.HandleWebSocket)
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		clientID = fmt.Sprintf("client_%d", time.Now().UnixNano())
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		ID:   clientID,
		Send: make(chan []byte, 256),
	}

	h.hub.Register(client)

	go h.writePump(conn, client)
	go h.readPump(conn, client)
}

func (h *WebSocketHandler) writePump(conn *websocket.Conn, client *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WebSocketHandler) readPump(conn *websocket.Conn, client *Client) {
	defer func() {
		h.hub.Unregister(client)
		conn.Close()
	}()

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
