package handler

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

const Label = "nvidia.GPU"

var Ex *Extender

type Extender struct {
	ClientSet *kubernetes.Clientset
}

func NewExtender() {
	clientset, err := NewClient()
	if err != nil {
		log.Fatalf("failed to create k8s clientset: %v", err)
	}

	Ex = &Extender{
		ClientSet: clientset,
	}
}

// NewClient connects to an API server.
func NewClient() (*kubernetes.Clientset, error) {
	kubeConfig := os.Getenv("KUBECONFIG")
	if kubeConfig == "" {
		kubeConfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			log.Fatalf("init clientset error %v\n", err)
			return nil, err
		}
	}
	client, err := kubernetes.NewForConfig(config)
	return client, err
}
