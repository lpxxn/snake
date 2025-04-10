package game

import (
	"math/rand"
)

// 食物结构
type Food struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// 创建随机位置的食物
func NewRandomFood() *Food {
	return &Food{
		X: rand.Intn(Width),
		Y: rand.Intn(Height),
	}
}
