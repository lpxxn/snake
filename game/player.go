package game

import (
	"github.com/lpxxn/snake/utils"
)

// 玩家结构
type Player struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Snake *Snake      `json:"snake"`
	Send  chan []byte `json:"-"`
}

// 创建新玩家
func NewPlayer(id string) *Player {
	return &Player{
		ID:    id,
		Name:  utils.GenerateRandomName(),
		Snake: NewSnake(),
		Send:  make(chan []byte, 256),
	}
}
