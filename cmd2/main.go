package main

import (
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

type point struct {
	x, y int
}

type snake struct {
	body []point
	dir  point
}

func (s *snake) move() {
	head := s.body[0]
	newHead := point{head.x + s.dir.x, head.y + s.dir.y}
	s.body = append([]point{newHead}, s.body[:len(s.body)-1]...)
}

func (s *snake) grow() {
	head := s.body[0]
	newHead := point{head.x + s.dir.x, head.y + s.dir.y}
	s.body = append([]point{newHead}, s.body...)
}

func (s *snake) collidesWith(p point) bool {
	for _, b := range s.body {
		if b == p {
			return true
		}
	}
	return false
}

func (s *snake) collidesWithSelf() bool {
	head := s.body[0]
	for _, b := range s.body[1:] {
		if b == head {
			return true
		}
	}
	return false
}

func (s *snake) collidesWithWall(w, h int) bool {
	head := s.body[0]
	if head.x < 0 || head.x >= w || head.y < 0 || head.y >= h {
		return true
	}
	return false
}

func (s *snake) draw() {
	for _, b := range s.body {
		termbox.SetCell(b.x, b.y, '█', termbox.ColorGreen, termbox.ColorDefault)
	}
}

func (s *snake) handleInput(ev termbox.Event) {
	switch ev.Key {
	case termbox.KeyArrowUp:
		if s.dir.y != 1 {
			s.dir = point{0, -1}
		}
	case termbox.KeyArrowDown:
		if s.dir.y != -1 {
			s.dir = point{0, 1}
		}
	case termbox.KeyArrowLeft:
		if s.dir.x != 1 {
			s.dir = point{-1, 0}
		}
	case termbox.KeyArrowRight:
		if s.dir.x != -1 {
			s.dir = point{1, 0}
		}
	}
}

func spawnFood(w, h int, s *snake) point {
	var p point
	for {
		p = point{rand.Intn(w), rand.Intn(h)}
		if !s.collidesWith(p) {
			break
		}
	}
	return p
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	rand.Seed(time.Now().UnixNano())

	w, h := termbox.Size()

	snake := snake{
		body: []point{{w / 2, h / 2}},
		dir:  point{1, 0},
	}

	food := spawnFood(w, h, &snake)

	termbox.SetInputMode(termbox.InputEsc)

mainLoop:
	for {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

		snake.draw()

		termbox.SetCell(food.x, food.y, '█', termbox.ColorRed, termbox.ColorDefault)

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			snake.handleInput(ev)
		case termbox.EventError:
			panic(ev.Err)
		}

		if snake.collidesWith(food) {
			snake.grow()
			food = spawnFood(w, h, &snake)
		} else {
			snake.move()
		}

		if snake.collidesWithSelf() || snake.collidesWithWall(w, h) {
			break mainLoop
		}

		termbox.Flush()

		time.Sleep(time.Second / 10)
	}

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termbox.SetCell(w/2-4, h/2, 'G', termbox.ColorGreen, termbox.ColorDefault)
	termbox.SetCell(w/2-3, h/2, 'A', termbox.ColorGreen, termbox.ColorDefault)
	termbox.SetCell(w/2-2, h/2, 'M', termbox.ColorGreen, termbox.ColorDefault)
	termbox.SetCell(w/2-1, h/2, 'E', termbox.ColorGreen, termbox.ColorDefault)
	termbox.SetCell(w/2, h/2, ' ', termbox.ColorGreen, termbox.ColorDefault)
	termbox.SetCell(w/2+1, h/2, 'O', termbox.ColorGreen, termbox.ColorDefault)
	termbox.SetCell(w/2+2, h/2, 'V', termbox.ColorGreen, termbox.ColorDefault)
	termbox.SetCell(w/2+3, h/2, 'E', termbox.ColorGreen, termbox.ColorDefault)
	termbox.Flush()

	time.Sleep(time.Second * 2)
}
