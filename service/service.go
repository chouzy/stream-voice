package service

import (
	"fmt"
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

func (s *Server) AsrServer(ctx *gin.Context) error {
	var (
		wg      sync.WaitGroup
		errSend error
		errRcv  error
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		errSend = s.asrSendMsgToClient(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		errRcv = s.asrReceiveMsgFromClient(ctx)
	}()

	wg.Wait()

	if errSend != nil && errRcv != nil {
		return fmt.Errorf("reveived asr result error: %v, received client message error: %v", errSend, errRcv)
	}
	if errSend != nil {
		return errSend
	}
	if errRcv != nil {
		return errRcv
	}

	return nil
}
