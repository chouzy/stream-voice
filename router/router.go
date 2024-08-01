package router

import (
	"github.com/gin-gonic/gin"
	"stream-voice/global"
	v1 "stream-voice/router/v1"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	if global.ServerSetting.Debug {
		r.Use(gin.Logger())
		r.Use(gin.Recovery())
	}

	groupV1 := r.Group("/stream-voice/v1")
	{
		groupV1.GET("/wx", v1.MiniProgramController)
	}

	return r
}
