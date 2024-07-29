package global

import (
	"stream-voice/pkg/logger"
	"stream-voice/pkg/setting"
)

var (
	ServerSetting *setting.ServerSettingS
	AsrSetting    *setting.AstSettingS
	LoggerSetting *setting.LogSettingS
	Log           *logger.Logger
)
