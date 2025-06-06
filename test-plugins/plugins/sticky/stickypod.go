package sticky

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// Name of the plugin used in the plugin registry and configurations.
	Name     = "StickyPod"
	stateKey = Name + "StateKey"

	kindNode = "Node"

	// Annotation key on VirtualMachine, value is the sticky node (if not empty)
	// Here we assume one VM has only one Pod.
	stickyAnnotationKey = "sticky-nodes"
)

var (
	_ framework.PreFilterPlugin = &StickyPod{}
	_ framework.FilterPlugin    = &StickyPod{}
	_ framework.PostBindPlugin  = &StickyPod{}
)

type StickyPod struct {
	//ClientSet *kubernetes.Clientset
	Handler framework.Handle
}
type stickyState struct {
	nodeExists bool
	// 多个node，逗号分隔
	//nodeList  []*v1.Node
	NodeNames []string
}

// Name returns name of the plugin
func (pl *StickyPod) Name() string {
	return Name
}

// NewPlugin New initializes a new plugin and returns it.
// PluginFactory is a function that builds a plugin.
// type PluginFactory = func(configuration runtime.Object, f framework.Handle) (framework.Plugin, error)
func NewPlugin(_ runtime.Object, handler framework.Handle) (framework.Plugin, error) {

	klog.Infof("Initializing StickyPod scheduling plugin")

	pl := StickyPod{
		Handler: handler,
	}

	return &pl, nil
}

// PreFilter invoked at the preFilter extension point.
func (pl *StickyPod) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	klog.Infof("Prefilter unscheduled pod: %s/%s", pod.Namespace, pod.Name)
	s := stickyState{false, nil}
	defer func() {
		state.Write(stateKey, &s)
	}()

	// Get pod owner reference
	// 这里应该应用 原preFilter的逻辑的
	// 或者这里就应该没有逻辑的，因为前端页面可以控制，不开启sticky node的pod不使用这个scheduler
	podOwnerRef := getPodOwnerRef(pod)
	if podOwnerRef == nil {
		klog.Infof("PreFilter: pod OwnerRef not found or doesn't meet expectation, skip sticky operations")
		//s.nodeList, s.NodeNames = localCommon.NodeCacheInfo.List()
		return framework.NewStatus(framework.Success, "Pod owner ref not found, return")
	}

	// Get sticky info
	ownerName := podOwnerRef.Name
	ns := pod.Namespace
	klog.Infof("PreFilter: parent is %s %s in %s namespace", podOwnerRef.Kind, ownerName, ns)

	// 这是不是能用反射合并下逻辑
	switch podOwnerRef.Kind {
	case "StatefulSet":
		statefulSet, err := pl.Handler.ClientSet().AppsV1().StatefulSets(ns).Get(context.TODO(), ownerName, metav1.GetOptions{ResourceVersion: "0"})
		if err != nil {
			klog.Infof("Get %s %s/%s failed: %v", podOwnerRef.Kind, ns, ownerName, err)
			return framework.NewStatus(framework.Error, "get vmi failed")
		}
		if _, ok := statefulSet.Annotations[stickyAnnotationKey]; !ok {
			return framework.NewStatus(framework.Success, "Pod don't stick nodes ")
		}
		s.nodeExists = true

		stickyNodeList := strings.Split(statefulSet.Annotations[stickyAnnotationKey], ",")
		s.NodeNames = stickyNodeList

		//s.nodeList = make([]*v1.Node, 0, len(stickyNodeList))
		//s.NodeNames = make([]string, 0, len(stickyNodeList))
		//for _, v := range stickyNodeList {
		//	node, exist := localCommon.NodeCacheInfo.GetNode(v)
		//	if !exist {
		//		continue
		//	}
		//	s.nodeList = append(s.nodeList, node)
		//	s.NodeNames = append(s.NodeNames, v)
		//}

		klog.Infof("PreFilter: pod  has sticky nodes %s ,write to scheduling context", s.NodeNames)
		return framework.NewStatus(framework.Success, "Check pod finish, return")
	case "ReplicaSet":
		replicaSet, err := pl.Handler.ClientSet().AppsV1().ReplicaSets(ns).Get(context.TODO(), ownerName, metav1.GetOptions{ResourceVersion: "0"})
		if err != nil {
			klog.Infof("Get %s %s/%s failed: %v", podOwnerRef.Kind, ns, ownerName, err)
			return framework.NewStatus(framework.Error, "get vmi failed")
		}
		if _, ok := replicaSet.Annotations[stickyAnnotationKey]; !ok {
			return framework.NewStatus(framework.Success, "Pod don't stick nodes ")
		}
		s.nodeExists = true

		stickyNodeList := strings.Split(replicaSet.Annotations[stickyAnnotationKey], ",")
		s.NodeNames = stickyNodeList

		//s.nodeList = make([]*v1.Node, 0, len(stickyNodeList))
		//s.NodeNames = make([]string, 0, len(stickyNodeList))
		//
		//for _, v := range stickyNodeList {
		//	node, exist := localCommon.NodeCacheInfo.GetNode(v)
		//	if !exist {
		//		continue
		//	}
		//	s.nodeList = append(s.nodeList, node)
		//	s.NodeNames = append(s.NodeNames, v)
		//}
		klog.Infof("PreFilter: pod  has sticky nodes %s ,write to scheduling context", s.NodeNames)
		return framework.NewStatus(framework.Success, "Check pod finish, return")

	default:
		klog.Infof("PreFilter: pod OwnerRef kind not found, skip sticky operations")
		return framework.NewStatus(framework.Success, "Pod owner ref kind not found, return")

	}

}

func (pl *StickyPod) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	klog.Infof("Filter %s/%s: start", pod.Namespace, pod.Name)
	klog.Infof(":::::::::::::::::::::::%s\n", nodeInfo.Node().Name)

	s, err := state.Read(stateKey)
	if err != nil {
		klog.Infof("Filter: pod %s/%s: read preFilter scheduling context failed: %v", pod.Namespace, pod.Name, err)
		return framework.NewStatus(framework.Error, fmt.Sprintf("read preFilter state fail: %v", err))
	}

	// 流弊...
	r, ok := s.(*stickyState)
	if !ok {
		klog.Errorf("Filter: pod %s/%s: scheduling context found but sticky node missing", pod.Namespace, pod.Name)
		return framework.NewStatus(framework.Error, fmt.Sprintf("convert %+v to stickyState fail", s))
	}

	if !r.nodeExists {
		klog.Infof("Filter: pod %v/%v, sticky node not exist, return success", pod.Namespace, pod.Name)
		return nil
	}
	flag := false
	for _, v := range r.NodeNames {
		if nodeInfo.Node().Name == v {
			flag = true
		}
	}
	if !flag {
		m := fmt.Sprintf("%s node not in sticky nodes list %v\n", nodeInfo.Node().Name, r.NodeNames)
		klog.Infof(m)
		return framework.NewStatus(framework.Unschedulable, m)
	}

	// 这是不是直接用nodeInfo *framework.NodeInfo就行
	//nodes, err := pl.Handler.SnapshotSharedLister().NodeInfos().List()
	//if err != nil {
	//	klog.Infof("Filter: found node info error: %v", err)
	//	return nil
	//}
	//nodeMap := make(map[string]bool)
	//for _, v := range nodes {
	//	nodeMap[v.Node().Name] = true
	//}
	//nodeCnt := 0
	//availableNode := make([]string, 0, len(r.NodeNames))
	//for _, node := range r.NodeNames {
	//	if _, ok := nodeMap[node]; ok {
	//		nodeCnt++
	//	}
	//}
	//if nodeCnt == 0 {
	//	return framework.NewStatus(framework.Unschedulable, "sticky node are not available")
	//}
	//
	//klog.Infof("Filter %s/%s: given node is the sticky node %s, use it", pod.Namespace, pod.Name, strings.Join(availableNode, ","))
	klog.Infof("Filter %s/%s: finish", pod.Namespace, pod.Name)
	return nil
}

func (pl *StickyPod) PostBind(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	klog.Infof("PostBind %s/%s: start", pod.Namespace, pod.Name)

	s, err := state.Read(stateKey)
	if err != nil {
		klog.Infof("PostBind: pod %s/%s: read preFilter scheduling context failed: %v", pod.Namespace, pod.Name, err)
		return
	}

	r, ok := s.(*stickyState)
	if !ok {
		klog.Errorf("PostBind: pod %s/%s: convert failed", pod.Namespace, pod.Name)
		return
	}

	if r.nodeExists {
		klog.Errorf("PostBind: Pod already has sticky annotation, return")
		return
	}

	// 可尝试，不指定sticky node，当pod第一次调度后填充当前节点到annotation，用于下次sticky

	klog.Infof("PostBind: annotating selected node %s to pod", nodeName)

	klog.Infof("PostBind %s/%s: finish", pod.Namespace, pod.Name)
}

// PreFilterExtensions returns prefilter extensions, pod add and remove.
func (pl *StickyPod) PreFilterExtensions() framework.PreFilterExtensions {
	return pl
}

// AddPod from pre-computed data in cycleState.
// no current need for this method.
func (pl *StickyPod) AddPod(ctx context.Context, cycleState *framework.CycleState, podToSchedule *v1.Pod, podToAdd *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	return framework.NewStatus(framework.Success, "")
}

// RemovePod from pre-computed data in cycleState.
// no current need for this method.
func (pl *StickyPod) RemovePod(ctx context.Context, cycleState *framework.CycleState, podToSchedule *v1.Pod, podToRemove *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	return framework.NewStatus(framework.Success, "")
}

// getPodOwnerRef returns the controller of the pod
func getPodOwnerRef(pod *v1.Pod) *metav1.OwnerReference {
	if len(pod.OwnerReferences) == 0 {
		return nil
	}

	// list？难道一下会有很多unscheduled pods
	klog.Info("num::::::::::::::", len(pod.OwnerReferences))
	for i := range pod.OwnerReferences {
		ref := &pod.OwnerReferences[i]
		if *ref.Controller && ref.Kind != kindNode {
			return ref
		}
	}

	return nil
}

// Clone don't really copy the data since there is no need
func (s *stickyState) Clone() framework.StateData {
	return s
}
