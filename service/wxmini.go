package service

import (
	"github.com/gin-gonic/gin"
	"stream-voice/global"
	"time"
)

// ReceiveClientMsg 接收客户端信息并处理数据
func (s *Server) ReceiveClientMsg(ctx *gin.Context) {
	msgCh := make(chan []byte)
	errCh := make(chan error, 1)

	s.conn.SetReadLimit(global.SocketSetting.ReadLimit)
	s.conn.SetPingHandler(func(appData string) error {
		return nil
	})

	// 开启协程读取客户端数据
	go func() {
		for {
			_, msg, err := s.conn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			msgCh <- msg
		}
	}()

	timer := time.NewTimer(global.SocketSetting.KeepAliveTime)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			global.Log.Errorf("timeout err")
			return
		case err := <-errCh:
			global.Log.Errorf("websocket err: %v", err)
			return
		case msg := <-msgCh:
			s.asrCh <- msg
		}
	}
}
