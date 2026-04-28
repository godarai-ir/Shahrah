> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# CrowdStrike Guardrails Integration

> Integrate CrowdStrike AIDR with TrueFoundry AI Gateway using Custom Guardrails and the guard_chat_completions endpoint.

This guide explains how to integrate CrowdStrike AIDR guardrails with TrueFoundry AI Gateway using the **Custom Guardrails** route.

<Info>
  This page replaces the older native CrowdStrike/Pangea integration flow. The recommended approach is now to use **Custom Guardrails** with a thin adapter service that calls CrowdStrike AIDR.
</Info>

## CrowdStrike AIDR endpoint used

* **Base URL**: `https://api.crowdstrike.com/aidr/aiguard`
* **Guard endpoint**: `POST /v1/guard_chat_completions`
* **Reference**: [CrowdStrike AIDR Guard LLM Chat Completions](https://aidr-docs.crowdstrike.com/docs/api/aidr#/v1/guard_chat_completions)

The endpoint analyzes prompt/response content and returns:

* `result.blocked` - whether the content should be blocked
* `result.transformed` - whether content was transformed/redacted
* `result.guard_output` - transformed structured output (when available)
* `result.detectors` - detector-level findings

## Integration architecture

TrueFoundry Custom Guardrails expect TrueFoundry request/response schemas. CrowdStrike expects `guard_input` payloads.\
So the recommended pattern is:

1. TrueFoundry AI Gateway calls your **Custom Guardrail adapter**
2. Adapter maps payload to CrowdStrike `guard_chat_completions`
3. Adapter maps CrowdStrike verdict back to TrueFoundry allow/block/mutate behavior

## Prerequisites

Before you begin, ensure you have:

1. A CrowdStrike account with AIDR access
2. A CrowdStrike bearer token with permission to call AIDR APIs
3. TrueFoundry AI Gateway access with permission to configure guardrails
4. A deployed adapter service endpoint reachable by the Gateway

## Quick start

<Steps>
  <Step title="Build and deploy the adapter service">
    Create two endpoints in your adapter:

    * `POST /crowdstrike/input` for LLM input guardrails
    * `POST /crowdstrike/output` for LLM output guardrails

    Both endpoints should:

    1. Read incoming TrueFoundry payload
    2. Build CrowdStrike payload under `guard_input` (especially `messages`)
    3. Call `POST https://api.crowdstrike.com/aidr/aiguard/v1/guard_chat_completions`
    4. Return block/mutate/pass behavior to TrueFoundry

    **Example adapter call to CrowdStrike:**

    ```python theme={"dark"}
    import requests

    def call_crowdstrike(guard_input: dict, token: str, event_type: str = "input") -> dict:
        url = "https://api.crowdstrike.com/aidr/aiguard/v1/guard_chat_completions"
        payload = {
            "guard_input": guard_input,
            "event_type": event_type
        }
        resp = requests.post(
            url,
            headers={
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json"
            },
            json=payload,
            timeout=10,
        )
        resp.raise_for_status()
        return resp.json()
    ```

    <Note>
      Keep the CrowdStrike token in your adapter environment (not in client code).\
      In TrueFoundry, authenticate only to your adapter endpoint.
    </Note>
  </Step>

  <Step title="Register Custom Guardrail integrations in TrueFoundry">
    Navigate to **AI Gateway → Controls → Guardrails** and create a guardrail group.

    Add two custom integrations:

    | Integration | URL                                         | Target     | Operation                | Enforcing Strategy |
    | ----------- | ------------------------------------------- | ---------- | ------------------------ | ------------------ |
    | Input       | `https://<your-adapter>/crowdstrike/input`  | `request`  | `validate` (or `mutate`) | `enforce`          |
    | Output      | `https://<your-adapter>/crowdstrike/output` | `response` | `validate` (or `mutate`) | `enforce`          |

    For details on the Custom Guardrail request/response schema and authentication options, see the [Custom Guardrails guide](/docs/ai-gateway/custom-guardrails).
  </Step>

  <Step title="Create guardrail rules">
    Bind the integrations to models through Guardrail Rules:

    ```yaml theme={"dark"}
    name: crowdstrike-guardrails
    type: gateway-guardrails-config
    rules:
      - id: crowdstrike-input
        when:
          target:
            operator: or
            conditions:
              model:
                values:
                  - openai-main/gpt-4o-mini
                condition: in
        llm_input_guardrails:
          - crowdstrike/crowdstrike-input
        llm_output_guardrails: []
      - id: crowdstrike-output
        when:
          target:
            operator: or
            conditions:
              model:
                values:
                  - openai-main/gpt-4o-mini
                condition: in
        llm_input_guardrails: []
        llm_output_guardrails:
          - crowdstrike/crowdstrike-output
    ```
  </Step>
</Steps>

## Validation logic

Your adapter should map CrowdStrike response to TrueFoundry behavior:

* If `result.blocked == true` -> return HTTP 400 (request blocked)
* If `result.transformed == true` and operation is `mutate` -> return transformed payload
* Otherwise -> return pass (no change)

## Example CrowdStrike response shape

```json theme={"dark"}
{
  "result": {
    "blocked": true,
    "transformed": false,
    "policy": "default",
    "detectors": {
      "malicious_prompt": {
        "detected": true,
        "data": {
          "action": "blocked"
        }
      }
    }
  }
}
```

## Testing checklist

1. Safe prompt should pass.
2. Prompt-injection prompt should return blocked response.
3. Sensitive data prompt should block or redact based on adapter behavior.
4. Verify guardrail traces in TrueFoundry request logs.

## Troubleshooting

<AccordionGroup>
  <Accordion title="Requests time out">
    Ensure your adapter timeout is below TrueFoundry guardrail timeout and optimize the payload sent to CrowdStrike.
  </Accordion>

  <Accordion title="Output guardrails not triggering">
    Confirm output integration is configured with `target: response` and bound under `llm_output_guardrails`.
  </Accordion>

  <Accordion title="Transformations are ignored">
    Use `operation: mutate` and return transformed payload from your adapter.
  </Accordion>
</AccordionGroup>

## Resources

* [CrowdStrike AIDR API Reference](https://aidr-docs.crowdstrike.com/docs/api/aidr#/v1/guard_chat_completions)
* [TrueFoundry Custom Guardrails guide](/docs/ai-gateway/custom-guardrails)
* [Configure Guardrail Rules](/docs/ai-gateway/guardrails-configuration)
