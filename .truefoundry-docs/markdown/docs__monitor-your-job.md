> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Monitor Your Job

> Monitor job execution using real-time logs, metrics dashboards, and events in the TrueFoundry platform.

TrueFoundry provides you with Logs, Metrics and Events to monitor your deployments to identify and debug issues.

## Browse Logs using UI

Logs are records of events that occur in your deployment, such as training result logs of your job, errors that occur, and system messages. TrueFoundry provides you with logs at 2 levels:

1. **Deployment Level Logs**: This will display aggregated logs from all job runs, including both ongoing and past executions
2. **Job Run / Pod Level Logs:** This allows you to view logs for an individual Job Run and the pods brought to execute that job run.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/axqgcBqu6mMYbfOL3uPyp" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/axqgcBqu6mMYbfOL3uPyp" style={{ border: "none", display: "flex", margin: "auto" }} />

## Metrics Dashboard

Metrics dashboard provides a visual representation of the metrics collected from your job like CPU Usage, GPU Usage, Network Usage etc. This allows you quick and easy way to identify fluctuations and potential performance bottlenecks. TrueFoundry provides you with metrics dashboards at two levels:

1. **Deployment Level Metrics**: This will display aggregated Metrics from all job runs, including both ongoing and past executions
2. **Job Run / Pod Level Metrics:** This allows you to view Metrics for an individual Job Run and the pods brought to execute that job run.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/izmPgnbVAyUGuNd1K52KD" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/izmPgnbVAyUGuNd1K52KD" style={{ border: "none", display: "flex", margin: "auto" }} />

## Events

Events are the occurrences that happen in your TrueFoundry deployment. They can be triggered by a variety of things, a few examples of which are:

* Deploying a new job version
* Starting or stopping a pod
* A pod crashing

Events can be used to track the progress of your deployment and to troubleshoot any problems that may occur. These are the standard Kubernetes events - you probably don't need to look at them often unless you are debugging your pod not starting.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/N45erEOH62c8WwEPBmuNI" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/N45erEOH62c8WwEPBmuNI" style={{ border: "none", display: "flex", margin: "auto" }} />

***
