package helmapi

import (
	"fmt"
	"github.com/softplan/tenkai-api/dbms/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/kubernetes/pkg/util/node"
	"time"
)

//DeletePod - Delete a Pod
func DeletePod(kubeconfig string, podName string, namespace string) error {
	_, client, err := getKubeClient(settings.KubeContext, kubeconfig)
	if err != nil {
		return err
	}
	opts := &metav1.DeleteOptions{}
	err = client.CoreV1().Pods(namespace).Delete(podName, opts)
	return err
}

//GetPods Method
func GetPods(kubeconfig string, namespace string) ([]model.Pod, error) {
	pods := make([]model.Pod, 0)
	_, client, err := getKubeClient(settings.KubeContext, kubeconfig)
	if err != nil {
		return pods, err
	}
	opts := &metav1.ListOptions{}
	list, err := client.CoreV1().Pods(namespace).List(*opts)
	var pod *model.Pod
	for _, element := range list.Items {

		pod = fillPod(element)

		pods = append(pods, *pod)
	}
	return pods, nil
}

func fillPod(pod v1.Pod) *model.Pod {

	result := &model.Pod{Name: pod.Name}

	restarts := 0
	totalContainers := len(pod.Spec.Containers)
	readyContainers := 0

	result.Status = string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		result.Status = pod.Status.Reason
	}

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					result.Status = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					result.Status = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				result.Status = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			result.Status = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			result.Status = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		restarts = 0
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				result.Status = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				result.Status = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					result.Status = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					result.Status = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
				readyContainers++
			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if result.Status == "Completed" && hasRunning {
			result.Status = "Running"
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == node.NodeUnreachablePodReason {
		result.Status = "Unknown"
	} else if pod.DeletionTimestamp != nil {
		result.Status = "Terminating"
	}

	result.Ready = fmt.Sprintf("%d/%d", readyContainers, totalContainers)
	result.Restarts = int(restarts)
	result.Age = translateTimestampSince(pod.CreationTimestamp)

	return result

}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}
