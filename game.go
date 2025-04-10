package main

import (
	"math/rand"
	"time"
)

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Position struct {
	X int
	Y int
}

type Snake struct {
	ID   string
	Name string
	Body []Position
	Dir  Direction
	Dead bool
}

type Food struct {
	Position Position
}

type Game struct {
	Players map[string]*Snake
	Foods   []Food
	Width   int
	Height  int
}

func NewGame(width, height int) *Game {
	return &Game{
		Players: make(map[string]*Snake),
		Foods:   make([]Food, 0),
		Width:   width,
		Height:  height,
	}
}

func (g *Game) AddPlayer(id string) *Snake {
	name := generateRandomName()
	snake := &Snake{
		ID:   id,
		Name: name,
		Body: []Position{{X: rand.Intn(g.Width), Y: rand.Intn(g.Height)}},
		Dir:  Direction(rand.Intn(4)),
	}
	g.Players[id] = snake
	return snake
}

func (g *Game) RemovePlayer(id string) {
	delete(g.Players, id)
}

func (g *Game) GenerateFood() {
	food := Food{
		Position: Position{
			X: rand.Intn(g.Width),
			Y: rand.Intn(g.Height),
		},
	}
	g.Foods = append(g.Foods, food)
}

func generateRandomName() string {
	names := []string{"Snake", "Python", "Cobra", "Viper", "Anaconda", "Boa", "Rattlesnake", "Mamba"}
	return names[rand.Intn(len(names))] + "-" + string(rand.Intn(1000)+1000)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
