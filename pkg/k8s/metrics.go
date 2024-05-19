package k8s

import (
	"context"
	"fmt"
	"math"

	"github.com/fatih/color"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func TopPod(node string, clientset *kubernetes.Clientset, metricsClient *versioned.Clientset) {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error fetching pods:", err)
		return
	}

	podNodeMap := make(map[string]string)
	maxNamespaceLen, maxPodNameLen, maxContainerNameLen := len("Namespace"), len("Pod"), len("Container")
	memoryUsages := []float64{}

	for _, pod := range pods.Items {
		podNodeMap[pod.Namespace+"/"+pod.Name] = pod.Spec.NodeName
		if len(pod.Namespace) > maxNamespaceLen {
			maxNamespaceLen = len(pod.Namespace)
		}
		if len(pod.Name) > maxPodNameLen {
			maxPodNameLen = len(pod.Name)
		}
		for _, container := range pod.Spec.Containers {
			if len(container.Name) > maxContainerNameLen {
				maxContainerNameLen = len(container.Name)
			}
		}
	}

	formatHeader := fmt.Sprintf("%%-%ds %%-%ds %%-%ds %%-8s %%-8s\n", maxNamespaceLen, maxPodNameLen, maxContainerNameLen)
	fmt.Printf(formatHeader, "Namespace", "Pod", "Container", "CPU", "Memory")

	metrics, err := metricsClient.MetricsV1beta1().PodMetricses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error fetching pod metrics:", err)
		return
	}

	for _, item := range metrics.Items {
		if podNodeMap[item.Namespace+"/"+item.Name] == node {
			for _, container := range item.Containers {
				memoryMi := float64(container.Usage.Memory().Value() / 1024 / 1024)
				memoryUsages = append(memoryUsages, memoryMi)
			}
		}
	}

	mean, stdDev := CalculateStats(memoryUsages)

	for _, item := range metrics.Items {
		if podNodeMap[item.Namespace+"/"+item.Name] == node {
			for _, container := range item.Containers {
				cpu := fmt.Sprintf("%vm", container.Usage.Cpu().MilliValue())
				memoryMi := float64(container.Usage.Memory().Value() / 1024 / 1024)
				memoryStr := fmt.Sprintf("%vMi", memoryMi)
				if ZScore(memoryMi, mean, stdDev) > 1.5 {
					memoryStr = color.RedString(memoryStr)
				}
				fmt.Printf(formatHeader, item.Namespace, item.Name, container.Name, cpu, memoryStr)
			}
		}
	}
}

func CalculateStats(data []float64) (mean float64, stdDev float64) {
	sum := 0.0
	for _, value := range data {
		sum += value
	}
	mean = sum / float64(len(data))

	totalSqDiff := 0.0
	for _, value := range data {
		totalSqDiff += math.Pow(value-mean, 2)
	}
	stdDev = math.Sqrt(totalSqDiff / float64(len(data)))

	return mean, stdDev
}

func ZScore(value, mean, stdDev float64) float64 {
	return (value - mean) / stdDev
}
