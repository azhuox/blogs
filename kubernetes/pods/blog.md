# Kubernetes (K8s) Pods

What is included in this blog:
- what is a pod
- Use of Pods
- Life cycle of a Pod

## What is A Pod

A Pod are smallest deploy-able unit in kubernetes. It is essentially is a group of one or more containers
(e.g. Docker containers) with share network (IP and port space) and storage. You can consider a pod an application
specific "host" which containers one or more "processes" (containers). These "processes" are tightly coupled to
provides some kind of service. This is also how a micro service is composed in kubernetes.

## Use of Pods

The following example demonstrates how to use Kubernetes Service and Pod to construct a micro-service:
```yaml

# ConfigMap for user micro-service
apiVersion: v1
kind: ConfigMap
metadata:
  name: user-msvc-configmap
data:
  mysql-host: 127.0.0.1:3306
---


# Pod for user micro-service
apiVersion: v1
kind: Pod
metadata:
  name: user-msvc-pod
  labels:
    app: user-msvc-pod
spec:
  volumes:
  - name: datadog-dir
    emptyDir: {}

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
          name: website-pro-configmap
          key: mysql-host
    volumeMounts:
      - name: named-pipe
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
      - name: named-pipe
        mountPath: "/var/log/monitor"
    requests:
      ...

# External Service for user micro-service
kind: Service
apiVersion: v1
metadata:
  name: user-msvc-service
spec:
  ports:
  - name: secure-port
    port: 443
    targetPort: 443
    protocol: TCP
  selector:
    app: user-msvc-pod
  type: LoadBalancer
  loadBalancerIP: 35.199.15.199 # This is a fake IP
---
```

### ConfigMap
Kubernetes ConfigMap is used to store configurations for other Kubernetes objects. It is designed for decoupling
configurations with containers. Kubernetes Secret is also designed for this purpose and it is majorly used for storing
sensitive data like SSL certificates or login credentials.

In this case, the config `mysql-host` is stored the ConfigMap `user-msvc-configmap` and gets used as an environmental
variables in the container `user-mscv`  inside the Pod `user-msvc-pod`


### pod template

The Pod Spec describes the components of a Pod. **It also demonstrates why a Pod is like an application-specific host while
containers inside are like "processes" in a "host".**

#### Required Fields

`apiVersion`, `kind: Pod` and `metadata` are required fields of a Pod as they are meta data for the Pod.
You can use `.metadata.name` to name your Pod and use `.metadata.labels` to add some labels to your Pod.

`.spec` field is also required as it defines all the components of a Pod, such as Volumes and containers.

#### Volumes

Volumes is the way that Kubernetes provides for sharing storage resources among the containers inside the same Pod.
Kubernetes supports several [types fo Volumes](https://kubernetes.io/docs/concepts/storage/volumes/#local), such as:
- emptyDir: It is an initially empty directory which is created when a Pod is created. It is deleted when the Pod is removed.
- nfs:  An `nfs` volume allows you to connect a NFS (Network File System) server to your Pod and to be shared among you
Pod's containers. Unlike `emptyDir`, the data in `nfs` Volume is permanent and won't be erased when a Pod is removed.
- local: A local Volume allows your Pod to access a node's local storage such as a disk, partition or directory.

In this example, an `emptyDir` volume is created when the Pod is created. The volume is respectively mounted to the path
`/var/log` and path `/var/log/monitor`  in container `user-msvc` and `monitor`. Both containers share the data of
this `emptyDir` volume. Because of this, the container `user-msvc` can create a file called `/var/log/user-msvc-error.log`
and writes error logs to the file, while the container `monitor` can read error logs from
file `/var/log/monitor/user-msvc-error.log` and then send them to backend data center.

You may need to setup Kubernetes Persistent Volumes (PV) and Kubernetes Persistent Volume Claims (PVC) in order to utilize
other kinds of Volumes (e.g. NFS server).
You can check this doc for more details about PV and PVC.

#### Network

Network is another shared resource in a Pod. This means the containers inside a Pod can reach each other
through `localhost` (or `127.0.0.1`). This also means a port cannot be used in two containers at the same time.
You need to deal with port assignment for your Pod's containers.

In this case, the container `cloudsql-proxy` exposes itself by opening the port `3306`. So other containers, like `user-msvc`,
is able to connect to the container `cloudsql-proxy` via `127.0.0.1:3306`. Moreover, the container `user-msvc` opens the port
`443`, which means it can handle the requests coming from `443`.

### Expose Your Pods

You can use Kubernetes Services to expose the containers inside your Pods. In this this case, the `user-msvc-service` is
a Kubernetes LoadBalancer Service that owns a public IP (35.199.15.199). It listens to port `443` and redirects any requests from
`35.199.15.199:443` to the pods with label `app: user-msvc`. The Pod `user-msvc-pod` has the label `app: user-msvc`. Therefore,
it can receive the requests redirected from the service `user-msvc-service`. Moreover, the container `user-msvc` opens
the port `443`， so that it can process these requests.

There are several kinds of Kubernetes Services designed for various use cases. You can check this document if you are interested.


## Pod Lifecycle

Each Pod has a `status` field which is used to indicates Pod status. It is essentially a
[PodStatus object](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#podstatus-v1-core) and it majorly
consists of fields `phase`, `conditions` and `containerConditions`

### Pod Phase

The following picture shows life cycle of a Pod.

[image]

The value of `phase` field can be one of these:
- Pending: The Pod creation request has been accepted by Kubernetes, but one or more of its Containers has not been created.
This includes time used by Pod scheduler and time spent in downloading Container images.
- Running: All of the Containers have been created and at least one container is still running.
- Failed: All Containers have terminated and at least one Container terminated with non-zero exit code.
- Unknown: The state of the Pod cannot be obtained. This is typically because the Node of the Pod becomes unavailable due
to network error.

Please note that the phase of the Pod is just high-level summary of where the Pod is in its lifecycle, which means
Pod phase is not equivalent to Pod status: A pod is in `Running` phase does not mean it is ready, instead it just means
at least one of its Containers is running. You should use Pod conditions or Container conditions of the Pod to
check the status of the Pod.

### Pod Conditions

A Pod's status consists of an array of [PodCondition], which is used to determine which conditions the Pod has
or has not passed. Here is am example of a Pod's status:

```yaml
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: 2019-01-14T22:34:21Z
    message: '0/4 nodes are available: 4 Insufficient cpu.'
    reason: Unschedulable
    status: "False"
    type: PodScheduled
```

You can check the doc of [PodCondition] object to understand the meaning of each PodCondition's field, here I just list
some  important fields:
- The field `status` indicates the status of the condition, its value can be "True", "False" or "Unknown".
- The field `type` indicates the type of the condition. Its possible values are:
    - `Unschedulable`： The Pod cannot be scheduled to any Node for some reasons, for example due to lacking of CPU resources.
    - `PodScheduled`: the Pod has been scheduled to a Node.
    - `Initialized`: all [init containers] have started successfully.
    - `ContainersReady`: all Containers in the Pod are ready.
    - `Ready`: the Pod is ready to serve requests from matching Services.

**A Pod is considered ready only when the status of the condition `Ready` is `True` and this condition is true only
 when the condition `PodScheduled` and the condition `ContainersReady` are both true**

### Container Conditions

A Pod uses an array of [ContainerStatus] object for indicates status of each Container. Here is an example:

```yaml
status:
  containerStatuses:
  - containerID: docker://4db376108b093542dbbdde7978e9df4c65a175096476c338fc06b773081f9c09
    image: gcr.io/cloudsql-docker/gce-proxy:1.12
    imageID: docker-pullable://gcr.io/cloudsql-docker/gce-proxy@sha256:4d8c6ea8039c23365e053582772bd8af69a2dc34924bde02d19e57b2e03fa3f3
    lastState: {}
    name: cloudsql-proxy
    ready: true
    restartCount: 0
    state:
      running:
        startedAt: 2019-01-14T20:54:09Z
  - ...
```

You can check [this API document] for details about ContainerStatus object, here I just list some important fields:
- The field `lastState` provides details about the container's last exit condition.
- The field `ready` is a boolean which indicates whether the container is ready
- The field `state` provides details about the container's current condition.

**Pod conditions and Container conditions are very useful for debugging as they tell you details about the Pod/Containers.
To get such information, simply just run `kubectl get pod <your_pod_name> -o yaml | grep "status:" -A 50`.**

Reference:
