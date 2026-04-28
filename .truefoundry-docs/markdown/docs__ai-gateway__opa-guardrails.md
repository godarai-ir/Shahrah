> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# OPA Guardrails

> Learn how to use Open Policy Agent (OPA) to create flexible, policy-driven guardrails for LLM inputs, outputs, and MCP tool invocations

OPA Guardrails enable you to implement flexible, policy-driven access control for LLM requests, responses, and MCP (Model Context Protocol) tool invocations using the [Open Policy Agent (OPA)](https://www.openpolicyagent.org/docs) and the Rego policy language. OPA provides a powerful, general-purpose policy engine that decouples policy decisions from your application logic, ensuring secure, and auditable AI workflows.

## Guardrail Overview

OPA Guardrails use the Rego policy language to evaluate authorization decisions across the entire LLM gateway lifecycle — from validating incoming requests and user identity, to filtering LLM responses for sensitive content, to controlling which MCP tools can be invoked. When a request passes through the gateway, the OPA guardrail sends the request context to your OPA server and enforces the policy decision in real time.

### Key Features

* **Full Lifecycle Coverage**: Apply policies at LLM input, LLM output and MCP pre-tool hooks
* **Flexible Policy Logic**: Write expressive policies in Rego to handle authentication, content filtering, role-based access, and more
* **Self-hosted Control**: Run your own OPA server, keeping all policy evaluation within your infrastructure
* **Default Deny**: All requests are denied unless explicitly permitted by a policy
* **Audit Mode**: Log policy violations without blocking requests, enabling safe policy rollout
* **Rich Context**: Policies receive full request context including user identity, metadata, request body, and response body

### When to Use OPA Guardrails

OPA Guardrails are ideal for:

* **Authentication & Authorization**: Restrict LLM access to specific users, service accounts, or teams
* **Model Access Control**: Allowlist or blocklist specific LLM models per user role
* **Input Validation**: Block prompt injection attempts, harmful content, or requests exceeding token limits
* **Output Filtering**: Detect and block sensitive data like PII, credentials, or SQL injection patterns in LLM responses
* **MCP Tool Access Control**: Restrict which tools can be invoked based on user identity or tool arguments
* **Compliance Enforcement**: Enforce organizational policies, data residency rules, and regulatory requirements

## Input Structure

OPA policies receive a single input object containing all the context needed to make authorization decisions. The structure of this object varies based on the hook type the guardrail is applied to. It varies slightly based on the hook type (LLM input, LLM output, MCP pre-tool, MCP post-tool). Here are the variations:

| Field      | Type   | Description                                                                            |
| ---------- | ------ | -------------------------------------------------------------------------------------- |
| `request`  | Object | The LLM request body or MCP tool invocation details                                    |
| `metadata` | Object | Additional metadata about the request, such as User identity and request context       |
| `context`  | Object | Custom context fields defined by the user, such as Hook type and streaming information |

### Structure of the input object when applied as an input LLM guardrail:

```json theme={"dark"}
  opainput: {
    "request":  { ... },
    "metadata": { ... },
    "context":  { ... }
  }

```

**Example input payload**

```json theme={"dark"}
{
  "request": {
    "max_tokens": 2500,
    "model": "vertex-main/gemini-2-5-flash",
    "messages": [
      {
        "role": "user",
        "content": "Who are you?"
      }
    ],
    "stream": true
  },
  "metadata": {
    "subject": {
      "id": "sample-id",
      "subjectSlug": "test@example.com",
      "subjectType": "user",
      "tenantName": "test-tenant",
      "teamName": [
        "test-team-1",
        "test-team-2"
      ]
    }
  },
  "context": {
    "hookType": "beforeRequestHook",
    "isStreaming": true
  }
}
```

### Structure of the input object when applied as an output LLM guardrail:

```json theme={"dark"}
{
  "request":  { ... },
  "metadata": { ... },
  "context":  { ... }
}
```

**Example Input Payload**

```json theme={"dark"}
{
  "request": {
    "id": "sample-id",
    "object": "chat.completion",
    "created": sample-timestamp,
    "model": "gemini-2.5-flash",
    "provider": "vertex",
    "choices": [
      {
        "message": {
          "role": "assistant",
          "content": "sample response content"
        },
        "index": 0,
        "finish_reason": "stop"
      }
    ],
    "usage": {
      "prompt_tokens": 9,
      "completion_tokens": 197,
      "total_tokens": 397,
      "completion_tokens_details": {
        "reasoning_tokens": 191,
        "audio_tokens": 0
      },
      "prompt_tokens_details": {
        "cached_tokens": 0,
        "audio_tokens": 0
      }
    }
  },
  "metadata": {
    "subject": {
      "id": "sample-id",
      "subjectSlug": "test@example.com",
      "subjectType": "user",
      "tenantName": "test-tenant",
      "teamName": [
        "test-team-1",
        "test-team-2"
      ]
    }
  },
  "context": {
    "hookType": "afterRequestHook",
    "isStreaming": false
  }
}    
```

### Structure of the input object when applied as an MCP tool pre-invoke guardrail:

```json theme={"dark"}
{
  "request":  { ... },
  "metadata": { ... },
  "context":  { ... }
}
```

\**Example Input Payload*

```json theme={"dark"}
opaInput {
  "request": {
    "sample-arg-1": "value1",
    "sample-arg-2": "value2"
  },
  "metadata": {
    "subject": {
      "id": "sample-id",
      "subjectSlug": "test@example.com",
      "subjectType": "user",
      "tenantName": "test-tenant",
      "teamName": [
        "test-team-1",
        "test-team-2"
      ]
    },
    "tool_name": "test-tool",
    "mcp_server": "test:mcp-server:sample-mcp",
    "mcp_server_name": "sample-mcp",
    "subjectType": "user"
  },
  "context": {
    "hookType": "mcpPreTool",
    "isStreaming": false
  }
}
```

### request

It contains the raw LLM request body sent by the caller. Here is an example of the request object from LLM input guardrail:

```json theme={"dark"}
{
    "max_tokens": 2500,
    "model": "vertex-main/gemini-2-5-flash",
    "messages": [
      {
        "role": "user",
        "content": "Who are you?"
      }
    ],
    "stream": true
}
```

### metadata

Contains enriched user identity information automatically populated by the gateway from the authenticated session and the X-TFY-METADATA request header. Here is an example of the metadata object from LLM input guardrail:

```json theme={"dark"}
{
    "subject": {
      "id": "sample-id",
      "subjectSlug": "test@example.com",
      "subjectType": "user",
      "tenantName": "test-tenant",
      "teamName": [
        "test-team-1",
        "test-team-2"
      ]
    }
}
```

### Context

Contains custom context fields defined by the user in the guardrail configuration. The `hookType` field indicates which lifecycle hook the request is being evaluated for, allowing you to write policies that behave differently based on the hook. Here is an example of the context object:

```json theme={"dark"}
{
    "hookType":    "beforeRequestHook",
    "isStreaming": false
}
```

## Output Structure

Your OPA policy must return a response in the one of the following format for the gateway to process it correctly:

```json theme={"dark"}
{
  "allow": boolean
}
```

```json theme={"dark"}
{
  "allow": boolean,
  "desc": string
}
```

| Field   | Type    | Mandatory | Description                                                                    |
| ------- | ------- | --------- | ------------------------------------------------------------------------------ |
| `allow` | boolean | Yes       | Indicates whether the request is allowed or denied                             |
| `desc`  | string  | No        | An optional description providing additional context about the policy decision |

## Example Policies

Here are practical examples of OPA policies.

### Example 1: Block Requests with Prompt Injection Keywords

This policy prevents a message containing blocked keywords from being processed by the LLM:

```opa theme={"dark"}
package authz_req

import future.keywords.if
import future.keywords.in

# ============================================================================
# CONFIGURATION
# ============================================================================

# Default deny
default allow_decision := false

# Blocked keywords in messages
blocked_keywords := [
    "hack",
    "exploit",
    "jailbreak",
    "ignore previous",
    "disregard instructions",
    "bypass security",
    "override system",
    "sudo mode",
    "admin access"
]

# ============================================================================
# HELPER FUNCTIONS
# ============================================================================

# Extract all user messages
get_user_messages(messages) := user_msgs if {
    user_msgs := [msg | 
        msg := messages[_]
        msg.role == "user"
    ]
}

# Check if text contains blocked keywords
contains_blocked_keywords(text) if {
    lower_text := lower(text)
    keyword := blocked_keywords[_]
    contains(lower_text, keyword)
}

# Check if any user message has blocked content
has_blocked_content(messages) if {
    user_msgs := get_user_messages(messages)
    msg := user_msgs[_]
    contains_blocked_keywords(msg.content)
}

# ============================================================================
# VALIDATION RULES
# ============================================================================

# Allow if messages have NO blocked keywords
allow_decision if {
    messages := input.request.messages
    not has_blocked_content(messages)
}

# ============================================================================
# DESCRIPTIONS
# ============================================================================

# Success description
description := "Request validation successful: no blocked keywords found" if {
    messages := input.request.messages
    not has_blocked_content(messages)
}

# Failure - blocked keywords
description := "Request blocked: message contains blocked keywords" if {
    messages := input.request.messages
    has_blocked_content(messages)
}

# ============================================================================
# RETURN FORMAT
# ============================================================================

# Return in format expected by OPA handler
allow := {
    "allow": allow_decision,
    "desc": description
}
```

**What this policy does:**

* Defines a list of blocked keywords commonly associated with prompt injection attempts
* Extracts user messages from the request
* Checks if user message contains any of the blocked keywords
* Denies the request if any blocked keywords are found, otherwise allows it
* Provides descriptive messages for both allowed and denied outcomes to aid debugging

### Additional Policy Examples

#### Permit All Users of a specific team

```opa theme={"dark"}
package authz_team_allow

# Default deny
default allow_decision := false

allow_decision if {
    input.metadata.subject.team_name == "test-team"
}

# Path authz_team/allow returns this object so gateway gets result.allow
allow := {"allow": allow_decision}
```

#### Forbid Access to Sensitive MCP Servers

```opa theme={"dark"}
package authz_fetch

# Default deny
default allow_decision := false

# Block fetch tool from nm fetch mcp server
allow_decision if {
    not is_blocked_tool
}

is_blocked_tool if {
    input.metadata.tool_name == "fetch"
    input.metadata.mcp_server_name == "nm fetch"
}

# Return format with policy path authz_fetch/allow
allow := {"allow": allow_decision}
```

## Safety and Enforcement

OPA Guardrails provide flexible, policy-driven enforcement across the entire LLM request lifecycle with configurable default behavior.

### Default Behavior

**OPA policies control their own default behavior**:

1. **Policy-Defined Defaults**: Each policy sets its own default using `default allow_decision := true` or `false`
2. **Flexible Logic**: You can implement allow-lists (default deny) or block-lists (default allow) based on your needs

**Recommended: Default Deny for Security**

```rego theme={"dark"}
# Recommended: Start with deny, explicitly allow 
# This ensures that nothing is permitted unless you explicitly write a rule allowing it

default allow_decision := false

allow_decision if {
    # Only allow specific conditions
    input.metadata.user_email == "trusted@example.com"
}
```

### Enforcement Points

OPA Guardrails support three enforcement strategies:

#### 1. Enforce But Ignore On Error (Default)

* Blocks requests when the OPA Server denies permission
* **Fails open** on OPA server errors (allows request)
* Use when availability is more critical than strict enforcement

#### 2. Enforce

* Blocks requests when the OPA Server denies permission
* Fails closed on OPA server errors (blocks request)

#### 3. Audit Mode

* Logs violations but **never blocks** requests
* Use for testing policies before enforcing them
* Violations are logged for review

### Configuration

**Configuring LLM Guardrails:**

1. Create an OPA guardrail integration with your OPA server URL and policy path with the appropriate enforcement strategy.
2. Add the guardrail as input or output guardrail while testing on play ground.
3. You can also add the guardrail to the appropriate hook in your guardrail rules:
   * `llm_input_guardrails` for request validation
   * `llm_output_guardrails` for response filtering

**Configuring MCP Guardrails:**

1. Create an OPA guardrail integration with your OPA server URL and policy path with the appropriate enforcement strategy.
2. Configure OPA guardrails in **AI Gateway → Controls → Guardrails**
3. Add the guardrail to the appropriate hook in your guardrail rules:
   * `mcp_tool_pre_invoke_guardrails` for tool access control
   * `mcp_tool_post_invoke_guardrails` for tool result filtering

### Best Practices for Safety

1. **Start with Audit Mode**: Test policies in audit mode before enforcing them to avoid blocking legitimate requests
2. **Default Deny for Security**: Use `default allow_decision := false` and explicitly allow only trusted conditions
3. **Use Descriptive Violations**: Include helpful `desc` fields to aid debugging when policies block requests

### Error Handling

When an OPA guardrail denies a request:

* The request is blocked at the configured hook (input/output/pre-tool)
* An error is returned to the caller: `"Guardrail violation detected"`
* The `desc` field from your policy is captured in logs
* Traces will show which guardrail integration blocked the request

When the OPA server is unreachable:

* If `failOnError: true` (default): Request is **blocked**
* If `failOnError: false`: Request is **allowed**
* If `audit_mode: true`: Request is **allowed** regardless of `failOnError`

<Tip>
  **Test Before Enforcing**: Always test new policies in audit mode first. Monitor the logs to ensure your policy allows legitimate requests and only blocks what you intend to block.
</Tip>

<Note>
  **OPA Server Availability**: Your OPA server becomes a critical dependency. Ensure it's highly available, monitored, and has appropriate timeouts configured. Consider using `failOnError: false` for non-critical policies if availability is a concern.
</Note>
