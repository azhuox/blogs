Config Auto Scalers for Microservices
Author(s): Aaron Zhuo
Approvers: [Approver's Name should be someone off the team]
Implemented in: <Go, Kubernetes>
TL;DR
A Kubernetes auto-scaler can help us auto scales the number of replicas of a microservice (a deployment or a stateful set) when it reaches the threshold defined in its config. This makes the microservice able to handle those unexpected requests.

This document talks about the current auto scalers we have and what we can do improve them.
Background
What We Have
Let us take iam-prod as an example:

We define three deployments for this service:


We define a Horizontal Pod Autoscaler (HPA) for each deployment:




You can see that the auto-scaler will be triggered only when the average CPU usage of all the pods of the deployment exceeds 50%. This means this auto-scaler only works well for CPU-intensive applications. Different types of applications may need different types of Autoscalers.
Auto Scalers in Kubernetes
There are two kinds of auto scalers in Kubernetes: Vertical Pod Autoscaler (VPA) and Horizontal Pod AutoScaler (HPA). The following picture vividly demonstrates the difference between these two Autoscalers.


Vertical Pod Autoscaler
VPA now is in beta state and can be tested in our k8s clusters. It performs auto-scaling by replacing old Pod with the new ones with more resource requests. when you apply a VPA to a microservice (a deployment or a stateful set), it automatically analyzes the CPU and memory needs of the containers and records those recommendations in its status field. Moreover, it can automatically apply these recommendations to the Pods of your microservice when you config `updateMode=auto`.

Here are the major benefits of VPAs:
Pods are scheduled onto nodes that have the appropriate resources available.
You don't have to run time-consuming benchmarking tasks to determine the correct values for CPU and memory requests.
Maintenance time is reduced because the Autoscaler can adjust CPU and memory requests over time without any action on your part.

Here are the major limitations of VPAs:
Vertical Pod Autoscaler should not be used with the Horizontal Pod Autoscaler (HPA) on CPU or memory at this moment. However, you can use VPA with HPA on custom and external metrics.
VPA reacts to most out-of-memory events, but not in all situations.
Vertical Pod Autoscaling supports a maximum of 200 VPA objects per cluster.
Do not enable Vertical Pod Autoscaling on clusters with more than 1000 nodes.
VPA updates resource requests for Pods by replacing old Pods with new ones.

Reference:
VPA API Spec
Configuring Vertical Pod Autoscaling in GKE Engine
Horizontal Pod Autoscaler
HPA performs auto-scaling by creating more Pods for a microservice so that it can handle more requests.

HPA can use the following matrix to trigger auto-scaling:
Resource-based matrix: for example CPU and memory.
Customer matrix: for example requests-per-second for Pods.
External matrix: for example a HPA based on a datadog matrix.

You can utilize multiple matrices in one HPA. The following example configs a multiple-matrices-based HPA:

apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
 name: iam
 namespace: iam-prod
spec:
 scaleTargetRef:
   apiVersion: extensions/v1beta1
   kind: Deployment
   name: iam
 minReplicas: 2
 maxReplicas: 4
 metrics:
   - type: Resource
     resource:
       name: cpu
       targetAverageUtilization: 50
   - type: Resource
     resource:
       name: memory
       targetAverageUtilization: 50
   - type: Pods
     pods:
       metric:
         name: requests-per-second
       targetAverageValue: 1k
   - type: External
     external:
       matricName: IAM.gRPC.Latency.avg
       metricSelector: 
         matchLabels:
           kube_container_name: iam
       targetAverageValue: 80  #ms

Reference:
Horizontal Por Autoscaler in Kubernetes
Horizontal Pod Autoscaler Walkthrough
Autoscale Kubernetes workloads with any Datadog metric

VPA v.s. HPA
As described above, VPA is suitable for the following cases:
You want k8s to analyze the resource recommendations for the containers running in the Pods of your microservices.
You want k8s to manage resource configs for the containers running in the Pods of your microservices.
You don’t want to increase the replicas of your microservice when unexpected requests come in. For example, the billing microservice has limits on how fast it can pull pubsubs, which means scaling horizontal hits quota issues. Therefore, it is better to keep the same number of pods with more resources and just do more work in each.
 
HPA is suitable for the following cases:
You want to perform auto-scaling with multiple matrices.
You want to do auto-scaling with some external matrices like datadog matrices.
You are OK with auto-scaling by creating more Pods.

Related Proposals:
A list of proposals this proposal builds on or supersedes.
Proposal
Try out all of the Autoscalers I mentioned above.
Add the support of those autoscalers in mscli if we decide to use them.
Use those autoscalers in our microservices.
Rationale
[A discussion of alternate approaches and the trade offs, advantages, and disadvantages of the specified approach.]
Implementation
Do a spike to test out all of the Autoscalers I mentioned above (VPA and HPA with resource, custom and external matrices).
Add the support of these Autoscalers in mscli: This should allows users to define one or more supported Autoscalers in microservice.yaml and use `mscli app provision` command to create such Autoscaler in the corresponding namespace.
Document the instructions about how to config & use these Autoscalers in mscli’s README.
Choose and use the right Autoscalers in your microservices based on your need. This may require every team’s effort. 
Open issues (if applicable)
[A discussion of issues relating to this proposal for which the author does not know the solution. This section may be omitted if there are none.]


