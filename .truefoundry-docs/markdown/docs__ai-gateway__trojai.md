> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# TrojAI DEFEND Integration

> Integrate TrojAI DEFEND with TrueFoundry AI Gateway for real-time AI firewall protection, PII detection, and prompt injection prevention.

This guide explains how to integrate [TrojAI DEFEND](https://www.troj.ai/) with TrueFoundry to add real-time AI firewall guardrails to your LLM applications.

## What is TrojAI DEFEND?

TrojAI DEFEND is an AI firewall that validates and enforces security policies on LLM inputs and outputs in real-time. It evaluates payloads against configurable rule chains — blocking, redacting, or flagging content based on your organization's security requirements.

### Key Features of TrojAI DEFEND

1. **Rule-Chain Firewall**: TrojAI DEFEND evaluates requests through a configurable chain of rules including PII detection, prompt injection prevention, blocklist matching, pattern matching, and content moderation. Each rule can independently block, redact, flag, or pass content, and the chain determines the final action.

2. **Flexible Operation Modes**: Support for both validation and mutation operations. Validation guardrails run in parallel for performance, while mutation guardrails run sequentially and can redact sensitive content (such as PII or credit card numbers) before the request reaches the model.

3. **Streaming and Multimodal Support**: Native support for streaming responses via Server-Sent Events with a sliding window approach for real-time evaluation. The firewall also processes multimodal content — extracting text features from base64-encoded payloads for rule evaluation.

### Adding TrojAI DEFEND Integration

To add TrojAI DEFEND to your TrueFoundry setup, follow these steps:

**Fill in the Guardrails Group Form**

* **Name**: Enter a name for your guardrails group.
* **Collaborators**: Add collaborators who will have access to this group.
* **TrojAI Config**:
  * **Name**: Enter a name for the TrojAI DEFEND configuration.
  * **Description** (Optional): A description for the guardrail (e.g., "TrojAI DEFEND firewall for real-time AI security").
  * **Operation**: The operation type for this guardrail.
    * **Validate**: Guardrails with this operation validate requests against your TrojAI firewall rules. These guardrails are run in parallel.
    * **Mutate**: Guardrails with this operation can both validate and mutate requests (e.g., redact PII). Mutate guardrails are run sequentially.
  * **Priority** (Optional): Execution priority for mutate guardrails (lower number = runs first).
  * **Enforcing Strategy**: Strategy for enforcing this guardrail:
    * **Enforce**: Guardrail is applied. If a violation is detected or an error occurs, the request is blocked.
    * **Enforce But Ignore On Error**: Guardrail is applied, but if an error occurs during execution, the guardrail is ignored and the request proceeds.
    * **Audit**: Request is never blocked. Violations are logged for review only.
* **TrojAI Client ID Auth**:
  * **Client Id**: The `x-eag-clientid` value used to authenticate and identify your firewall policy. This determines which rule chain is applied to requests. Obtain this from your TrojAI DEFEND configuration.
* **Base URL**: The URL of your TrojAI DEFEND firewall instance (e.g., `https://trojaifirewall.your-domain.com`).

<Frame caption="Fill in the TrojAI DEFEND Guardrail Form">
  <img src="https://mintcdn.com/truefoundry/u3eFxtfIcMtyvKm0/images/2026-04-10_22.30.01.png?fit=max&auto=format&n=u3eFxtfIcMtyvKm0&q=85&s=383242132b75b8832521cd94b1f43d3d" alt="TrueFoundry interface for configuring TrojAI DEFEND with fields for name, description, client ID authentication, operation type, enforcing strategy, and base URL" width="1053" height="861" data-path="images/2026-04-10_22.30.01.png" />
</Frame>

#### How TrojAI DEFEND Evaluates Requests

TrueFoundry integrates with TrojAI DEFEND using the `/v1/validateParsedText` endpoint. This endpoint accepts structured LLM payloads (e.g., OpenAI chat completion format), parses them using the firewall policy's handler configuration, and runs input or output rules without calling a downstream model — TrueFoundry handles model invocation separately.

The rule direction (input vs output) is determined automatically:

* **Input guardrails**: Rules evaluate the user prompt before it reaches the model.
* **Output guardrails**: Rules evaluate the model response before it reaches the user.

#### Response Structure

The TrojAI DEFEND API returns a response with the full rule evaluation results:

<AccordionGroup>
  <Accordion title="Example Response: Content Passed">
    All rules passed — the content is safe to proceed.

    ```json theme={"dark"}
    {
      "ModelInputStrings": ["What is the capital of France?"],
      "ModelOutputStrings": [],
      "ConsolidatedFinalOutput": {},
      "Action": "PASS",
      "InputRuleResults": [
        {
          "ruleName": "block_list",
          "action": "PASS",
          "modifiedString": "What is the capital of France?",
          "alias": "",
          "metadata": {},
          "blocked": false,
          "redacted": false,
          "description": ""
        },
        {
          "ruleName": "pii_monitor",
          "action": "PASS",
          "modifiedString": "What is the capital of France?",
          "alias": "SDE",
          "metadata": {},
          "blocked": false,
          "redacted": false,
          "description": ""
        }
      ],
      "OutputRuleResults": [],
      "RawModelOutput": {}
    }
    ```
  </Accordion>

  <Accordion title="Example Response: PII Blocked">
    A credit card number was detected by the `pii_monitor` rule and the request was blocked.

    ```json theme={"dark"}
    {
      "InputString": "1234123412341234",
      "ModifiedInputString": "Message was blocked by the TrojAI moderation system. If this seems like a mistake, please contact the administrator",
      "Action": "BLOCK",
      "InputRuleResults": {
        "ruleResults": [
          {
            "ruleName": "block_list",
            "action": "PASS",
            "modifiedString": "1234123412341234",
            "alias": "",
            "metadata": {},
            "blocked": false,
            "redacted": false,
            "description": ""
          },
          {
            "ruleName": "pii_monitor",
            "action": "BLOCK",
            "modifiedString": "",
            "alias": "SDE",
            "metadata": {
              "foundMatches": ["CREDIT_CARD: 1234123412341234"]
            },
            "blocked": true,
            "redacted": false,
            "description": ""
          }
        ],
        "originalString": "1234123412341234",
        "blocked": true,
        "redactions": null,
        "modifiedString": "1234123412341234"
      }
    }
    ```
  </Accordion>
</AccordionGroup>

| Field                | Type   | Description                                                                |
| -------------------- | ------ | -------------------------------------------------------------------------- |
| `Action`             | string | Overall action: `PASS`, `BLOCK`, `REDACT`, or `FLAG`                       |
| `InputRuleResults`   | array  | Results from each input rule evaluation (populated for input guardrails)   |
| `OutputRuleResults`  | array  | Results from each output rule evaluation (populated for output guardrails) |
| `ModelInputStrings`  | array  | Strings extracted from the payload when running input rules                |
| `ModelOutputStrings` | array  | Strings extracted from the payload when running output rules               |

#### Validation Logic

TrueFoundry uses the TrojAI DEFEND response to determine content safety:

* If the `Action` is `BLOCK`, the request is blocked and a 400 error is returned to the caller.
* If the `Action` is `REDACT` and the operation is set to **Mutate**, the redacted content replaces the original and the request proceeds.
* If the `Action` is `FLAG`, the request proceeds but the flag is logged for audit.
* If the `Action` is `PASS`, the original content is passed through unchanged.

#### Request Logs

When a TrojAI DEFEND guardrail triggers, you can inspect the full request flow in the TrueFoundry request logs. The logs show the guardrail evaluation call to your TrojAI DEFEND base URL, the rule chain results, the final action, and the downstream model request status.
