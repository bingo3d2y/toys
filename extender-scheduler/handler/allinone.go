package handler

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	"sort"
	"strconv"
)

type NodeScore struct {
	Node  v1.Node
	Score int64
}

type NodeScoreList struct {
	NodeList []*NodeScore
}

func (l NodeScoreList) Len() int {
	return len(l.NodeList)
}

func (l NodeScoreList) Swap(i, j int) {
	l.NodeList[i], l.NodeList[j] = l.NodeList[j], l.NodeList[i]
}

func (l NodeScoreList) Less(i, j int) bool {
	return l.NodeList[i].Score < l.NodeList[j].Score
}

// FilterOnlyOne 过滤掉不满足条件的节点,并将其余节点打分排序，最终只返回得分最高的节点以实现完全控制调度结果
func (ex *Extender) FilterOnlyOne(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	// 过滤掉不满足条件的节点
	nodeScores := &NodeScoreList{NodeList: make([]*NodeScore, 0)}

	for _, node := range args.Nodes.Items {
		_, ok := node.Labels[Label]
		if !ok { // 排除掉不带指定标签的节点
			continue
		}
		// 对剩余节点打分
		score := ComputeScore(node)
		nodeScores.NodeList = append(nodeScores.NodeList, &NodeScore{Node: node, Score: score})
	}
	// 没有满足条件的节点就报错
	if len(nodeScores.NodeList) == 0 {
		return &extenderv1.ExtenderFilterResult{
			Nodes: args.Nodes,
			//NodeNames: args.NodeNames,
			NodeNames: nil,
		}, nil
	}
	// 排序
	sort.Sort(nodeScores)
	// 然后取最后一个，即得分最高的节点，这样由于 Filter 只返回了一个节点，因此最终肯定会调度到该节点上
	m := (*nodeScores).NodeList[len((*nodeScores).NodeList)-1]

	// 组装一下返回结果
	args.Nodes.Items = []v1.Node{m.Node}

	return &extenderv1.ExtenderFilterResult{
		Nodes:     args.Nodes,
		NodeNames: &[]string{m.Node.Name},
	}, nil
}

func ComputeScore(node v1.Node) int64 {
	// 获取 Node 上的 Label 作为分数
	priorityStr, ok := node.Labels[Label]
	if !ok {
		klog.Errorf("node %q does not have label %s", node.Name, Label)
		return 0
	}

	priority, err := strconv.Atoi(priorityStr)
	if err != nil {
		klog.Errorf("node %q has priority %s are invalid", node.Name, priorityStr)
		return 0
	}
	return int64(priority)
}
