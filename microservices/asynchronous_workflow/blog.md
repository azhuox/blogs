# Status-driven and Task-based Asynchronous Workflow in Microservices

# Preface

When developing a microservice in a container-orchestration system like Kubernetes, there are several things that we need to keep in mind all the time:

1. The Pods of the microservice are mortal and they won’t be able to survive scheduling failures, node failures or other evictions.
2. ANYTHING can fail ANY TIME in the microservice for whatever naughty reasons.

These factors normally only have very minor impact on simple APIs or jobs, for example, a synchronous API which only has one step to create a user in the system. The worst case is user is not created due to the API failure. You can see the APIs or jobs likes this would not create any garbage in the system even if they failed the operations they perform are `all or nothing`. However, computer systems are always complicated and some jobs may involve with multiple operations, multiple databases and multiple sub systems, which makes it difficult to achieve "all or nothing" when processing such jobs in an asynchronous way. The purpose of this blog is to demonstrate how to utilize a status-driven and task-based asynchronous workflow to process complicated jobs.

# An Example
Suppose you are working on a WordPress-hosting platform and you need to build an API in a Go microservice (say `site-manager`) to allow customers to create sites in the system. The following are the major components of the `site-manager` microservice:

```
app:
    site-manager:   # The module for dealing with all the site-related requests.
        apiserver:  # site-manager's API server, which exposes the service through APIs.
        service:    # site-manager's service, which handles all the business logic.
        repo:       # site-manager's repo, which takes charge of all the database work.
```

The following picture demonstrates the workflow of the synchronous version this API that was already built before you joined the team:

1. A customer logs in the system and opens the `create a site` page.
2. The customer inputs the site name, tagline and the name of a free domain he wants, for example, `free-domain-part`.we-host-sites-for-you-and-this-domain-is-too-long.com. Then the customer submits the request and waits by watching an animation which indicates the site is being built.
3. The front end sends the requests (with 30 seconds timeout) to the site-manager microservice.
4. The site-manager's API server receives the request, performs the permission checks and calls the site-manager's service to process the request if the checks are passed.
5. The site-manager's service has no business logic to process so it just delegates the request to the site-manager's repo.
6. The site-manager's repo executes the following steps to process the request:
    a. It validates all the arguments and will abort the request if some arguments are invalid.
    b. It generates an UUID as the site ID and saves it with other site metadata to the database. It will abort the request if an error occurs.
    c. It then creates the site in the file system. It will abort the request and trigger a rollback routine if an error occurs.
    d. It creates a database for the site. It will abort the request and trigger a rollback routine if an error occurs.
    e. It calls a WordPress API to bootstrap the site in the database. It will abort the request and trigger a rollback routine if an error occurs.
    f. It will return the site metadata to the service if all the above steps succeed.
7. The site-manager's service returns the result to the API server and the API server returns the result back to the front end.
8. The front end redirects the customer to the site's dashboard if the site is successfully created.


[image]

# The Problems
Now let us take several seconds to "judge" this API: what will happen if a Pod gets "murdered" when it is executing this API? What is the next to do to rescue the whole job when the API successfully saves the site metadata in the database but fails to creates the file in the File System? What if the API fails and the rollback fails as well?  What if it takes a long time to bootstrap a site in the database and causes timeout? The simple answer is the API will fail. The more sophisticated answer is the API will fail and potentially leave some piece of dangling data in the system and the customer may not be happy especially after waiting for like 30 seconds.

# An Solution
From the above example, you can see that it is not a good idea to process such a complicated and time-consuming job in a synchronous API, as it can fail any time but cannot handle the failure properly. We need to build an robust and transaction-safe API to replace the synchronous one. There are several principles we should consider when developing such an asynchronous workflow:

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

Compared to the synchronous API, this asynchronous API has several changes:

- **The API returns 202 when the request is accepted, which means the site-manager microservice needs to provide another API for checking site status.**
- **There is no roll back when something fails. Instead, each task worker retries forever until it succeeds and the success of these task workers leads to the success of the API.** The only way to revert site create operations is to invoke a `delete site` API call.

Apparently the above workflow is not comprehensive enough to explain this API. Now let us go through every part of the API explore more details.

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

**The key point of make an asynchronous workflow robust is to make each task worker robust enough to finish its tasks. The success of the workflow is built on the success of every task worker.** Now let us take the `save site metadata` task worker as an example to discuss how to build a robust worker.

The following pseudo code shows how to launch the `save site metadata` task workers in Go. The `repo.processSaveSiteMetadataTask()` method is the task handler It can process at most 20 tasks at the same time. The back off configuration specifies that a failed task will be retried in 3 seconds, 3 * 2 seconds, 3 * 2 * 2 seconds... until it succeeds. **A task is only considered successful when the `repo.processSaveSiteMetadataTask()` exits without any error. Any error or any aborted operation will trigger a retry.**

```go
// LaunchSaveSiteMetadataWorkers launches the workers for saving site metadata to the system.
func (r *repo) LaunchSaveSiteMetadataWorkers() {
    // Backoff is a time.Duration counter used for retry
    backoff := &cloudtasks.Backoff{
        Min: 3 * time.Second,       // Min is the minimum time to wait before retry.
        Max: 0,                     // Max == 0 means there is no limitation of `wait_time * Factor`.
        Factor: 2,                  // Factor is the multiplying factor for each increment step.
    }

	for i := 0; i < 20; i++ {
	    // Start 20 workers
	    w := cloudtasks.NewWorker(
    		r.QueueSaveSiteMetadata,    // Where to fetch the tasks.
    		r.processSaveSiteMetadataTask,         // The task handler for processing the tasks of saving a site metadata to the database.
    		cloudtasks.WithLeaseDuration(1*time.Minute),    // The timeout for each task.
    		cloudtasks.WithMaximumTasksPerLease(1),        // The maximum tasks that be fetched by the worker at one time.
    		cloudtasks.WithHandlerBackoff(SendGAExecReportTaskBackOff),
    	)
    	go w.Work(m.ctx)
	}
}
```

The following pseudo code shows the major workflow of the `repo.processSaveSiteMetadataTask()` method.

```go
// processSaveSiteMetadataTask processes the tasks of saving site metadata to the system.
// The `task` object contains an ID, a payload which has all the site metadata, and a number to track how many attempts
// have been made to complete this task.
func (r *repo) processSaveSiteMetadataTask(task *cloudtasks.Task) error {
    // Parse the site metadata from the task payload.
    siteMetadata, err := parseMetadataFromPayload(task.Payload)
    if err != nil {
        return fmt.Errorf("error parsing the payload from the task %s, err: %s", task.ID, err.Error())
    }

    // Save the metadata to the database.
    if err = r.saveSiteMetadata(siteMetadata); err != nil {
        return fmt.Errorf("error saving the metadata to the system for site %s, err: %s", siteMetadata.ID, err.Error())
    }

    // Distribute a `create site FS` task
    if err = r.distributeCreateSiteFSTask(siteMetadata); err != nil {
        return fmt.Errorf("error distributing a task to create FS for site %s, err: %s", siteMetadata.ID, err.Error())
    }

     // Distribute a `create site DB` task
    if err = r.distributeCreateSiteDBTask(siteMetadata); err != nil {
        return fmt.Errorf("error distributing a task to create DB for site %s, err: %s", siteMetadata.ID, err.Error())
    }

    return nil
}
```

You can that the `repo.processSaveSiteMetadataTask()` method consists of four sequential steps. Any failure in any step will trigger the retry and it will retry forever until it succeeds. This sounds pretty robust right? The answer is yes and **NO**. The retry mechanism does make task workers really robust but there is a side affect: Suppose the front end times out and send a request to delete the site that is being created so that the customer can retry, while the `repo.processSaveSiteMetadataTask()` method failed several times but is still trying to save the site metadata to the system. The delete API finishes quickly as nothing has been created. Now the `repo.processSaveSiteMetadataTask()` finally succeeds: it saves the site metadata to the database and distributes the `create site FS` and `create site DB` tasks. This causes two problems: 1. The free domain that the customer wants is locked. 2. Some garbage data was created and never gets a chance to be cleaned up. **Based on the above discussion, you can see that these task workers need to abort the work in some scenarios.** But how to do that? We need to introduce a variable, the state, to control this.

## The State

The following pseudo code describes how a state looks like:

```go
type SiteState{
    Status        string    // Site status, available options: [creating | deleting | running]
    MetadataReady bool
    FSReady       bool
    DBReady       bool
    Bootstrapped  bool
}
```

In this Go strcut, the `MetadataReady`, `FSReady`, `DBReady`, `Bootstrapped` is used to indicate whether the corresponding task (e.g. the `save site metadata` task) has been successfully completed. The `Status` field is used to indicate the site status and it can be either `creating`, `running`, `deleting` or `deleted`. The site status can only be `running` when those `*ready` or `*ed` fields are both true. Because of this, now we need another task worker, the `update site status` task worker, to do this job:

```go
// processUpdateSiteStatusTask processes the tasks of updating site status for given site.
func (r *repo) processUpdateSiteStatusTask(task *cloudtasks.Task) error {
    // Parse the site metadata from the task payload.
    siteMetadata, err := parseMetadataFromPayload(task.Payload)
    if err != nil {
        return fmt.Errorf("error parsing the payload from the task %s, err: %s", task.ID, err.Error())
    }

    // Fetch site state
    siteState, err := r.fetchSiteState(siteMetadata.ID)
    if err != nil {
        return fmt.Errorf("error getting the site state for site %s, err: %s", siteMetadata.ID, err.Error())
    }

    if siteState.Status != "creating" {
        sendAlert("error updating site status to from `creating` to `running` site %s, unexpected site status %s. Abort the whole workflow.",
            siteMetadata.ID, siteState.Status)
        return nil  // Return nil to abort the whole workflow.
    }

    if siteState.MetadataReady && siteState.FSReady && siteState.DBReady && siteState.Bootstrapped {
        // Update the site status to `running` if everything is ready
        if err = r.updateSiteSate(r.stateStatus, "running"); err != nil {
            // Return an error to trigger a retry
            return fmt.Errorf("error updating site status from creating to running for site %s, err %s",
                siteMetadata.ID, err.Error())
        }
    }
```


The following pseudo code is the refactored version of the `repo.processSaveSiteMetadataTask()` method. The major changes are:

- The whole workflow will be aborted if the siteState.Status is not `creating`. **This allows us to safely abort the creation workflow when the site is being deleting.**
- The `save site metadata` step will be skipped if `siteState.MetadataReady == true`.
- It covers all the edge cases.

```go
// processSaveSiteMetadataTask processes the tasks of saving site metadata to the system.
func (r *repo) processSaveSiteMetadataTask(task *cloudtasks.Task) error {

    // Distribute a `update site status` task

    // Parse the site metadata from the task payload.
    siteMetadata, err := parseMetadataFromPayload(task.Payload)
    if err != nil {
        return fmt.Errorf("error parsing the payload from the task %s, err: %s", task.ID, err.Error())
    }

    // Fetch site state
    siteState, err := r.fetchSiteState(siteMetadata.ID)
    if err != nil {
        return fmt.Errorf("error getting the site state for site %s, err: %s", siteMetadata.ID, err.Error())
    }

    if siteState.Status != "creating" {
        // Unexpected site status
        sendAlert("error saving site metadata for site %s, unexpected site status %s. Abort the whole workflow.", siteMetadata.ID, siteState.Status)
        return nil  // Return nil to abort the whole workflow.
    }

    if !siteState.MetadataReady {
        // The site metadata has not been saved or `siteState.MetadataReady` did not get successfully updated in the last run.

        // Save the metadata to the database.
        if err = r.saveSiteMetadata(siteMetadata); err != nil {
            if r.alreadyExistsError(err) {
                if task.Attempts == 1 {
                    // Site metadata already exists but no attempt has been made. This should never happen.
                    sendAlert("error saving site metadata for site %s, the site metadata already exists. Abort the whole workflow.", siteMetadata, siteState.Status)
                    return nil  // Abort the whole workflow.
                }
                // Site metadata is already saved but siteState.MetadataReady is not updated
                goto updateMetadataReady
            }

            // Failed to save the site metadata, return an error to trigger a retry.
            return fmt.Errorf("error saving the metadata to the system for site %s, err: %s", siteMetadata.ID, err.Error())
        }

        // Update the MetadataReady
        updateMetadataReady:
        if err = r.updateSiteState(r.stateMetadataReady, true); err != nil {
            // Return an error to trigger a retry
            return fmt.Errorf("error updating MetadataReady state for site %s, err %s", siteMetadata.ID, err.Error())
        }
    }

    // Skip saving site metadata if `siteState.MetadataReady == true`.

    // Distribute a `create site FS` task

    // Distribute a `create site DB` task
    return nil
}
```

The following picture shows the final workflow:

[image]

## The worst case

Now let us talk about worse case that can happen: The front end times out and sends a `delete site` request request in order to free the free domain that the customer inputs. Suppose the site-manager's `delete site` API and sets `siteState.Status` to `deleting` and triggers the delete workflow while the site is still being created. Any pending task in the `create site` workflow will be aborted as the `siteState.Status` is not `creating` anymore. However, those ongoing task workers, for example, the `create site metadata` task worker, may not notice the change of `siteState.Status`, which may lead to a situation where the `delete site metadata` task worker finishes first before the `create site metadata` task worker. We can alleviate this problem by delaying the execution the `delete site metadata` task worker (e.g. 3 seconds) after setting `siteState.Status` to `deleting`. But what if it still happened? Well, it is right time to contact the support...

# Summary

This blog discusses a status driven asynchronous workflow which can be used to process complicated jobs. The following are the key points of method:

- A complicated job is separated to multiple smaller tasks. One or more task workers are introduced to processed these tasks.
- Each task work retries forever until it finishes its task and there is no roll back.
- A status driven assembly line is constructed by these task workers to process the job: each task worker processes one task of the job, updates the (status) of the job and move it forward. The job is considered completed only when it successfully goes through the whole assembly line.
- One assembly line for one kind of job, which makes it easier to realize the whole system.

This workflow can be very helpful for building robust software. However, it is a little bit complicated that you may be wondering whether it is worth to build utilize a workflow? Well, it is definitely not necessary to apply this workflow to those simple CRUD operations. However, I think it is totally worth to use it to process complicated jobs. Take the `create site` job as an example, The synchronous API may have 99.5% availability while the asynchronous version realized this workflow may increase the availability to 99.99%. It is a your choice whether to get 0.49% improvement or not.

It reminds me of a famous Chinese cuisine called [Buddha Jumps Over the Wall](https://en.wikipedia.org/wiki/Buddha_Jumps_Over_the_Wall) when I am writing this blog. This cuisine needs a dozen of ingredients and has a dozen steps. It takes almost a day to cook and any step's failure can ruin the dish. Building a software is very similar to cooking a cuisine where every piece needs to be well designed and realized. The quality of the software highly depends on how you "cook" it.

At the end, I hope you enjoy reading this "receipt" _>

# Reference
- [Choosing Between Cloud Tasks and Cloud Pub/Sub](https://cloud.google.com/tasks/docs/comp-pub-sub)
- [Google Cloud Tasks](https://cloud.google.com/tasks/)
- [Buddha Jumps Over the Wall](https://en.wikipedia.org/wiki/Buddha_Jumps_Over_the_Wall)



example: 佛跳墙



