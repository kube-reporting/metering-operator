package main

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
)

// NewPodUsage creates a representation of usage using the objects for a Pod and it's associated Node.
func NewPodUsage(pod *v1.Pod, node *v1.Node) (PodUsage, error) {
	if pod == nil || node == nil {
		return PodUsage{}, fmt.Errorf("neither pod nor node may be nil when creating PodUsage, pod: %v, node: %v", pod, node)
	}

	// get pod data
	podMeta := pod.GetObjectMeta()
	if podMeta == nil {
		return PodUsage{}, fmt.Errorf("metadata of Pod must be set to create PodUsage, pod: %v", pod)
	}

	// get node metadata
	nodeMeta := node.GetObjectMeta()
	if nodeMeta == nil {
		return PodUsage{}, fmt.Errorf("metadata of Node must be set to create PodUsage, node: %v", node)
	}

	// sum total Pod resource requests across containers
	podCPU := resource.Quantity{}
	podMemory := resource.Quantity{}
	for _, c := range pod.Spec.Containers {
		if cpu := c.Resources.Requests.Cpu(); cpu != nil {
			podCPU.Add(*cpu)
		}
		if memory := c.Resources.Requests.Memory(); memory != nil {
			podMemory.Add(*memory)
		}
	}

	usage := PodUsage{
		UID:            string(podMeta.GetUID()),
		Name:           podMeta.GetName(),
		Namespace:      podMeta.GetNamespace(),
		StartTime:      pod.Status.StartTime.Time.UTC().Format(time.RFC3339),
		CurrentTime:    time.Now().UTC().Format(time.RFC3339), // should probably use other timesource
		NodeName:       node.GetName(),
		NodeExternalID: node.Spec.ExternalID,
	}

	var ok bool
	if usage.RequestedCPU, ok = podCPU.AsDec().Unscaled(); !ok {
		return PodUsage{}, fmt.Errorf("failed to convert CPU requests to an integer: %v", podCPU)
	}

	if usage.RequestedMemory, ok = podMemory.AsDec().Unscaled(); !ok {
		return PodUsage{}, fmt.Errorf("failed to convert memory requests to an integer: %v", podMemory)
	}

	if usage.NodeCPUCapacity, ok = node.Status.Capacity.Cpu().AsDec().Unscaled(); !ok {
		return PodUsage{}, fmt.Errorf("failed to convert CPU capacity to an integer: %v", node.Status.Capacity.Cpu())
	}

	if usage.NodeMemoryCapacity, ok = node.Status.Capacity.Memory().AsDec().Unscaled(); !ok {
		return PodUsage{}, fmt.Errorf("failed to convert memory capacity to an integer: %v", node.Status.Capacity.Memory())
	}

	if usage.NodeCPUCapacity != 0 {
		usage.CPUPercent = float64(usage.RequestedCPU) / float64(usage.NodeCPUCapacity)
	}

	if usage.NodeMemoryCapacity != 0 {
		usage.MemoryPercent = float64(usage.RequestedMemory) / float64(usage.NodeMemoryCapacity)
	}

	return usage, nil
}

// PodUsage is a selected collection of the data required for Pod utilization statistics.
type PodUsage struct {
	// - Pod info

	// UID is identi***REMOVED***er assigned to Pod by the API server.
	UID string `csv:"uid"`
	// Name of the Pod.
	Name string `csv:"name"`
	// Namespace of the Pod.
	Namespace string `csv:"namespace"`
	// StartTime is when the Kubelet started creating a Pod. This time could be before an image was pulled.
	StartTime string `csv:"startTime"`
	// Time is the current
	CurrentTime string `csv:"currentTime"`

	// - Pod resource requests

	// RequestedCPU is the amount of CPU the Pod requested be minimally available in millicores.
	RequestedCPU int64 `csv:"requestedCPU"`
	// RequestedMemory is the amount of CPU the Pod requested be minimally available in bytes.
	RequestedMemory int64 `csv:"requestedMemory"`

	// - Node info

	// NodeName is the name of the Node the Pod has been assigned to.
	NodeName string `csv:"nodeName"`
	// NodeExternalID is the identi***REMOVED***er for the Node assigned by the cloud provider
	NodeExternalID string `csv:"nodeExternalName"`
	// NodeCPUCapacity is the total CPU capabilities of the Node.
	NodeCPUCapacity int64 `csv:"nodeCPUCapacity"`
	// NodeMemoryCapacity is the total memory capabilities of the Node.
	NodeMemoryCapacity int64 `csv:"nodeMemoryCapacity"`
	// MemoryPercent is the percent of total memory being used by the Pod.
	MemoryPercent float64 `csv:"memoryPercent"`
	// CPUPercent is the percent of total CPU being used by the Pod.
	CPUPercent float64 `csv:"cpuPercent"`
}
