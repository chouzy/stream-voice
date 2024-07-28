package global

import (
	"stream-voice/pkg/logger"
	"stream-voice/pkg/setting"
)

var (
	Server string
	Asr    *setting.Ast
	Log    *logger.Zap
)
