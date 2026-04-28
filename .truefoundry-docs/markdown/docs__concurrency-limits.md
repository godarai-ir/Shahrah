> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Set Concurrency Limit

> Step-by-step guide for concurrency limits, explaining configuration, best practices, and real-world usage on TrueFoundry.

TrueFoundry allows you to set a limit on how many runs of a Job can run concurrently (provided sufficient resources are available in the cluster). By default, this is set to `None` which allows an unlimited number of runs concurrently.

You might want to limit the maximum number of concurrent runs of a job in case we don't want to bring the resources up at the same time - for e.g. if you only have quotas for 3 GPU machines, you can only run 3 jobs in parallel. So the concurrent limit can be set to 3 in that case.

<Info>
  ### Please note that if you have to run 100 jobs, triggering 100 job runs at the same time will cost you roughly the same as running 10 jobs in parallel and waiting for 10 such batches to get all the 100 job runs done.

  This will be true in most cases unless the docker image pull time is close to or greater than the job execution time.
</Info>

This can also be a requirement if only one run of a job should run at a single point - this can be true if you are reading / writing from a volume and only one source can be writing at a point to avoid corruption. We can set the concurrency limit to 1 in this case.

If more instances of a job are triggered than the specified concurrency limit, the excess jobs are placed in a queue and executed sequentially once available slots become free.

## Configuring Concurrency Limit for Job

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/Qgrk6xDk6UeaJWPCgL46Z" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/Qgrk6xDk6UeaJWPCgL46Z" style={{ border: "none", display: "flex", margin: "auto" }} />

***
