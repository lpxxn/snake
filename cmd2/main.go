package main

import (
	"math/rand"
	"time"

	"github.com/rthornton128/goncurses"
)

type point struct {
	x, y int
}

type snake struct {
	head point
	tail []point
}

type board struct {
	width, height int
	snake         snake
	food          point
	window        *goncurses.Window
}

func main() {
	rand.Seed(time.Now().UnixNano())

	stdscr, _ := goncurses.Init()
	defer goncurses.End()

	height, width := stdscr.MaxYX()

	brd := board{
		width:  width,
		height: height,
		window: stdscr,
		snake: snake{
			head: point{width / 2, height / 2},
			tail: make([]point, 0),
		},
	}

	brd.spawnFood()
	brd.drawBoard()

	direction := "right"

	for {
		key := stdscr.GetChar()
		if key == 'q' {
			break
		}

		switch key {
		case 'w':
			direction = "up"
		case 'a':
			direction = "left"
		case 's':
			direction = "down"
		case 'd':
			direction = "right"
		}

		if !brd.updateSnake(direction) {
			break
		}
		brd.drawBoard()
	}
}

func (b *board) spawnFood() {
	b.food = point{rand.Intn(b.width), rand.Intn(b.height)}
}

func (b *board) drawBoard() {
	b.window.Clear()

	b.window.Box(0, 0)
	b.window.MovePrint(b.snake.head.y, b.snake.head.x, "O")

	for _, pt := range b.snake.tail {
		b.window.MovePrint(pt.y, pt.x, "o")
	}

	b.window.MovePrint(b.food.y, b.food.x, "*")

	b.window.Refresh()
}

func (b *board) updateSnake(direction string) bool {
	var newHead point

	switch direction {
	case "up":
		newHead = point{b.snake.head.x, b.snake.head.y - 1}
	case "down":
		newHead = point{b.snake.head.x, b.snake.head.y + 1}
	case "left":
		newHead = point{b.snake.head.x - 1, b.snake.head.y}
	case "right":
		newHead = point{b.snake.head.x + 1, b.snake.head.y}
	}

	if newHead.x < 1 || newHead.x >= b.width-1 || newHead.y < 1 || newHead.y >= b.height-1 {
		return false
	}

	if newHead == b.food {
		b.spawnFood()
	} else {
		b.snake.tail = append([]point{b.snake.head}, b.snake.tail...)
		if len(b.snake.tail) > 0 {
			b.window.MovePrint(b.snake.tail[len(b.snake.tail)-1].y, b.snake.tail[len(b.snake.tail)-1].x, " ")
			b.snake.tail = b.snake.tail[:len(b.snake.tail)-1]
		}
	}

	b.snake.head = newHead
	return true
}
