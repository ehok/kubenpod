# Kubenpod
> Note that everything is experimental and may change significantly at any time.

This tool aims to make it easier to list the pods that work in node on Kubernetes Cluster.
This can be done with a node name or a pod name as you can see at examples below.
If you encounter any problem or complexity, you can post it here as a comment.

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
$ kubenpod -l ip-10-0-0-1.us-east-1.compute.internal

hoke              hoke-deployment-cda628x58-p1vzz           1/1   Running     0      6d3h
kube-system       aws-node-r6n1m                            1/1   Running     0      58d
kube-system       coredns-165c7a19dd-st08z                  1/1   Running     0      56d
kube-system       kube-proxy-ty1m6                          1/1   Running     0      26d
kube-system       spot-interrupt-handler-f7syd              1/1   Running     0      5d22h
monitoring        prometheus-node-exporter-d91th            1/1   Running     0      58d
nginx-ingress     nginx-ingress-controller-vqf8q            1/1   Running     0      58d
```
```
$ kubenpod -pl hoke-deployment-cda628x58-p1vzz

hoke              hoke-deployment-cda628x58-p1vzz           1/1   Running     0      6d3h
kube-system       aws-node-r6n1m                            1/1   Running     0      58d
kube-system       coredns-165c7a19dd-st08z                  1/1   Running     0      56d
kube-system       kube-proxy-ty1m6                          1/1   Running     0      26d
kube-system       spot-interrupt-handler-f7syd              1/1   Running     0      5d22h
monitoring        prometheus-node-exporter-d91th            1/1   Running     0      58d
nginx-ingress     nginx-ingress-controller-vqf8q            1/1   Running     0      58d
```
```
$ kubenpod --help

USAGE:
  kubenpod <NODE_NAME>                 : show metrics of all pods of the <NODE_NAME>
  kubenpod -p <POD_NAME>               : show metrics of pods of the <POD_NAME>'s node 
  kubenpod                             : show this message
  kubenpod -h,--help                   : show this message
  kubenpod -l,--list <NODE_NAME>       : list all pods of the <NODE_NAME>
  kubenpod -lp,--listpod <POD_NAME>    : list all pods of the <POD_NAME>'s node 
```

## Special Thanks To

Mert Öngengil [@mertongngl](http://github.com/mertongngl)
