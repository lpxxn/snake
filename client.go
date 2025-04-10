package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Game *Game
	Send chan []byte
}

func NewClient(id string, conn *websocket.Conn, game *Game) *Client {
	return &Client{
		ID:   id,
		Conn: conn,
		Game: game,
		Send: make(chan []byte, 256),
	}
}

func (c *Client) Read() {
	defer func() {
		c.Game.RemovePlayer(c.ID)
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// TODO: 处理玩家输入指令
	}
}

func (c *Client) Write() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
