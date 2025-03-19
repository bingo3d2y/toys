package routers

import (
	"extender-scheduler/apis"
	"github.com/gin-gonic/gin"
)

func MyCustomScheduler(r *gin.Engine) {
	r.POST("/filter", apis.Filter)
	r.POST("/prioritize", apis.Prioritize)
	// r.POST("/bind", apis.Bind)
	r.POST("/allinone", apis.AllInOne)
}
