# Kubernetes Services

## What Is A Kubernetes Service

A [Service](https://kubernetes.io/docs/concepts/services-networking/service/) is a Kubernetes object that exposes a set of Pods as a network service. 
Moreover, it provides service discovery mechanism that dynamically adds or removes IP addresses of Pods to its endpoint list based on the creation or deletion of these Pods.

## Service Types

Kubernetes provides many types of services but here only those frequently used ones are introduced. 
You can check [this document](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types) for more details.
Here only commonly used Services will be introduced.

### LoadBalancer

A LoadBalancer exposes a set of Pods externally. A LoadBalancer is an L4 (Layer 4) load balancer, 
which means it can only utilize the information at the [transport layer](https://en.wikipedia.org/wiki/Transport_layer) (Layer 4) to determine how to distribute client requests across a group of Pods. 


Here is an example of LoadBalancer that make the `foo` Kubernetes applications public in demo environment:

```
apiVersion: v1
kind: Service
metadata:
  name: foo-service
  namespace: foo-demo
spec:
  type: LoadBalancer
  selector:
    app: foo
    environment: demo
  ports:
  - protocol: TCP
    port: 443
    targetPort: 8443
status:
  loadBalancer:
    ingress:
    - ip: 10.254.2.127
```

From the spec, you can see that:

* It relies on the `spec.selector` field to select Pods. 

* The `status.loadBalancer` field shows the external IP address that is automatically assigned by Kubernetes.

* The `spec.ports` defines the ports that this service opens for the `foo` application. 

* With the external IP address and then open port, the `foo` service in the demo environment can be accessed with address `10.254.2.127:443`. 
When a request is sent to this address, the LoadBalancer will redirect it to the port 8443 of one of the `foo` Pods. 

### ClusterIP

A ClusterIP is a Service that exposes a set of Pods on a cluster-internal IP, which means this Service is only reachable from within the cluster. 
It is also a L4 load balancer that can only provide simple load balancing functionality based on information at the transport layer.

Here is an example of ClusterIP for the `foo` application:

```
apiVersion: v1
kind: Service
metadata:
  name: default-grpc
  namespace: foo-demo
spec:
  type: ClusterIP
  selector:
    app: foo
    environment: demo
  clusterIP: 10.0.54.223
  ports:
  - name: grpc
    port: 8443
    protocol: TCP
    targetPort: 8443
```

From the spec, you can see that:

* Like a LoadBalancer Service, a ClusterIP Service also relies on the `spec.selector` field to select Pods. 

* The `spec.clusterIP` field shows the internal IP address that is automatically allocated by Kubernetes. 
Only workloads within the same cluster can utilize this Service to access the `foo` application.

* The `spec.ports` field defines the ports that this service opens for the `foo` application.


Kubernetes will allocate a unique DNS address to a Service when it is created. The format of the DNS address is `service-name.namespace.svc.cluster.local`. 
For example, the DNS address for above ClusterIP Service is `default-grpc.foo-demo.svc.cluster.local`. 

### Ingress

An Ingress is an object that manage external access to one or more Kubernetes applications in a cluster. 
It is not a Kubernetes Service, but it does provides load balancing, SSL termination and name-based virtual hosting. 

Unlike a Kubernetes Service which is L4 load balancer and can only manage one Kubernetes applications, 
an Ingress is a L7 (application layer) load balancer and can manage multiple Kubernetes applications based on path or host names. 
For example, the following shows an example of path based Ingress. 
With this Ingress, requests with the `foo.bar.com/foo` URL will be redirected to service1 (with the 8000 port) while 
requests with the `foo.bar.com/bar` URL will be redirected to service2 (with the 9000 port). service1 and service2 can either be ClusterIP or NodePort Services.

```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: simple-fanout-example
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - path: /foo
        pathType: Prefix
        backend:
          service:
            name: service1
            port:
              number: 8000
      - path: /bar
        pathType: Prefix
        backend:
          service:
            name: service2
            port:
              number: 9000
```

## What is Next

This is the last blog of my Kubernetes introduction blogs. I highly recommend you reading [official docs](https://kubernetes.io/docs/concepts/overview/) if you want to learn Kubernetes deeply.

## Reference

- [Kubernetes Services](https://kubernetes.io/docs/concepts/services-networking/service/)
- [Kubernetes Service Types](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types)
- [transport layer](https://en.wikipedia.org/wiki/Transport_layer)
- [Kubernetes Overview](https://kubernetes.io/docs/concepts/overview/)
