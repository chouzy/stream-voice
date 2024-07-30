package router

import "github.com/gin-gonic/gin"

func NewRouter() *gin.Engine {
	r := gin.New()

	groupV1 := r.Group("/stream-voice/v1")
	{
		groupV1.GET("/wx")
	}

	return r
}
