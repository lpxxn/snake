package game

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// GameServer 管理多人游戏服务器
type GameServer struct {
	game      *MultiplayerGame
	clients    map[*websocket.Conn]*Player
	mutex      sync.RWMutex
	minPlayers int
}

// Player 表示一个玩家
type Player struct {
	ID      string
	Name    string
	Snake   *Snake
	Conn    *websocket.Conn
	IsAlive bool
	IsReady bool
}

// MultiplayerGame 表示多人游戏状态
type MultiplayerGame struct {
	*Game
	Players    map[string]*Player
	Foods      []Point
	MaxPlayers int
	Started    bool
	mutex      sync.RWMutex
}

// GameMessage 表示游戏消息
type GameMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// NewGameServer 创建新的游戏服务器
func NewGameServer(width, height int) *GameServer {
	return &GameServer{
		game: &MultiplayerGame{
			Game:       NewGame(width, height),
			Players:    make(map[string]*Player),
			Foods:      make([]Point, 0),
			MaxPlayers: 4,
			Started:    false,
		},
		clients:    make(map[*websocket.Conn]*Player),
		minPlayers: 2,
	}
}

// HandleNewPlayer 处理新玩家连接
func (s *GameServer) HandleNewPlayer(conn *websocket.Conn) {
	// 检查游戏状态和玩家数量
	s.mutex.Lock()
	if s.game.Started {
		s.mutex.Unlock()
		conn.WriteJSON(map[string]string{"type": "error", "message": "游戏已经开始"})
		conn.Close()
		return
	}

	if len(s.clients) >= s.game.MaxPlayers {
		s.mutex.Unlock()
		conn.WriteJSON(map[string]string{"type": "error", "message": "游戏人数已满"})
		conn.Close()
		return
	}

	// 创建新玩家
	player := &Player{
		ID:      generatePlayerID(),
		Name:    generatePlayerName(),
		Conn:    conn,
		IsAlive: true,
		Snake:   s.createNewSnake(),
	}

	// 添加玩家到游戏
	s.clients[conn] = player
	s.game.Players[player.ID] = player
	s.mutex.Unlock()

	// 通知所有玩家有新玩家加入并更新玩家列表
	s.broadcastGameState()
	s.checkAllPlayersReady()

	// 开始处理玩家消息
	go s.handlePlayerMessages(player)
}

// handlePlayerMessages 处理玩家的WebSocket消息
func (s *GameServer) handlePlayerMessages(player *Player) {
	fmt.Printf("Handling player messages for player %s\n", player.ID)
	for {
		var msg GameMessage
		err := player.Conn.ReadJSON(&msg)
		if err != nil {
			s.removePlayer(player)
			fmt.Printf("Error reading message: %v", err)
			break
		}
		fmt.Printf("%+v\n", fmt.Sprintf("Received message: %+v", msg))
		switch msg.Type {
		case "ready":
			fmt.Printf("Player %s(%s) is ready\n", player.Name, player.ID)
			s.mutex.Lock()
			player.IsReady = true
			s.mutex.Unlock()
			s.checkAllPlayersReady()
			s.broadcastGameState()
			fmt.Printf("Broadcast ready state completed for player %s\n", player.ID)
		case "direction":
			var dir Direction
			json.Unmarshal(msg.Payload, &dir)
			s.game.mutex.Lock()
			player.Snake.direction = dir
			s.game.mutex.Unlock()
		case "startGame":
			s.mutex.Lock()
			allReady := true
			for _, p := range s.clients {
				if !p.IsReady {
					allReady = false
					break
				}
			}
			if allReady && (len(s.clients) == 1 || len(s.clients) >= s.minPlayers) {
				s.mutex.Unlock()
				go s.startGame()
			} else {
				s.mutex.Unlock()
				player.Conn.WriteJSON(map[string]string{"type": "error", "message": "无法开始游戏：玩家未准备就绪或人数不足"})
			}
		}
	}
}

// removePlayer 移除玩家
func (s *GameServer) removePlayer(player *Player) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.clients, player.Conn)
	delete(s.game.Players, player.ID)
	player.Conn.Close()

	// 如果游戏已经开始且玩家数量不足，结束游戏
	if s.game.Started && len(s.clients) < s.minPlayers {
		s.endGame()
	}

	s.broadcastGameState()
}

// checkAllPlayersReady 检查是否所有玩家都准备就绪
func (s *GameServer) checkAllPlayersReady() {
	s.mutex.RLock()
	// 检查是否所有玩家都准备就绪
	allReady := true
	playerList := make([]map[string]interface{}, 0)
	for _, player := range s.clients {
		if !player.IsReady {
			allReady = false
		}
		playerList = append(playerList, map[string]interface{}{
			"id": player.ID,
			"name": player.Name,
			"isReady": player.IsReady,
		})
	}
	s.mutex.RUnlock()

	// 广播当前准备状态
	s.broadcastMessage(map[string]interface{}{
		"type": "readyState",
		"allReady": allReady,
		"playerCount": len(s.clients),
		"players": playerList,
	})
}

// startGame 开始游戏
func (s *GameServer) startGame() {
	s.game.mutex.Lock()
	s.game.Started = true
	s.game.mutex.Unlock()

	// 生成初始食物
	for i := 0; i < 3; i++ {
		s.game.generateFood()
	}

	// 广播游戏开始
	s.broadcastMessage(map[string]string{"type": "gameStart"})

	// 开始游戏循环
	go s.gameLoop()
}

// gameLoop 游戏主循环
func (s *GameServer) gameLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		s.game.mutex.Lock()

		if !s.game.Started {
			s.game.mutex.Unlock()
			break
		}

		// 移动所有活着的蛇
		for _, player := range s.game.Players {
			if player.IsAlive {
				s.moveSnake(player)
			}
		}

		// 检查游戏是否结束
		aliveCount := 0
		var winner *Player
		for _, player := range s.game.Players {
			if player.IsAlive {
				aliveCount++
				winner = player
			}
		}

		if aliveCount <= 1 {
			s.game.mutex.Unlock()
			s.endGame()
			if winner != nil {
				s.broadcastMessage(map[string]interface{}{
					"type":    "gameOver",
					"winner":  winner.ID,
				})
			}
			break
		}

		s.game.mutex.Unlock()
		s.broadcastGameState()
	}
}

// moveSnake 移动一条蛇并处理碰撞
func (s *GameServer) moveSnake(player *Player) {
	head := player.Snake.body[0]
	newHead := Point{X: head.X, Y: head.Y}

	// 根据方向移动
	switch player.Snake.direction {
	case Up:
		newHead.Y--
	case Down:
		newHead.Y++
	case Left:
		newHead.X--
	case Right:
		newHead.X++
	}

	// 检查是否撞墙
	if newHead.X < 0 || newHead.X >= s.game.Width ||
		newHead.Y < 0 || newHead.Y >= s.game.Height {
		player.IsAlive = false
		return
	}

	// 检查是否撞到其他蛇
	for _, otherPlayer := range s.game.Players {
		if !otherPlayer.IsAlive {
			continue
		}
		for _, p := range otherPlayer.Snake.body {
			if p.X == newHead.X && p.Y == newHead.Y {
				player.IsAlive = false
				return
			}
		}
	}

	// 移动蛇
	player.Snake.body = append([]Point{newHead}, player.Snake.body...)

	// 检查是否吃到食物
	ateFood := false
	for i, food := range s.game.Foods {
		if newHead.X == food.X && newHead.Y == food.Y {
			ateFood = true
			// 移除被吃掉的食物
			s.game.Foods = append(s.game.Foods[:i], s.game.Foods[i+1:]...)
			// 生成新食物
			s.game.generateFood()
			break
		}
	}

	if !ateFood {
		// 如果没有吃到食物，删除尾部
		player.Snake.body = player.Snake.body[:len(player.Snake.body)-1]
	}
}

// broadcastGameState 广播游戏状态给所有玩家
func (s *GameServer) broadcastGameState() {
	s.game.mutex.RLock()
	state := map[string]interface{}{
		"type":    "gameState",
		"players": make(map[string]interface{}),
		"foods":   s.game.Foods,
	}

	for id, player := range s.game.Players {
		state["players"].(map[string]interface{})[id] = map[string]interface{}{
			"snake":   player.Snake.body,
			"isAlive": player.IsAlive,
		}
	}
	s.game.mutex.RUnlock()

	s.broadcastMessage(state)
}

// broadcastMessage 广播消息给所有玩家
func (s *GameServer) broadcastMessage(msg interface{}) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for conn := range s.clients {
		err := conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Error broadcasting message: %v", err)
		}
	}
}

// endGame 结束游戏
func (s *GameServer) endGame() {
	s.game.mutex.Lock()
	s.game.Started = false
	s.game.mutex.Unlock()
}

// createNewSnake 为新玩家创建蛇
func (s *GameServer) createNewSnake() *Snake {
	// 随机生成起始位置，确保不靠近边界
	rand.Seed(time.Now().UnixNano())
	// 设置安全边距
	margin := 3
	startX := margin + rand.Intn(s.game.Width-2*margin)
	startY := margin + rand.Intn(s.game.Height-2*margin)

	// 检查是否与其他蛇重叠
	for {
		valid := true
		for _, player := range s.game.Players {
			if player.Snake == nil {
				continue
			}
			for _, p := range player.Snake.body {
				// 检查是否太靠近其他蛇
				if abs(p.X-startX) < 2 && abs(p.Y-startY) < 2 {
					valid = false
					break
				}
			}
			if !valid {
				break
			}
		}
		if valid {
			break
		}
		// 如果位置无效，重新生成
		startX = margin + rand.Intn(s.game.Width-2*margin)
		startY = margin + rand.Intn(s.game.Height-2*margin)
	}

	return &Snake{
		body: []Point{
			{X: startX, Y: startY},
		},
		direction: Right,
	}
}

// generatePlayerID 生成玩家ID
func generatePlayerID() string {
	return fmt.Sprintf("player_%d", time.Now().UnixNano())
}

// abs 计算绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// generatePlayerName 生成玩家名字
func generatePlayerName() string {
	adjectives := []string{"快乐的", "勇敢的", "聪明的", "可爱的", "机智的"}
	animals := []string{"蛇", "龙", "虎", "豹", "狮"}
	return adjectives[rand.Intn(len(adjectives))] + animals[rand.Intn(len(animals))]
}