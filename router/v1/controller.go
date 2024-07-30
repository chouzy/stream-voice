package v1

import (
	"encoding/json"
	"net/http"
	"stream-voice/global"
	"stream-voice/model"
	err_code "stream-voice/pkg/err-code"
	"stream-voice/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func initConn(ctx *gin.Context) (*websocket.Conn, *model.Request, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		global.Log.Errorf("websocket conn error: %w", err)
		response.NewResponse(ctx, conn, err_code.WebSocketConnErr).AbortWithJson(http.StatusNotAcceptable)
		return nil, nil, err
	}

	params, ok := ctx.GetQuery("params")
	if !ok {
		response.NewResponse(ctx, conn, err_code.RequestFormatError).SendJson().End(websocket.CloseInvalidFramePayloadData, "param not found")
		return nil, nil, err
	}

	request := model.Request{}
	if err = json.Unmarshal([]byte(params), &request); err != nil {
		response.NewResponse(ctx, conn, err_code.RequestFormatError).SendJson().End(websocket.CloseInvalidFramePayloadData, err.Error())
		return nil, nil, err
	}

	return conn, &request, nil
}

func MiniProgramController(ctx *gin.Context) {
	conn, request, err := initConn(ctx)
	if err != nil {
		return
	}

}