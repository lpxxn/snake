package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/eiannone/keyboard"
)

const (
	width  = 40
	height = 20
)

type point struct {
	x, y int
}

type game struct {
	snake   []point
	apple   point
	running bool
	score   int
}

func (g *game) init() {
	g.snake = []point{{width / 2, height / 2}}
	g.randomApple()
	g.running = true
	g.score = 0
}

func (g *game) randomApple() {
	rand.Seed(time.Now().UnixNano())
	for {
		g.apple = point{rand.Intn(width), rand.Intn(height)}
		for _, p := range g.snake {
			if p == g.apple {
				continue
			}
		}
		break
	}
}

func (g *game) draw() {
	fmt.Print("\033[2J") // clear the terminal
	fmt.Print("\033[H")  // move the cursor to the top left corner
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			p := point{x, y}
			if p == g.apple {
				fmt.Print("@")
			} else if contains(g.snake, p) {
				fmt.Print("o")
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println()
	}
	fmt.Printf("Score: %d\n", g.score)
}

func (g *game) update(direction point) {
	head := g.snake[len(g.snake)-1]
	newHead := point{head.x + direction.x, head.y + direction.y}
	if newHead == g.apple {
		g.score++
		g.snake = append(g.snake, newHead)
		g.randomApple()
	} else if contains(g.snake, newHead) || newHead.x < 0 || newHead.x >= width || newHead.y < 0 || newHead.y >= height {
		g.running = false
	} else {
		g.snake = append(g.snake[1:], newHead)
	}
}

func contains(points []point, p point) bool {
	for _, q := range points {
		if p == q {
			return true
		}
	}
	return false
}

func main() {
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()
	g := game{}
	g.init()
	direction := point{1, 0}
	for g.running {
		g.draw()
		time.Sleep(time.Second / 1000)
		g.update(direction)
		switch direction {
		case point{1, 0}:
			if isKeyPressed("up") {
				direction = point{0, -1}
			} else if isKeyPressed("down") {
				direction = point{0, 1}
			}
		case point{-1, 0}:
			if isKeyPressed("up") {
				direction = point{0, -1}
			} else if isKeyPressed("down") {
				direction = point{0, 1}
			}
		case point{0, 1}:
			if isKeyPressed("left") {
				direction = point{-1, 0}
			} else if isKeyPressed("right") {
				direction = point{1, 0}
			}

		case point{0, -1}:
			if isKeyPressed("left") {
				direction = point{-1, 0}
			} else if isKeyPressed("right") {
				direction = point{1, 0}
			}
		}
	}
	fmt.Printf("Game over! Final score: %d\n", g.score)
}

func isKeyPressed(key string) bool {
	// Check if the specified key is pressed.
	// This implementation uses the "github.com/eiannone/keyboard" package,
	// which you may need to install with "go get github.com/eiannone/keyboard".
	// Alternatively, you could use a different package or implementation
	// to handle keyboard input.
	_, keyEvent, _ := keyboard.GetKey()
	if key == "up" {
		return keyEvent == keyboard.KeyArrowUp
	} else if key == "down" {
		return keyEvent == keyboard.KeyArrowDown
	} else if key == "left" {
		return keyEvent == keyboard.KeyArrowLeft
	} else if key == "right" {
		return keyEvent == keyboard.KeyArrowRight
	}
	return false
}

const (
	upArrowKeyCode    = keyboard.KeyArrowUp
	downArrowKeyCode  = keyboard.KeyArrowDown
	leftArrowKeyCode  = keyboard.KeyArrowLeft
	rightArrowKeyCode = keyboard.KeyArrowRight
)
