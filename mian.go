package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"stream-voice/global"
	"stream-voice/pkg/logger"
	"stream-voice/pkg/setting"
	"time"

	"github.com/natefinch/lumberjack"
)

var (
	config string
)

func setupFlag() error {
	flag.StringVar(&config, "config", "./conf/config.yaml", "配置文件路径")
	flag.Parse()
	return nil
}

func setupSetting() error {
	s, err := setting.NewSetting(config)
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
	err = s.ReadSection("zap", &global.LoggerSetting)
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

func setupLogger() error {
	fileName := global.LoggerSetting.LogSavePath + "/" +
		global.LoggerSetting.LogFileName + "_" + time.Now().Format("20060102150405") + global.LoggerSetting.LogFileExt
	global.Log = logger.NewLogger(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    global.LoggerSetting.MaxSize,
		MaxAge:     global.LoggerSetting.MaxAge,
		MaxBackups: global.LoggerSetting.MaxBackups,
		Compress:   global.LoggerSetting.Compress,
	}, "")
	return nil
}

func initSetting() error {
	err := setupFlag()
	if err != nil {
		return fmt.Errorf("init.setupFlag err: %w", err)
	}
	err = setupSetting()
	if err != nil {
		return fmt.Errorf("init.setupSetting err: %w", err)
	}
	err = setupLogger()
	if err != nil {
		return fmt.Errorf("init.setupLogger err: %w", err)
	}
	return nil
}

func main() {
	err := initSetting()
	if err != nil {
		log.Fatal(err)
	}

}
