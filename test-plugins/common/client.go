package common

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
)

var K8sClientSet *kubernetes.Clientset

func NewClientSet() {

	kubeConfig := os.Getenv("KUBECONFIG")
	if kubeConfig == "" {
		kubeConfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			klog.Fatalf("error creating inClusterConfig %v\n", err)
		}
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("init client set error %v\n", err)
	}
	K8sClientSet = client
	klog.Infof("Initializing k8s client successful")
}
