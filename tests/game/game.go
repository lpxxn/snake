package game

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Direction 表示蛇的移动方向
type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

// Point 表示坐标点
type Point struct {
	X, Y int
}

// Snake 表示蛇
type Snake struct {
	body     []Point
	direction Direction
}

// Game 表示游戏状态
type Game struct {
	Width     int
	Height    int
	Snake     *Snake
	Food      Point
	Score     int
	GameOver  bool
}

// NewGame 创建新游戏
func NewGame(width, height int) *Game {
	// 初始化蛇的位置在中心
	startX := width / 2
	startY := height / 2

	snake := &Snake{
		body: []Point{
			{X: startX, Y: startY},
		},
		direction: Right,
	}

	game := &Game{
		Width:  width,
		Height: height,
		Snake:  snake,
		Score:  0,
	}

	// 生成第一个食物
	game.generateFood()

	return game
}

// generateFood 生成新的食物
func (g *Game) generateFood() {
	rand.Seed(time.Now().UnixNano())
	for {
		x := rand.Intn(g.Width)
		y := rand.Intn(g.Height)
		
		// 确保食物不会生成在蛇身上
		valid := true
		for _, p := range g.Snake.body {
			if p.X == x && p.Y == y {
				valid = false
				break
			}
		}
		
		if valid {
			g.Food = Point{X: x, Y: y}
			break
		}
	}
}

// Move 移动蛇
func (g *Game) Move() {
	if g.GameOver {
		return
	}

	// 获取蛇头
	head := g.Snake.body[0]
	newHead := Point{X: head.X, Y: head.Y}

	// 根据方向移动
	switch g.Snake.direction {
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
	if newHead.X < 0 || newHead.X >= g.Width ||
		newHead.Y < 0 || newHead.Y >= g.Height {
		g.GameOver = true
		return
	}

	// 检查是否撞到自己
	for _, p := range g.Snake.body {
		if p.X == newHead.X && p.Y == newHead.Y {
			g.GameOver = true
			return
		}
	}

	// 移动蛇
	g.Snake.body = append([]Point{newHead}, g.Snake.body...)

	// 检查是否吃到食物
	if newHead.X == g.Food.X && newHead.Y == g.Food.Y {
		g.Score++
		g.generateFood()
	} else {
		// 如果没有吃到食物，删除尾部
		g.Snake.body = g.Snake.body[:len(g.Snake.body)-1]
	}
}

// ChangeDirection 改变蛇的方向
func (g *Game) ChangeDirection(d Direction) {
	// 防止180度转向
	if (d == Up && g.Snake.direction == Down) ||
		(d == Down && g.Snake.direction == Up) ||
		(d == Left && g.Snake.direction == Right) ||
		(d == Right && g.Snake.direction == Left) {
		return
	}
	g.Snake.direction = d
}

// String 返回游戏当前状态的字符串表示
func (g *Game) String() string {
	// 创建游戏面板
	board := make([][]string, g.Height)
	for i := range board {
		board[i] = make([]string, g.Width)
		for j := range board[i] {
			board[i][j] = " "
		}
	}

	// 绘制食物
	board[g.Food.Y][g.Food.X] = "*"

	// 绘制蛇
	for i, p := range g.Snake.body {
		if i == 0 {
			board[p.Y][p.X] = "@" // 蛇头
		} else {
			board[p.Y][p.X] = "o" // 蛇身
		}
	}

	// 构建输出字符串
	var result string
	result += fmt.Sprintf("Score: %d\n", g.Score)
	
	// 添加上边界
	result += "+" + strings.Repeat("-", g.Width*2) + "+\n"
	
	// 添加游戏面板
	for _, row := range board {
		result += "|" + strings.Join(row, " ") + "|\n"
	}
	
	// 添加下边界
	result += "+" + strings.Repeat("-", g.Width*2) + "+\n"

	if g.GameOver {
		result += "Game Over!\n"
	}

	return result
}