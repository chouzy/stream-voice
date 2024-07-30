package router

import (
	"github.com/gin-gonic/gin"
	v1 "stream-voice/router/v1"
)

func NewRouter() *gin.Engine {
	r := gin.New()

	groupV1 := r.Group("/stream-voice/v1")
	{
		groupV1.GET("/wx", v1.MiniProgramController)
	}

	return r
}
