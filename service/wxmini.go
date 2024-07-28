package service

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func StartService() {
	r := gin.Default()
	r.GET("/ws", handleWebSocket)
	r.Run("localhost:8080")
}

func handleWebSocket(c *gin.Context) {
	// 升级HTTP连接为WebSocket连接
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// 处理WebSocket连接
	for {
		msgType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Received msg: ", string(p))

		// 发送消息
		err = conn.WriteMessage(msgType, []byte("hello, world"))
		if err != nil {
			log.Println(err)
			return
		}
	}
}
