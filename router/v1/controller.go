package v1

import (
	"net/http"
	"stream-voice/global"
	"stream-voice/model"
	err_code "stream-voice/pkg/err-code"
	"stream-voice/pkg/response"
	"stream-voice/service"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func initConn(ctx *gin.Context) (*websocket.Conn, *model.Request, error) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  10280,
		WriteBufferSize: 10280,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		global.Log.Errorf("websocket conn error: %v", err)
		response.NewResponse(ctx, conn, err_code.WebSocketConnErr).AbortWithJson(http.StatusNotAcceptable)
		return nil, nil, err
	}
	if conn == nil {
		global.Log.Errorf("websocket conn is nil")
	}
	global.Log.Debug("websocket conn success")

	// params, ok := ctx.GetQuery("params")
	// if !ok {
	// 	response.NewResponse(ctx, conn, err_code.RequestFormatError).SendJson().End(websocket.CloseInvalidFramePayloadData, "param not found")
	// 	return nil, nil, fmt.Errorf("param not found")
	// }
	//
	request := model.Request{}
	// if err = json.Unmarshal([]byte(params), &request); err != nil {
	// 	response.NewResponse(ctx, conn, err_code.RequestFormatError).SendJson().End(websocket.CloseInvalidFramePayloadData, err.Error())
	// 	return nil, nil, err
	// }

	return conn, &request, nil
}

func MiniProgramController(ctx *gin.Context) {
	conn, _, err := initConn(ctx)
	if err != nil {
		return
	}
	server := service.NewServer(conn)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.SendASRMsgToClient(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		server.ReceiveClientMsg(ctx)
	}()

	wg.Wait()
	response.NewResponse(ctx, conn, err_code.Success).End(websocket.CloseNormalClosure, "websocket close")
}
