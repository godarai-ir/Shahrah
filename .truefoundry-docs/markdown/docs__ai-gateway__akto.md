> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Akto Integration

> Learn how to use Akto with TrueFoundry AI Gateway for LLM security, prompt injection detection, and policy violation monitoring.

This guide explains how to integrate [Akto](https://www.akto.io/) with TrueFoundry to enhance the security and compliance of your LLM applications.

## What is Akto?

Akto is an API security and LLM observability platform that provides native guardrail capabilities for LLM deployments. It analyzes incoming requests and outgoing responses for prompt injection, policy violations, and other security threats, then blocks or allows traffic based on your configured policies.

### Key Features of Akto

1. **Prompt Injection Detection**: Identifies and blocks prompt injection attempts before they reach the LLM, preventing adversarial inputs from manipulating model behavior.

2. **Policy Violation Monitoring**: Continuously monitors LLM traffic against configurable security policies and ingests blocked events automatically for audit purposes.

3. **Streaming Support**: Native support for streaming LLM responses, allowing guardrails to operate correctly whether the model is in streaming or non-streaming mode.

### Adding Akto Integration

To add Akto to your TrueFoundry setup, follow these steps:

**Fill in the Guardrails Group Form**

* **Name**: Enter a name for your guardrails group.
* **Collaborators**: Add collaborators who will have access to this group.
* **Akto Config**:
  * **Name**: Enter a name for the Akto configuration (e.g., `akto`).
  * **Description** (Optional): A description for the guardrail (e.g., "Akto guardrail for LLM security, prompt injection detection, and policy violation monitoring").
  * **Operation**: The operation type for this guardrail. Akto guardrails can only be used for **Validate** — requests are validated against your Akto security policies.
* **Akto Authentication Data**:
  * **Token**: The JWT token used to authenticate requests to your Akto guardrails service. You can obtain this token from your Akto dashboard. Keep it secure, as it grants access to your Akto guardrails resources.
* **Enforcing Strategy**: Strategy for enforcing this guardrail:
  * **Enforce**: Guardrail is applied. If a violation is detected or an error occurs during execution, the request is blocked.
  * **Enforce But Ignore On Error**: Guardrail is applied, but if an error occurs during execution, the guardrail is ignored and the request proceeds.
  * **Audit**: Request is never blocked. Violations are logged for review only.
* **Base URL**: The URL of your Akto guardrails service (e.g., `https://<your-instance>-guardrails.akto.io`).

<Frame caption="Fill in the Akto Guardrail Form">
  <img src="https://mintcdn.com/truefoundry/yR_clVDeJDlQkXKY/images/2026-03-23_12.06.17.png?fit=max&auto=format&n=yR_clVDeJDlQkXKY&q=85&s=0b607a014e7e36a8ad58feab97ff91ca" alt="TrueFoundry interface for configuring Akto with fields for name, description, token authentication, operation, enforcing strategy, and base URL" width="1073" height="796" data-path="images/2026-03-23_12.06.17.png" />
</Frame>

#### Response Structure

The Akto guardrails API (`POST /api/http-proxy/truefoundry`) returns a response with the following structure:

<AccordionGroup>
  <Accordion title="Example Response: Allowed (HTTP 200)">
    The request passed all guardrail checks. TrueFoundry proceeds to forward the request to the LLM.

    ```json theme={"dark"}
    {}
    ```
  </Accordion>

  <Accordion title="Example Response: Blocked (HTTP 400)">
    The request was blocked by Akto guardrails. TrueFoundry returns a 400 error and does not forward the request to the LLM. Akto automatically ingests the blocked event for audit purposes.

    ```json theme={"dark"}
    {
      "actionErrors": ["Request blocked by guardrails: prompt injection detected"]
    }
    ```
  </Accordion>

  <Accordion title="Example Response: Unauthorized (HTTP 401)">
    The request was rejected because the provided token was invalid or missing. Verify that the token in your Akto configuration is correct.

    ```json theme={"dark"}
    {
      "actionErrors": ["Unauthorized: invalid token"]
    }
    ```
  </Accordion>
</AccordionGroup>

| Field          | Type  | Description                                                                                           |
| -------------- | ----- | ----------------------------------------------------------------------------------------------------- |
| `actionErrors` | array | List of error messages explaining why the request was blocked. Present only on 400 and 401 responses. |

#### Validation Logic

TrueFoundry uses the HTTP status code returned by the Akto guardrails API to determine whether to allow or block the request:

* If Akto returns **HTTP 200**, the request is considered safe and will be forwarded to the LLM.
* If Akto returns **HTTP 400**, the request is considered a policy violation — it will be blocked and a 400 error is returned to the caller.
* If Akto returns **HTTP 401**, the request is blocked due to an authentication failure. Verify your token configuration.

#### Request Logs

When an Akto guardrail triggers, you can inspect the full request flow in the TrueFoundry request logs. The logs show the guardrail evaluation call to your Akto service URL, the block result, and the downstream model request status.
