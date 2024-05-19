package root

import (
	"fmt"
	"os"

	"github.com/ehok/kubenpod/pkg/k8s"
	"github.com/ehok/kubenpod/version"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	podName       string
	namespace     string
	clientset     *kubernetes.Clientset
	metricsClient *versioned.Clientset
)

func init() {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		fmt.Println("Error building kubeconfig:", err)
		os.Exit(1)
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Error creating Kubernetes client:", err)
		os.Exit(1)
	}

	metricsClient, err = versioned.NewForConfig(config)
	if err != nil {
		fmt.Println("Error creating metrics client:", err)
		os.Exit(1)
	}

	cmdServiceResources.Flags().StringVarP(&namespace, "namespace", "n", "", "Specify the namespace of the service")

	rootCmd.PersistentFlags().StringVarP(&podName, "pod-name", "p", "", "Specify the name of the pod to focus on a specific node")
}

var rootCmd = &cobra.Command{Use: "kubenpod"}

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of kubenpod and the Kubernetes server",
	Run: func(cmd *cobra.Command, args []string) {
		version.PrintVersion()
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
		nodeName := k8s.GetNodeName(podName, args, clientset)
		if nodeName == "" {
			fmt.Println("No node selected.")
			return
		}
		k8s.TopPod(nodeName, clientset, metricsClient)
	},
}

var cmdList = &cobra.Command{
	Use:   "list [NODE_NAME]",
	Short: "List all pods on the NODE_NAME",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeName := k8s.GetNodeName(podName, args, clientset)
		if nodeName == "" {
			fmt.Println("No node selected.")
			return
		}
		k8s.ListPod(nodeName, clientset)
	},
}

var cmdServiceResources = &cobra.Command{
	Use:   "service-resources [SERVICE_NAME]",
	Short: "Show detailed information about a service and related resources like deployments and pods",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var serviceName string
		if len(args) == 0 {
			serviceName, namespace = k8s.PromptForService(clientset)
			if serviceName == "" {
				fmt.Println("No service selected.")
				return
			}
		} else {
			serviceName = args[0]
			if namespace == "" {
				namespace = k8s.FindServiceNamespace(serviceName, clientset)
			}
		}
		k8s.ShowServiceDetailsAndRelatedResources(serviceName, namespace, clientset)
	},
}

func Execute() {
	rootCmd.AddCommand(cmdTop, cmdList, cmdVersion, cmdServiceResources)
	// rootCmd.Execute()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
