package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lpxxn/snake/game"
)

const (
	// 写入超时时间
	writeWait = 10 * time.Second

	// 读取超时时间
	pongWait = 60 * time.Second

	// 发送ping的间隔时间
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有CORS请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 客户端命令结构
type Command struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// 方向命令
type DirectionCommand struct {
	Direction game.Direction `json:"direction"`
}

// 处理WebSocket连接
func HandleWebSocket(w http.ResponseWriter, r *http.Request, g *game.Game) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket升级失败:", err)
		return
	}

	// 创建新玩家
	playerID := r.URL.Query().Get("id")
	if playerID == "" {
		playerID = generateID()
	}
	log.Printf("新WebSocket连接: 玩家ID=%s", playerID)

	player := game.NewPlayer(playerID)

	// 注册玩家
	g.Register <- player

	// 启动goroutine处理连接
	go writePump(conn, player, g)
	go readPump(conn, player, g)
}

// 从客户端读取消息
func readPump(conn *websocket.Conn, player *game.Player, g *game.Game) {
	defer func() {
		g.Unregister <- player
		conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// 解析命令
		var cmd Command
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Println("Error parsing command:", err)
			continue
		}

		// 处理命令
		switch cmd.Type {
		case "direction":
			var dirCmd DirectionCommand
			if err := json.Unmarshal(cmd.Payload, &dirCmd); err != nil {
				log.Println("Error parsing direction command:", err)
				continue
			}

			// 更改蛇的方向
			if player.Snake != nil {
				player.Snake.ChangeDirection(dirCmd.Direction)
			}
		}
	}
}

// 向客户端发送消息
func writePump(conn *websocket.Conn, player *game.Player, g *game.Game) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-player.Send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// 通道已关闭
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case message := <-g.Broadcast:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 生成唯一ID
func generateID() string {
	return time.Now().Format("20060102150405") +
		fmt.Sprintf("%d", rand.Intn(1000))
}
