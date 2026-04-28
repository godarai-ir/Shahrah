> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Redirect And Mirror Traffic

> Understand intercepts and how they modify request and execution flow.

Intercepts provide a way to attach ad-hoc routing rules for a service deployed on the TrueFoundry platform. We can use it to redirect traffic arriving at a service to another one on the basis of a header. It can be useful to test a new release for a service by intercepting the traffic on the main service and redirecting it to the the copied service. This is useful in development purposes for rapid debugging specially if a lot of microservices are involved. Let's consider the scenario below:

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/yRoKH_fkKi2nPtuV/images/f6fba865-a07b8ee-image.png?fit=max&auto=format&n=yRoKH_fkKi2nPtuV&q=85&s=540cba28cd3b380a50dcfcf37013183e" width="1471" height="722" data-path="images/f6fba865-a07b8ee-image.png" />
</Frame>

Let's imagine we are working on the model microservice and currently the v1 version is deployed. While we can test the model microservice separately by issuing API calls, sometimes its more convenient to check the end to end flow with a new version of the model. To test this, we can swap the model-v1 with model-v2, but this will affect the experience of everyone in case model-v2 doesn't work well. Or we need to duplicate frontend, backend and db to be able to test the entire stack separately. This can become very cost intensive and also a lot of manual work - specially if a lot of developers are testing things independently.

Intercepts provide an easy solution to the problem by redirect a small percentage of the traffic to the model-v2 or redirect based on headers in the incoming request.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c45c7857-bad8e02-image.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=92fef11843e92114959285de827e30c7" width="1472" height="996" data-path="images/c45c7857-bad8e02-image.png" />
</Frame>

In this case, we can just make a request from the existing frontend with some header like MODEL\_VERSION: v2 and configure the interceptor to redirect all requests going to model-v1 with that header to redirect to model-v2. This way the other requests to frontend continue to get routed to model-v1 and only the request that we are testing end up going to model-v2. This can be useful to test end to end scenarios without redeploying the entire stack for every single developer.

## Using UI

1. Before intercepts can be enabled for a service, we need to allow intercepts for that service.

<iframe href="https://app.supademo.com/embed/-NEDH1pMYvatwRsmmRbpY" typeofembed="iframe" height="475px" width="100%" provider="app.supademo.com" src="https://app.supademo.com/embed/-NEDH1pMYvatwRsmmRbpY" style={{ border: "none", display: "flex", margin: "auto" }} />

2. We now can start adding intercept rules

<iframe href="https://app.supademo.com/embed/AEINnszejpoMHpmKZrQmQ " typeofembed="iframe" height="475px" width="100%" provider="app.supademo.com" src="https://app.supademo.com/embed/AEINnszejpoMHpmKZrQmQ " style={{ border: "none", display: "flex", margin: "auto" }} />

3. To test the intercept you can send a request to the service with the header set to the value in intercept rule and verify that the request indeed lands on the service set as destination. Eg - `curl -H "<header name>: <header value>" https://<service endpoint>` . If the intercept rules don't match, the request will continue to go to the original service.

<Info>
  ### Note:

  A service if associated with an intercept cannot be deleted until the intercept is deleted first
</Info>

***
