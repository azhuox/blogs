# K8s Network Overview

The documents in this folder are meant to explain network in Kubernetes, including pod network, service network and ingress network.

Please note that these docs now are still in the `draft` state. I will try finishing them when I have enough time...

## Example: Topology of Demo GKE Cluster

THe following picture shows the the topology of a [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine/) cluster which is used to demonstrate K8s network.

Explanation:

- The cluster is a regional (us-central1) GKE cluster;
- The cluster has a two-nodes node pool, one in us-central1-a while another one is in us-central1-b
- The node IP address range (CIDR block) is `10.128.0.0/20` while the gateway is `10.128.0.1`. This means every node in the cluster will get an IP from this IP range and **all the network events inside the GKE cluster are essentially translated to the events in this IP range.**
- The Pod IP address range (CIDR block) is `10.36.0.0/14`. This means every Pod in the cluster will get an IP from this range.
- The Pod IP address range (CIDR block) is `10.40.0.0/20`. This means every K8s Service in the cluster will get an IP (aka. Cluster IP or Internal IP) from this IP range.

[image]  

## Next

[This doc](https://github.com/azhuox/blogs/blob/master/kubernetes/network/pod.md) explains how the network is "constructed" inside a K8s node.

[This blog](https://github.com/azhuox/blogs/blob/master/kubernetes/network/int-svc-network.md) explains how Pods communicate with each other inside the cluster.

Reference:
- [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/)
