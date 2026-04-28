> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Sticky Routing

> Pin user requests using consistent hash based routing

Imagine a situation where a user is performing some expensive calculation in a request and wants to do so repeatedly over a bunch of API calls. Some of these expensive calculations might have some common computation that you might want to cache between requests. In such a case, you would like to route all the requests from a specific user to a specific replica.

One such example is **Prefix Caching** with model servers like [vLLM](https://docs.vllm.ai/en/stable/automatic_prefix_caching/apc.html) and [SGLang](https://lmsys.org/blog/2024-01-17-sglang/#backend-automatic-kv-cache-reuse-with-radixattention) where generating the next reply in a converstation can benefit from the cached computation of the conversation history so far and result in significantly lower latencies.

<Frame caption="">
  <img src="https://mintcdn.com/truefoundry/4MAaF__cLD4iud16/images/502eb0f9-da45c05cfc33c9af97408e2c9919c7096403ceed55bcbbf547ceffbc941be8f3-image.png?fit=max&auto=format&n=4MAaF__cLD4iud16&q=85&s=f43b7865c3720a9d18bdc6a2add6855b" width="2081" height="802" data-path="images/502eb0f9-da45c05cfc33c9af97408e2c9919c7096403ceed55bcbbf547ceffbc941be8f3-image.png" />
</Frame>

User 1's requests are always routed to Pod 1 while User 2's requests are always routed to Pod 2

<Warning>
  ### Pods can be transitory and sticky routing is best-effort basis

  It is important to remember that Kubernetes pods and nodes can be short lived and can be re-scheduled due to external events. It is not advised to put any critical stateful logic in Services.

  Considering this, sticky routing is best effort basis. In the event of a pinned pod getting deleted / evicted, all sessions pinned to that pod will be re-assigned randomly to other pods. Similarly, in the event of a scale up or scale down, the sessions **might** be re-assigned to evenly distribute the load.
</Warning>

## Enabling Sticky Routing

To enable sticky routing, first in the Service `labels` section add a `special` label `tfy_sticky_session_header_name`and add a header name against it. For e.g. `x-truefoundry-sticky-session-id`

<Frame caption="Sticky Routing from UI">
  <img src="https://mintcdn.com/truefoundry/OHzlp6GY5G-JfKle/images/c5a520ad-a1d57bc25d620b3f1f5fea0565ac7a028cbdab69ebc1eb718e01c52d26ac468c-image.png?fit=max&auto=format&n=OHzlp6GY5G-JfKle&q=85&s=6e27613ab5bd76f708c151b217ea0871" width="1888" height="1590" data-path="images/c5a520ad-a1d57bc25d620b3f1f5fea0565ac7a028cbdab69ebc1eb718e01c52d26ac468c-image.png" />
</Frame>

### Passing the header when making requests

Now, your clients can send this header with a unique value for a "session". For e.g. this value can be a conversation id or user id, anything that identifies a unique session that can benefit from stickyness.

For e.g.

<CodeGroup>
  ```swift cURL lines theme={"dark"}
  curl -X POST https://my-service.example.com/v1/chat/completions \
      -H 'Content-Type: application/json' \
      -H 'x-truefoundry-sticky-session-id: session-id-qnfjk' \
      -d '{"messages": [{"role": "user", "content": "Hi"}]}'
  ```

  ```python Python lines theme={"dark"}
  import requests

  response = requests.post(
      "https://my-service.example.com/v1/chat/completions",
      headers={"x-truefoundry-sticky-session-id": "session-id-qnfjk"},
      json={"messages": [{"role": "user", "content": "Hi"}]}
  }
  response.raise_for_status()
  print(response.json())
  ```

  ```python OpenAI SDK lines theme={"dark"}
  from openai import OpenAI

  client = OpenAI(
      base_url="https://my-service.example.com/v1",
      api_key="<YOUR API KEY>",
      default_headers={"x-truefoundry-sticky-session-id": "session-id-qnfjk"},
  )
  ```
</CodeGroup>

After the first request, any request that has header `x-truefoundry-sticky-session-id: session-id-qnfjk` will be routed to the same pod the first request was routed to.

<Info>
  ### What happens if header is not added?

  The request will be routed according to default load balancing policy (random / round robin / least load)
</Info>

***
