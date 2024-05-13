package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

func main() {
	const cliVersion = "1.2.0"
	var podName, namespace string
	var rootCmd = &cobra.Command{Use: "kubenpod"}
	rootCmd.PersistentFlags().StringVarP(&podName, "pod-name", "p", "", "Specify the name of the pod to focus on a specific node")

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

	var cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of kubenpod and the Kubernetes server",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kubenpod version %s\n", cliVersion)
			version, err := clientset.Discovery().ServerVersion()
			if err != nil {
				fmt.Printf("Error fetching Kubernetes server version: %s\n", err)
				return
			}
			fmt.Printf("Kubernetes server version: %s\n", version.GitVersion)
		},
	}

	var cmdTop = &cobra.Command{
		Use:   "top [NODE_NAME]",
		Short: "Show metrics of all pods on the NODE_NAME or the node of a specified pod",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nodeName := ""
			if podName != "" {
				nodeName = getNodeNameFromPod(podName, clientset)
				if nodeName == "" {
					fmt.Println("Specified pod not found or failed to retrieve node.")
					return
				}
			} else {
				nodeName = getNodeName(args, clientset)
				if nodeName == "" {
					fmt.Println("No node selected.")
					return
				}
			}
			topPod(nodeName, clientset, metricsClient)
		},
	}

	var cmdList = &cobra.Command{
		Use:   "list [NODE_NAME]",
		Short: "List all pods on the NODE_NAME",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nodeName := ""
			if podName != "" {
				nodeName = getNodeNameFromPod(podName, clientset)
				if nodeName == "" {
					fmt.Println("Specified pod not found or failed to retrieve node.")
					return
				}
			} else {
				nodeName = getNodeName(args, clientset)
				if nodeName == "" {
					fmt.Println("No node selected.")
					return
				}
			}
			listPod(nodeName, clientset)
		},
	}

	var cmdServiceResources = &cobra.Command{
		Use:   "service-resources [SERVICE_NAME]",
		Short: "Show detailed information about a service and related resources like deployments and pods",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var serviceName string
			if len(args) == 0 {
				serviceName, namespace = promptForService(clientset)
				if serviceName == "" {
					fmt.Println("No service selected.")
					return
				}
			} else {
				serviceName = args[0]
				if namespace == "" {
					namespace = findServiceNamespace(serviceName, clientset)
				}
			}
			showServiceDetailsAndRelatedResources(serviceName, namespace, clientset)
		},
	}
	cmdServiceResources.Flags().StringVarP(&namespace, "namespace", "n", "", "Specify the namespace of the service")

	rootCmd.AddCommand(cmdTop, cmdList, cmdVersion, cmdServiceResources)
	rootCmd.Execute()
}

func getNodeNameFromPod(podName string, clientset *kubernetes.Clientset) string {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", podName),
	})
	if err != nil {
		fmt.Printf("Error retrieving pods: %v\n", err)
		return ""
	}
	if len(pods.Items) == 0 {
		fmt.Println("No pods found with the specified name.")
		return ""
	}

	nodeName := pods.Items[0].Spec.NodeName
	fmt.Printf("Node name retrieved from pod '%s': %s\n", podName, nodeName)
	return nodeName
}

func getNodeName(args []string, clientset *kubernetes.Clientset) string {
	if len(args) > 0 && args[0] != "" {
		return args[0]
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("Failed to fetch nodes:", err)
		return ""
	}

	nodeNames := []string{}
	for _, node := range nodes.Items {
		nodeNames = append(nodeNames, node.Name)
	}

	prompt := promptui.Select{
		Label: "Select Node",
		Items: nodeNames,
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Println("Prompt failed:", err)
		return ""
	}

	return result
}

func topPod(node string, clientset *kubernetes.Clientset, metricsClient *versioned.Clientset) {
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
				// memoryStr := fmt.Sprintf("%vMi", memoryMi)
				// if memoryMi > 300 { // Example threshold for highlighting
				// 	memoryStr = color.RedString(memoryStr)
				// }
				// fmt.Printf(formatHeader, item.Namespace, item.Name, container.Name, cpu, memoryStr)
			}
		}
	}

	mean, stdDev := calculateStats(memoryUsages)

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

	totalSqDiff := 0.0 // Total squared difference
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
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node,
	})
	if err != nil {
		fmt.Println("Error fetching pods:", err)
		return
	}

	var maxPodNameLen int = len("Pod")
	var maxNamespaceLen int = len("Namespace")

	for _, pod := range pods.Items {
		if len(pod.Name) > maxPodNameLen {
			maxPodNameLen = len(pod.Name)
		}
		if len(pod.Namespace) > maxNamespaceLen {
			maxNamespaceLen = len(pod.Namespace)
		}
	}

	podFormat := fmt.Sprintf("%%-%ds %%-%ds\n", maxNamespaceLen+2, maxPodNameLen)
	fmt.Printf(podFormat, "Namespace", "Pod")

	for _, pod := range pods.Items {
		fmt.Printf(podFormat, pod.Namespace, pod.Name)
	}
}

func promptForService(clientset *kubernetes.Clientset) (string, string) {
	svcList, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing services across all namespaces: %v\n", err)
		return "", ""
	}

	var services []string
	var serviceMap = make(map[string]string)
	for _, svc := range svcList.Items {
		serviceIdentifier := fmt.Sprintf("%s (%s)", svc.Name, svc.Namespace)
		services = append(services, serviceIdentifier)
		serviceMap[serviceIdentifier] = fmt.Sprintf("%s %s", svc.Name, svc.Namespace)
	}

	if len(services) == 0 {
		fmt.Println("No services available to select.")
		return "", ""
	}

	prompt := promptui.Select{
		Label: "Select Service",
		Items: services,
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed: %v\n", err)
		return "", ""
	}

	parts := strings.Split(serviceMap[result], " ")
	return parts[0], parts[1]
}

func findServiceNamespace(serviceName string, clientset *kubernetes.Clientset) string {
	svcList, err := clientset.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing services across all namespaces: %v\n", err)
		return ""
	}
	for _, svc := range svcList.Items {
		if svc.Name == serviceName {
			return svc.Namespace
		}
	}
	return ""
}

func showServiceDetailsAndRelatedResources(serviceName, namespace string, clientset *kubernetes.Clientset) {
	if namespace == "" {
		fmt.Printf("Service '%s' not found in any namespace or namespace not specified.\n", serviceName)
		return
	}
	service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error fetching service '%s' in namespace '%s': %v\n", serviceName, namespace, err)
		return
	}

	fmt.Printf("Service Name: %s in Namespace: %s\n", service.Name, namespace)
	fmt.Printf("Type: %s\n", service.Spec.Type)
	fmt.Printf("Cluster IP: %s\n", service.Spec.ClusterIP)
	fmt.Println("Ports:")
	for _, port := range service.Spec.Ports {
		fmt.Printf("- Port: %d -> TargetPort: %s Protocol: %s\n", port.Port, port.TargetPort.String(), port.Protocol)
	}

	selector := labels.SelectorFromSet(service.Spec.Selector).String()

	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err == nil && len(deployments.Items) > 0 {
		fmt.Println("Related Deployments:")
		for _, d := range deployments.Items {
			fmt.Printf("- %s\n", d.Name)
		}
	}

	replicaSets, err := clientset.AppsV1().ReplicaSets(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err == nil && len(replicaSets.Items) > 0 {
		fmt.Println("Related ReplicaSets:")
		for _, rs := range replicaSets.Items {
			fmt.Printf("- %s\n", rs.Name)
		}
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err == nil && len(pods.Items) > 0 {
		fmt.Println("Related Pods:")
		for _, p := range pods.Items {
			fmt.Printf("- %s\n", p.Name)
		}
	}
}
