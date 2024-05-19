package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func PromptForService(clientset *kubernetes.Clientset) (serviceName string, namespace string) {
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

func FindServiceNamespace(serviceName string, clientset *kubernetes.Clientset) string {
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

func ShowServiceDetailsAndRelatedResources(serviceName, namespace string, clientset *kubernetes.Clientset) {
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
