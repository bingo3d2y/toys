package apis

import (
	"extender-scheduler/handler"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	"net/http"
)

func Filter(c *gin.Context) {
	klog.Info("begin to [Filter]...")

	var args extenderv1.ExtenderArgs
	if err := c.BindJSON(&args); err != nil {
		klog.Errorf("[filter] failed to decode result: %v", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	klog.Infof("begin to schedule %s pod in %s namespace...\n", args.Pod.Name, args.Pod.Namespace)
	res, _ := handler.Ex.Filter(args)
	c.JSON(http.StatusOK, res)
	return
}
