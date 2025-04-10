package game

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

const (
	// 游戏区域大小
	Width  = 50
	Height = 50

	// 游戏刷新率 (毫秒)
	TickRate = 100
)

// 方向常量
type Direction int

const (
	Up Direction = iota
	Right
	Down
	Left
)

// 游戏状态消息类型
type MessageType string

const (
	PlayerJoin  MessageType = "player_join"
	PlayerLeave MessageType = "player_leave"
	GameState   MessageType = "game_state"
	PlayerList  MessageType = "player_list"
	PlayerDead  MessageType = "player_dead"
)

// 消息结构
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// 游戏主结构
type Game struct {
	Players    map[string]*Player
	Foods      []*Food
	mutex      sync.Mutex
	Broadcast  chan []byte
	Register   chan *Player
	Unregister chan *Player
}

// 创建新游戏
func NewGame() *Game {
	return &Game{
		Players:    make(map[string]*Player),
		Foods:      make([]*Food, 10), // 初始10个食物
		Broadcast:  make(chan []byte),
		Register:   make(chan *Player),
		Unregister: make(chan *Player),
	}
}

// 启动游戏循环
func (g *Game) Start() {
	// 初始化食物
	g.initFoods()
	log.Println("游戏已启动，初始化了", len(g.Foods), "个食物")

	ticker := time.NewTicker(TickRate * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case player := <-g.Register:
			log.Printf("玩家 %s (%s) 已注册", player.Name, player.ID)
			g.registerPlayer(player)
		case player := <-g.Unregister:
			log.Printf("玩家 %s (%s) 已注销", player.Name, player.ID)
			g.unregisterPlayer(player)
		case <-ticker.C:
			g.update()
		}
	}
}

// 初始化食物
func (g *Game) initFoods() {
	for i := 0; i < len(g.Foods); i++ {
		g.Foods[i] = NewRandomFood()
	}
}

// 注册新玩家
func (g *Game) registerPlayer(player *Player) {
	g.mutex.Lock()
	g.Players[player.ID] = player
	g.mutex.Unlock()

	// 广播玩家加入消息
	g.broadcastPlayerJoin(player)

	// 发送当前玩家列表给所有人
	g.broadcastPlayerList()

	// 立即广播游戏状态，确保新玩家能看到蛇和食物
	g.broadcastGameState()

	// 记录玩家注册成功的日志
	log.Printf("玩家 %s (%s) 已成功注册，当前玩家数: %d", player.Name, player.ID, len(g.Players))
}

// 注销玩家
func (g *Game) unregisterPlayer(player *Player) {
	g.mutex.Lock()
	if _, ok := g.Players[player.ID]; ok {
		// 如果玩家死亡，将蛇身转换为食物
		if player.Snake != nil && len(player.Snake.Body) > 0 {
			for _, segment := range player.Snake.Body {
				g.Foods = append(g.Foods, &Food{X: segment.X, Y: segment.Y})
			}
		}

		delete(g.Players, player.ID)
		close(player.Send)
	}
	g.mutex.Unlock()

	// 广播玩家离开消息
	g.broadcastPlayerLeave(player)

	// 更新玩家列表
	g.broadcastPlayerList()
}

// 更新游戏状态
func (g *Game) update() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// 更新所有蛇的位置
	for _, player := range g.Players {
		if player.Snake != nil {
			player.Snake.Move()

			// 检查是否吃到食物
			g.checkFoodCollision(player)

			// 检查是否撞到墙
			if g.checkWallCollision(player) {
				g.handlePlayerDeath(player)
				continue
			}

			// 检查是否撞到其他蛇
			if g.checkSnakeCollision(player) {
				g.handlePlayerDeath(player)
				continue
			}
		}
	}

	// 广播游戏状态
	g.broadcastGameState()
}

// 检查食物碰撞
func (g *Game) checkFoodCollision(player *Player) {
	head := player.Snake.Body[0]

	for i, food := range g.Foods {
		if head.X == food.X && head.Y == food.Y {
			// 吃到食物，蛇长度增加
			player.Snake.Grow()

			// 移除这个食物并生成新的
			g.Foods[i] = NewRandomFood()
			break
		}
	}
}

// 检查墙壁碰撞
func (g *Game) checkWallCollision(player *Player) bool {
	head := player.Snake.Body[0]
	return head.X < 0 || head.X >= Width || head.Y < 0 || head.Y >= Height
}

// 检查蛇之间的碰撞
func (g *Game) checkSnakeCollision(player *Player) bool {
	head := player.Snake.Body[0]

	// 检查是否撞到自己
	for i := 1; i < len(player.Snake.Body); i++ {
		segment := player.Snake.Body[i]
		if head.X == segment.X && head.Y == segment.Y {
			return true
		}
	}

	// 检查是否撞到其他蛇
	for id, otherPlayer := range g.Players {
		if id == player.ID || otherPlayer.Snake == nil {
			continue
		}

		for _, segment := range otherPlayer.Snake.Body {
			if head.X == segment.X && head.Y == segment.Y {
				return true
			}
		}
	}

	return false
}

// 处理玩家死亡
func (g *Game) handlePlayerDeath(player *Player) {
	// 将蛇身转换为食物
	for _, segment := range player.Snake.Body {
		g.Foods = append(g.Foods, &Food{X: segment.X, Y: segment.Y})
	}

	// 重置玩家的蛇
	player.Snake = NewSnake()

	// 广播玩家死亡消息
	g.broadcastPlayerDeath(player)
}

// 广播游戏状态
func (g *Game) broadcastGameState() {
	// 构建游戏状态
	type GameStatePayload struct {
		Players []PlayerState `json:"players"`
		Foods   []*Food       `json:"foods"`
	}

	playerStates := make([]PlayerState, 0, len(g.Players))
	for _, player := range g.Players {
		if player.Snake != nil {
			playerStates = append(playerStates, PlayerState{
				ID:    player.ID,
				Name:  player.Name,
				Snake: player.Snake,
			})
		}
	}

	payload := GameStatePayload{
		Players: playerStates,
		Foods:   g.Foods,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error marshaling game state:", err)
		return
	}

	message := Message{
		Type:    GameState,
		Payload: payloadBytes,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshaling message:", err)
		return
	}

	// 添加日志记录游戏状态
	log.Printf("广播游戏状态: 玩家数=%d, 食物数=%d", len(playerStates), len(g.Foods))

	g.Broadcast <- messageBytes
}

// 广播玩家列表
func (g *Game) broadcastPlayerList() {
	type PlayerInfo struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	players := make([]PlayerInfo, 0, len(g.Players))
	for _, player := range g.Players {
		players = append(players, PlayerInfo{
			ID:   player.ID,
			Name: player.Name,
		})
	}

	payloadBytes, err := json.Marshal(players)
	if err != nil {
		log.Println("Error marshaling player list:", err)
		return
	}

	message := Message{
		Type:    PlayerList,
		Payload: payloadBytes,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshaling message:", err)
		return
	}

	g.Broadcast <- messageBytes
}

// 广播玩家加入
func (g *Game) broadcastPlayerJoin(player *Player) {
	payload, err := json.Marshal(map[string]string{
		"id":   player.ID,
		"name": player.Name,
	})
	if err != nil {
		log.Println("Error marshaling player join:", err)
		return
	}

	message := Message{
		Type:    PlayerJoin,
		Payload: payload,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshaling message:", err)
		return
	}

	g.Broadcast <- messageBytes
}

// 广播玩家离开
func (g *Game) broadcastPlayerLeave(player *Player) {
	payload, err := json.Marshal(map[string]string{
		"id":   player.ID,
		"name": player.Name,
	})
	if err != nil {
		log.Println("Error marshaling player leave:", err)
		return
	}

	message := Message{
		Type:    PlayerLeave,
		Payload: payload,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshaling message:", err)
		return
	}

	g.Broadcast <- messageBytes
}

// 广播玩家死亡
func (g *Game) broadcastPlayerDeath(player *Player) {
	payload, err := json.Marshal(map[string]string{
		"id":   player.ID,
		"name": player.Name,
	})
	if err != nil {
		log.Println("Error marshaling player death:", err)
		return
	}

	message := Message{
		Type:    PlayerDead,
		Payload: payload,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshaling message:", err)
		return
	}

	g.Broadcast <- messageBytes
}

// 玩家状态
type PlayerState struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Snake *Snake `json:"snake"`
}
