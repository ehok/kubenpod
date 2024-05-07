# KubenPod CLI Tool
> Note that everything is experimental and may change significantly at any time.

- KubenPod is a command-line interface tool for Kubernetes that aims to make it easier to list the pods that work in node on Kubernetes Cluster. This tool uses the Kubernetes and Metrics APIs to fetch data about pods and their resource usage.
- This can be done with a node name or a pod name as you can see at examples below.
If you encounter any problem or complexity, open an issue.

## Installation
Before using the KubenPod CLI, ensure you have configured your Kubernetes context and have appropriate permissions to access the resources.
```
sudo curl -LO https://raw.githubusercontent.com/engincanhoke/toppod/master/kubenpod > kubenpod;
sudo chmod +x kubenpod;
sudo mv kubenpod /usr/local/bin/kubenpod
```

## Usage
The KubenPod CLI provides two main commands:
1. `top` This command is used to show resource metrics (CPU and memory) of all pods on a specified node or the node of a specified pod.
```bash
kubenpod top [NODE_NAME]
kubenpod top --pod-name [POD_NAME]
```
2. `list` This command lists all pods on a specified node or the node of a specified pod.
```bash
kubenpod list [NODE_NAME]
kubenpod list --pod-name [POD_NAME]
```

### `top` Command
To display the metrics of all pods on a specific node:
```
$ kubenpod top nodename123
```
To display the metrics of all pods on the node where a specific pod is running:
```
$ kubenpod top --pod-name mypodname
```

### `list` Command
To list all pods on a specific node:
```
$ kubenpod list nodename123
```
To list all pods on the node where a specific pod is running:
```
$ kubenpod list --pod-name mypodname
```

### `version` Command
To display version information for CLI & Kubernetes server:
```
$ kubenpod version
```

## Additional Notes
- Ensure that your Kubernetes context is correctly set. `kubenpod` uses the default kubeconfig file path for authentication.
- The commands require at least cluster-level read permissions to fetch the node and pod metrics.

## Contributing
Contributions to improve Kubenpod are welcome. Please feel free to fork the repository, make changes, and submit a pull request.

## License
This project is licensed under the MIT License - see the LICENSE file for details.

## Special Thanks To
Mert Öngengil [@mertongngl](http://github.com/mertongngl)