package response

import (
	"errors"
	"io"
	"stream-voice/global"
	"stream-voice/model"
	err_code "stream-voice/pkg/err-code"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Response struct {
	Ctx      *gin.Context
	Response *model.Response
	Conn     *websocket.Conn
	isClosed bool
	lock     sync.Mutex
}

func NewResponse(ctx *gin.Context, conn *websocket.Conn, err *err_code.Error) *Response {
	resp := &model.Response{}
	resp.Statue.Code = err.Code()
	resp.Statue.ErrMsg = err.Msg()
	return &Response{
		Ctx:      ctx,
		Response: resp,
		Conn:     conn,
	}
}

func (r *Response) SetData(data model.AsrRespData) *Response {
	r.Response.Data = data
	return r
}

func (r *Response) SendJson() *Response {
	if r.isClosed {
		return r
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	err := r.Conn.WriteJSON(r.Response)
	if err != nil && !errors.Is(err, io.EOF) {
		r.isClosed = true
		global.Log.Errorf("response.SendJson error: %v", err)
	}
	return r
}

func (r *Response) End(code int, reason string) {
	if r.isClosed {
		return
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	err := r.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason))
	if err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, websocket.ErrCloseSent) {
			return
		}
		global.Log.Errorf("websocket close error: %v, code: %v, reason: %v", err, code, reason)
	}
}

func (r *Response) AbortWithJson(httpCode int) {
	r.Ctx.AbortWithStatusJSON(httpCode, r.Response)
}
