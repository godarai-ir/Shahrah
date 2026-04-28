> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Pause / Resume Service

> Pause services to shut down all replicas and save costs, then resume them without losing your configuration.

Pausing a service means that all replicas of your Service will be shut down, and your service will no longer consume any resources or serve requests - hence saving costs. This can be useful for dev environments when you don't want to run the service but also don't want to delete the configuration since it might be used later.

### Pause a Service

<iframe href="https://app.supademo.com/embed/VQPzzhrnf2mXxm1U7tGt9" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/VQPzzhrnf2mXxm1U7tGt9" style={{ border: "none", display: "flex", margin: "auto" }} />

<Warning> If you have Autoscaling enabled for your service, the option to pause a service will not be available to you. </Warning>

### Resume a Service

Resume brings up all the replicas of the service that were originally configured. So if a service was initially configured with 5 replicas and then paused, resuming will bring back all the 5 pods of the service.

<iframe href="https://app.supademo.com/embed/thaU0fkjhhStGv2LjCs0E " typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/thaU0fkjhhStGv2LjCs0E " style={{ border: "none", display: "flex", margin: "auto" }} />
