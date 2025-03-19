package main

import (
	"extender-scheduler/routers"
)

// 如果不实现nodeCacheCapable 就不用初始化这个client-go ClientSet
func init() {
	//handler.NewExtender()
	//common.NewNodeCache(handler.Ex.ClientSet)
}

func main() {
	r := routers.InitMgrRouter()

	r.Run(":32080")
}
