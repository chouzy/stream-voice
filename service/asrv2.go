package service

import (
	"context"
	"fmt"
	"stream-voice/global"
	"stream-voice/model"
	err_code "stream-voice/pkg/err-code"
	"stream-voice/pkg/logger"
	"stream-voice/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
)

// const (
// 	StatusFirstFrame    = 0
// 	StatusContinueFrame = 1
// 	StatusLastFrame     = 2
// )

// 和客户端进行通信
func (s *Server) asrConn2Client(ctx *gin.Context) error {
	s.conn.SetReadLimit(global.SocketSetting.ReadLimit)
	s.conn.SetPingHandler(func(appData string) error {
		return nil
	})

	var (
		errRead error
	)

	context := ctx.Value("ctx").(context.Context)
	// 开启协程读取客户端数据
	go func() {
		for {
			select {
			case <-ctx.Done():
				global.Log.Info("client conn is done and finish read from client")
				return
			case <-context.Done():
				global.Log.Info("content is done")
				return
			default:
			}

			_, msg, err := s.conn.ReadMessage()
			if err != nil {
				errRead = err
				global.Log.Error(err.Error())
				return
			}

			req := model.Request{}
			if err = json.Unmarshal(msg, &req); err != nil {
				errRead = err
				global.Log.Error(err.Error())
				return
			}

			global.Log.WithFields(logger.Fields{"request": len(msg)}).Info("read client message")
			s.reqCh <- req
		}
	}()

	// 向客户端发送数据
	for {
		select {
		case <-ctx.Done():
			global.Log.Info("client conn is done and finish send to client")
			return nil
		case <-context.Done():
			global.Log.Info("content is done")
			return nil
		default:
		}

		resp := <-s.respCh
		response.NewResponse(ctx, s.conn, err_code.Success).SetData(resp.Data.Result.String()).SendJson()
		if resp.Data.Status == 2 {
			global.Log.Info("client message send finish")
			break
		}
	}

	// wg.Wait()
	return errRead
}

// 和服务端进行通信
func (s *Server) asrConn2Xunfei(ctx *gin.Context) error {
	conn, err := initConn()
	if err != nil {
		global.Log.Error("xunfei conn err")
		return err
	}
	defer conn.Close()

	var (
		errRead error
	)

	context := ctx.Value("ctx").(context.Context)
	// 向讯飞发送数据
	go func() {
		status := StatusFirstFrame
		for {
			select {
			case <-ctx.Done():
				global.Log.Info("client conn is done and finish send to xunfei")
				return
			case <-context.Done():
				global.Log.Info("content is done")
				return
			case req := <-s.reqCh:
				// global.Log.WithFields(logger.Fields{"size": len(req)}).Info("send client message")
				if req.IsLast {
					status = StatusLastFrame
				}
				switch status {
				case StatusFirstFrame: // 发送第一帧音频，带business 参数
					frameData := map[string]interface{}{
						"common": map[string]interface{}{
							"app_id": global.AsrSetting.Appid, // appid 必须带上，只需第一帧发送
						},
						"business": map[string]interface{}{ // business 参数，只需一帧发送
							"language": "zh_cn",
							"domain":   "iat",
							"accent":   "mandarin",
						},
						"data": map[string]interface{}{
							"status":   StatusFirstFrame,
							"format":   "audio/L16;rate=16000",
							"audio":    req.Data,
							"encoding": "raw",
						},
					}
					conn.WriteJSON(frameData)
					global.Log.Info("send first frame")
					status = StatusContinueFrame
				case StatusContinueFrame:
					frameData := map[string]interface{}{
						"data": map[string]interface{}{
							"status":   StatusContinueFrame,
							"format":   "audio/L16;rate=16000",
							"audio":    req.Data,
							"encoding": "raw",
						},
					}
					conn.WriteJSON(frameData)
				case StatusLastFrame:
					frameData := map[string]interface{}{
						"data": map[string]interface{}{
							"status":   StatusLastFrame,
							"format":   "audio/L16;rate=16000",
							"audio":    req.Data,
							"encoding": "raw",
						},
					}
					conn.WriteJSON(frameData)
					global.Log.Info("send last frame")
					return
				}
			}
		}
	}()

	//获取返回的数据
	for {
		select {
		case <-ctx.Done():
			global.Log.Info("ctx is done")
			return nil
		case <-context.Done():
			global.Log.Info("content is done")
			return nil
		default:
		}

		var resp = model.AsrRespData{}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			global.Log.Error("read xunfei reaponse err")
			break
		}
		json.Unmarshal(msg, &resp)
		global.Log.WithFields(logger.Fields{"response": resp}).Info("xunfei response")
		s.respCh <- resp
		if resp.Code != 0 {
			global.Log.WithFields(logger.Fields{"respCode": resp.Code}).Error("xunfei response code err")
			errRead = fmt.Errorf("xunfei response code err: %v", resp.Code)
			break
		}
		if resp.Data.Status == 2 {
			global.Log.WithFields(logger.Fields{"response": resp.Message}).Info("xunfei response end")
			break
		}
	}

	return errRead
}
