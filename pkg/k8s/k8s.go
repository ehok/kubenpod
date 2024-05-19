package k8s

import (
	"context"
	"fmt"

	"github.com/manifoldco/promptui"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetNodeName(podName string, args []string, clientset *kubernetes.Clientset) string {
	if podName != "" {
		return GetNodeNameFromPod(podName, clientset)
	}
	return SelectNode(args, clientset)
}

func GetNodeNameFromPod(podName string, clientset *kubernetes.Clientset) string {
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

func SelectNode(args []string, clientset *kubernetes.Clientset) string {
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

func ListPod(node string, clientset *kubernetes.Clientset) {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node,
	})
	if err != nil {
		fmt.Println("Error fetching pods:", err)
		return
	}

	var maxPodNameLen = len("Pod")
	var maxNamespaceLen = len("Namespace")

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
