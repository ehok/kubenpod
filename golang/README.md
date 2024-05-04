# Kubenpod CLI Tool

Kubenpod is a command-line interface (CLI) tool built in Go, designed to interact with a Kubernetes cluster to display information about pods. It can list all pods on a specified node and show detailed metrics such as CPU and memory usage for these pods.

## Features

- **List Pods**: List all pods running on a specified Kubernetes node.
- **Top Pods**: Show detailed metrics like CPU and memory usage for all pods on a specified node, with dynamic highlighting for pods using resources above a certain threshold.

## Prerequisites

Before you run the Kubenpod tool, make sure you have the following installed:
- Go (version 1.15 or higher)
- Access to a Kubernetes cluster
- Kubernetes configuration file set up (usually found at `~/.kube/config`)

## Installation

To use the Kubenpod tool, clone the repository and build the application:

```bash
git clone https://github.com/your-repository/kubenpod.git
cd kubenpod
go build -o kubenpod
```

## Usage

### List Pods

To list all pods on a specific node:

```bash
./kubenpod list <NODE_NAME>
```

This command displays a list of all pods running on the specified node `<NODE_NAME>`, showing the namespace and pod name.

### Show Top Pods

To display metrics for all pods on a specific node:

```bash
./kubenpod top <NODE_NAME>
```

This command provides detailed metrics for each pod on the node, including CPU and memory usage. It highlights pods with memory usage that significantly deviates from the average based on a calculated z-score.

## How It Works

- **Kubernetes Client**: Connects to your Kubernetes cluster using the kubeconfig file to fetch data.
- **Metrics Client**: Gathers metrics from the Kubernetes Metrics Server for pods, calculating average usage and deviations.
- **Commands**:
  - `list`: Fetches pods from the Kubernetes API and formats the output.
  - `top`: Fetches metrics for pods, calculates statistical thresholds, and formats the output with conditional coloring for high usage.

## Contributing

Contributions to improve Kubenpod are welcome. Please feel free to fork the repository, make changes, and submit a pull request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.