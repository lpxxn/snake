package main

import (
	"fmt"
	"time"

	"github.com/lpxxn/snake/game"
	"github.com/nsf/termbox-go"
)

func main() {
	// 初始化终端
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	// 创建游戏实例
	g := game.NewGame(40, 20)

	// 创建事件通道
	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	// 游戏主循环
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey {
				switch ev.Key {
				case termbox.KeyArrowUp:
					g.ChangeDirection(game.Up)
				case termbox.KeyArrowDown:
					g.ChangeDirection(game.Down)
				case termbox.KeyArrowLeft:
					g.ChangeDirection(game.Left)
				case termbox.KeyArrowRight:
					g.ChangeDirection(game.Right)
				case termbox.KeyEsc:
					return
				}
			}
		case <-ticker.C:
			// 清屏
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

			// 移动蛇
			g.Move()

			// 绘制游戏状态
			fmt.Print(g.String())

			// 刷新屏幕
			termbox.Flush()

			// 检查游戏是否结束
			if g.GameOver {
				time.Sleep(2 * time.Second)
				return
			}
		}
	}
}