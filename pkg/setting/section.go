package setting

import "time"

type ServerSettingS struct {
	Type     string
	HttpPort string
	Debug    bool
}

type WebSocketSettingS struct {
	KeepAliveTime   time.Duration
	ReadBufferSize  int
	WriteBufferSize int
	ReadLimit       int64
}

type AstSettingS struct {
	HostUrl   string
	Appid     string
	ApiSecret string
	ApiKey    string
}

type LoggerSettingS struct {
	LogFileName string
	LogFileExt  string
	LogSavePath string
	MaxSize     int
	MaxAge      int
	MaxBackups  int
	Compress    bool
}
