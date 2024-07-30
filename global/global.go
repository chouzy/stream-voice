package global

import (
	"stream-voice/pkg/logger"
	"stream-voice/pkg/setting"
)

var (
	ServerSetting *setting.ServerSettingS
	SocketSetting *setting.WebSocketSettingS
	AsrSetting    *setting.AstSettingS
	LoggerSetting *setting.LoggerSettingS

	Log *logger.Logger
)
