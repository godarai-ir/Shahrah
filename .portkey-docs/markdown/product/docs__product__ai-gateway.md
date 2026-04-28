> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# AI Gateway

> The world's fastest AI Gateway with advanced routing & integrated Guardrails.

## Features

<CardGroup cols={3}>
  <Card title="Universal API" href="/product/ai-gateway/universal-api">
    Use any of the supported models with a universal API (REST and SDKs)
  </Card>

  <Card title="Cache (Simple & Semantic)" href="/product/ai-gateway/cache-simple-and-semantic">
    Save costs and decrease latencies by using a cache
  </Card>

  <Card title="MCP Support" href="/product/ai-gateway/remote-mcp">
    Connect to Remote MCP severs, allowing you to connect external tools and data sources.
  </Card>

  <Card title="Fallbacks" href="/product/ai-gateway/fallbacks">
    Fallback between providers and models for resilience
  </Card>

  <Card title="Conditional Routing" href="/product/ai-gateway/conditional-routing">
    Route to different targets based on custom conditional checks
  </Card>

  <Card title="Multimodality" href="/product/ai-gateway/multimodal-capabilities">
    Use vision, audio, image generation, and more models
  </Card>

  <Card title="Automatic Retries" href="/product/ai-gateway/automatic-retries">
    Setup automatic retry strategies
  </Card>

  <Card title="Circuit Breaker" href="/product/ai-gateway/circuit-breaker">
    Configure per-strategy circuit protection and failure handling
  </Card>

  <Card title="Load Balancing" href="/product/ai-gateway/load-balancing">
    Load balance between various API Keys to counter rate-limits
  </Card>

  <Card title="Canary Testing" href="/product/ai-gateway/canary-testing">
    Canary test new models in production
  </Card>

  <Card title="gRPC (Beta)" href="/product/ai-gateway/grpc">
    Use gRPC transport for lower latency and efficient binary serialization
  </Card>

  <Card title="Request Timeout" href="/product/ai-gateway/request-timeouts">
    Easily handle unresponsive LLM requests
  </Card>

  <Card title="Budget Limits" href="/product/model-catalog/integrations#3-budget-%26-rate-limits">
    Set usage limits based on costs incurred or tokens used
  </Card>

  <Card title="Rate Limits" href="/product/model-catalog/integrations#3-budget-%26-rate-limits">
    Set hourly, daily, or per minute rate limits on requests or tokens sent
  </Card>

  <Card title="Custom Hosts" href="/product/ai-gateway/custom-hosts">
    Route requests to privately hosted or local models using custom host URLs
  </Card>
</CardGroup>

## Using the Gateway

The various gateway strategies are implemented using Gateway configs. You can read more about configs below.

<Card title="Configs" href="/product/ai-gateway/configs" />

## Open Source

We've open sourced our battle-tested AI gateway to the community. You can run it locally with a single command:

```sh theme={"system"}
npx @portkey-ai/gateway
```

[**Contribute here**](https://github.com/portkey-ai/gateway).

While you're here, why not [give us a star](https://git.new/ai-gateway-docs)? It helps us a lot!

<Frame>
  <img src="https://mintcdn.com/portkey-docs/VWP2Y8zxPP5N4jE6/images/product/ai-gateway/ai-10-gateway.png?fit=max&auto=format&n=VWP2Y8zxPP5N4jE6&q=85&s=74d524a588b870420b7de2f13a1267e3" width="1280" height="640" data-path="images/product/ai-gateway/ai-10-gateway.png" />
</Frame>

You can also [self-host](https://github.com/Portkey-AI/gateway/blob/main/docs/installation-deployments.md) the gateway and then connect it to Portkey. Please reach out on [hello@portkey.ai](mailto:hello@portkey.ai) and we'll help you set this up!
