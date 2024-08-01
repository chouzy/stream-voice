package test

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestAsr(t *testing.T) {
	file := ""
	frameSize := 1280
	audioFile, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer audioFile.Close()

	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	conn, resp, err := d.Dial("wx://127.0.0.1:8080/", nil)
	if err != nil || resp.StatusCode != 101 {
		t.Fatal(err)
	}
	defer conn.Close()
	t.Log("websocket conn success")

	buffer := make([]byte, frameSize)
	end := false
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
			"data": buffer[:len],
		})
		if end {
			return
		}
	}
}
