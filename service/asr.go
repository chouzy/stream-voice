package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/url"
	"stream-voice/global"
	"stream-voice/model"
	err_code "stream-voice/pkg/err-code"
	"stream-voice/pkg/logger"
	"stream-voice/pkg/response"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	StatusFirstFrame    = 0
	StatusContinueFrame = 1
	StatusLastFrame     = 2
)

// 和客户端进行通信
func (s *Server) asrConn2Client(ctx *gin.Context) error {
	s.conn.SetReadLimit(global.SocketSetting.ReadLimit)
	s.conn.SetPingHandler(func(appData string) error {
		return nil
	})

	errCh := make(chan error, 1)

	// 接收客户端数据
	go func() {
		defer global.Log.Info("receive client message end")
		for {
			_, msg, err := s.conn.ReadMessage()
			if err != nil {
				global.Log.WithFields(logger.Fields{"reason": err}).Error("receive client message error")
				errCh <- err
				return
			}

			req := &model.Request{}
			if err := json.Unmarshal(msg, req); err != nil {
				global.Log.WithFields(logger.Fields{"reason": err}).Error("parse client message error")
				errCh <- err
				return
			}

			s.reqCh <- req
		}
	}()

	timer := time.NewTimer(global.SocketSetting.KeepAliveTime)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			global.Log.Error("receive client message timeout")
			return nil
		case err := <-errCh:
			return err
		case resp := <-s.respCh:
			response.NewResponse(ctx, s.conn, err_code.Success).SetData(resp.Data.Result.String()).SendJson()
			if resp.Data.Status == 2 {
				return nil
			}
		}
		timer.Reset(global.SocketSetting.KeepAliveTime)
	}
}

// 和服务端进行通信
func (s *Server) asrConn2Xunfei(ctx *gin.Context) error {
	conn, err := initConn()
	if err != nil {
		global.Log.WithFields(logger.Fields{"reason": err}).Errorf("xunfei conn failed")
		return err
	}
	defer conn.Close()

	errCh := make(chan error, 1)
	endCh := make(chan struct{}, 1)

	go func() {
		defer global.Log.Info("receive xunfei message end")
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				global.Log.WithFields(logger.Fields{"reason": err}).Error("receive xunfei message error")
				errCh <- err
				return
			}

			resp := &model.AsrRespData{}
			if err = json.Unmarshal(msg, resp); err != nil {
				global.Log.WithFields(logger.Fields{"reason": err}).Error("parse xunfei message error")
				errCh <- err
				return
			}

			s.respCh <- resp
			global.Log.WithFields(logger.Fields{"status": resp.Data.Status, "data": resp.Data.Result.String()}).Info("xunfei message")
			if resp.Data.Status == 2 {
				endCh <- struct{}{}
				return
			}
		}
	}()

	timer := time.NewTimer(global.SocketSetting.KeepAliveTime)
	defer timer.Stop()

	status := StatusFirstFrame
	for {
		select {
		case <-timer.C:
			global.Log.Error("receive xunfei message timeout")
			return nil
		case err = <-errCh:
			return err
		case <-endCh:
			s.closed <- struct{}{}
			return nil
		case req := <-s.reqCh:
			if req.IsLast {
				status = StatusLastFrame
			}
			switch status {
			case StatusFirstFrame: // 发送第一帧音频，带business 参数
				frameData := map[string]interface{}{
					"common": map[string]interface{}{
						"app_id": global.AsrSetting.Appid,
					},
					"business": map[string]interface{}{
						"language": "zh_cn",
						"domain":   "iat",
						"accent":   "mandarin",
						// "vad_eos":  10000, // 端点等待时间
						// "dwa":      "wpgs", // 开启动态修正
					},
					"data": map[string]interface{}{
						"status":   StatusFirstFrame,
						"format":   "audio/L16;rate=16000",
						"audio":    req.Data,
						"encoding": "raw",
					},
				}
				if err = conn.WriteJSON(frameData); err != nil {
					global.Log.WithFields(logger.Fields{"frameData": frameData}).Error("first send frame error")
					return err
				}
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
				if err = conn.WriteJSON(frameData); err != nil {
					global.Log.WithFields(logger.Fields{"frameData": frameData}).Error("continue send frame error")
					return err
				}
			case StatusLastFrame:
				frameData := map[string]interface{}{
					"data": map[string]interface{}{
						"status":   StatusLastFrame,
						"format":   "audio/L16;rate=16000",
						"audio":    req.Data,
						"encoding": "raw",
					},
				}
				if err = conn.WriteJSON(frameData); err != nil {
					global.Log.WithFields(logger.Fields{"frameData": frameData}).Error("last send frame error")
					return err
				}
				global.Log.Info("send last frame")
				// time.Sleep(10 * time.Second)
				// return nil
				continue
			}
		}
		timer.Reset(global.SocketSetting.KeepAliveTime)
	}
}

func initConn() (*websocket.Conn, error) {
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	// 握手并建立websocket 连接
	authUrl, err := assembleAuthUrl(global.AsrSetting.HostUrl, global.AsrSetting.ApiKey, global.AsrSetting.ApiSecret)
	if err != nil {
		return nil, err
	}
	conn, resp, err := d.Dial(authUrl, nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 101 {
		return nil, fmt.Errorf("conn resp code error and code is %v", resp.StatusCode)
	}
	return conn, nil
}

// 创建鉴权url  apikey 即 hmac username
func assembleAuthUrl(hostUrl string, apiKey, apiSecret string) (string, error) {
	ul, err := url.Parse(hostUrl)
	if err != nil {
		return "", err
	}
	date := time.Now().UTC().Format(time.RFC1123)
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	sign := strings.Join(signString, "\n")
	sha := hmacSha256(sign, apiSecret)
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	callUrl := hostUrl + "?" + v.Encode()
	return callUrl, nil
}

func hmacSha256(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	encodeData := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(encodeData)
}
