> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create a Virtual Model

> Step-by-step guide to creating virtual models, configuring targets, and using them in your application

This page walks through creating a virtual model in the TrueFoundry dashboard and using it from your application. For an overview of routing strategies, health detection, and advanced features like sticky routing, see the [Virtual Models overview](/docs/ai-gateway/virtual-model).

## Create a virtual model

<Steps>
  <Step title="Open Virtual Models in AI Gateway">
    From the TrueFoundry dashboard, go to **AI Gateway** → **Models** → **Virtual Model**.

    <Frame caption="Navigate to Virtual Models in AI Gateway">
      <img src="https://mintcdn.com/truefoundry/HhWQUTTSjIJ8Poha/images/virtual-model-navigate-1.jpeg?fit=max&auto=format&n=HhWQUTTSjIJ8Poha&q=85&s=a8725c93b970f98d71abea7f7c753192" alt="Navigate to Virtual Models in AI Gateway" width="1280" height="529" data-path="images/virtual-model-navigate-1.jpeg" />
    </Frame>

    <Note>
      Virtual models live inside **Virtual Model Provider Groups**. You can add models to an existing group or create a new group when you start.
    </Note>
  </Step>

  <Step title="Create or select a provider group and set access controls">
    Give the group a unique name (3–64 characters, alphanumeric and hyphens, cannot start with a number). Configure collaborators:

    * **User** — May call the virtual models in this group for inference.
    * **Manager** — May change virtual model configuration.

    <Frame caption="Create Virtual Model Provider Group and configure access controls">
      <img src="https://mintcdn.com/truefoundry/HhWQUTTSjIJ8Poha/images/virtual-model-navigate-2.jpeg?fit=max&auto=format&n=HhWQUTTSjIJ8Poha&q=85&s=2f7150510199688619fa991296092786" alt="Create Virtual Model Provider Group and configure access controls" width="1280" height="656" data-path="images/virtual-model-navigate-2.jpeg" />
    </Frame>

    See [Gateway access control](/docs/ai-gateway/gateway-access-control) for details.
  </Step>

  <Step title="Define the virtual model, strategy, and targets">
    For each virtual model in the group, set:

    * **Name** — Identifier used in the full path `group-name/virtual-model-name` (for example `gpt-4-production`).

    * **Model types** — Operation kinds this virtual model supports (chat, completion, embedding, and so on). All targets must support the operation you invoke.

    * **Routing strategy** — Choose one of three strategies:

      | Strategy           | When to use                                                                       |
      | ------------------ | --------------------------------------------------------------------------------- |
      | **Weight-based**   | Canaries, fixed capacity splits, A/B allocation. Assign weights that sum to 100.  |
      | **Latency-based**  | Automatic performance chasing. No weights needed — the gateway picks the fastest. |
      | **Priority-based** | Primary + backup topologies. Assign priority numbers (0 = highest).               |

      For how each strategy works, see the [overview](/docs/ai-gateway/virtual-model#routing-strategies).

    * **Target models** — For each target, configure:

      | Field                     | Description                                                                                                                                                                                 |
      | ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
      | **Target**                | A real model from the catalog (not another virtual model).                                                                                                                                  |
      | **Weight**                | Traffic share (weight-based only). Weights across targets should sum to 100.                                                                                                                |
      | **Priority**              | Priority level (priority-based only). Lower number = higher priority.                                                                                                                       |
      | **Retry config**          | Attempts, delay (ms), and status codes that trigger retries. Defaults: 2 attempts, 100 ms delay, retry on `429`, `500`, `502`, `503`.                                                       |
      | **Fallback status codes** | HTTP codes that cause fallback to another target. Defaults: `401`, `403`, `404`, `429`, `500`, `502`, `503`.                                                                                |
      | **Fallback candidate**    | Whether this target may receive traffic when another target fails. Default: `true`.                                                                                                         |
      | **Override parameters**   | Per-target request parameters like `temperature`, `max_tokens`, or `prompt_version_fqn` for [model-specific prompts](/docs/ai-gateway/virtual-model#model-specific-prompt-overrides).       |
      | **Header overrides**      | Inject or remove HTTP headers for this target only. Use `set` to add/overwrite headers and `remove` to strip them. See [header overrides](/docs/ai-gateway/virtual-model#header-overrides). |

          <Note>
            `prompt_version_fqn` override does not apply when using agents with MCP/tools; it is supported for standard chat completion requests.
          </Note>

    * **Slug** (optional) — Short global alias for this virtual model. See [Slugs](#virtual-model-slugs).

    <Frame caption="Configure virtual model details, routing strategy, and target models">
      <img src="https://mintcdn.com/truefoundry/HhWQUTTSjIJ8Poha/images/virtual-model-navigate-3.jpeg?fit=max&auto=format&n=HhWQUTTSjIJ8Poha&q=85&s=93e96647ec4e21e08f5ea4ce904ca04e" alt="Configure virtual model details, routing strategy, and target models" width="1052" height="585" data-path="images/virtual-model-navigate-3.jpeg" />
    </Frame>

    <Note>
      Configure the slug in the Virtual Model Provider Group settings. Slugs must be **unique across all virtual models in the tenant**.
    </Note>
  </Step>
</Steps>

## Common patterns

The following YAML sketches show the `routing_config` shape used inside a virtual model. In the dashboard, the same fields are set in the UI.

<AccordionGroup>
  <Accordion title="Priority chain — fail over when rate limited">
    ```yaml theme={"dark"}
    routing_config:
      type: priority-based-routing
      load_balance_targets:
        - target: azure/gpt4
          priority: 0
          fallback_status_codes: ["429"]
        - target: openai/gpt4
          priority: 1
          fallback_status_codes: ["429"]
        - target: anthropic/claude-3-opus
          priority: 2
    ```
  </Accordion>

  <Accordion title="Canary rollout with weights">
    ```yaml theme={"dark"}
    routing_config:
      type: weight-based-routing
      load_balance_targets:
        - target: azure/gpt4-v1
          weight: 90
        - target: azure/gpt4-v2
          weight: 10
    ```
  </Accordion>

  <Accordion title="On-prem primary with cloud fallback">
    ```yaml theme={"dark"}
    routing_config:
      type: priority-based-routing
      load_balance_targets:
        - target: onprem/llama
          priority: 0
          fallback_status_codes: ["429", "500", "502", "503"]
        - target: bedrock/llama
          priority: 1
          retry_config:
            attempts: 2
            delay: 100
    ```
  </Accordion>

  <Accordion title="Latency race with limited retries per target">
    ```yaml theme={"dark"}
    routing_config:
      type: latency-based-routing
      load_balance_targets:
        - target: azure/gpt4
          retry_config:
            attempts: 1
        - target: openai/gpt4
          retry_config:
            attempts: 1
    ```
  </Accordion>

  <Accordion title="Different prompt versions per provider">
    ```yaml theme={"dark"}
    routing_config:
      type: weight-based-routing
      load_balance_targets:
        - target: openai/gpt4
          weight: 70
          override_params:
            prompt_version_fqn: chat_prompt:internal/my-app/gpt4-optimized-prompt:1
        - target: anthropic/claude-3-opus
          weight: 30
          override_params:
            prompt_version_fqn: chat_prompt:internal/my-app/claude-optimized-prompt:1
    ```
  </Accordion>

  <Accordion title="Sticky routing for multi-turn conversations">
    ```yaml theme={"dark"}
    routing_config:
      type: weight-based-routing
      sticky_routing:
        ttl_seconds: 3600
        session_identifiers:
          - key: x-user-id
            source: headers
          - key: x-conversation-id
            source: headers
      load_balance_targets:
        - target: provider-a/model-a
          weight: 70
          fallback_candidate: true
        - target: provider-b/model-b
          weight: 30
          fallback_candidate: true
    ```
  </Accordion>

  <Accordion title="Region-based routing using SaaS gateway metadata">
    Route to region-specific model deployments based on which SaaS gateway handled the request. The gateway automatically adds `tfy_gateway_region` and `tfy_gateway_zone` to request metadata — no client changes needed.

    ```yaml theme={"dark"}
    routing_config:
      type: priority-based-routing
      load_balance_targets:
        - target: azure-us/gpt-4o
          priority: 0
          metadata_match:
            tfy_gateway_region: US
        - target: azure-eu/gpt-4o
          priority: 0
          metadata_match:
            tfy_gateway_region: EU
        - target: openai/gpt-4o
          priority: 1
    ```

    US gateway traffic goes to the Azure US deployment, EU traffic to Azure EU, and everything else falls back to OpenAI. See [Metadata Keys](/docs/ai-gateway/globally-distributed-saas#metadata-keys) for all available region and zone values.
  </Accordion>

  <Accordion title="Per-target header overrides">
    Inject or remove headers on specific targets — useful when one provider needs extra headers the others don't.

    ```yaml theme={"dark"}
    routing_config:
      type: weight-based-routing
      load_balance_targets:
        - target: azure/gpt-4o
          weight: 80
          headers_override:
            set:
              x-deployment-id: gpt4o-eastus
              api-version: "2024-06-01"
            remove:
              - x-internal-trace
        - target: openai/gpt-4o
          weight: 20
    ```
  </Accordion>

  <Accordion title="Metadata filtering with enterprise tier routing">
    Route enterprise-tier traffic to a dedicated deployment while standard traffic uses a shared pool.

    ```yaml theme={"dark"}
    routing_config:
      type: weight-based-routing
      load_balance_targets:
        - target: azure-dedicated/gpt-4o
          weight: 100
          metadata_match:
            tier: enterprise
        - target: openai/gpt-4o
          weight: 60
        - target: azure-shared/gpt-4o
          weight: 40
    ```

    Requests with `x-tfy-metadata: {"tier":"enterprise"}` go exclusively to the dedicated Azure deployment. All other requests are split 60/40 between OpenAI and the shared Azure deployment.
  </Accordion>

  <Accordion title="Priority chain with SLA cutoff and retries">
    Full configuration using priority-based routing with SLA thresholds, custom retries, and controlled fallback eligibility.

    ```yaml theme={"dark"}
    routing_config:
      type: priority-based-routing
      load_balance_targets:
        - target: azure/gpt-4o
          priority: 0
          sla_cutoff:
            time_per_output_token_ms: 50
          retry_config:
            attempts: 3
            delay: 200
            on_status_codes: ["429", "500", "503"]
          fallback_status_codes: ["429", "500", "502", "503"]
          fallback_candidate: true
        - target: openai/gpt-4o
          priority: 1
          retry_config:
            attempts: 2
            delay: 100
          fallback_candidate: true
        - target: anthropic/claude-sonnet
          priority: 2
          fallback_candidate: false
    ```
  </Accordion>
</AccordionGroup>

### Environment- or segment-specific routing

Use **different virtual model names** per environment or segment (for example `booking-app/gpt-prod` vs `booking-app/gpt-dev`) and have your client pass the appropriate `model`. You can still send metadata and headers for observability, rate limits, and other gateway features; routing for a given virtual model name is always defined on that virtual model.

<Tip>
  SaaS gateway metadata keys (`tfy_gateway_region`, `tfy_gateway_zone`) are available in request metadata and can be used for metadata-driven virtual model routing rules. See [Metadata Keys](/docs/ai-gateway/globally-distributed-saas#metadata-keys) for all available values and [metadata-based target filtering](/docs/ai-gateway/virtual-model#metadata-based-target-filtering) for worked examples.
</Tip>

## Use a virtual model from your application

Once created, use the full path **`virtual-model-group-name/virtual-model-name`** as the `model` value in API requests — it works like any other model in the gateway.

### Try in the Playground

* Click **Try in playground** on the virtual model row after creation, or
* Open the Playground and pick the virtual model from the model dropdown.

<Frame caption="Try in playground button next to virtual model">
  <img src="https://mintcdn.com/truefoundry/HhWQUTTSjIJ8Poha/images/virtual-model-play-button.jpeg?fit=max&auto=format&n=HhWQUTTSjIJ8Poha&q=85&s=21e1aa922e4d6ec0e2352881d4efc2c7" alt="Try in playground button next to virtual model" width="1280" height="998" data-path="images/virtual-model-play-button.jpeg" />
</Frame>

<Frame caption="Select virtual model from playground dropdown">
  <img src="https://mintcdn.com/truefoundry/HhWQUTTSjIJ8Poha/images/virtual-model-playground-dropdown.jpeg?fit=max&auto=format&n=HhWQUTTSjIJ8Poha&q=85&s=09b65310f6af9ce2d7d7d85af2e226d8" alt="Select virtual model from model dropdown in playground" width="1280" height="576" data-path="images/virtual-model-playground-dropdown.jpeg" />
</Frame>

## Virtual model slugs

Slugs are optional short names that refer to a single virtual model. They are unique **across the tenant**. You can use either the slug or the full `group/model` path in requests.

<Frame caption="Configure slug in Virtual Model Provider Group settings">
  <img src="https://mintcdn.com/truefoundry/kSjZe8MkhpWsNvu-/images/Screenshot2025-12-09at11.12.34PM.png?fit=max&auto=format&n=kSjZe8MkhpWsNvu-&q=85&s=ad87fbfd68a2f50e8ab809e80e288853" alt="Configure slug in Virtual Model Provider Group settings" width="2522" height="2008" data-path="images/Screenshot2025-12-09at11.12.34PM.png" />
</Frame>

If the virtual model is `my-first-virtual-account/model-1` with slug `virtual-model-1`, both bodies are valid:

```json theme={"dark"}
{
  "model": "my-first-virtual-account/model-1",
  "messages": []
}
```

```json theme={"dark"}
{
  "model": "virtual-model-1",
  "messages": []
}
```
