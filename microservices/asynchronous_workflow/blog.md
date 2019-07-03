# Asynchronous Task Driven Workflow in Microservices

# Preface

When developing a microservice in a container-orchestration system, for example, Kubernetes, there are several things that we need to keep in mind all the time:

1. Pods are mortal and they won’t be able to survive scheduling failures, node failures or other evictions.
2. ANYTHING can fail ANY TIME for whatever naughty reasons.

These normally only have very minor impact on those simple operations, like a simple synchronous API which only has one step to create a resource, for example, a user in the system. The worst case is the resource is not created because of the API failure. You can see the API would not create any garbage in the system even if it failed as it is a `all or nothing` operation. However, computer systems are normally complicated sometimes you are very "unlucky" to make some contribution to this part. Let us take a look at an interesting example.

Suppose you are working on a WordPress-hosting platform and you need to build an API in a Golang microservice, say `site-manager`, for customers to create their sites in the system. The following are all the tools you have:

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

The following picture demonstrates the workflow of the synchronous `site create` API:

1. A customer creates an account in the system and logs in. Well this involves with another complicated workflow. So let us assume everything works well and the customer finishes this within 5 seconds.
2. The customer goes to the front end and `site creation` page.
3. The customer inputs the site name, tagline and the name of a free domain he wants, for example, `free-domain-part`.we-host-sites-for-you-and-this-domain-is-too-long.com. The customer submits the site creation page and starts enjoying an animation which indicates the site is being built, and prepares to be excited about exploring his site.
4. The front end sends the requests to the site-manager microservice.
5. The site-manager's API server receives the request, performs the authorization and authentication checks and calls the site-manager's service to process the request if the checks are passed.
6. The site-manager's service has no business logic to process so it just delegates the request to the site-manager's repo.
7. The site-manager's repo execute the following steps to process the request:
    a. It validates all the arguments. It will return an error if some arguments are invalid.
    b. It saves the site metadata to the database. It will abort the operation and return an error if an error occurs.
    c. It creates the site in the file system. It will abort the operation, trigger rollback and return an error if an error occurs.
    d. It creates a database for the site. It will abort the operation, trigger rollback return an error if an error occurs.
    e. It calls a WordPress API to bootstrap the site in the database. It will abort the operation, trigger rollback return an error if an error occurs.
    f. It will return the site metadata to the service if all the above steps succeed.
8. The site-manager's service returns the result to the API server while the API server returns the result back to the front end.
9. The front end redirects the customer to site's dashboard if the site is successfully created.
10. The customer starts exploring his site.


[image]

Now let us take several seconds to "judge" this API: what will happen if a Pod gets "murdered" when it is executing this API? What is the next to do when the API successfully saves the site metadata in the database but fails to creates the file in the File System? What if the API fails and the rollback fails as well?  What if it takes a long time to bootstrap the site in the database and causes timeout? The simple answer is the API will fail. The more sophisticated answer is the API will fail and potentially leave some piece of dangling data in the system and the customer may not be happy especially after waiting for like 30 seconds.

From the above example, you can see that it is not a good idea to perform such a complicated and time-consuming task (creating a site in the system) in a synchronous API as it is so fragile in this scenario. We need to build an asynchronous and transaction safe API to replace the synchronous one. There are several principles we should consider when developing such an asynchronous workflow:

1. We should break this complicated task into multiple smaller and simpler tasks;
2. Each task should have the retry mechanism to ensure its final success;
3. Task B should be driven by task A if task B depends on task A;
3. Some tasks can be executed simultaneously if they do not depend on each other;
4. A group of flags (or a `struct`) should be used to indicate the state of each task and the state of the whole workflow.


example: 佛跳墙



