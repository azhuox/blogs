# Kubernetes StatefulSets

## prerequisites
I recommend you know the basic knowledge Kubernetes Pods before reading this blog. You can check [this doc](https://kubernetes.io/docs/concepts/workloads/pods/pod/) for details about Kubernetes Pods.

## What Is A StatefulSet

StatefulSet is a Kubernetes object designed to manage stateful applications. 
Like a Deployment, a StatefulSet scales up a set of pods to a desired number that you define in a config file. 
Pods in a StatefulSet runs the same containers defined in the `Pod spec` inside the `StatefulSet spec`.  
Unlike a Deployment, every Pod of a StatefulSet owns a sticky and stable identity. 
A StatefulSet also provides the guarantee about ordered deployment, deletion, scaling and rolling updates for its Pods.

## A StatefulSet Example
A complete StatefulSet consists of two components:
- A Headless Service used to control the network ID of the StatefulSet's Pods
- A StatefulSet object used to create and manage its Pods. 

The following example demonstrates how to use a StatefulSet to create a ZooKeeper Server. 
Please note that the following `StatefulSet Spec` is simplified for demo purpose. 
You can check [this yaml file](https://github.com/kubernetes/contrib/blob/master/statefulsets/zookeeper/zookeeper.yaml "The StatefulSet Configuration for Setting Up A ZooKeeper Service") for the complete configuration of this ZooKeeper Service.

### ZooKeeper Service

A [Zookeeper service](https://zookeeper.apache.org/) is a distributed coordination system for distributed applications.  
It allows to you read, write data and observe data updates. 
Data is stored and replicated in each ZooKeeper server and these servers work together as a ZooKeeper Ensemble. 

The following picture shows the overview of a ZooKeeper service with five ZooKeeper servers. 
You can see each server in a ZooKeeper service has a stable network ID for potential leader elections. 
Moreover, one of the ZooKeeper servers needs to be selected as leader for managing the service topology and processing write requests. 
StatefulSets is suitable for running such an application as it guarantees uniqueness for Pods.

![A five-nodes Zookeeper Service](https://raw.githubusercontent.com/azhuox/blogs/master/kubernetes/statefulsets/assets/zookeeper-svc-in-statefulset.jpeg "A five-nodes Zookeeper Service")  
  
### Headless Service
A Headless Service is responsible for controlling the network domain for a StatefulSet. The way to create a headless service is to specify `clusterIP == None`. 

The following spec is for creating a Headless Service for the ZooKeeper service. 
This Headless Service is used to manage [Pod Identify](#pod-identify) for the following StatefulSet.

```yaml
apiVersion: v1
kind: Service
metadata:
  namespace: default
  name: zk-hs
  labels:
    app: zk-hs
spec:
  ports:
  - port: 2888
    name: server
  - port: 3888
    name: leader-election
  clusterIP: None
  selector:
    app: zk
```

Unlike a [`ClusterIP` Service](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types) or 
a [`LoadBalancer` Service](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer), a Headless Service does not provide load-balancing. 
Based on my experience, any request to `zk-hs.default.svc.cluster.local` is always redirected to the first StatefulSet Pod (`zk-0` in the example).
Therefore, A Kubernetes Service that provides load balancing or an Ingress is required if you need to load-balance traffic for your StatefulSet. 

### StatefulSet Spec

The following spec demonstrates how to use a StatefulSet to run a ZooKeeper service:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  namespace: default
  name: zk

# StatefulSet spec
spec:
  serviceName: zk-hs
  selector:
    matchLabels:
      app: zk # It has to match .spec.template.metadata.labels
  replicas: 5
  podManagementPolicy: OrderedReady
  updateStrategy:
    type: RollingUpdate

  # volumeClaimTemplates creates a Persistent Volume for each StatefulSet Pods.
  volumeClaimTemplates:
  - metadata:
      name: datadir
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: standard
      resources:
        requests:
          storage: 10Gi

  # Pod spec
  template:
    metadata:
      labels:
        app: zk
    spec:
      affinity:
        nodeAffinity:
          ...
        podAntiAffinity:
          ...
      # Containers running in each Pod
      containers:
      - name: k8szk
        image: gcr.io/google_samples/k8szk:v3
        ports:
        - containerPort: 2181
          name: client
        - containerPort: 2888
          name: server
        - containerPort: 3888
          name: leader-election
        env:
        ...
        readinessProbe:
          exec:
            command:
            - "zkOk.sh"
          initialDelaySeconds: 10
          timeoutSeconds: 5
        livenessProbe:
          ...
        volumeMounts:
        - name: datadir
          mountPath: /var/lib/zookeeper
```

#### Metadata

The `metadata` field contains metadata of this Stateful Set, which includes the name of this Stateful Set and the Namespace it belongs to. 
You can also put [labels and annotations](https://kubernetes.io/docs/reference/kubernetes-api/labels-annotations-taints/) in this field.

#### Stateful Set Spec and Pod Template

The `spec` field defines the specification of this Stateful Set and the `spec.template` defines a template for creating the Pods this Stateful Set manages.

#### Pod Selector

Like a Deployment, a StatefulSet uses the `spec.selctor` field to find which Pods to manage. 
You can check [this doc](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#labelselector-v1-meta) for details about the usage of Pod Selector.  

#### Replica

The field `spec.replica` specifies the desired number of Pods for the StatefulSet. It is recommended to run an odd number of Pods for some stateful applications 
like ZooKeepers, based on the consideration of the efficiency of some operations. 
For example, a ZooKeeper service marks a data write complete only when more than half of its servers send an acknowledgment back to the leader. 
Take a six pods ZooKeeper service as an example. The service remains available as long as at least four servers (ceil(6/2 + 1)) are available, 
which means you service can tolerate the failure of two servers. Nevertheless, it can still tolerate two-servers failure when the server number is lowered down to five. 
Meanwhile, this also improves write efficiency as now it only needs 3 servers' acknowledgment to complete a write request. 
Therefore, having the sixth server, in this case, does not give you any additional advantage in terms of write efficiency and server availability.

#### Pod Identify
A StatefulSet Pod is assigned a unique ID (aka. Pod Name) from its Headless Service when it is created. 
This ID sticks to the Pod during the life cycle of the StatefulSet. 
The pattern of constructing ID is `${statefulSetName}-${ordinal}`. For example, Kubernetes will create five Pods with five unique IDs `zk-0`, `zk-1`, `zk-2`, `zk-3` and `zk-4` for the above ZooKeeper service. 

The ID of a StatefulSet Pod is also its hostname. The sub domain takes the form 
`${podID}.${headlessServiceName}.{$namespace}.svc.cluster.local` where `cluster.local` is the cluster domain. 
For example, the sub domain of the first ZooKeeper Pod is `zk-0.zk-hs.default.svc.cluster.local`. 
It is recommended to use a Stateful Pod's sub domain other than its IP to reach the Pod as the sub domain is unique within the whole cluster.

#### podManagementPolicy

You can choose whether to create/update/delete a StatefulSet's Pod in order or in parallel 
by specifying `spec.podManagementPolicy == OrderedReady` or `spec.podManagementPolicy == Parallel`. 
`OrderedReady` is the default setting and it controls the Pods to be created with the order `0, 1, 2, ..., N` and to be deleted with the order `N, N-1, ..., 1, 0`. 
In addition, it has to wait for the current Pod to become Ready or terminated prior to terminating or launching the next Pod. 
`Parallel` launches or terminates all the Pods simultaneously. It does not rely on previous Pod's state to lunch or terminate the next Pod.

#### updateStrategy

There are several rolling update strategies available for StatefulSets. 
`RollingUpdate` is the default strategy and it deletes and recreates each Pod for a StatefulSet when a rolling update occurs.

Doing rolling updates for the stateful applications like ZooKeepers is a little bit tricky: 
Other Pods need enough time to elect a new leader when the StatefulSet Controller is recreating the leader. 
Therefore, You should consider configuring [`readnessProbe`](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/#define-readiness-probes) and 
`readnessProbe.initialDelaySeconds` for the containers inside a StatefulSet Pods to delay the new Pod to be ready, 
thus delaying the rolling update of the next Pod and giving other running Pods enough time to update the service topology.
 This should give your stateful applications, for example, a ZooKeeper service, enough time to handle the case where a Pod is lost and back.

#### Pod Affinity

Like a Deployment, the ideal scenario of running a StatefulSet is distribute its Pods to different nodes in different zones and 
avoid running multiple Pods in the same node. The  `spec.template.spec.affinity` field allows you to specify node affinity and inter-pod affinity (or anti-affinity) for the SatefulSet Pods. 
You can check [this doc](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/) for details about using node/pod affinity in Kubernetes

#### volumeClaimTemplates

The `spec.volumeClaimTemplates` filed is used to provide stable storage for StatefulSets. 
As shown in the following picture, the `spec.volumeClaimTemplates` field creates a Persistent Volume Claim (`datadir-zk-0`), a Persistent Volume (`pv-0000`), 
and a 10 GB standard persistent disk for Pod `zk-0`. These storage settings have the same life cycle with the StatefulSet, which means the storage for a Stateful Pod 
is stable and persistent. Any StatefulSet Pod will not lose its data whenever it is terminated and recreated.

![The Persistent Storage in the Zookeeper Service](https://raw.githubusercontent.com/azhuox/blogs/master/kubernetes/statefulsets/assets/pvs-zookeeper-service.jpeg)

## What is Next



# Reference
- [Pods in Kubernetes](https://kubernetes.io/docs/concepts/workloads/pods/pod/)
- [The StatefulSet Configuration for Setting Up A ZooKeeper Service](https://github.com/kubernetes/contrib/blob/master/statefulsets/zookeeper/zookeeper.yaml)
- [Official Website of ZooKeeper](https://zookeeper.apache.org/)
- [`ClusterIP` Services in Kubernetes](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types)
- [`LoadBalancer` Services in Kubernetes](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer)
- [LabelSelector in Kubernetes](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#labelselector-v1-meta)
- [Readiness Probe in Kubernetes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/#define-readiness-probes)
- [Assigning Pods to Nodes](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/)
