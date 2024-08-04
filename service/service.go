package service

import (
	"fmt"
	"stream-voice/global"
	"stream-voice/model"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
	conn   *websocket.Conn
	reqCh  chan *model.Request
	respCh chan *model.AsrRespData

	closed chan struct{}
}

func NewServer(conn *websocket.Conn) *Server {
	return &Server{
		conn:   conn,
		reqCh:  make(chan *model.Request, 5),
		respCh: make(chan *model.AsrRespData, 5),
		closed: make(chan struct{}, 1),
	}
}

func (s *Server) AsrServer(ctx *gin.Context) error {
	var (
		wg        sync.WaitGroup
		errClient error
		errXunfei error
	)

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			global.Log.Info("s.asrConn2Client end")
		}()
		errClient = s.asrConn2Client(ctx)
	}()

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			global.Log.Info("s.asrConn2Xunfei end")
		}()
		errXunfei = s.asrConn2Xunfei(ctx)
	}()

	wg.Wait()

	if errClient != nil && errXunfei != nil {
		return fmt.Errorf("client error: %v, xunfei error: %v", errClient, errXunfei)
	}
	if errClient != nil {
		global.Log.Error("client err")
		return errClient
	}
	if errXunfei != nil {
		global.Log.Error("xunfei err")
		return errXunfei
	}

	return nil
}
