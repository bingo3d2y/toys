package handler

import (
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

var DataDict = map[string]int{
	"tesla-t4":    50,
	"ampere-a100": 80,
}

// Prioritize 给 Pod 打分
// 注意：此处返回得分 Scheduler 会将其与其他插件打分合并后再选择节点，因此这里的逻辑不能完全控制最终的调度结果。
// 想要完全控制调度结果，只能在 Filter 接口中实现，过滤掉不满足条件的节点，并对剩余节点进行打分，最终 Filter 接口只返回得分最高的那个节点
func (ex *Extender) Prioritize(args extenderv1.ExtenderArgs) (*extenderv1.HostPriorityList, error) {
	var result extenderv1.HostPriorityList
	for _, node := range args.Nodes.Items {
		klog.Info("...............", node.Name)

		if _, ok := node.Labels["test-label"]; ok {
			klog.Info("found label,:::", node.Name)
			result = append(result, extenderv1.HostPriority{
				Host:  node.Name,
				Score: 100000,
			})
			continue
		}

		// 获取 Node 上的 Label 作为分数
		priorityStr, ok := node.Labels[Label]
		if !ok {
			klog.Errorf("node %q does not have label %s", node.Name, Label)
			continue
		}

		priority := DataDict[priorityStr]

		result = append(result, extenderv1.HostPriority{
			Host:  node.Name,
			Score: int64(priority),
		})
	}

	klog.Info("res info:::", result)
	return &result, nil
}
