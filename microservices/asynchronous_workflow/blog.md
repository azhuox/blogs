# Status-driven and Task-based Asynchronous Workflow in Microservices

# Preface

When developing a microservice in a container-orchestration system like Kubernetes, there are several things that we need to keep in mind all the time:

1. The Pods for running the microservice are mortal and they won’t be able to survive scheduling failures, node failures or other evictions.
2. ANYTHING can fail ANY TIME in the microservice for whatever naughty reasons.

These factors normally only have a very minor impact on simple APIs or jobs, for example, a synchronous API which only has one step to create a user in the system. The worst case is the user is not created due to the API failure. You can see the APIs or jobs likes this would not create any garbage in the system even if they failed, as the operations they perform are `all or nothing`. However, computer systems are always complicated and some jobs may involve with multiple operations, multiple databases and multiple subsystems, which makes it difficult to achieve "all or nothing" when processing such jobs in an asynchronous way. The purpose of this blog is to demonstrate how to utilize a status-driven and task-based asynchronous workflow to process complicated jobs.

# An Example
Suppose you are working on building a WordPress-hosting platform and you need to build an API in a Go microservice (say `site-manager`) to allow customers to create sites in the system. The following are the major components of the `site-manager` microservice that you can use:

```go
app:
    site-manager:   # The module for dealing with all the site-related requests.
        apiserver:  # site-manager's API server, which exposes the service through APIs.
        service:    # site-manager's service, which handles all the business logic.
        repo:       # site-manager's repo, which takes charge of all the database work.
```

The following picture demonstrates the workflow of the synchronous version this API that was already built before you joined the team:

1. A customer logs in the system and opens the site creation page.
2. The customer inputs the site name, tagline and the name of a free domain he wants, for example, `free-domain-part`.we-host-sites-for-you-and-this-domain-is-too-long.com. Then the customer submits the request and waits by watching an animation which indicates the site is being built.
3. The front end sends the requests (with 30 seconds timeout) to the site-manager microservice.
4. The site-manager's API server receives the request, performs the permission checks and calls the site-manager's service to process the request if the checks are passed.
5. The site-manager's service has no business logic to process so it just delegates the request to the site-manager's repo.
6. The site-manager's repo executes the following steps to process the request:
    a. It validates all the arguments and will abort the request if some arguments are invalid.
    b. It generates a UUID as the site ID and saves it with other site metadata to the database. It will abort the request if an error occurs.
    c. It then creates the site in the file system. It will abort the request and trigger a rollback routine if an error occurs.
    d. It creates a database for the site. It will abort the request and trigger a rollback routine if an error occurs.
    e. It calls a WordPress API to bootstrap the site in the database. It will abort the request and trigger a rollback routine if an error occurs.
    f. It will return the site metadata to the service if all the above steps succeed.
7. The site-manager's service returns the result to the API server and the API server returns the result back to the front end.
8. The front end redirects the customer to the site's dashboard if the site is successfully created.


![The Synchronous Site Creation API](https://github.com/azhuox/blogs/blob/master/microservices/asynchronous_workflow/assets/synchronous-site-creation-api.png?raw=true)

# Problems
Now let us take several seconds to "judge" this API: what will happen if a Pod gets "murdered" when it is executing this API? What is the next to do to rescue the whole job when the API successfully saves the site metadata in the database but fails to create the file in the File System? What if the API fails and the rollback fails as well?  What if it takes a long time to bootstrap a site in the database and causes timeout? The simple answer is the API will fail. The more sophisticated answer is the API will fail and potentially leave some piece of dangling data in the system and the customer may not be happy especially after waiting for like 30 seconds.

# An Solution
From the above example, you can see that it is not a good idea to process such a complicated and time-consuming job in a synchronous API, as it can fail any time but cannot handle the failure properly. We need to build a robust and transaction-safe asynchronous API to replace the synchronous one. There are several principles we should consider when developing this new API:

1. We should break this complicated job into multiple smaller and simpler tasks.
2. We can construct an assembly line where each task worker processes one task to move the job to the completion state.
3. Each task worker should have the retry mechanism to ensure that it can successfully process the tasks it is assigned.
4. Task worker B should be driven by task worker A if B depends on A.
5. Some tasks can be executed simultaneously if they do not depend on each other.
6. We can create an object in the database for tracking the status of the resource (e.g. a site) that is being process.

## Technology Choices: Task Queue v.s. Pub/Sub

We need to choose a technology for the system to distribute, store, fetch and execute those tasks. There are two options we have: Task Queue and Pub/Sub. [This article](https://cloud.google.com/tasks/docs/comp-pub-sub) compares the difference between these two technologies using Google Cloud Tasks and Google Cloud Pub/Sub. From my viewpoint, the key differences are:
1. Pub/Sub aims to decouple publishers and subscribers. This means when a publisher publishes an event to a topic, he does not care who subscribes the topic and what subscribers will do to handle this event.
2. Task Queue is aimed at explicit invocation where a publisher (aka. scheduler) retains full control of execution. More specifically the scheduler specifies where each message (task) is delivered and when each task to be delivered, while the workers accept the tasks and process them.

**All in all, Pub/Sub emphasizes on decoupling a publisher and its subscribers, while Task Queue is meant to chain a publisher (task scheduler) and its subscribers (task workers) together through messages (tasks). Pub/sub is more suitable for the communication between microservice, while Task Queue is more suitable for constructing asynchronous workflow within a microservice.**

I think Task Queue is the better choice in this case as some tasks in the workflow of this new API depend on the other ones. Additionally, Task Queue normally provides retry while Pub/Sub does not, and retry is the key to ensure each task to reach the final success. Plus, Task Queue can provide task/message creation deduplication to avoid repeated execution of the same tasks.

Now assume that [Google Cloud Tasks] is adopted and let us discuss how to utilize this technology and the above principles to implement the asynchronous site creation API.

# Task-based Asynchronous Workflow

## Overview

The following picture shows the workflow of this asynchronous site creation API:

![The Asynchronous Site Creation API](https://github.com/azhuox/blogs/blob/master/microservices/asynchronous_workflow/assets/asynchronous-site-creation-api.png?raw=true)

Here is the brief workflow of this API:

1. The customer fills the site creation form and submits the request.
2. The front end sends the request (with 30 seconds timeout) to the site-manager microservice.
3. The site-manager's API server receives the request, performs the permission check and calls the site-manager's service to process the request if the permission check is passed.
4. The site-manager's service delegates the request to the site-manager's repo.
5. The site-manager's repo performs the following steps to process the request:
 a. It validates all the arguments and will return an error if some arguments are invalid.
 b. It creates an object (let us call it `site status`) in the database for tracking the status of the creating site, and will return an error if fails.
 c. It distributes a `create site metadata` task.
6. The site-manager's API server returns 202 (request accepted) to the front end.
7. The related task workers in the site-manager's repo start working together to create the site in the system.

Compared to the synchronous version, this asynchronous API has several significant changes:

- **The API returns 202 when the request is accepted, which means the site-manager microservice needs to provide another API for checking whether the site is created through the site status**
- **There is no rollback when something fails. Instead, when processing a task, a task worker will retry forever until it succeeds. The only way to revert the site creation operations is to invoke a site deletion API call. The reason for doing this is the site creation workflow is already very complicated and injecting the rollback into it will make it way more complicated. Instead, we want this API to be more declarative: The API should always try its best to move a new site from the initialize state (`creating`) to the expected state (`running`).**

Apparently, the above workflow is not comprehensive enough to explain this API. Now let us go through every part of this workflow to explore more details.

## Synchronous to asynchronous

You may notice that this API consists of two parts: the synchronous part and the asynchronous part. The synchronous part includes the steps of performing the permission check, validating the arguments, creating the site status in the database and distributing a `create site metadata` task to trigger the asynchronous part. The synchronous part plays two important roles. Firstly, it acts as a gateway and filters out bad (4**) requests. For example, while validating the arguments, it needs to ensure that the domain that the customer provides has not existed in the system. Secondly, it triggers the asynchronous workflow to process the site creation request. **The synchronous part is very fragile as it may fail in any step as there is no retry mechanism to ensure its final success. Moreover, the failure of any step in this part will lead to the failure of the whole API. Therefore, the synchronous part should just do as less work as possible.**

The asynchronous part consists of multiple task workers and each worker processes just one task to move the creating site to the desired state. **A site creation request gains the guarantee to be processed only when it reaches the asynchronous part.**

## Task Schedulers and Task Workers

The following picture demonstrates the relationship between each task worker. You can see that a task worker can be a task scheduler for other workers. Moreover, the chain of these workers represents the site creation workflow: an assembly line is constructed by these task workers to process the site creation job.

![Better Dependencies Between Task Workers](https://github.com/azhuox/blogs/blob/master/microservices/asynchronous_workflow/assets/better-task-worker-dependencies.png?raw=true)

### Parallel Execution

Parallel Execution is meant to execute multiple tasks simultaneously to short the API execution time. As shown in the following picture, the `create site metadata`, `create site FS` and `create site DB` task can be all distributed in the `repo.CreateSiteAsync()` method. However, there is a potential problem if we realize the API in this way. That is, the site creation process will be out of control when the `repo.CreateSiteAsync()` method successfully triggers the asynchronous workflow by distributing the `create site metadata` task but fails to distribute other tasks. In this case, the `repo.CreateSiteAsync()` method returns an internal error (5**) back to the front end. The customer can retry later but he may not be able to use the same domain. This is because the `save site metadata` task has been triggered in the last API and may have been successfully executed, which means the domain that the customer chooses already exists in the system and cannot be used anymore. Therefore, as shown in the above picture, it is better to let the `repo.CreateSiteAync()` method starts the asynchronous workflow with just one "key": whether the request is accepted depends on whether the `save site metadata` task is successfully distributed. The downside is now the `create site FS` and `create site DB` task has to rely on the `save site metadata` task although they do not logically depend on each other. However, this is totally worth as it makes the system way safer.

![Dependencies Between Task Workers](https://github.com/azhuox/blogs/blob/master/microservices/asynchronous_workflow/assets/task-worker-dependencies.png?raw=true)

### Sequential Execution

In this API, the `bootstrap site in DB` task needs to be triggered by the `create site DB` task worker. This is because a site can only be bootstrapped (initialized) in the database when its database is created.

## Task Execution

**The key point of making this asynchronous workflow robust is to make each task worker robust enough to finish its tasks. The workflow's success is built on the success of every task worker.** Now let us take the `save site metadata` task worker as an example to see how to build a robust task worker.

The following pseudo-code shows how to launch the `save site metadata` task workers in Go. The `repo.processSaveSiteMetadataTask()` method is the task handler for processing the `save site metadata` tasks. The back-off configuration specifies that a failed task will be retried in 3 seconds, 3 * 2 seconds, 3 * 2 * 2 seconds... until it succeeds. Any error or any aborted operation will trigger a retry. **A task is only considered successfully completed when the `repo.processSaveSiteMetadataTask()` exits without any error.**

```go
// LaunchSaveSiteMetadataWorkers launches the workers for processing the `saving site metadata` works
func (r *repo) LaunchSaveSiteMetadataWorkers() {
    // Backoff is a time.Duration counter used for retry
    backOff := &cloudtasks.Backoff{
        Min: 3 * time.Second,       // Min is the minimum time to wait before retry.
        Max: 0,                     // Max == 0 means there is no limitation of `wait_time * Factor`.
        Factor: 2,                  // Factor is the multiplying factor for each increment step.
    }

	for i := 0; i < 20; i++ {
	    // Start 20 workers
	    w := cloudtasks.NewWorker(
    		r.QueueSaveSiteMetadata,                        // Where to fetch the tasks.
    		r.processSaveSiteMetadataTask,                  // The task handler.
    		cloudtasks.WithLeaseDuration(1*time.Minute),    // The timeout for each task.
    		cloudtasks.WithMaximumTasksPerLease(1),         // The maximum tasks that be fetched by the worker at one time.
    		cloudtasks.WithHandlerBackOff(backOff),
    	)
    	go w.Work(m.ctx)
	}
}
```

The following pseudo-code shows the major workflow of the `repo.processSaveSiteMetadataTask()` method.

```go
// processSaveSiteMetadataTask processes the tasks of saving site metadata to the system.
// The `task` object contains a task ID, a payload which has all the site metadata, and a number to track how many attempts
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
        return fmt.Errorf("error distributing a `create site FS` task for site %s, err: %s", siteMetadata.ID, err.Error())
    }

     // Distribute a `create site DB` task
    if err = r.distributeCreateSiteDBTask(siteMetadata); err != nil {
        return fmt.Errorf("error distributing a `create site DB` task for site %s, err: %s", siteMetadata.ID, err.Error())
    }

    return nil
}
```

The `repo.processSaveSiteMetadataTask()` method consists of four **sequential** steps and any step’s failure will trigger the retry and it will retry forever until it succeeds. This seems pretty robust right? The answer is yes and **NO**. The retry mechanism does make the task workers really robust but there is a side effect: Suppose the front end times out. In order to allow the customer retry with the same domain, it sends a request to delete the site that is being created. At the same time, the `repo.processSaveSiteMetadataTask()` method triggered from the last API call failed several times but is still retrying. The site deletion API finishes quickly as nothing has been created. And then the `repo.processSaveSiteMetadataTask()` method finally succeeds: it saves the site metadata to the database and distributes the `create site FS` and `create site DB` task. This will cause two problems: 1. The free domain that the customer wants is locked. 2. Some garbage data is created and never gets a chance to be cleaned up. **Based on the above discussion, you can see that these task workers need to be permanently aborted in some scenarios. We need some extra flags to indicate these scenarios and the `site status` object is created for this purpose.**

## Status Driven Asynchronous Workflow

The following pseudo code shows how a `site status` object looks like in Go:

```go
type SiteStatus{
    State         string    // Site state, available options: [creating | deleting | running].
    MetadataReady bool      // A field used to indicate whether site metadata is ready.
    FSReady       bool      // A field used to indicate whether site FS is ready.
    DBReady       bool      // A field used to indicate whether site DB is ready.
    Bootstrapped  bool      // A field used to indicate whether site is bootstrapped in the database.
}
```

In this Go structure, The `SiteStatus.State` field is used to indicate which state the site is in and it can be either `creating`, `running` or `deleting`. The site state should be converted from `creating` to `running` when those `*Ready` or `*ed` fields are both true. Therefore, we need an `update site state` task worker to do this work:

```go
// processUpdateSiteStatusTask processes the tasks of updating site status for given site.
func (r *repo) processUpdateSiteStatusTask(task *cloudtasks.Task) error {
    // Parse the site metadata from the task payload.
    siteMetadata, err := parseMetadataFromPayload(task.Payload)
    if err != nil {
        return fmt.Errorf("error parsing the payload from the task %s, err: %s", task.ID, err.Error())
    }

    // Fetch site state
    siteStatus, err := r.fetchSiteStatus(siteMetadata.ID)
    if err != nil {
        return fmt.Errorf("error getting the site status for site %s, err: %s", siteMetadata.ID, err.Error())
    }

    if siteStatus.State != "creating" {
        sendAlert("error updating site state to from `creating` to `running` for site %s, unexpected site state %s. Abort the whole workflow.",
            siteMetadata.ID, siteStatus.State)
        return nil  // Return nil to abort the whole workflow.
    }

    if siteStatus.MetadataReady && siteStatus.FSReady && siteStatus.DBReady && siteStatus.Bootstrapped {
        // Update the site state to `running` if everything is ready
        if err = r.updateSiteStatus(r.siteStatusState, "running"); err != nil {
            // Return an error to trigger a retry
            return fmt.Errorf("error updating the site state from creating to running for site %s, err %s",
                siteMetadata.ID, err.Error())
        }
    }
```


The following pseudo -ode demonstrates the refactored version of the `repo.processSaveSiteMetadataTask()` method with the `site status` object being used. The major changes are:

- The whole workflow will be aborted if the `siteStatus.State` is not `creating`. **This allows us to safely abort the creation workflow when the site is being deleted.**
- The `save site metadata` step will be skipped if `siteStatus.MetadataReady == true`.
- With the help `site status` object, each task worker now can have more accurate control of the tasks it is processing.
- **The workflow becomes status driven: A new site's state is moved from `creating` towards `running` by each task worker.**

```go
// processSaveSiteMetadataTask processes the tasks of saving site metadata to the system.
func (r *repo) processSaveSiteMetadataTask(task *cloudtasks.Task) error {

    // Distribute a `update site status` task

    // Parse the site metadata from the task payload.
    siteMetadata, err := parseMetadataFromPayload(task.Payload)
    if err != nil {
        return fmt.Errorf("error parsing the payload from the task %s, err: %s", task.ID, err.Error())
    }

    // Fetch site status
    siteStatus, err := r.fetchSiteStatus(siteMetadata.ID)
    if err != nil {
        return fmt.Errorf("error getting the site status for site %s, err: %s", siteMetadata.ID, err.Error())
    }

    if siteStatus.State != "creating" {
        // Unexpected site state
        sendAlert("error saving site metadata for site %s, unexpected site state %s. Abort the whole workflow.", siteMetadata.ID, siteState.Status)
        return nil  // Return nil to abort the whole workflow.
    }

    if !siteStatus.MetadataReady {
        // The site metadata has not been saved or `siteState.MetadataReady` did not get successfully updated in the last run.

        // Save the metadata to the database.
        if err = r.saveSiteMetadata(siteMetadata); err != nil {
            if r.alreadyExistsError(err) {
                if task.Attempts == 1 {
                    // Site metadata already exists but no attempt has been made. This should never happen.
                    sendAlert("error saving site metadata for site %s, the site metadata already exists unexpectedly. Abort the whole workflow.", siteMetadata, siteState.Status)
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
        if err = r.updateSiteStatus(r.siteStatusMetadataReady, true); err != nil {
            // Return an error to trigger a retry
            return fmt.Errorf("error updating the MetadataReady field for site %s, err %s", siteMetadata.ID, err.Error())
        }
    }

    // Skip saving site metadata if `siteState.MetadataReady == true`.

    // Distribute a `create site FS` task

    // Distribute a `create site DB` task
    return nil
}
```

The following picture shows the final workflow:

![Status Driven Site Creation Workflow](https://github.com/azhuox/blogs/blob/master/microservices/asynchronous_workflow/assets/status-driven-site-creation-workflow.png?raw=true)

## The worst case

Now let us talk about worst case that can happen: The front end times out and sends a site deletion request to delete the site that is being created. Suppose the site-manager's API server and sets `siteStatus.State` to `deleting` and triggers the site deletion workflow while the site is still being created. Any pending task in the site creation workflow will be aborted as the `siteStatus.State` is not `creating` anymore. However, the ongoing tasks may not notice the change of `siteStatus.State`, which may lead to a situation where the site deletion workflow finishes before the site creation workflow. We can alleviate this problem by delaying the execution the `delete site metadata` task (e.g. 3 seconds) after setting `siteState.Status` to `deleting` in site deletion workflow. But this problem may still happen (very unlikely though). What if it still happened?  Well, it is the right time to contact the support...

# Summary

This blog discusses a status-driven and task-based asynchronous workflow which can be used to process complicated jobs. The following are the key points of this workflow:

- A complicated job is separated into multiple smaller tasks. One or more task workers are introduced to processed these tasks.
- Each task work retries forever until it successfully processes its tasks.
- An assembly line is constructed by these task workers to process the job: each task worker processes one task of the job to move it towards the completion. The job is considered completed only when it successfully goes through the whole assembly line.
- An object is introduced to keeps the status of resource that is being processed by the assembly line: The resource is slowly moved from the initial state (e.g. `creating`) to the desired state (e.g. `running`) by task workers. There is no rollback so this state transformation can only be successful or aborted.

This workflow can be very helpful for building robust software. However, it is a little bit complicated that you may be wondering whether it is worth to utilize such a workflow in your software? Well, it is definitely not necessary to apply this workflow to those simple CRUD operations. However, I think it is totally worth to use it to process complicated jobs. Take the site creation job as an example, the synchronous API may have 99.8% availability while the asynchronous version using this workflow may increase the availability to 99.999%.

It reminds me of a famous Chinese cuisine called [Buddha Jumps Over the Wall](https://en.wikipedia.org/wiki/Buddha_Jumps_Over_the_Wall) when I am writing this blog. This cuisine needs a dozen of ingredients and has a dozen of steps. It takes almost a day to cook and any step's failure can ruin the dish. In my opinion, building  a software is very similar to cooking a cuisine where every piece needs to be well designed and implemented. The quality of the software highly depends on how you "cook" it.

In the end, I hope you enjoy reading this "receipt" _>

# Reference
- [Choosing Between Cloud Tasks and Cloud Pub/Sub](https://cloud.google.com/tasks/docs/comp-pub-sub)
- [Google Cloud Tasks](https://cloud.google.com/tasks/)
- [Buddha Jumps Over the Wall](https://en.wikipedia.org/wiki/Buddha_Jumps_Over_the_Wall)

