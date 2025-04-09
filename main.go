package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/lpxxn/snake/game"
)

var (
	addr     = flag.String("addr", ":8981", "http service address")
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源的WebSocket连接
		},
	}
)

func main() {
	flag.Parse()

	// 创建游戏服务器
	gameServer := game.NewGameServer(30, 20)

	// 处理WebSocket连接
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("升级WebSocket连接失败: %v\n", err)
			return
		}
		gameServer.HandleNewPlayer(conn)
	})

	// 提供静态文件服务
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// 启动HTTP服务器
	log.Printf("服务器启动在 %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
