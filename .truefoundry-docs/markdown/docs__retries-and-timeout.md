> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Set Retries And Timeout

> Configure retry policies and timeout settings for jobs to improve reliability and manage compute resources.

## Retries

Retries allow you to retry a job run if it fails due to temporary errors or conditions. This can be helpful in handling temporary disruptions and improving the overall reliability of your job executions. Consider a scenario where your job communicates with an intermittent service. If the service is unavailable or experiencing issues, retrying the job can increase the likelihood of successful communication.

Another example might involve a job that processes data from an external source. If the source is temporarily unavailable due to maintenance or overload, retries can ensure that the job eventually retrieves the data and proceeds with processing.

By specifying the retries parameter, you can define the maximum number of times the job will be retried before it is marked as failed. The default retry count is 1.

### Configuring Retries for Job

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/QlsGf68HCUG6cG0_UNose" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/QlsGf68HCUG6cG0_UNose" style={{ border: "none", display: "flex", margin: "auto" }} />

## Timeouts

Timeouts establish a maximum duration for a job's execution, ensuring that jobs don't get stuck in an infinite loop or consume excessive resources. By specifying the timeout parameter, you can define the maximum time (in seconds) a job is allowed to run regardless of its retry attempts.

For example, if your machine learning training code typically takes 10 minutes (600 seconds) to train, you can set the timeout to around 15-20 minutes or a more appropriate value. This ensures that if the job gets stuck in an infinite loop or takes an unreasonably long time due to some issue, the timeout will terminate it and you don't incur costs indefinitely.

<Info>
  ### Note:

  '`timeout` will take precedence over the `retries`(retry limit),

  For instance, if you set the timeout to 300 seconds (5 minutes) and retries to 3, the job will terminate after 5 minutes regardless of how many times it attempted to run.
</Info>

### Configuring Timeout for Job

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/clp2poskq3y5apebtik9w3oq3" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/clp2poskq3y5apebtik9w3oq3" style={{ border: "none", display: "flex", margin: "auto" }} />

***
