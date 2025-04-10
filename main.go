package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/lpxxn/snake/game"
	"github.com/lpxxn/snake/server"
)

func main() {
	port := flag.String("port", "8090", "服务器端口")
	flag.Parse()

	// 初始化游戏
	gameInstance := game.NewGame()
	go gameInstance.Start()

	// 设置WebSocket处理器
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(w, r, gameInstance)
	})

	// 提供静态文件
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// 启动服务器
	log.Printf("服务器启动在 http://localhost:%s", *port)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
