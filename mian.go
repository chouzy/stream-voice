package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"stream-voice/global"
	"stream-voice/pkg/logger"
	"stream-voice/pkg/setting"
	"stream-voice/router"
	"syscall"
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
	err = s.ReadSection("Server", &global.ServerSetting)
	if err != nil {
		return err
	}
	err = s.ReadSection("WebSocket", &global.SocketSetting)
	if err != nil {
		return err
	}
	err = s.ReadSection("Asr", &global.AsrSetting)
	if err != nil {
		return err
	}
	err = s.ReadSection("Logger", &global.LoggerSetting)
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

	gin.SetMode(gin.DebugMode)
	r := router.NewRouter()
	s := http.Server{
		Addr:           "127.0.0.1:" + global.ServerSetting.HttpPort,
		Handler:        r,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("s.ListenAndServe err: %v", err)
		}
	}()

	// 待所有服务请求完成后再关闭服务
	// 适用于k8s和docker的服务重启
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 接收系统信号量
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}

	log.Println("server exiting")
}
