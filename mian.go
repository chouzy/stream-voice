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
	err = s.ReadSection("server", &global.Server)
	if err != nil {
		return err
	}
	err = s.ReadSection("asr", &global.Asr)
	if err != nil {
		return err
	}
	err = s.ReadSection("zap", &global.Log)
	if err != nil {
		return err
	}

	global.Asr.Appid = os.Getenv("XF_ASR_APP_ID")
	global.Asr.ApiSecret = os.Getenv("XF_ASR_API_SECRET")
	global.Asr.ApiKey = os.Getenv("XF_ASR_API_KEY")

	return nil
}

func main() {
	err := initConfig()
	if err != nil {
		panic(err)
	}
}
