package err_code

var (
	Success = NewError(0, "成功")

	WebSocketConnErr    = NewError(500, "websocket连接异常")
	WebSocketReadMsgErr = NewError(501, "websocket读取数据异常")

	RequestFormatError = NewError(1000, "请求格式错误")
	ServerError        = NewError(5000, "服务异常")
)
