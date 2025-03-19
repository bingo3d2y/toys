package routers

import "github.com/gin-gonic/gin"

func InitMgrRouter() *gin.Engine {
	r := gin.New()

	MyCustomScheduler(r)

	return r
}
