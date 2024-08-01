package service

import (
	"stream-voice/model"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
	conn  *websocket.Conn
	asrCh chan model.Request
}

func NewServer(conn *websocket.Conn) *Server {
	return &Server{
		conn:  conn,
		asrCh: make(chan model.Request, 1),
	}
}

func (s *Server) AsrServer(ctx *gin.Context) (err error) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = s.asrSendMsgToClient(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = s.asrReceiveMsgFromClient(ctx)
	}()

	wg.Wait()
	return nil
}
