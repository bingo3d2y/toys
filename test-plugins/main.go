package main

import (
	"k8s.io/klog/v2"
	"os"

	_ "k8s.io/component-base/metrics/prometheus/clientgo" // for rest client metric registration
	_ "k8s.io/component-base/metrics/prometheus/version"  // for version metric registration
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"test-plugins/plugins/sticky"
)

func init() {
	// 感觉：k8s scheduling framework即kube-scheduler启动时会根据kube-scheduler.conf初始化 k8s clientSet
	//localCommon.NewClientSet()
	//localCommon.NewNodeCache(localCommon.K8sClientSet)
}
func main() {
	// Register custom plugins to the scheduler framework.
	command := app.NewSchedulerCommand(
		// WithPlugin creates an Option based on plugin name and factory. Please don't remove this function: it is used to register out-of-tree plugins,
		// hence there are no references to it from the kubernetes scheduler code base.
		app.WithPlugin(sticky.Name, sticky.NewPlugin),
	)

	err := command.Execute()
	if err != nil {
		klog.Errorf("%v\n", err)
		os.Exit(1)
	}
}
