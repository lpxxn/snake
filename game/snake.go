package game

import (
	"math/rand"
)

// 蛇的一个节点
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// 蛇结构
type Snake struct {
	Body      []Point   `json:"body"`
	Direction Direction `json:"direction"`
	Growing   bool      `json:"growing"`
}

// 创建新蛇
func NewSnake() *Snake {
	// 随机初始位置
	x := rand.Intn(Width-10) + 5
	y := rand.Intn(Height-10) + 5

	// 初始长度为3
	body := []Point{
		{X: x, Y: y},
		{X: x - 1, Y: y},
		{X: x - 2, Y: y},
	}

	return &Snake{
		Body:      body,
		Direction: Right,
		Growing:   false,
	}
}

// 改变蛇的方向
func (s *Snake) ChangeDirection(newDirection Direction) {
	// 防止180度转向
	if (s.Direction == Up && newDirection == Down) ||
		(s.Direction == Down && newDirection == Up) ||
		(s.Direction == Left && newDirection == Right) ||
		(s.Direction == Right && newDirection == Left) {
		return
	}

	s.Direction = newDirection
}

// 移动蛇
func (s *Snake) Move() {
	// 获取头部位置
	head := s.Body[0]

	// 根据方向计算新头部位置
	var newHead Point
	switch s.Direction {
	case Up:
		newHead = Point{X: head.X, Y: head.Y - 1}
	case Right:
		newHead = Point{X: head.X + 1, Y: head.Y}
	case Down:
		newHead = Point{X: head.X, Y: head.Y + 1}
	case Left:
		newHead = Point{X: head.X - 1, Y: head.Y}
	}

	// 将新头部添加到身体前面
	s.Body = append([]Point{newHead}, s.Body...)

	// 如果不是在生长，则移除尾部
	if !s.Growing {
		s.Body = s.Body[:len(s.Body)-1]
	} else {
		s.Growing = false
	}
}

// 让蛇生长
func (s *Snake) Grow() {
	s.Growing = true
}
