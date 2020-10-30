# Kubernetes Pods
What is included in this blog:
- An introduction of Kubernetes Pods
- Use of Pods
- The life cycle of a Pod

# Kubernetes Pods
A [Kubernetes (K8s) Pod](https://kubernetes.io/docs/concepts/workloads/pods/pod/) is the smallest deployable unit in K8s.
It is essentially is a group of one or more containers (e.g. Docker containers) which share the computing resources including storage and network
and each container can request to have its own CPU and memory resources.

As shown in the following picture, a Pod is very similar to an application specific "host" running one or more "processes" (containers). These "processes" work together to provides some kind of service.

![Hosts V.S. Pods](https://github.com/azhuox/blogs/blob/master/kubernetes/pods/assets/host_vs_pod.png?raw=true)

# Use of Pods
The following example demonstrates how to use a K8s Pod to construct a single-replica microservice. This Pod consists of three containers: the `user-usvc` container has all the microservice's business logic, the `cloudsql-proxy` container proxies all the MySQL requests to a Google Cloud SQL instance, while the `datadog-agent` container sends the logs to datadog.

```yaml

# ConfigMap for user-msvc
apiVersion: v1
kind: ConfigMap
metadata:
  name: user-msvc
data:
  mysql-host: 127.0.0.1:3306
---


# Pod for user-msvc
apiVersion: v1
kind: Pod
metadata:
  name: user-msvc
  labels:
    app: user-msvc
spec:
  # "Processes"
  containers:
  - name: user-msvc
    image: gcr.io/path/to/user-msvc:1.0.0
    ports:
    - name: secure-port
      containerPort: 443
      protocol: TCP
    env:
    - name: MYSQL_HOST
      valueFrom:
        configMapKeyRef:
          name: user-msvc
          key: mysql-host
    volumeMounts:
      - name: ndatadog-data
        mountPath: "/var/log/user-msvc"
    requests:
      cpu: 100m
      memory: 100Mi
    limits:
        cpu: 200m
        memory: 200Mi

  - name: cloudsql-proxy
    image: gcr.io/cloudsql-docker/gce-proxy:1.12
    ports:
    - name: connect-port
      containerPort: 3306
      protocol: TCP
    requests:
      ...

  - name: datadog-agent
    image: gcr.io/path/to/monitor:1.0.0
    volumeMounts:
      - name: datadog-data
        mountPath: "/var/log/monitor"
    requests:
      ...

  # Storage
  volumes:
  - name: datadog-data
    emptyDir: {}
---


# External Service for user-msvc
kind: Service
apiVersion: v1
metadata:
  name: user-msvc
spec:
  ports:
  - name: secure-port
    port: 443
    targetPort: 443
    protocol: TCP
  selector:
    app: user-msvc
  type: LoadBalancer
  loadBalancerIP: 12.34.56.78 # This is a fake IP
---
```

## ConfigMap
K8s ConfigMaps are used to store configurations for other K8s objects. It is designed for decoupling configurations from the configurations of other K8s objects. Kubernetes Secret is also designed for this purpose and it is used for storing sensitive data like SSL certificates.

In this case, the `mysql-host` config is stored in the `user-msvc` ConfigMap and gets used as an environmental variable in the `user-mscv` container.


## Pod Template

### Required Fields
The `apiVersion`, `kind: Pod` and `metadata` fields are required as they are the Pod's metadata. The `.spec` field is also required as it defines all the components of the Pod, including volumes and containers.

### Containers
The `.spec.containers` defines all the containers of the Pod. For each container, you need to configure which image it is going to run, all the environmental variables and computing resources it needs, including storage, CPU, memory, and network.

### Storage Configuration
The `.spec.volumes` field specifies the shared storage resources for all the containers of the Pod. Kubernetes supports many types of Volumes. you can check [this doc](https://kubernetes.io/docs/concepts/storage/volumes/) for more details.

In this example, an [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume is created when the Pod is created. The volume is respectively mounted to the `/var/log` and `/var/log/monitor` path in the `user-msvc` and `monitor` container. Both containers share the data in this Volume. Because of this, the `user-msvc` container can create a file called `/var/log/user-msvc-error.log` and writes logs to this file, while the `datadog-agent` container can read the logs from the `/var/log/monitor/user-msvc-error.log` file and then sends them to datadog.

Kubernetes provides Persistent Volumes and Persistent Volume Claims for isolating storage configurations. You can check [this blog](https://www.aaronzhuo.com/persistent-volumes-and-persistent-volume-claims-in-kubernetes/) for more details.

### Network
The network is another shared resource among the containers running in the same Pod, which means these containers can reach each other through `localhost` (`127.0.0.1`). This also means a port cannot be used in two containers at the same time. You need to deal with the port assignment for each container.

In this example, the `cloudsql-proxy` container exposes itself by opening the port `3306`. So that the `user-msvc` container is able to connect to it through `127.0.0.1:3306`. Moreover, the `user-msvc` container opens the `443` port and a K8s [LoadBalancer Service](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer) `user-msvc` is in charge of redirecting HTTPs requests to this port so that the `user-msvc` container can process them.

The following picture shows what the `user-msvc` microservice looks like in Kubernetes world.

![user-msvc Topology](https://github.com/azhuox/blogs/blob/master/kubernetes/pods/assets/user_msvc_topology.png?raw=true)

# What is Next

It is not a good idea to directly utilize K8s Pods to run applications as Pods are mortal. They cannot be resurrected when they are killed for whatever reason. Because of this, you should use K8s Deployments to run stateless applications and K8s StatefulSets to run stateful applications.

You can take a look at this [this blog](https://www.aaronzhuo.com/kubernetes-deployment/) if you are curious about K8s Deployments

You can take a look at this [this blog](https://www.aaronzhuo.com/kubernetes-statefulsets/) if you are curious about K8s StatefulSets.


Reference:

- [Kubernetes Pods](https://kubernetes.io/docs/concepts/workloads/pods/pod/)
- [Kubernetes EmptyDir]((https://kubernetes.io/docs/concepts/storage/volumes/#emptydir))
- [Kubernetes Persistent Volumes and Persistent Volume Claims]((https://www.aaronzhuo.com/persistent-volumes-and-persistent-volume-claims-in-kubernetes/))
- [Kubernetes Deployments](https://www.aaronzhuo.com/kubernetes-deployments/)
- [Kubernetes StatefulSets](https://www.aaronzhuo.com/kubernetes-statefulsets/)
