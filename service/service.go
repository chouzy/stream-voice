package service

import (
	"github.com/gorilla/websocket"
)

type Server struct {
	conn  *websocket.Conn
	asrCh chan []byte
}

func NewServer(conn *websocket.Conn) *Server {
	return &Server{
		conn:  conn,
		asrCh: make(chan []byte, 1280),
	}
}
