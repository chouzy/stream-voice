package service

import (
	"github.com/gorilla/websocket"
)

type Controller interface {
	Asr()
}

type Server struct {
	conn *websocket.Conn
	asr  chan []byte
}

func NewServer(conn *websocket.Conn) *Server {
	return &Server{
		conn: conn,
		asr:  make(chan []byte, 1024),
	}
}
