# Kubernetes Network: Pods

[This blog](https://medium.com/google-cloud/understanding-kubernetes-networking-pods-7117dd28727) already examines the networking of Pods in K8s. But I think the info in this blog is somehow out of date as GKE is no longer uses Docker bridge networking to manage the network for Pods. In this blog I am going to try my best to explore the networking of Pods in GKE.

The following picture shows the overview of Pod networking

[image]

From the picture, you can see that:

- Each node gets an internal IP from the IP range `10.128.0.0/`
- Each node is assigned a subnet (CIDR block) of the IP range `10.36.0.0/14`, such as `10.36.0.0/24` or `10.36.1.0/24`. This means each Pod running in the same node will be in the same CIDR block.
- GKE uses [Calico](https://docs.projectcalico.org/v3.0/introduction/) as its Container Network Interface (CNI) plugin to manage the network for the containers. Calico creates a veth pair for each Pod. One side is inserted into the Pod's network space while another side gets directly exposed to the node's network. Additionally, all the containers in the same Pod share the veth pair. That is why they can talk to each other through `localhost`. For example, the `busybox` container can talk to the `simple-http-server` container by `curl http://localhost:80`.

Reference:

- [Calico](https://docs.projectcalico.org/v3.0/introduction/)
