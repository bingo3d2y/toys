package common

import (
	"k8s.io/client-go/tools/cache"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

// NodeCache 用于存储节点信息
type NodeCache struct {
	sync.RWMutex
	nodes map[string]*v1.Node
}

func (c *NodeCache) AddNode(node *v1.Node) {
	c.Lock()
	defer c.Unlock()
	c.nodes[node.Name] = node
}

func (c *NodeCache) GetNode(name string) (*v1.Node, bool) {
	c.RLock()
	defer c.RUnlock()
	node, exists := c.nodes[name]
	return node, exists
}

func NewNodeCache(clientset *kubernetes.Clientset) *NodeCache {
	cacheInfo := &NodeCache{nodes: make(map[string]*v1.Node)}

	// 使用 Informer 监听 Node 资源
	// 默认 0，表示不定期同步，只依赖 Watch 机制
	// 如果指定 Resync 间隔，则周期性地重新获取所有对象
	factory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
	informer := factory.Core().V1().Nodes().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			cacheInfo.AddNode(node)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			node := newObj.(*v1.Node)
			cacheInfo.AddNode(node)
		},
		DeleteFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			cacheInfo.Lock()
			defer cacheInfo.Unlock()
			delete(cacheInfo.nodes, node.Name)
		},
	})

	go informer.Run(make(chan struct{})) // 启动 Informer
	return cacheInfo
}
