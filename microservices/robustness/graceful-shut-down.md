Gracefully Shutdown
Author(s): Aaron Zhuo
Approvers: [Approver's Name should be someone off the team]
Implemented in: <Go>
TL;DR
This doc talks about something we can do to improve the graceful shutdown mechanism in our microservices.
Background
I recommend you reading this article before proceeding if you have no brief concept about k8s components and architecture.
Our microservices are implemented in Go and they are essentially k8s deployments or stateful sets. Let’s take the `sre-reporting` deployment in `sre-reporting-prod` namespace as an example. It is a two-replicas deployment and each pod runs two containers: sre-reporting and google_auth_proxy. When a rolling update occurs, it spins up two new pods and then kills the old pods. 
The followings are what happens when a old pod is being murdered:
The k8s API server gets the request of rolling update.
The k8s API server starts deleting with a grace period, which is 30 seconds in our case. It then changes the pod state from `running` to `terminating`.
Pod shows up as “Terminating” when listed in client commands
(simultaneous with 3) The k8s API server tells endpoint controller, kube-proxy on every node and the node on which the pod is running to start killing the pod.
(simultaneous with 3) The endpoint controller removes the pod’s IP from the endpoint list of any service that binds to the `sre-reporting` deployment.
(simultaneous with 3) The kube-proxy running on every node removes the pod’s IP from iptables.
(simultaneous with 3) The kubelet on the node where the pod is running starts the following steps for each container:
The preStop hook is executed if it is defined.
The container is sent the TERM signal.
The `sre-reporting` container 1. Marks healthz dead so that each Stormbreaker pod starts refreshing its envoy configs; 2. Sleep 20 seconds to wait for Stormbreaker; 3. Call `grcpServer.GracefulStop()` to gracefully shuts down the grpc server; 4. Call `httpServer.Close()` to shuts down the HTTP Server.
Each Stormbreaker pod performs readiness probe every 5 seconds and it won't realize the pod is being killed until it continuously receives three probe failures. This means it takes at least 15 seconds for each Stormbreaker pod to start refreshing its configs.
When the grace period expires, any container still running in the pod are violently killed with SIGKILL. 
There are some problems I found in the current workflow:
The ambassador container, the `google_auth_proxy` container gets killed intermediately as it has no `preStop` hook and no gracefully shut down workflow. 
Sleep 20 seconds is not enough. I found Stormbreaker sometimes still sent requests to the old pod even after the `sre-reporting` container has shut down the grpc server and the HTTP server.
`httpServer.Close()` does not perform gracefully shut down.
Related Proposals:
A list of proposals this proposal builds on or supersedes.
Proposal
We need to do the following steps to fix the above problems:
Change `terminationGracePeriodSeconds` from 30 seconds to 60 seconds so that each container can have enough time to gracefully shut down.
Define a `preStop` hook for ambassador or sidecar containers. It doesn’t have to be complex but just like `sleep 59 seconds` to prevent them from being killed intermediately during the termination period. 
Extend sleep time from 20 seconds to 30 seconds so that each Stormbreaker pod can have about 15 seconds other than 5 seconds to refresh its envoy configs. This also means the gRPC server and HTTP server have 30 seconds to terminate themselves. Any slow synchronous API that takes more 30 seconds to finish should be converted to asynchronous using cloud tasks or odin.
Replace `httpServer.Close()` with `httpServer.Shutdown(ctx)` in order to gracefully shut down the HTTP server.
Rationale
[A discussion of alternate approaches and the trade offs, advantages, and disadvantages of the specified approach.]
Implementation
In mscli, change `terminationGracePeriodSeconds` from 30 seconds to 60 seconds.
In mscli, change `initialDelaySeconds` from 0 seconds to 5 seconds for readiness probe and liveness probe. This gives new pods some time to breathe before serving requests.
Update mscli in mission control so that a microservice can adopt the above changes after you deploy a new version via mission control, or by running `mscli app deploy ...`.
Replace `httpServer.Close()` in gosdks/serverconfig/server.go with `httpServer.Shutdown(ctx)`
Update sleep time in gosdks/serverconfig/server.go to 30 seconds.
Update gosdks in every golang microservice. At least for those important services.
Define `preStop` hook for sidecar or ambassador containers, for example, `google_auth_proxy`.
Convert slow synchronous APIs to asynchronous using cloud tasks or odin.
Open issues (if applicable)
[A discussion of issues relating to this proposal for which the author does not know the solution. This section may be omitted if there are none.]



