package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
)

type websocketHub struct {
	clients    map[*websocketClient]bool
	broadcast  chan []byte
	register   chan *websocketClient
	unregister chan *websocketClient
}

func newHub() *websocketHub {
	return &websocketHub{
		broadcast:  make(chan []byte),
		register:   make(chan *websocketClient),
		unregister: make(chan *websocketClient),
		clients:    make(map[*websocketClient]bool),
	}
}

func (hub *websocketHub) sendResult(result *monitorResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Unable to marshal result: %v", err)
	}

	hub.broadcast <- data

	return nil
}

func (hub *websocketHub) run() {
	for {
		select {
		case client := <-hub.register:
			log.Printf("[WebSocket] Adding client")
			hub.clients[client] = true

		case client := <-hub.unregister:
			log.Printf("[WebSocket] Removing client")
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				close(client.send)
			}

		case message := <-hub.broadcast:
			log.Printf("[WebSocket] Broadcasting message")
			for client := range hub.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(hub.clients, client)
				}
			}
		}
	}
}

type websocketClient struct {
	hub  *websocketHub
	conn *websocket.Conn
	send chan []byte
}

func (c *websocketClient) close() {
	c.hub.unregister <- c
	c.conn.Close()
}

func (c *websocketClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
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
