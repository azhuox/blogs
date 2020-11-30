# Kubernetes Overview

## What is Kubernetes
Kubernetes is a portable, extensible system for running and coordinating containerized workloads and services across a cluster of machines. 
It is designed to manage the life cycle of containerized applications with the guarantee of stability, scalability and high availability.

## Topology of Kubernetes
The following picture shows a Kubernetes cluster running on [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine):

![Kubernetes Overview](https://github.com/azhuox/blogs/blob/master/kubernetes/overview/assets/k8s-overview.png?raw=true)

From the picture, you can see that:

* A Kubernetes cluster has one or more master nodes (for highly availability purpose). A master node consists of the following core components:

    - `etcd` is used as backing store for all the Kubernetes’s cluster data.

    - `controller manager` is in charge of maintaining the state of the cluster, such as rolling updates, fault detection, auto scaling, etc.

    - `scheduler` is responsible for scheduling Pods to proper worker nodes according to a predetermined scheduling policy.

    - `apiserver` provides a uniform entry for authentication, authorization, API registration and resource operations. 
    As a developer, this is the only component you need to interactive with in a Kubernetes master. 
    When you run a Kubernetes CLI (`kubectl`) command in a terminal, such as `kubectl get pods`, what happens is `kubectl` converts the command to an API request, 
    sends it to `apiserver` and displays the results in the terminal. 

* A Kubernetes cluster has one or more worker nodes that are running in multiple zones. Each worker node is a [Google Compute Engine (GCE)](https://cloud.google.com/compute) virtual machine and has the following core components:

    - `kubelet` is responsible for maintaining life cycle of Pods and containers.

    - `kube-proxy` provides service discovery and load balancing functionality for Kubernetes Services.

There are actually more components in a Kubernetes cluster. But as a developer, you do not need to dive deeply to be familiar with any of these components. It is enough for you to know that a Kubernetes cluster is a distributed system that consists of multiple machines running across multiple zones.

## Kubernetes Objects and Workloads
While Kubernetes utilizes containers ([docker containers](https://www.docker.com/resources/what-container) as default for GKE clusters) 
as the underlying mechanism to deploy & run applications, it builds additional layers of abstraction on the top of the container interface to 
make it easier for developers to containerize their applications. Instead of managing and interacting with containers directly, 
developers use objects and workloads that Kubernetes provides to construct their applications.

Here are the main objects and workloads that Kubernetes provides:

* [Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/): A Namespace is considered a virtual cluster in a Kubernetes cluster. It is used to separate projects and teams. 
In our system, each of our applications (a.k.a microservices) has two Namespaces, one for the demo environment and another one for the prod environment.

* [ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/): A ConfigMap is a Kubernetes object for storing non-confidential data in key-value pairs. 
It allows you to decouple environment-specific configuration from containers in your application, which makes your application easily portable. 
A Pod can consume data in a ConfigMap as environment variables or mount a ConfigMap as a Volume. 

* [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/): A Secret is a Kubernetes object for storing sensitive configuration, such as passwords, API keys and OAuth tokens, in key-value pairs.

* [Pods](https://kubernetes.io/docs/concepts/workloads/pods/): A Pod is the smallest deployable unit that you can create in Kubernetes. 
A Pod consists of one or more containers. These containers have their own CPU & memory resources but need to share other computing resources, including storage and network. 
When a Pod is created, it will be scheduled to a proper Kubernetes worker node based on its predetermined policy. 
You can consider a Pod a virtual machine running in a Kubernetes cluster. We will take about this in the next chapter.

* [ReplicaSets](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/): A ReplicaSet is used to maintain a stable set of replica Pods running at the same state at any given time. 
It is also used to guarantee the availability of a specified number of replica Pods.

* [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) & [Stateful Sets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/): A Deployment provides declarative updates, such as rolling updates and rolling back, for Pod ReplicaSets.  
In other words, A Deployment ensures a group of identical Pods to achieve a desired state at any time. StatefulSets provides the same functionality. 
However, Deployments are used to manage stateless applications while StatefulSets are used to manage stateful applications.

* [Services](https://kubernetes.io/docs/concepts/services-networking/service/): A Service is an object that exposes an application running as a set of Pods as a network service. A Service is responsible for load-balancing requests to your application’s Pods.

## Kubernetes CLI Command

When working with Kubernetes, you need to know how to utilize its CLI tool `kubectl` to interact with Kubernetes clusters. Here are some frequently used `kubectl` commands that you may use in your daily work.

* Connecting a GKE cluster: `gcloud container clusters get-credentials vendasta-central --region us-central1 --project repcore-prod` is the command for connecting to our major GKE cluster called vendasta-central. You only need to run this command once.

* Switching current namespace: `kubectl config set-context --current --namespace=<your_namespace>` switches your current context (Namespace) to the given Namespace. This command is useful when you need to switch between demo and prod environment for a microservice.

* Create one or more Kubernetes objects: `kubectl apply -f <a_yaml_or_json_file>` can help you create one or more Kubernetes objects. The given file must be yaml or json formatted. You can put the specification of many Kubernetes objects into a single file as long as each object is separated with ---.

* Get Kubernetes objects: `kubectl get <object_type>` can help you get certain type of Kubernetes objects. Here are two examples of this command:

    - `kubectl get pods` lists all the Kubernetes Pods in your current Namespace. You can specify a Namespace by using `-n <your_namespace>` parameter.

    - `kubectl get pod <your_pod> -o yaml` prints out details of given Pod in yaml format. 

## What Is Next

Check this blog if you are curious about the minimal deploy-able unit in Kubernetes: [Kubernetes Pods](https://github.com/azhuox/blogs/blob/master/kubernetes/pods/README.md).

## Reference:

- [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine)
- [Google Compute Engine (GCE)](https://cloud.google.com/compute)
- [Docker Containers](https://www.docker.com/resources/what-container)
- [Kubernetes Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
- [Kubernetes ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Kubernetes Pods](https://kubernetes.io/docs/concepts/workloads/pods/)
- [Kubernetes ReplicaSets](https://kubernetes.io/docs/concepts/workloads/controllers/replicaset/)
- [Kubernetes Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Kubernetes Stateful Sets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
- [Kubernetes Services](https://kubernetes.io/docs/concepts/services-networking/service/)
