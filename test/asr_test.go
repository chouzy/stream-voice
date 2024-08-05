package test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"stream-voice/model"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestAsr(t *testing.T) {
	file := "../doc/16k_10.pcm"
	frameSize := 1280

	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, resp, err := d.Dial("ws://127.0.0.1:8080/stream-voice/v1/wx", nil)
	if err != nil || resp.StatusCode != 101 {
		panic(err)
	}
	defer conn.Close()
	t.Log("websocket conn success")

	go func() {
		audioFile, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer audioFile.Close()

		buffer := make([]byte, frameSize)
		end := false
		index := 0
		t.Log("message send start")
		for {
			len, err := audioFile.Read(buffer)
			if err != nil {
				if err == io.EOF {
					end = true
					t.Log("audio file read finish")
				} else {
					t.Fatal("audio file read fatal")
				}
			}
			conn.WriteJSON(map[string]interface{}{
				"data":   base64.StdEncoding.EncodeToString(buffer[:len]),
				"isLast": end,
			})
			index++
			if end {
				t.Log("message send end")
				return
			}
			time.Sleep(40 * time.Millisecond)
		}
	}()

	for {
		var resp = model.Response{}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Log("read message error:", err)
			break
		}
		json.Unmarshal(msg, &resp)
		t.Logf("response: %+v", resp)
	}
}
