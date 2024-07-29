package main

import (
	"os"
	"stream-voice/global"
	"stream-voice/pkg/setting"
)

func initConfig() error {
	s, err := setting.NewSetting()
	if err != nil {
		return err
	}
	err = s.ReadSection("server", &global.ServerSetting)
	if err != nil {
		return err
	}
	err = s.ReadSection("asr", &global.AsrSetting)
	if err != nil {
		return err
	}
	err = s.ReadSection("zap", &global.LogSetting)
	if err != nil {
		return err
	}

	if global.ServerSetting.Debug {
		global.AsrSetting.Appid = os.Getenv("XF_ASR_APP_ID")
		global.AsrSetting.ApiSecret = os.Getenv("XF_ASR_API_SECRET")
		global.AsrSetting.ApiKey = os.Getenv("XF_ASR_API_KEY")
	}

	return nil
}

func main() {
	err := initConfig()
	if err != nil {
		panic(err)
	}

}
