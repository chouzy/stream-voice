package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"stream-voice/global"
	"stream-voice/model"
	err_code "stream-voice/pkg/err-code"
	"stream-voice/pkg/logger"
	"stream-voice/pkg/response"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	StatusFirstFrame    = 0
	StatusContinueFrame = 1
	StatusLastFrame     = 2
)

// asrReceiveMsgFromClient 接收客户端信息并处理数据
func (s *Server) asrReceiveMsgFromClient(ctx *gin.Context) error {
	msgCh := make(chan []byte)
	errCh := make(chan error, 1)

	s.conn.SetReadLimit(global.SocketSetting.ReadLimit)
	s.conn.SetPingHandler(func(appData string) error {
		return nil
	})

	// 开启协程读取客户端数据
	go func() {
		select {
		case <-ctx.Done():
			global.Log.Info("client conn is done and finish client read goroutine")
			return
		default:
		}
		for {
			_, msg, err := s.conn.ReadMessage()
			if err != nil {
				global.Log.Errorf("client read message err: %v", err)
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
			global.Log.Error("client read message timeout")
			return fmt.Errorf("client read message timeout")
		case err := <-errCh:
			return err
		case msg := <-msgCh:
			s.asrCh <- msg
		}
	}
}

// asrSendMsgToClient 发送信息给客户端
func (s *Server) asrSendMsgToClient(ctx *gin.Context) error {
	conn, err := initConn()
	if err != nil {
		global.Log.Errorf("xunfei websocket conn err: %v", err)
		return err
	}
	defer conn.Close()

	// 发送数据
	go func() {
		status := StatusFirstFrame
		for {
			data := <-s.asrCh
			if string(data) == "--end--" {
				status = StatusLastFrame
			}
			select {
			case <-ctx.Done():
				global.Log.Info("client conn is done and finish xunfei server goroutine")
				return
			default:
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
						"audio":    base64.StdEncoding.EncodeToString(data),
						"encoding": "raw",
					},
				}
				conn.WriteJSON(frameData)
				status = StatusContinueFrame
			case StatusContinueFrame:
				frameData := map[string]interface{}{
					"data": map[string]interface{}{
						"status":   StatusContinueFrame,
						"format":   "audio/L16;rate=16000",
						"audio":    base64.StdEncoding.EncodeToString(data),
						"encoding": "raw",
					},
				}
				conn.WriteJSON(frameData)
			case StatusLastFrame:
				frameData := map[string]interface{}{
					"data": map[string]interface{}{
						"status":   StatusLastFrame,
						"format":   "audio/L16;rate=16000",
						"audio":    base64.StdEncoding.EncodeToString(data),
						"encoding": "raw",
					},
				}
				conn.WriteJSON(frameData)
				return
			}

			time.Sleep(40 * time.Millisecond)
		}
	}()

	// 获取返回数据
	for {
		// TODO: 是否需要监听ctx.Done()
		var resp = model.AsrRespData{}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			global.Log.Errorf("xunfei conn read message err: %v", err)
			return err
		}
		json.Unmarshal(msg, &resp)
		if resp.Code != 0 {
			global.Log.WithFields(logger.Fields{"response": resp}).Errorf("xunfei response err: %v", err)
			return fmt.Errorf("xunfei response err: %v", err)
		}
		global.Log.WithFields(logger.Fields{"response": string(msg)}).Info("xunfei response")
		response.NewResponse(ctx, s.conn, err_code.Success).SetData(resp).SendJson()
		if resp.Data.Status == 2 {
			break
		}
	}
	return nil
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
	if err != nil || resp.StatusCode != 101 {
		return nil, err
	}
	return conn, nil
}

// 创建鉴权url  apikey 即 hmac username
func assembleAuthUrl(hostUrl string, apiKey, apiSecret string) (string, error) {
	ul, err := url.Parse(hostUrl)
	if err != nil {
		return "", err
	}
	// 签名时间
	date := time.Now().UTC().Format(time.RFC1123)
	// date = "Tue, 28 May 2019 09:10:42 MST"
	// 参与签名的字段 host ,date, request-line
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	// 拼接签名字符串
	sign := strings.Join(signString, "\n")
	// 签名结果
	sha := hmacSha256(sign, apiSecret)
	// 构建请求参数 此时不需要urlencoding
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
		"hmac-sha256", "host date request-line", sha)
	// 将请求参数使用base64编码
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	// 将编码后的字符串url encode后添加到url后面
	callUrl := hostUrl + "?" + v.Encode()
	return callUrl, nil
}

func hmacSha256(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	encodeData := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(encodeData)
}
