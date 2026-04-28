> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Truefoundry Gateway Plane Architecture

> Architecture of the TrueFoundry AI Gateway, designed for high availability, low latency, and scalable LLM integration in production environments.

The gateway plane comprises of one or more deployments of the Truefoundry AI Gateways.

The AI Gateway is like a proxy between your application and the LLM provider. It lies in the critical path of production traffic and hence needs to be highly available and performant. TrueFoundry AI gateway has been designed to be a low-latency, high throughput AI gateway.

The gateway is written using the [Hono](https://hono.dev/) framework which is ultra-fast, minimalistic and is designed for the edge. The architecture of gateway is based on the following principles:

1. There are no external calls in the pathway of executing a request from the client to the LLM provider (unless we are using cache).
2. All checks for ratelimits, loadbalancing, authentication and authorization are done in memory.
3. The logs and metrics are written to a queue in an async manner.
4. Gateway never fails a request even if the external queue is down.
5. The gateway is designed to be horizontally scalable and CPU bound so that it scaled nicely as number of requests increases.
6. It has a control-plane and proxy separation so that multiple gateways can be deployed cross-region and configuration can be managed from a single place.

<Note>
  The gateway plane is completely stateless and subscribes to the control-plane queue to fetch the configuration. If the queue is down, the gateway will try to fetch the configuration from the control-plane backend service via HTTP request.
</Note>

<Frame caption="TrueFoundry AI Gateway Architecture">
  <img src="https://mintcdn.com/truefoundry/Gz34IEy99_S9XhHC/images/docs/platform/gateway-plane-architecture.png?fit=max&auto=format&n=Gz34IEy99_S9XhHC&q=85&s=69c594438268a63376d17c818ddffc67" alt="TrueFoundry AI Gateway Architecture" width="3972" height="2327" data-path="images/docs/platform/gateway-plane-architecture.png" />
</Frame>

The key components in the above diagram are:

1. **Global Autentication/Licensing Server**: This is a global licensing/authentication server. It is used to authenticate every user who logs into the control plane. In case, you are installing the control-plane on your end, the Global Authentication server remains within Truefoundry.

   <Note>
     The only information that flows to global licensing server is the emails of the employees logging into the control-plane and the number of requests flowing through the gateway.
   </Note>
2. **Control-Plane**: This is the control plane that manages the configuration for the gateway like models, virtual models, users, teams, virtual accounts, rate-limiting config, legacy routing configuration, and related settings. The configuration is synced to the gateway via the NATS queue.
3. **Gateway Plane**: The gateway plane handles all the requests from the users/applications to the LLM/MCP servers.

## Request Flow

Here's how a LLM request flows through the gateway:

<img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/ai-gateway-request-flow-dia-2.png?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=3544ba2cabb27b90d902cd55382fffaa" alt="Diagram illustrating the request flow through the AI Gateway, from user to LLM provider and back" width="6179" height="3992" data-path="images/ai-gateway-request-flow-dia-2.png" />

The key steps in the request flow are:

<Steps>
  <Step title="User makes a request to the AI Gateway">
    The user will make a request to the LLM gateway with a valid JWT key and in the OpenAI compatible API format.
  </Step>

  <Step title="Gateway validates the request">
    The gateway will validate the request to check if the JWT token is valid. It also validated if the user has access to the model. The gateway subscribes to all the authorization rules and updates from control-plane and performs in-memory checks to ensure that the request is authenticated and authorized.
  </Step>

  <Step title="Gateway checks for budget-limiting and rate-limiting rules">
    The gateway will check for rate-limiting and budget-limiting rules to ensure that the user is not exceeding any of the threshold limits.
  </Step>

  <Step title="Gateway selects the model based on the load-balancing rules">
    Gateway then checks the load-balancing rules to select the appropriate model for the request.
  </Step>

  <Step title="Router/Adapter translates the request to the appropriate model">
    Since the Gateway unifies the API of different LLM providers, it needs to translate the request to the format expected by the target LLM model. It then forwards the translated request to the model.
  </Step>

  <Step title="Handle Response from the model">
    If the model responds with a successful response, the response if forwarded to the user. The request data is sent for logging and metrics collection to the queue in an async manner. If the model responds with an error, the gateway will check if any fallback model is configured. It will then retry the request to the fallback model.
  </Step>

  <Step title="Log the request and response">
    The request and response will be stored in the queue from where it is picked by the backend service and inserted in Clickhouse database (which is backed by BlobStorage). **The aggregation service calculates the metrics required for rate-limiting and budget-limiting**. The metrics are sent back to the queue to which the gateway subscribes. Hence, all the gateway pods refresh the updated rate-limiting, latency and budget metrics in an async way.
  </Step>
</Steps>

## Gateway Performance and Benchmarks

The gateway has been benchmarked to handle 250 RPS on a single pod with 1 CPU and 1 GB of RAM. The key results of the benchmark are:

**Benchmarking Results:**

* **High Scalability**: AI Gateway can scale without any degradation in performance until about 350 RPS with 1 vCPU & 1 GB machine before the CPU utilisation reaches 100% and latencies start to get affected. With more CPU or more replicas, the LLM Gateway can scale to tens of thousands of requests per second.
* **Minimal Overhead**: TrueFoundry AI Gateway is continuously optimized and benchmarked to add as little latency as possible.

| RPS       | Tracing Disabled | Tracing Enabled (w/o Request Logs) | Complete Tracing |
| --------- | ---------------- | ---------------------------------- | ---------------- |
| 200 - 220 | +7 ms            | +7 ms                              | +8 ms            |
| 350 - 370 | +7 ms            | +7 ms                              | +12 ms           |

* There is minimal overhead on the gateway even after applying multiple authorization rules, load-balancing and rate-limiting configs.

<Info>
  ### 🚀 Learn more on AI Gateway Benchmarks here: [Read more](https://www.truefoundry.com/blog/truefoundry-llm-gateway-is-blazing-fast)
</Info>

## FAQ

<AccordionGroup>
  <Accordion title="What role does NATS play in the architecture?">
    [NATS](https://nats.io/) is used as the message broker between the control-plane and the gateway pods. The control-plane publishes all updates to this queue which is instantly propagated to all the gateway pods. This data includes the users, authorization data between users and models, the load-balancing configs, and the aggregated data for each segment and model for the gateway to perform ratelimiting with in-memory checks. NATS also serves as a cache to store the data temporarily since it provides [key-value store with TTL functionality](https://docs.nats.io/nats-concepts/jetstream/key-value-store).

    After every request that passes through the gateway, the gateway pod publishes an event to NATS containing the input tokens, output tokens, latency, cost, model name, etc. The control-plane has an aggregator service that subscribes to NATS for all these updates. It then calculates aggregates of latency per model, total cost per user, model, team or any other custom segment defined by the user in the Loadbalancing / ratelimiting configs. The aggregated data is then again published to NATS by the aggregator service. The Gateway pods subscribe to NATS for this aggregated data. The aggregated data is updated in the in-memory store of the gateway pods, using which all the checks for ratelimiting, load-balancing, and budget-limiting are performed.
  </Accordion>

  <Accordion title="How does the Gateway enforce rate-limiting?">
    Truefoundry AI gateway uses the [Sliding Window Token Bucket](https://blog.algomaster.io/p/rate-limiting-algorithms-explained-with-code) algorithm to enforce rate-limiting. Since the minimum unit of rate-limiting is on a per-minute window for LLM traffic, the gateway maintains a sliding window of 60 seconds to track the number of requests. The gateway also maintains a token bucket for each user, model, team, or any other custom segment defined by the user in the Loadbalancing / ratelimiting configs for every 5 seconds. So to calculate the counter in the last 50 seconds, it sums up the tokens in the last 12 buckets. The older buckets are removed from the sliding window.
  </Accordion>

  <Accordion title="How is the Gateway updated about the configuration changes in control-plane?">
    The gateway uses the control-plane to fetch all the configuration data for models, users, virtual accounts, ratelimiting, routing configs, SSO configuration when it starts up. It also subscribes to the NATS queue to remain updated about any configuration changes realtime. If the NATS queue is down for some reason and the gateway cannot connect to it, it also tries to get the configuration from the control-plane backend service via HTTP request. If this also fails, the gateway will not startup and the readiness probe of gateway pods will fail - which will instruct Kubernetes to not direct any traffic to the gateway pod. **Getting the configuration from control-plane via the REST API call is implemented as an additional fail-safe mechanism in case the control-plane is up partially.**

    If NATS queue is up, the gateway will sync configs in realtime. If NATS queue is down, and backend service is up, the sync will happen only at startup and no further sync will happen till NATS comes back online. Once it comes back up, all configuration will be automatically synced to Gateway and it will be fully reconciled with th control-plane configuration.
  </Accordion>

  <Accordion title="Will gateway continue to work if control-plane is down?">
    The gateway continues to work with whatever configuration it has already fetched from the control-plane and doesn't go down even if the control-plane is down. Once control-plane is back up, all the configuration that hasn't synced automatically syncs up with the gateway. If the gateways pods restart while the control-plane is down and both NATS and backend service are down in the control-plane, the gateway pod that has restarted will not startup. **We recommend to keep multiple pods of gateway in a deployment so that the chances of all of them restarting when the control-plane is down are minimal**. The gateway pods' readiness probe succeeds only when all configuration has been synced from the control-plane - before that, no traffic is directed by Kubernetes to the gateway pod.
  </Accordion>

  <Accordion title="Is there a chance of configuration drift between control-plane and gateway?">
    We use NATS with at least once guarantee. The payload in NATS is idempotent - which means that the entire state of configuration is published from the control-plane to NATS. So publishing the same payload multiple times has no impact on the state of the configuration and hence gateway functionality.

    However, in some failure scenarios, its possible that gateway misses one of the updates from the control-plane. To mitigate this case further, the control-plane publishes the entire configuration every 10 mins. This makes sure that all gateways are synced every 10 mins even if one intermediate update was missed. This ensures eventual consistency and makes sure that there is no configuration drift between control-plane and gateway.
  </Accordion>

  <Accordion title="How does authentication and authorization work in the gateway?">
    The gateway syncs the SSO configuration from the control-plane via NATS. Once the gateway knows the Issuer URl, we download the public keys to verify the JWT token and cache them in the gateway pod. Every request that hits the gateway is verified using the public key cached and hence doesn't require any external calls to verify the JWT token. For authorization, the entire map of users, models and their associations are kept in memory in the gateway pod. The authorization check is done in memory and hence doesn't require any external calls. Since users and models are essentially small strings, this doesn't take any significant memory and the authorization checks are very fast.
  </Accordion>

  <Accordion title="How to configure forward proxy for air-gapped clusters?">
    In air-gapped environments where the gateway cannot directly access the internet, you can configure a forward proxy to enable external connectivity. The forward proxy acts as an intermediary that routes all external requests (model calls, MCP requests, guardrails) from the gateway to the internet.

    <Frame caption="Air-gapped Environment with Forward Proxy Architecture">
      <img src="https://mintcdn.com/truefoundry/hSvkwO1g5HqmgwD5/images/Screenshot2025-12-03at6.33.34PM.png?fit=max&auto=format&n=hSvkwO1g5HqmgwD5&q=85&s=fa713c4de8b5cbd135519d3239f21afb" alt="Air-gapped Environment with Forward Proxy Architecture" width="2024" height="1050" data-path="images/Screenshot2025-12-03at6.33.34PM.png" />
    </Frame>

    ## Configure Gateway Environment Variables

    <Note>
      Ensure you have a forward proxy deployed and accessible before configuring these environment variables.
    </Note>

    Set these environment variables in your Gateway deployment:

    ```bash theme={"dark"}
    HTTP_PROXY=http://forward-proxy.namespace.svc.cluster.local:3128
    HTTPS_PROXY=http://forward-proxy.namespace.svc.cluster.local:3128
    NO_PROXY=servicefoundry-server.truefoundry.svc.cluster.local,nats.truefoundry.svc.cluster.local
    ```

    <Info>
      * `HTTP_PROXY`: Proxy endpoint URL for HTTP requests
      * `HTTPS_PROXY`: Proxy endpoint URL for HTTPS requests
      * `NO_PROXY`: Comma-separated list of endpoints (hostnames, domains, or IPs) that should bypass the proxy
      * For same-cluster proxy: Use Kubernetes service DNS format: `<service-name>.<namespace>.svc.cluster.local:<port>`
      * For external proxy: Use the exposed proxy URL and ensure TCP traffic is allowed from gateway pods to the proxy endpoint
    </Info>

    <Accordion title="I don't have a forward proxy. How do I deploy one?">
      If you don't have an existing forward proxy, you can use the [TrueFoundry Squid Proxy Helm chart](https://github.com/truefoundry/infra-charts/tree/main/charts/tfy-squid-proxy) to deploy a Squid proxy.

      Example Squid configuration:

      ```squid theme={"dark"}
      # Listens on 3128
      http_port 3128

      # No caching
      cache deny all

      # Allow all by default
      http_access allow all
      ```

      <Note>
        Customize the Squid config based on your security requirements (authentication, domain restrictions).
      </Note>
    </Accordion>
  </Accordion>
</AccordionGroup>
