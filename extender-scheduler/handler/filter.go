package handler

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

// Filter 过滤掉不满足条件的节点
// 当 NodeCacheCapable 设置为 true 时, default scheduler 填充的是： ExtenderArgs.nodeNames
// 当 NodeCacheCapable 设置为 false 时,  default scheduler 填充的是： ExtenderArgs.nodes
func (ex *Extender) Filter(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	nodes := make([]v1.Node, 0)
	nodeNames := make([]string, 0)

	if args.Nodes == nil && args.NodeNames == nil {
		return &extenderv1.ExtenderFilterResult{
			Nodes:     args.Nodes,
			NodeNames: &nodeNames,
		}, nil
	}
	for _, node := range args.Nodes.Items {
		klog.Infof("node name: %s not found %s, skip\n", node.Name, Label)
		_, ok := node.Labels[Label]
		if !ok { // 排除掉不带指定标签的节点
			continue
		}
		nodes = append(nodes, node)
		nodeNames = append(nodeNames, node.Name)
	}

	// 没有满足条件的节点,也不报错，继续调度
	if len(nodes) == 0 {
		klog.Error("custom scheduler not found valid nodes, turn to default scheduler...")
		return &extenderv1.ExtenderFilterResult{
			Nodes: args.Nodes,
			//NodeNames: &nodeNames,
			NodeNames: nil,
		}, nil
	}

	args.Nodes.Items = nodes

	return &extenderv1.ExtenderFilterResult{
		Nodes:     args.Nodes,
		NodeNames: &nodeNames,
	}, nil
}
