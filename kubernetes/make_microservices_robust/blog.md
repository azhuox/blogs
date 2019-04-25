# How to make your micro services robust in Kubernetes

What is included in this blog:

- A discussion about how to make a microservice robust in k8s

# An example

Suppose you want to user Kubernetes manager a user micro-service (written in Golang) which provides the following APIs:

```go
Get a user details: GET https://user.micro-service.com/users/v1/{ID}/
Delete a user: DELETE https://user.micro-service.com/users/v1/{ID}/
Create a user: POST https://user.micro-service.com/users/v1/ json:{user info}
Modify a user info: PUT  https://user.micro-service.com/users/v1/ json:{user info}
```

Here is pseudo code of launching a server and registering the APIs above in Golang:
```go
func main(){
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/users/v1/{ID}/", userv1.Get)
	router.HandleFunc("/users/v1/{ID}/", userv1.Delete)
	router.HandleFunc("/users/v1/", userv1.CreateOrUpdate)

	var tlsConfig *tls.Config
	var server *http.Server

	// Init tls config and server
    // ...
	server = &http.Server{
        Addr:      ":443",
        Handler:   logging.HTTPMiddleware(router),
        TLSConfig: tlsConfig,
    }

    // Listen to port 443
    conn, err := net.Listen("tcp", ":443")
    if err != nil {
        err = fmt.Errorf("error listing to 443 port, err: %s", err.Error())
        return
    }
    listener = tls.NewListener(conn, server.TLSConfig)

    // Launch the server
    logging.Infof(ctx, "ready to serve...")
    if err := server.Serve(listener); err != nil {
        logging.Criticalf(ctx, "error serving requests, err: %s", err.Error())
    }
}
```

The logic of each of these APIs is straightforward. Take API `Get a user details` as an example, the service connects to the
database via a proxy and then look for a user based on the `userID` from the request. It returns `200` with the user details
if the user is found. Otherwise, it returns `404` with error message `User ${userID} not found`.

Let's take a look how to run such micro-service in Kubernetes

## Pod Spec

Here is the Pod spec for creating Pods used to run the user micro-service.

```yaml

```

As shown in the following picture, the container `user-msvc` connects to the container `cloudsql-proxy` via
`tcp://127.0.0.1:3306`. Then `cloudsql-proxy` proxies all the requests from `user-msvc` to a mySQL instance running
in Google Cloud Platform.
[image]


## Make your pods `permanent`

Pods in kubernetes are not durable entities. Directly created pods using Pod templates
won't be able to survive from any disaster, like node failures or scheduling failures. Therefore, in general, you should
avoid creating pods directly unless you can tolerate potential data lost. Instead, you should always use k8s controllers
to create pods even for singleton.

Kubernetes provides several controllers to offer self-healing for Pods, they are:
- Deployment: Deployment is a Kubernetes object used to managing stateless applications. It provides a convenient way
for you to scale up/down and roll update/back your Pods. You can check [this doc](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) for more details.
- StatefulSet: StatefulSet is designed for managing stateful applications. Like a Deployment, a StatefulSet can
scales up a set of Pods and perform rolling updates. Unlike a Deployment, a StatefulSet assigns a sticky and stable
identify to each of its Pods. You can check [this doc](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) for more details.
- DaemonSet: `Kubernetes DaemonSet` is designed for running daemons in kubernetes.
A DaemonSet ensures that all Nodes of your Kubernetes Cluster (by default) run a replica of a pod. You can check
[this doc](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) for more details.

### Config replicas


### Distribute Pods to different Zones

[image]

## Disruption Budget


## Make your pods robust


### Config readnessprobe and livenessprobe

[image]

### Gracefully shut down


[image]

