package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

const (
	width  = 40
	height = 20
)

type point struct {
	x, y int
}

type snake struct {
	body      []point
	direction point
}

func main() {
	rand.Seed(time.Now().UnixNano())

	s := newSnake()

	var food point
	s.generateFood(&food)

	score := 0

	for {
		clearScreen()
		printBoard(s, food, score)
		time.Sleep(time.Second / 10)

		if !s.move() {
			break
		}

		if s.head() == food {
			score++
			s.grow()
			s.generateFood(&food)
		}
	}

	fmt.Println("Game over!")
}

func newSnake() *snake {
	s := &snake{
		body: []point{
			{width / 2, height / 2},
			{width/2 - 1, height / 2},
		},
		direction: point{1, 0},
	}
	return s
}

func (s *snake) head() point {
	return s.body[0]
}

func (s *snake) grow() {
	tail := s.body[len(s.body)-1]
	s.body = append(s.body, tail)
}

func (s *snake) move() bool {
	head := s.head()

	// calculate new head position
	newHead := point{head.x + s.direction.x, head.y + s.direction.y}

	// check for collision with walls or own body
	if newHead.x < 0 || newHead.x >= width || newHead.y < 0 || newHead.y >= height {
		return false
	}

	for _, p := range s.body {
		if newHead == p {
			return false
		}
	}

	// move the snake
	s.body = append([]point{newHead}, s.body[:len(s.body)-1]...)

	return true
}

func (s *snake) generateFood(f *point) {
	for {
		*f = point{rand.Intn(width), rand.Intn(height)}
		if !s.contains(*f) {
			break
		}
	}
}

func (s *snake) contains(p point) bool {
	for _, v := range s.body {
		if v == p {
			return true
		}
	}
	return false
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func printBoard(s *snake, food point, score int) {
	board := make([][]byte, height)
	for i := range board {
		board[i] = make([]byte, width)
	}

	// draw the snake
	for _, p := range s.body {
		board[p.y][p.x] = '*'
	}

	// draw the food
	board[food.y][food.x] = '@'

	// print the board
	fmt.Println("Score:", score)
	for _, line := range board {
		fmt.Println(string(line))
	}
}
