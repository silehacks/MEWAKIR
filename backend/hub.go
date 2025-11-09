package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Hub coordinates WebSocket clients in meetings.
type Hub struct {
	store      *MeetingStore
	register   chan *Client
	unregister chan *Client
	broadcast  chan *signalMessage
	rooms      map[string]map[*Client]bool
}

// Client represents a meeting participant via WebSocket.
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	meetingID string
	name      string
}

type signalMessage struct {
	meetingID string
	sender    *Client
	payload   []byte
}

// NewHub creates a new Hub instance.
func NewHub(store *MeetingStore) *Hub {
	return &Hub{
		store:      store,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *signalMessage),
		rooms:      make(map[string]map[*Client]bool),
	}
}

// Run starts the hub event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			room := h.rooms[client.meetingID]
			if room == nil {
				room = make(map[*Client]bool)
				h.rooms[client.meetingID] = room
			}
			room[client] = true
			log.Printf("client %s joined meeting %s", client.name, client.meetingID)
		case client := <-h.unregister:
			if room, ok := h.rooms[client.meetingID]; ok {
				if _, exists := room[client]; exists {
					delete(room, client)
					close(client.send)
					log.Printf("client %s left meeting %s", client.name, client.meetingID)
					if len(room) == 0 {
						delete(h.rooms, client.meetingID)
					}
				}
			}
		case message := <-h.broadcast:
			if room, ok := h.rooms[message.meetingID]; ok {
				for client := range room {
					if client == message.sender {
						continue
					}
					select {
					case client.send <- message.payload:
					default:
						close(client.send)
						delete(room, client)
					}
				}
			}
		}
	}
}

// ServeWS upgrades the HTTP connection to WebSocket and registers the client.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	meetingID := strings.TrimSpace(r.URL.Query().Get("meetingId"))
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if meetingID == "" || name == "" {
		http.Error(w, "meetingId and name are required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:       h,
		conn:      conn,
		send:      make(chan []byte, 256),
		meetingID: meetingID,
		name:      name,
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(5120)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read error: %v", err)
			}
			break
		}
		c.hub.broadcast <- &signalMessage{
			meetingID: c.meetingID,
			sender:    c,
			payload:   message,
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(40 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
