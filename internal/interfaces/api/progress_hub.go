package api

import (
	"encoding/json"
	"sync"
)

type ProgressMessage struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Payload interface{} `json:"payload"`
}

type SearchProgressPayload struct {
	Scanned    int    `json:"scanned"`
	Found      int    `json:"found"`
	CurrentDir string `json:"current_dir"`
}

type OperationProgressPayload struct {
	Operation string `json:"operation"`
	Done      int    `json:"done"`
	Total     int    `json:"total"`
}

type Client struct {
	ID   string
	Send chan []byte
}

type ProgressHub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

var hub = &ProgressHub{
	clients: make(map[string]*Client),
}

func GetProgressHub() *ProgressHub {
	return hub
}

func (h *ProgressHub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
}

func (h *ProgressHub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client.ID]; ok {
		close(client.Send)
		delete(h.clients, client.ID)
	}
}

func (h *ProgressHub) SendToClient(clientID string, msg ProgressMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.clients[clientID]
	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case client.Send <- data:
	default:
	}
}

func (h *ProgressHub) BroadcastSearchProgress(clientID string, scanned int, found int, currentDir string) {
	h.SendToClient(clientID, ProgressMessage{
		Type: "search_progress",
		ID:   clientID,
		Payload: SearchProgressPayload{
			Scanned:    scanned,
			Found:      found,
			CurrentDir: currentDir,
		},
	})
}

func (h *ProgressHub) BroadcastOperationProgress(clientID string, operation string, done int, total int) {
	h.SendToClient(clientID, ProgressMessage{
		Type: "operation_progress",
		ID:   clientID,
		Payload: OperationProgressPayload{
			Operation: operation,
			Done:      done,
			Total:     total,
		},
	})
}
