package main

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	var rootCmd = &cobra.Command{Use: "kubenpod"}

	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		fmt.Println("Error building kubeconfig:", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Error creating Kubernetes client:", err)
		os.Exit(1)
	}

	metricsClient, err := versioned.NewForConfig(config)
	if err != nil {
		fmt.Println("Error creating metrics client:", err)
		os.Exit(1)
	}

	var cmdTop = &cobra.Command{
		Use:   "top [NODE_NAME]",
		Short: "Show metrics of all pods on the NODE_NAME",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			topPod(args[0], clientset, metricsClient)
		},
	}

	var cmdList = &cobra.Command{
		Use:   "list [NODE_NAME]",
		Short: "List all pods on the NODE_NAME",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			listPod(args[0], clientset)
		},
	}

	rootCmd.AddCommand(cmdTop, cmdList)
	rootCmd.Execute()
}

func topPod(node string, clientset *kubernetes.Clientset, metricsClient *versioned.Clientset) {
	// Fetch all pods across all namespaces
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error fetching pods:", err)
		return
	}

	// Build a map to relate each pod to its node
	podNodeMap := make(map[string]string)
	// maxNamespaceLen := len("Namespace")
	// maxPodNameLen := len("Pod")
	// maxContainerNameLen := len("Container")
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

	// Define the format for output
	formatHeader := fmt.Sprintf("%%-%ds %%-%ds %%-%ds %%-8s %%-8s\n", maxNamespaceLen, maxPodNameLen, maxContainerNameLen)
	fmt.Printf(formatHeader, "Namespace", "Pod", "Container", "CPU", "Memory")

	// Fetch metrics for all pods across all namespaces
	metrics, err := metricsClient.MetricsV1beta1().PodMetricses("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Error fetching pod metrics:", err)
		return
	}

	for _, item := range metrics.Items {
		if podNodeMap[item.Namespace+"/"+item.Name] == node { // Ensure we only print metrics for the specified node
			for _, container := range item.Containers {
				// cpu := fmt.Sprintf("%vm", container.Usage.Cpu().MilliValue())
				memoryMi := float64(container.Usage.Memory().Value() / 1024 / 1024)
				memoryUsages = append(memoryUsages, memoryMi)
				// memoryStr := fmt.Sprintf("%vMi", memoryMi)
				// if memoryMi > 300 { // Example threshold for highlighting
				// 	memoryStr = color.RedString(memoryStr)
				// }
				// fmt.Printf(formatHeader, item.Namespace, item.Name, container.Name, cpu, memoryStr)
			}
		}
	}

	mean, stdDev := calculateStats(memoryUsages)

	// Print each pod metric with dynamic highlighting based on z-score
	for _, item := range metrics.Items {
		if podNodeMap[item.Namespace+"/"+item.Name] == node {
			for _, container := range item.Containers {
				cpu := fmt.Sprintf("%vm", container.Usage.Cpu().MilliValue())
				memoryMi := float64(container.Usage.Memory().Value() / 1024 / 1024)
				memoryStr := fmt.Sprintf("%vMi", memoryMi)
				if zScore(memoryMi, mean, stdDev) > 1.5 { // Current z-score threshold is 1.5 for highlighting
					memoryStr = color.RedString(memoryStr)
				}
				fmt.Printf(formatHeader, item.Namespace, item.Name, container.Name, cpu, memoryStr)
			}
		}
	}
}

func calculateStats(data []float64) (mean float64, stdDev float64) {
	sum := 0.0
	for _, value := range data {
		sum += value
	}
	mean = sum / float64(len(data)) // Find the mean

	totalSqDiff := 0.0 // total squared difference
	for _, value := range data {
		totalSqDiff += math.Pow(value-mean, 2)
	}
	stdDev = math.Sqrt(totalSqDiff / float64(len(data)))

	return mean, stdDev
}

func zScore(value, mean, stdDev float64) float64 {
	return (value - mean) / stdDev
}

func listPod(node string, clientset *kubernetes.Clientset) {
	// Fetch all pods
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node,
	})
	if err != nil {
		fmt.Println("Error fetching pods:", err)
		return
	}

	var maxPodNameLen int = len("Pod")
	var maxNamespaceLen int = len("Namespace")

	// Determine the maximum string lengths for pod names and namespaces
	for _, pod := range pods.Items {
		if len(pod.Name) > maxPodNameLen {
			maxPodNameLen = len(pod.Name)
		}
		if len(pod.Namespace) > maxNamespaceLen {
			maxNamespaceLen = len(pod.Namespace)
		}
	}

	// Formatting string based on the max lengths
	podFormat := fmt.Sprintf("%%-%ds %%-%ds\n", maxNamespaceLen+2, maxPodNameLen)
	fmt.Printf(podFormat, "Namespace", "Pod")

	// Print each pod's namespace and name with dynamic formatting
	for _, pod := range pods.Items {
		fmt.Printf(podFormat, pod.Namespace, pod.Name)
	}

}
