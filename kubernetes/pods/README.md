# Kubernetes Pods

This blog talks about the basic knowledge of Kubernetes [Pods](https://kubernetes.io/docs/concepts/workloads/pods/). 
But before exploring Kubernetes Pods, let us first go through what a [Docker container](https://docs.docker.com/get-started/overview/) is as it is the major container technology that we use to run our applications.

## Docker Containers

Docker is an open platform that allows you to package and run applications in a loosely isolated environment called a container. 
A Docker container is a runnable instance of a Docker image and a Docker read-only template for creating a Docker container. 
In other words, we need to construct a Docker image before running a Docker container. 
Take the service `foo` as an example, here is the Docker file for creating its Docker images:

```
# Suppose the CI tool already build the `foo` service as a executable binary file and saved it in `./go-server`.

FROM alpine:3.5

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

COPY ./go-server /usr/bin/server

ENTRYPOINT ["/usr/bin/server"]
```

From the file, you can see a minimal Docker file can be as simple as only containing the following directives:

* FROM directive specifies alpine:3.5, which essentially is a mini Linux system, as its parent container.
* RUN apk update… updates container’s dependencies.
* COPY command copies an executable binary file that contains all the business logic of the service `foo` from ./go-server to /user/bin/server.
* ENTRYPOINT tells Docker to executes /user/bin/server as this container’s entry point.

With this Docker file, our continuous integration (CI) tool is able to build and push a new image to the service `foo's` Docker image warehouse 
every time when there is a new change in the service `foo's` repo. Then with this Docker image, we can run and deploy the service `foo` as a containerized application in Kubernetes.

## Pod Overview

A Pod is the smallest deployable unit in Kubernetes. It consists of one or more containers. 
These containers have their own CPU and memory resources but need to share other resources, including storage and network. 
As shown in the following picture, a Pod is very similar to an application-specific "host" running one or more "processes" (containers). 
These "processes" work together to construct a containerized workload or service.c

![Hosts V.S. Pods](https://github.com/azhuox/blogs/blob/master/kubernetes/pods/assets/host_vs_pod.png?raw=true)

## Use of Pods

The following example demonstrates how to use a K8s Pod to construct a single-replica microservice. 
This Pod consists of three containers: the container `user-usvc` has all the microservice's business logic, the container `cloudsql-proxy` proxies all the MySQL requests to a Google Cloud SQL instance, 
while the container `datadog-agent` sends the logs to datadog server.

```yaml

# ConfigMap for user-msvc
apiVersion: v1
kind: ConfigMap
metadata:
  name: user-msvc
  namespace: user-msvc
data:
  mysql-host: 127.0.0.1:3306

---  
# Secret of user-msvc
apiVersion: v1
kind: Secret
metadata:
  name: user-msvc
  namespace: user-msvc
type: Opaque
data:
  datadog-api-key: XKSSAKCKANmOWQ5NDU1NWU2MWE2ZDI=
  github-access-token: Yzk4MDZkNzAxNTQwMjkwOA==
  # All other configs are omitted.  

---
# Pod for user-msvc
apiVersion: v1
kind: Pod
metadata:
  name: user-msvc
  namespace: user-msvc
  labels:
    app: user-msvc
spec:
  # Storage
  volumes:
  - name: datadog-data
    emptyDir: {}
  - name: user-msvc-secret
    secret:
      defaultMode: 420
      secretName: user-msvc-secret
    
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
        mountPath: /var/log/user-msvc
      - name: user-msvc-secret
        mountPath: /etc/user-msvc/secret
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
        mountPath: /var/log/monitor
      - name: user-msvc-secret
        mountPath: /etc/datadog-agent/secret
    requests:
      ...
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

The following picture shows the topology of the Pod `user-msvc` in Kubernetes.

![Topology of user-msvc Pod](https://github.com/azhuox/blogs/blob/master/kubernetes/pods/assets/user_msvc_topology.png?raw=true)

From the above Pod configuration, we can see that:

* The ConfigMap `user-msvc` and Secret `user-msvc` are created for storing configurations and sensitive data. 

* The field `spec.containers` defines all the containers of the Pod. For each container, you need to configure which image it is going to run, 
all the environmental variables and computing resources it needs, including storage, CPU, memory, and network.

* Containers within the same Pod share the network, which means these containers reach each other through 127.0.0.1. However, a port can only be exclusively occupied by a container. In this case, the container `cloudsql-proxy` exposes itself by opening the port `3306`. Therefore, the container `user-msvc` is able to connect to it through `127.0.0.1:3306`. 
Moreover, the container `user-msvc` opens the `443` port for processing incoming requests from the container `user-msvc` [Kubernetes LoadBalancer Service](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer).

* The field `spec.volumes` specifies the shared storage resources for all the containers of the Pod. 
Kubernetes supports many types of Volumes. you can check [this doc](https://kubernetes.io/docs/concepts/storage/volumes/) for more details about Kubernetes Volumes.

In this example, an [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume is created when the Pod is created. 
The volume is respectively mounted to the path `/var/log` and `/var/log/monitor` in the container `user-msvc` and `monitor`. 
Both containers share the data in this Volume. Because of this, the container `user-msvc` can create a file called `/var/log/user-msvc-error.log` 
and writes logs to this file, while the container `datadog-agent` can read the logs from the file `/var/log/monitor/user-msvc-error.log` and then sends them to datadog.

The Secret `user-msvc` is also used as Volume in this example. 
Then it is mounted in the path `/etc/user-msvc/secret` in the container `user-msvc` and the path `/etc/datadog-agent/secret` in the container `datadog-agent`.
When a Secret is mounted into a directory in a container, each of its data will be created as an individual file in that directory. Moreover, a Secret Volume should be read only.

A ConfigMap can be directly used in a Pod's environmental variables or can be used as a Volume as well.  

All in all, this example demonstrates that a Pod is like an application-specific “host“ that coordinates one or more “processes“ (containers) to work together to provide some kind of service.


## What Is Next

It is not a good idea to directly utilize K8s Pods to run applications as Pods are mortal. They cannot be resurrected when they are killed for whatever reason. 
Because of this, you should use K8s Deployments to run stateless applications and K8s StatefulSets to run stateful applications.

Check [this blog](https://github.com/azhuox/blogs/blob/master/kubernetes/deployments/README.md) if you are curious about how to utilize K8s Deployments to run stateless applications.

Check [this blog](https://github.com/azhuox/blogs/blob/master/kubernetes/statefulsets/README.md) if you are curious about ho to utilize K8s StatefulSets to run stateful applications.


## Reference

- [Kubernetes Pods](https://kubernetes.io/docs/concepts/workloads/pods/)
- [Docker Containers](https://docs.docker.com/get-started/overview/)
- [Kubernetes Volumes](https://kubernetes.io/docs/concepts/storage/volumes/)
- [Kubernetes EmptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir)
- [Kubernetes LoadBalancer Service](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer)
- [Kubernetes Deployments](https://www.aaronzhuo.com/kubernetes-deployments/)
- [Kubernetes StatefulSets](https://www.aaronzhuo.com/kubernetes-statefulsets/)
