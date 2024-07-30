package setting

type ServerSettingS struct {
	Type  string
	Debug bool
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
