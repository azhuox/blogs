# Asynchronous Message Driven Workflow in Microservices

# Preface

When developing a microservice in a container-orchestration system like Kubernetes using a [Kubernetes Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/), there are several things that we need to keep in mind all the time:

1. Pods are mortal and they won’t be able to survive scheduling failures, node failures or other evictions.
2. ANYTHING can fail ANY TIME for whatever naughty reasons.

These normally only have very minor impact on those simple operations, like a simple synchronous API which only has one step to create a resource, for example, a user in the system. The worst case is the resource is not created because of the API failure. You can see the API would not create any garbage in the system even if it failed as it is a `all or nothing` operation. However, computer systems are normally complicated sometimes you are very "unlucky" to make some contribution in terms of complexity. Well, do not panic, as the purpose of this blog is to demonstrate how to build an asynchronous task driven system which may help you to deal with complex tasks in a nice way.

# An Example
Suppose you are working on a WordPress-hosting platform and you are required to build an API for customers to create their sites in the system in a Golang microservice, say `site-manager`. The following are all the tools you have:

```
app:
    site-manager:   # The module for dealing with all the site-related requests.
        apiserver:  # site-manager's API server, which exposes the service through APIs.
        service:    # site-manager's service, which handles all the business logic.
        repo:       # site-manager's repo, which takes charge of all the database stuff.
            redisClient:    # Redis client, used to manage sites' metadata in the database (A Redis HA cluster in this case).
            fsClient:       # Filesystem client, used to manage sites in file system .
            mysqlClient:    # mysql client, used to manager sites in the database (each site has its onw mysql database).
```

The following picture demonstrates the workflow of a synchronous `site creation` API which was built by someone else:

1. A customer creates an account in the system and logs in. Well this involves with another complicated workflow. So let us assume everything works well and the customer finishes this within 5 seconds.
2. The customer goes to the front end and `site creation` page.
3. The customer inputs the site name, tagline and the name of a free domain he wants, for example, `free-domain-part`.we-host-sites-for-you-and-this-domain-is-too-long.com. Then the customer submits the site create request and starts enjoying an animation which indicates the site is being built, and prepares to be excited about exploring his site.
4. The front end sends the requests to the site-manager microservice.
5. The site-manager's API server receives the request, performs the authorization and authentication checks and calls the site-manager's service to process the request if the checks are passed.
6. The site-manager's service has no business logic to process so it just delegates the request to the site-manager's repo.
7. The site-manager's repo execute the following steps to process the request:
    a. It validates all the arguments. It will return an error if some arguments are invalid.
    b. It generate an UUID as the site ID and saves it with other site metadata to the database. It will abort the operation and return an error if an error occurs.
    c. It creates the site in the file system. It will abort the operation, trigger rollback and return an error if an error occurs.
    d. It creates a database for the site. It will abort the operation, trigger rollback return an error if an error occurs.
    e. It calls a WordPress API to bootstrap the site in the database. It will abort the operation, trigger rollback return an error if an error occurs.
    f. It will return the site metadata to the service if all the above steps succeed.
8. The site-manager's service returns the result to the API server while the API server returns the result back to the front end.
9. The front end redirects the customer to site's dashboard if the site is successfully created.
10. The customer starts exploring his site.


[image]

# The Problems
Now let us take several seconds to "judge" this API: what will happen if a Pod gets "murdered" when it is executing this API? What is the next to do when the API successfully saves the site metadata in the database but fails to creates the file in the File System? What if the API fails and the rollback fails as well?  What if it takes a long time to bootstrap the site in the database and causes timeout? The simple answer is the API will fail. The more sophisticated answer is the API will fail and potentially leave some piece of dangling data in the system and the customer may not be happy especially after waiting for like 30 seconds.

# An Solution
From the above example, you can see that it is not a good idea to perform such a complicated and time-consuming task (creating a site in the system) in a synchronous API as it is so fragile in this scenario. We need to build an asynchronous and transaction safe API to replace the synchronous one. There are several principles we should consider when developing such an asynchronous workflow:

1. We should break this complicated task into multiple smaller and simpler tasks;
2. Each task should have the retry mechanism to ensure its final success;
3. Task B should be driven by task A if task B depends on task A;
3. Some tasks can be executed simultaneously if they do not depend on each other;
4. A group of flags (or a `struct`) should be used to indicate the status of each task and the state of the whole workflow.

## Technology Choices: Task Queue v.s. Pub/Sub

From the above discussion, you can see that we need to utilize a technology for the system to distribute, store, fetch and execute the tasks. There are options we have: Task Queue and Pub/Sub. [This article](https://cloud.google.com/tasks/docs/comp-pub-sub) compares the difference between these two technologies using Google Cloud Tasks and Google Cloud Pub/Sub. From my view point, the key differences are:
1. Pub/Sub aims to decouple publishers and subscribers. This means when a publisher publishes a event to a topic, he does not care who subscribed the topic and what subscribers will do to handle this event.
2. Task Queue is aimed at explicit invocation where the publisher (aka. scheduler) retains full control of execution. More specifically the scheduler specifies where each message (task) is delivered and when each task to be delivered.

**All in all, Pub/Sub emphasizes on decoupling a publisher and his subscribers, while Task Queue is meant to chain a publisher (task scheduler) and its subscribers (task workers) together through messages (tasks).**

Let us go back the example, I think Task Queue is the better choice in this case as some steps in this workflow depend on other ones. Additionally, Task Queue normally provides retries while Pub/Sub does not. Retires is key to ensure each task to reach the final success.

Let us assume that [Google Cloud Tasks] is adopted to realize this workflow. Excepts the advantages mentioned above, Google Cloud Tasks also provides the following features:

- Configurable retries with back off options
- Access and management of individual tasks in a queue
- Task/message creation deduplication

Now let us discuss how to realize an asynchronous `site creation` API using Google Cloud Tasks and the principles mentioned above.

# Asynchronous Message Driven Workflow

## Overview

The following picture shows the workflow of the asynchronous `site creation` API:

[image]

Here is the brief workflow of this asynchronous API:

1. The customer fills the `create a site` form and submits the request.
2. The front end sends the requests to the site-manager microservice.
3. The site-manager's API server receives the request, performs the authorization and authentication checks and calls the site-manager's service to process the request if the checks are passed.
4. The site-manager's service delegates the request to the site-manager's repo.
5. The site-manager's repo performs the following steps:
 a. It validates all the arguments and will return an error if some arguments are invalid.
 b. It creates the state data in the database for the following task workers, and will return an error if fails.
 c. It distributes a `create site metadata` task.
6. The site-manager's API server returns 202 (request accept) to the front end.
7. The task workers in the site-manager's repo start working to create the site in the system.

Apparently the above workflow is too simple to explain this API. Now let us go through every part of the API explore more details.

## Synchronous to asynchronous

You may notice that this API consists of two parts: the synchronous part and the asynchronous part.

The synchronous part includes the steps to perform the auth check, validate the arguments, create the state and distribute a `create site metadata` task (to trigger the asynchronous workflow). The synchronous part normally plays two important roles. Firstly, it acts as a gateway and filters out those bad (4**) requests. For example, while validating the arguments, it needs to ensure that the domain that the customer provides has not existed in the system. Secondly it triggers the asynchronous workflow to process the request. **The synchronous part is very fragile as the failure of of any step in this part will lead to the failure of the API. Additionally, it may fail in any step as there is no retry mechanism to ensure its final success. Therefore the synchronous part should not do too much work.**

The asynchronous part is the key of this API. It consists of multiple task workers and each worker does one part of job to make its contribution to complete the request. A request is guaranteed to be processed and completed when it reaches the asynchronous part.

## Task Schedulers and Task Workers

The following picture demonstrates the relationship between each task worker. You can see that a task worker can be the task scheduler for other workers and the chain of these workers indicates the site creation workflow. It is like an assembly line where each worker finishes its task to move the creating job forward until the project is made.

[image]

### Parallel Execution

Parallel Execution is meant to execute multiple tasks simultaneously to short the API execution time. As shown in the following picture, the creation of site's metadata, file system and database can be performed in parallel in this workflow. However, there will a potential problem if we realize the API in this way. That is, the workflow will be out of control when the `repo.CreateSiteAsync()` method successfully distribute the `save site metadata` task but fails to other tasks. The `repo.CreateSiteAsync()` method returns an internal error (5**) back to the front end. The customer can retry later but he may not be able to use the same domain. This is because the `save site metadata` task has been trigger in the last API and may have been executed successfully, which means the domain that the customer chooses already exists in the system and cannot be used anymore. Therefore, as shown in the above picture, instead of giving the `repo.CreateSiteAync()` three "keys" to start three workers, it is better to reduce the key number to just one: whether the request is accepted depends on whether the `save site metadata` task is successfully distributed. The downside is now the `create site FS` and `create site DB` tasks need to rely on the `save site metadata` task although they do not logically depend on each other. However, this is totally worth as it makes the system safer.

[image]

### Sequential Execution

In this API, the `bootstrap site in DB` task needs to be triggered by the `create site DB` task. This is because a site can only be bootstrapped (initialized) when its database is created.

## Task Execution

## The State

## Site status

## Roll Back


# Reference
- [Kubernetes Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Choosing Between Cloud Tasks and Cloud Pub/Sub](https://cloud.google.com/tasks/docs/comp-pub-sub)
- [Google Cloud Tasks](https://cloud.google.com/tasks/)
- [Buddha Jumps Over the Wall](https://en.wikipedia.org/wiki/Buddha_Jumps_Over_the_Wall)



example: 佛跳墙



