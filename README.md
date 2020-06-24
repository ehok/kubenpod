# Kubenpod
> Note that everything is experimental and may change significantly at any time.

This repository collects pods that are running on specified node's.

## Installation
```
sudo curl -LO https://raw.githubusercontent.com/engincanhoke/toppod/master/kubenpod > kubenpod;
sudo chmod +x kubenpod;
sudo mv kubenpod /usr/local/bin/kubenpod
```
## Usage
```
$ kubenpod ip-10-0-0-1.us-east-1.compute.internal

hoke              hoke-deployment-cda628x58-p1vzz                                   176m   267Mi
kube-system       aws-node-r6n1m                                                    3m     28Mi
kube-system       coredns-165c7a19dd-st08z                                          3m     20Mi
kube-system       kube-proxy-ty1m6                                                  1m     14Mi
kube-system       spot-interrupt-handler-f7syd                                      2m     4Mi
monitoring        prometheus-node-exporter-d91th                                    2m     9Mi
nginx-ingress     nginx-ingress-controller-vqf8q                                    36m    151Mi
```
```
$ kubenpod -p hoke-deployment-cda628x58-p1vzz

kubenpod ip-10-0-0-1.us-east-1.compute.internal
hoke              hoke-deployment-cda628x58-p1vzz                                   176m   267Mi
kube-system       aws-node-r6n1m                                                    3m     28Mi
kube-system       coredns-165c7a19dd-st08z                                          3m     20Mi
kube-system       kube-proxy-ty1m6                                                  1m     14Mi
kube-system       spot-interrupt-handler-f7syd                                      2m     4Mi
monitoring        prometheus-node-exporter-d91th                                    2m     9Mi
nginx-ingress     nginx-ingress-controller-vqf8q                                    36m    151Mi
```
```
$ kubenpod --help

USAGE:
  kubenpod <NODE_NAME>          : list all pods of the <NODE_NAME>
  kubenpod -p <POD_NAME>        : list all pods of the <POD_NAME>'s node 
  kubenpod                      : show this message
  kubenpod -h,--help            : show this message
```
