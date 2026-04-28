> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Autoscaling

> Step-by-step guide for autoscaling overview, explaining configuration, best practices, and real-world usage on TrueFoundry.

While deploying a service, you can provision multiple instances of your container (pod in the kubernetes world) so that the incoming traffic is load-balanced between the multiple instances. We refer to each such instance as a replica in TrueFoundry. While deploying at TrueFoundry, the default value of a replica is 1. While this is fine for development purposes, a single replica is not the best choice for production workloads because of the following reasons:

1. If one replica goes down because of any reason, the entire service goes down. Having 2-3 replicas at minimum helps provide more fault-tolerance.
2. A single replica might not be able to take all of the incoming traffic and hence cause client side errors.

To increase the number of replicas, we can either set it to a fixed higher value or enable autoscaling so that the number of replicas is automatically adjusted based on some configuration of incoming traffic, or cpu usage.

## Setting Fixed Number of Replicas

This can be a good choice if the incoming traffic remains constant and we have a good idea of the number of replicas needed. To set the replicas using the UI, you can follow the steps below:

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/Xzq4uxmpMRkzxaNe7II7r " typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/Xzq4uxmpMRkzxaNe7II7r " style={{ border: "none", display: "flex", margin: "auto" }} />

## Enable Autoscaling

When the traffic or resource usage of the service is not constant, we ideally want the number of replicas to go up and down based on the incoming traffic. We need to define the minimum and maximum number of replicas in this case and the autoscaling strategy will decide what should be the number of replicas between the min and max replicas. We can enable autoscaling based on the following parameters:

1. **CPU Usage:** If the application's cpu usage goes up when traffic increases, this can be a good parameter to autoscale on. For e.g., let's say your application (one replica) consumes 0.3 vCPU on steady state - however, as traffic goes up, the CPU usage starts increasing to a max of 1 CPU. In this case, setting autoscaling to trigger when CPU usage is greater than 0.6 can be a good idea.
2. **Requests Per Second (RPS)**: This is the easiest to reason about and calculate. We can benchmark our service to decide how many requests per second can be served by one replica using benchmarking tools like [Locust](https://docs.locust.io/en/stable/installation.html). Let's say one replica can serve 10 requests per second without an degradation in quality (increase in latency or errors). If we expect the traffic to vary from 50 requests per second to 200 requests per second, we can set the minimum replicas to 5, maximum to 20 and set the rps autoscaling metric to 10.
3. **Time Based Autoscaling**: This can be useful if the traffic patterns shift based on timezones, or you want to shut down dev workloads during non-office hours. For e.g, if you want to scale up replicas only between Monday to Friday 9AM to 9PM, you can set time based scheduling with cron start schedule as `0 9 * * 1-5` and end schedule as `0 21 * * 1-5`.

<Tip>
  In most ML workloads, scaling based on requests per second is the most effective metric of scaling. This is because most ML libraries are optimized to max out CPU and GPU usage.
</Tip>

## Configure Autoscaling for an Application

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/hvP_d4sEaAvx4ow2lcqfS" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/hvP_d4sEaAvx4ow2lcqfS" style={{ border: "none", display: "flex", margin: "auto" }} />

The key things to configure for autoscaling are:

1. **Min Replicas**: The minimum number of replicas to scale down to.
2. **Max Replicas**: The maximum number of replicas to scale up to.
3. **Autoscaling Metric**: The metric to autoscale on.
4. **Autoscaling Value**: The value of the metric to autoscale on.
5. **Polling Interval (Advanced)**: This defines the time interval at which the metrics are checked to decide if the service needs to be scaled up/down. The default values is 30 seconds. You can make it lower to 10 seconds in case you want to scale more quickly.

<Note> Getting the autoscaling metric and value can be tricky and requires us to benchmark the service.
Read below on how to figure this out. </Note>

## How to decide the autoscaling metric value

If you are using **CPU as the autoscaling metric**, it is usually a safe bet to autoscale when CPU usage is greater than 70%. You can further tune this value based on your application's behaviour like time to startup (We might want to scale up at a lower cpu percentage like 60% in this case) or reliability requirements.

If you are using **time as the autoscaling metric**, you can derive the cron expression of the start and end times to scale up and down.

<Tip>
  You can use [this tool](https://crontab.guru/) to derive the cron expression for your use case. LLM's can also provide the answer with a small prompt like:

  > I want to scale up my services every Monday morning at 9Am and shut them down at Friday 6 PM. Can you provide me the start and ending cron expressions?

  The LLM should respond back with Start: 0 9 \* \* 1 and End: 0 18 \* \* 5
</Tip>

For setting up requests/second as the autoscaling metric, we need to benchmark the service to decide the number of requests per second that can be served by one replica. Benchmarking in general is useful because of the following reasons:

* Determining the service's capacity in terms of handling requests per second.
* Assessing the threshold of concurrent requests, indicating the number of users the service can accommodate simultaneously.
* Analyzing how latency fluctuates with increasing traffic volume.

### How to benchmark a service

There are multiple tools to benchmark your service like [Locust](https://locust.io/), [Gatling](https://gatling.io/), [K6](https://k6.io/), [hey](https://github.com/rakyll/hey). This guide will outline **how to use Locust to benchmark your service** - since its quite easy to setup, use and is widely used by datascientists and engineers. You can follow the steps mentioned below:

<Steps>
  <Step title="Install Locust">
    ```
    pip install locust
    ```
  </Step>

  <Step title="Write the Locust File">
    We need to write a simple class which implements a function to send requests to your deployed service. You will need to replace

    ```python locust_benchmark.py lines theme={"dark"}
    from locust import FastHttpUser, task

    class HelloWorldUser(FastHttpUser):
        @task
        def hello_world(self):
            self.client.post(
                # No need to mention host here
                "/path/to/your/endpoint",  
                # Add your request payload here
                json={
                    "model": "<Add your deployed service name>",
                    "prompt": [
                        "This is a test prompt"
                    ],
                    "max_tokens":80,
                    "ignore_eos":True
                }
            )
    ```

    In the code above, we are sending a POST request to the deployed service with the payload mentioned above.

    <Note>
      If your service is deployed at [https://service-name.example.com](https://service-name.example.com) and you want to benchmark the API at `https://<service-name>.example.com/path/to/your/endpoint`, you only need to mention `/path/to/your/endpoint` in the code above.
    </Note>

    You can read this [doc](https://docs.locust.io/en/stable/writing-a-locustfile.html#writing-a-locustfile) to understand more about writing a locust file.
  </Step>

  <Step title="Run the Locust File locally">
    ```
    locust -f locust_benchmark.py
    ```

    Once you run this, you can open [http://localhost:8089](http://localhost:8089) in a browser window. You will find the UI like this:

    <Frame caption="">
      <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/68fa0d06-669e928-image.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=1487511e6565a64abde5f4daeef099f3" width="3024" height="1302" data-path="images/68fa0d06-669e928-image.png" />
    </Frame>

    In the section of host, paste the endpoint of your service: `https://service-name.example.com/` in the example above.
  </Step>

  <Step title="Send requests to the service to benchmark">
    Once you paste the link, you can click on "Start Swarming" after setting the following parameters:

    * **Number of Users:** Number of concurrent users that will bombard your service
    * **Spawn Rate:** If multiple users are selected, the rate at which it will create new users (this can be 1 by default)

    Once you start swarming, you can see the results on the dashboard:

    <Frame caption="">
      <img src="https://mintcdn.com/truefoundry/DdP_2rhue4AQQlob/images/2939f013-8ea70db-image.png?fit=max&auto=format&n=DdP_2rhue4AQQlob&q=85&s=77d22f1c10b67602d35cafab508cf670" width="3024" height="646" data-path="images/2939f013-8ea70db-image.png" />
    </Frame>

    Once this is setup, you can edit and increase the number of users from top by clicking on Edit.
    You can view the detailed charts by clicking on "Charts" tab as shown in the picture. The results look something like this.

    <Frame caption="">
      <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c42eddd9-385fe2e-image.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=22675c6e937fee9d84e48774d1103403" width="3014" height="1712" data-path="images/c42eddd9-385fe2e-image.png" />
    </Frame>
  </Step>

  <Step title="Deploy the locust service on Truefoundry">
    While you can benchmark your service by running Locust locally, its usually recommended to deploy the locust service on the same cluster in which your service is deployed. This helps in reducing the latency from the locust code to your service and helps provide more accurate results.

    <Note>
      The latencies depicted in Locust while running on your local machine might not be accurate since its dependent on your internet speed. Moving it to the cluster helps provide a more stable and fast network.
    </Note>

    You can deploy your locust code as a service on truefoundry. To deploy, place the following 2 files in a directory (Let's call it `locust_service`).

    * `locust_benchmark.py`(shown above)
    * `requirements.txt`

    <CodeGroup>
      ```python locust_benchmark.py lines theme={"dark"}
      from locust import FastHttpUser, task

      class HelloWorldUser(FastHttpUser):
          @task
          def hello_world(self):
              self.client.post(
                  # No need to mention host here
                  "/path/to/your/endpoint",  
                  # Add your request payload here
                  json={
                      "model": "<Add your deployed service name>",
                      "prompt": [
                          "This is a test prompt"
                      ],
                      "max_tokens":80,
                      "ignore_eos":True
                  }
              )
      ```

      ```python requirements.txt lines theme={"dark"}
      locust
      # add other dependencies if used in your locust script
      ```
    </CodeGroup>

    Once you have this ready, please go to TrueFoundry UI and follow the following steps:

    1. Click on `New Deployment` button on top right of your screen.
    2. Select your workspace
    3. Select `Code from Laptop`
    4. Click on `Next`
    5. Follow the guide from the UI
    6. In the deployment form, name the service as `locust-service`, choose   `Python Code (I don't have Dockerfile)` option and set the command as `locust -f locust_benchmark.py`.

    <Frame caption="">
      <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/59bbcc67-fd89e39-image.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=d8f617601c9798f00542991d7d5856c3" width="3024" height="1708" data-path="images/59bbcc67-fd89e39-image.png" />
    </Frame>

    <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/locust-service-deployment.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=2031f3f3454498bb887fadd58ab60253" width="3024" height="1714" data-path="images/locust-service-deployment.png" />

    <Warning>
      Run locust with only 1 replica. Running it with more than 1 replica will not be useful since only 1 pod can run benchmarking and it will cause wierd issues in the Locust UI.
    </Warning>

    Once the deployment of the locust service completes. you can open the endpoint of the locust service and see the same UI that we saw above while running it locally.
  </Step>

  <Step title="Benchmark your service to get the autoscaling metric value">
    Once the deployment of the locust service completes, you can open the endpoint of the locust service and see the same UI that we saw above while running it locally. To benchmark your service (let's call it my-service), follow the guidelines below:

    1. Run 1 replica of service-a. **Its important to benchmark with 1 replica to understand how many requests per second can be served by 1 replica.**
    2. Start with a few users (1-10) and gradually keep increasing it.
    3. We should increase the number of users till either the latency starts increasing beyond our point of tolerance or we start seeing errors from the service.
    4. Observe the CPU and memory usage of the service at this time from the metrics on Truefoundry portal.

    <img src="https://mintcdn.com/truefoundry/5CkapnZ7CyjQJ4bx/images/docs/metrics-benchmarking.png?fit=max&auto=format&n=5CkapnZ7CyjQJ4bx&q=85&s=c0a049c5080197990ae4675f12145a53" width="3024" height="1712" data-path="images/docs/metrics-benchmarking.png" />

    5. The CPU usage should be going up as we increase the number of users and hence requests per second.

    <Tip>
      In most cases, if CPU usage is not going up with requests per second going up, or the service is incurring high latencies or errors even when CPU usage is not close to requested value - it denotes an inefficiency in the code or configuration.
    </Tip>

    6. We can try increasing the cpu requested if we want per pod to allow it to handle more requests per second. Let's say we can handle 10 requests per second with 1 CPU. Going more than 10 requests per seconds results in higher latencies. In this case we can set the autoscaling metric to requests per second and the value to somewhere around 7-8 to have some buffer in each replica.
    7. Set the autoscaling criteria and redeploy the service. Now if you increase the number of users in Locust, you should be able to see the pods autoscale to multiple replicas and requests per second goes above 10.
  </Step>
</Steps>
