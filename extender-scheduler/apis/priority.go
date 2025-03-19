package apis

import (
	"extender-scheduler/handler"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	"net/http"
)

func Prioritize(c *gin.Context) {
	klog.Info("begin to [Prioritize]...")

	var args extenderv1.ExtenderArgs

	if err := c.BindJSON(&args); err != nil {
		klog.Errorf("[priority] failed to decode result: %v", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	res, _ := handler.Ex.Prioritize(args)

	c.JSON(http.StatusOK, res)
	klog.Info("prioritize::::", res)
	return
}
