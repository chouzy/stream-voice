package main

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"runtime"
	"stream-voice/global"
	"stream-voice/pkg/logger"
	"stream-voice/pkg/setting"
	"time"
)

func setupSetting() error {
	s, err := setting.NewSetting("./conf/config.yaml")
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
	err := setupSetting()
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
		panic(err)
	}

	// 使用量监控日志
	go func() {
		for {
			time.Sleep(time.Minute * 5)

			stats := runtime.MemStats{}
			runtime.ReadMemStats(&stats)
			alloc, syss, idle, inuse, stack := float64(stats.HeapAlloc)/1024/1024, float64(stats.HeapSys)/1024/1024,
				float64(stats.HeapIdle)/1024/1024, float64(stats.HeapInuse)/1024/1024, float64(stats.StackInuse)/1024/1024
			global.Log.WithFields(logger.Fields{
				"GoRoutines": runtime.NumGoroutine(),
				"Memory":     fmt.Sprintf("heap_alloc=%.2fMB heap_sys=%.2fMB heap_idle=%.2fMB heap_inuse=%.2fMB stack=%.2fMB\n", alloc, syss, idle, inuse, stack),
			}).Infof("监控日志")
		}
	}()
}
