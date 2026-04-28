> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Scale Service to 0

> Auto-scale services to zero replicas when idle and automatically restart them on incoming traffic using Elasti.

A lot of services and models that are deployed are not used all the time. A very common pattern is dev services not being used during the weekends or off-office hours. While Truefoundry provides a [pause/resume service](/docs/pause-resume-service) feature, it relies on developers to manually pause the service. This can lead to human errors and cost leakage.

Another option to automatically shut down a service if there are no requests for a period and then automatically start it when a request is received is to use the scale to 0 feature. You can configure that the service will shut down once there are no requests for let's say 10 minutes. After 10 mins of no requests, the service will be scaled to 0. Then if we make a request to the service, the service will automatically be scaled up.

<Warning>
  The start of the service will have a delay (cold start time) since it might first need to download the image and then start the service. This feature is most commonly used and best suited for dev environments.
</Warning>

## Configuring Scale to 0

In the Service deployment form, enable the `Advanced Fields` toggle and then you will see the `Auto Shutdown` section as shown below.

<img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/b8b90ea5-d684b01a06afea52360c6aa6f7d1b7469d69fe13596de47c3833061697e76b31-image.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=21c4e050dd4fe1923f4a6546564e47bb" width="1120" height="611" data-path="images/b8b90ea5-d684b01a06afea52360c6aa6f7d1b7469d69fe13596de47c3833061697e76b31-image.png" />

You will need to configure the idle timeout parameter. This is the number of minutes after which the service will be scaled to 0 if there are no requests.

## How does Scale to 0 work?

Scale to 0 is powered by the [Elasti](https://github.com/truefoundry/elasti). Its an open-source Kubernetes-native solution built by Truefoundry that offers scale-to-zero functionality on Kubernetes when there is no traffic and automatic scale up to 0 when traffic arrives. A brief summary about Elasti is:

> Most Kubernetes autoscaling solutions like HPA or Keda can scale from 1 to n replicas based on cpu utilization or memory usage. However, these solutions do not offer a way to scale to 0 when there is no traffic. Elasti solves this problem by dynamically managing service replicas based on real-time traffic conditions. It only handles scaling the application down to 0 replicas and scaling it back up to 1 replica when traffic is detected again. The scaling after 1 replica is handled by the autoscaler like HPA or Keda.

Elasti uses a proxy mechanism that queues and holds requests for scaled-down services, bringing them up only when needed. The proxy is used only when the service is scaled down to 0. When the service is scaled up to 1, the proxy is disabled and the requests are processed directly by the pods of the service.

<img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/elasti-architecture.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=849a2e852df60883727da2f05f250612" width="1253" height="874" data-path="images/docs/elasti-architecture.png" />
