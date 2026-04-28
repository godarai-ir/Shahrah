> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Introduction

> Learn about guardrails for ensuring safety, compliance, and quality in LLM interactions and MCP tool invocations

## Why Guardrails?

Once AI applications go to production, they handle real user data and — in the case of agents — call external tools on their own. Things can go wrong fast:

* **A customer support chatbot** leaks a user's credit card number because PII wasn't stripped from the context.
* **A coding agent** runs `rm -rf /` through an MCP tool after hallucinating a shell command — and nothing stopped it.
* **A healthcare assistant** makes up drug dosage numbers. The response reaches the patient unchecked.
* **An internal Q\&A bot** gets jailbroken through prompt injection, leaking confidential company data.

Guardrails prevent these scenarios. They sit between your application and the LLM (or MCP tool), inspecting and — when needed — blocking or rewriting data before it causes damage. You can attach them to **LLM requests** (check the prompt going in, check the response coming out) and to **MCP tool calls** (check the arguments before the tool runs, check the results after it returns).

## How a TrueFoundry Guardrail Works

Each guardrail has two settings you configure: what it *does* with the data, and how strictly it *enforces* its decisions.

### Operation Mode

| Mode         | Behavior                                                                                                                                                                                                  | Execution                                         |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------- |
| **Validate** | Looks at the data, blocks the request if something is wrong. Doesn't touch the data itself. <br /> `E.g., a content moderation guardrail sees hate speech in the prompt and blocks the request outright.` | Runs in **parallel** (faster)                     |
| **Mutate**   | Looks at the data *and* rewrites it. Can also block. <br /> `E.g., a PII guardrail rewrites `"My SSN is 123-45-6789"`to`"My SSN is REDACTED"` and lets the request through.`                              | Runs **sequentially** by priority (lower = first) |

### Enforcement Strategy

This decides what happens when a guardrail catches a violation — and also what happens if the guardrail *itself* has a problem (like a timeout or a provider outage).

| Strategy                        | On Violation           | On Guardrail Error                 |
| ------------------------------- | ---------------------- | ---------------------------------- |
| **Enforce**                     | Block                  | Block                              |
| **Enforce But Ignore On Error** | Block                  | Let through (graceful degradation) |
| **Audit**                       | Let through (log only) | Let through                        |

<Tip>
  **How to roll out safely**:

  * Start with **Audit** so you can see what guardrails would catch without affecting users.
  * Once things look right, switch to **Enforce But Ignore On Error** — you get protection, but a guardrail provider outage won't take your app down.
  * Move to **Enforce** when you need strict compliance.
</Tip>

## How Truefoundry AI Gateway Runs Guardrails

Where guardrails run depends on whether you're making an LLM call or invoking an MCP tool.

<Tabs>
  <Tab title="LLM Requests">
    LLM requests have two hooks — **Input** (before the model sees the prompt) and **Output** (after the model responds):

    <CardGroup cols={2}>
      <Card title="LLM Input" icon="arrow-right-to-bracket">
        Runs **before** the prompt reaches the LLM:

        * PII masking and redaction
        * Prompt injection detection
        * Content moderation
      </Card>

      <Card title="LLM Output" icon="arrow-right-from-bracket">
        Runs **after** the LLM responds:

        * Hallucination detection
        * Secrets detection
        * Content filtering
      </Card>
    </CardGroup>

    <Frame caption="LLM Request Guardrail Flow">
      <img src="https://mintcdn.com/truefoundry/Fc6z3zZxeYI2aatT/images/llm_guardrails_flow.png?fit=max&auto=format&n=Fc6z3zZxeYI2aatT&q=85&s=ae40c961eed9a85488d594f74528777a" alt="Diagram showing the flow of LLM requests through input and output guardrails" width="958" height="481" data-path="images/llm_guardrails_flow.png" />
    </Frame>

    Here's the order of operations when a request hits the gateway:

    1. **Input Mutation** guardrails run first and block until they finish (e.g., redacting PII from the prompt).
    2. **Input Validation** kicks off in the background — it checks for things like prompt injection *while* the model request is already in flight.
    3. The **model request starts** with the mutated prompt.
    4. If **Input Validation fails** while the model is still running, the gateway **cancels the model request** right away so you don't pay for it.
    5. Once the model responds, **Output Mutation** guardrails process the response (e.g., stripping secrets).
    6. **Output Validation** checks the final result. If it fails, the response is blocked — though model costs have already been incurred at this point.
    7. The clean response goes back to the client.

    | Hook                  | Execution                           | What Happens on Failure |
    | --------------------- | ----------------------------------- | ----------------------- |
    | **Input Validation**  | Async (parallel with model request) | Model request cancelled |
    | **Input Mutation**    | Sync (before model request)         | Request blocked         |
    | **Output Mutation**   | Sync (after model response)         | Response blocked        |
    | **Output Validation** | Sync (after output mutation)        | Response rejected       |
  </Tab>

  <Tab title="MCP Tool Invocations">
    MCP tool calls have two hooks — **Pre Tool** (before the tool runs) and **Post Tool** (after it returns):

    <CardGroup cols={2}>
      <Card title="MCP Pre Tool" icon="screwdriver-wrench">
        Runs **before** the tool is called:

        * SQL injection prevention
        * Parameter validation
        * Permission checks (Cedar/OPA policies)
      </Card>

      <Card title="MCP Post Tool" icon="check-double">
        Runs **after** the tool returns:

        * Code safety checks
        * Secrets detection in outputs
        * PII redaction from results
      </Card>
    </CardGroup>

    <Frame caption="MCP Tool Guardrail Flow">
      <img src="https://mintcdn.com/truefoundry/Fc6z3zZxeYI2aatT/images/mcp_guardrails_flow.png?fit=max&auto=format&n=Fc6z3zZxeYI2aatT&q=85&s=af6c9d29abcfa2bb6bd4b73d39fbf129" alt="Diagram showing the flow of MCP tool invocations through pre and post tool guardrails" width="958" height="491" data-path="images/mcp_guardrails_flow.png" />
    </Frame>

    The flow is straightforward:

    1. **Pre Tool** guardrails all run before the tool is called. If any of them fail, the tool never executes.
    2. The **tool runs** only after all pre-tool guardrails pass.
    3. **Post Tool** guardrails check (and optionally rewrite) the tool's output.
    4. The validated result goes back to the model.

    | Hook              | Execution                     | What Happens on Failure    |
    | ----------------- | ----------------------------- | -------------------------- |
    | **MCP Pre Tool**  | Sync (before tool invocation) | Tool doesn't run           |
    | **MCP Post Tool** | Sync (after tool returns)     | Result withheld from model |

    <Note>
      Guardrails run on **every tool call separately**. If an agent calls five tools in a row, each one gets its own guardrail checks.
    </Note>

    <Tip>
      MCP hooks matter most in agentic setups where the model decides which tools to call on its own. Use **Pre Tool** to stop bad actions before they happen (e.g., blocking a `DROP TABLE` query), and **Post Tool** to clean up results before the model sees them (e.g., redacting secrets from a database response).
    </Tip>
  </Tab>
</Tabs>

## Latency Impact of Guardrails

Guardrails add processing time — but the gateway is designed to keep that impact small.

<Tabs>
  <Tab title="LLM Requests">
    * **Input Validation runs in parallel** with the model request, so in the happy path, it adds no extra wait time before you see the first token.
    * **Input Mutation runs before** the model request, so its processing time is added directly.
    * When **Input Validation fails**, the model request gets cancelled immediately — you don't pay for a response you were going to throw away.

    **Execution Flow Examples**

    <AccordionGroup>
      <Accordion title="All Guardrails Pass">
        Input validation runs in parallel with the model request. Output guardrails process the response before it's returned.

        ```mermaid theme={"dark"}
        gantt
            title All Guardrails Pass
            dateFormat X
            axisFormat  
            section Input
            Input Mutation              :done, 0, 2
            section Parallel Execution
            Input Validation ✓          :done, 2, 5
            Model Request               :active, 2, 8
            section Output
            Output Mutation             :done, 8, 10
            Output Validation ✓         :done, 10, 12
        ```
      </Accordion>

      <Accordion title="Input Validation Failure">
        Input validation fails while the model is running — the model request gets **cancelled immediately** to save costs.

        ```mermaid theme={"dark"}
        gantt
            title Input Validation Failure — Model Request Cancelled
            dateFormat X
            axisFormat  
            section Input
            Input Mutation              :done, 0, 2
            section Parallel Execution
            Input Validation ✗          :crit, 2, 5
            Model Request (cancelled)   :crit, 2, 5
        ```
      </Accordion>

      <Accordion title="Output Validation Failure">
        The model finishes, but output validation fails — the response is rejected. Model costs are already incurred at this point.

        ```mermaid theme={"dark"}
        gantt
            title Output Validation Failure — Response Rejected
            dateFormat X
            axisFormat  
            section Input
            Input Mutation              :done, 0, 2
            section Parallel Execution
            Input Validation ✓          :done, 2, 5
            Model Request               :active, 2, 8
            section Output
            Output Mutation             :done, 8, 10
            Output Validation ✗         :crit, 10, 12
        ```
      </Accordion>
    </AccordionGroup>
  </Tab>

  <Tab title="MCP Tool Invocations">
    * **All MCP guardrails run synchronously** — pre-tool guardrails block before the tool executes, and post-tool guardrails block after.
    * **Pre-tool guardrail latency adds directly** to the tool call time since the tool cannot start until all pre-tool checks pass.
    * When a **pre-tool guardrail fails**, the tool never executes — you avoid the cost and side effects of a bad tool call entirely.
  </Tab>
</Tabs>

You can track the latency impact of each guardrail in **AI Gateway → Monitor → Request Traces**. Each guardrail span shows its execution time, result, scope, and which entity it was applied on.

<Frame caption="Guardrail latency visible in Request Traces">
  <img src="https://mintcdn.com/truefoundry/jcJfesAzCw8ThHaY/images/guardrail-request-traces.png?fit=max&auto=format&n=jcJfesAzCw8ThHaY&q=85&s=0c36d4b75b93f70049344ff8278dae13" alt="Request Traces view showing guardrail execution latency, result type, scope, and applied entity for each span" width="1024" height="560" data-path="images/guardrail-request-traces.png" />
</Frame>

## How to Apply Guardrails

You can attach guardrails in two ways:

<AccordionGroup>
  <Accordion title="Per-Request via Headers">
    Pass the `X-TFY-GUARDRAILS` header to apply guardrails on a single request. Handy for testing or when different requests need different guardrails.

    ```json theme={"dark"}
    {
      "llm_input_guardrails": ["my-group/pii-redaction"],
      "llm_output_guardrails": ["my-group/secrets-detection"],
      "mcp_tool_pre_invoke_guardrails": ["my-group/sql-sanitizer"],
      "mcp_tool_post_invoke_guardrails": ["my-group/code-safety"]
    }
    ```
  </Accordion>

  <Accordion title="Gateway-Level via Policies">
    Set up guardrail rules in **AI Gateway → Controls → Guardrails** to apply guardrails automatically based on who's making the request, which model they're calling, or which MCP tool is being used. This is the way to go for org-wide enforcement.

    For a step-by-step walkthrough, see the [Getting Started](/docs/ai-gateway/guardrails-getting-started) guide. For the full policy reference, see [Guardrails Configuration](/docs/ai-gateway/guardrails-configuration).
  </Accordion>
</AccordionGroup>

## Supported Guardrails

The AI Gateway ships with built-in guardrails and integrates with a range of external providers — all managed through a single interface.

### TrueFoundry Built-in Guardrails

Ready to use out of the box — no external credentials needed.

<Columns cols={3}>
  <Card title="Secrets Detection" href="/docs/ai-gateway/secrets-detection" icon="key">
    Catches and redacts credentials like AWS keys, API keys, JWT tokens, and private keys.
  </Card>

  <Card title="Code Safety Linter" href="/docs/ai-gateway/code-safety-linter" icon="code">
    Flags unsafe code patterns — eval, exec, os.system, subprocess calls, dangerous shell commands.
  </Card>

  <Card title="SQL Sanitizer" href="/docs/ai-gateway/sql-sanitizer" icon="database">
    Catches risky SQL: DROP, TRUNCATE, DELETE/UPDATE without WHERE, string interpolation.
  </Card>

  <Card title="Regex Pattern Matching" href="/docs/ai-gateway/regex-pattern-matching" icon="brackets-curly">
    Matches and redacts sensitive patterns (PII, payment cards, credentials) using built-in or custom regex.
  </Card>

  <Card title="Prompt Injection" href="/docs/ai-gateway/tfy-prompt-injection" icon="shield">
    Detects prompt injection attacks and jailbreak attempts using model-based analysis.
  </Card>

  <Card title="PII Detection" href="/docs/ai-gateway/tfy-pii" icon="user-shield">
    Finds and redacts personally identifiable information with configurable entity categories.
  </Card>

  <Card title="Content Moderation" href="/docs/ai-gateway/tfy-content-moderation" icon="ban">
    Blocks harmful content across hate, self-harm, sexual, and violence categories with adjustable thresholds.
  </Card>

  <Card title="Cedar Guardrails" href="/docs/ai-gateway/cedar-guardrails" icon="shield">
    Fine-grained access control for MCP tools using Cedar policy language with default-deny security.
  </Card>

  <Card title="OPA Guardrails" href="/docs/ai-gateway/opa-guardrails" icon="shield">
    Fine-grained access control with full policy lifecycle management using Open Policy Agent.
  </Card>
</Columns>

### External Providers

We also integrate with third-party guardrail providers. Don't see yours? Reach out — we're happy to add it.

<Columns cols={3}>
  <Card title="OpenAI Moderations" href="/docs/ai-gateway/openai-moderations">
    OpenAI's moderation API for detecting violence, hate speech, harassment, and other policy violations.
  </Card>

  <Card title="AWS Bedrock Guardrail" href="/docs/ai-gateway/bedrock-guardrails">
    AWS Bedrock's guardrail capabilities for AI models.
  </Card>

  <Card title="Azure PII" href="/docs/ai-gateway/azure-pii">
    Azure's PII detection service for identifying and redacting personal data.
  </Card>

  <Card title="Azure Content Safety" href="/docs/ai-gateway/azure-content-safety">
    Azure Content Safety for detecting harmful or inappropriate content.
  </Card>

  <Card title="Azure Prompt Shield" href="/docs/ai-gateway/azure-prompt-shield">
    Azure Prompt Shield for blocking prompt injection and jailbreak attempts.
  </Card>

  <Card title="Enkrypt AI" href="/docs/ai-gateway/enkrypt-ai">
    Advanced moderation and compliance — toxicity, bias, and sensitive data detection.
  </Card>

  <Card title="Palo Alto Prisma AIRS" href="/docs/ai-gateway/palo-alto-airs">
    Palo Alto AI Risk for content safety and threat detection.
  </Card>

  <Card title="Fiddler" href="/docs/ai-gateway/fiddler">
    Fiddler-Safety and Fiddler-Response-Faithfulness guardrails.
  </Card>

  <Card title="CrowdStrike" href="/docs/ai-gateway/crowstrike">
    API-based security — content moderation, prompt injection detection, toxicity analysis. (Formerly Pangea, acquired by CrowdStrike in 2025.)
  </Card>

  <Card title="Patronus AI" href="/docs/ai-gateway/patronus">
    Hallucination detection, prompt injection, PII leakage, toxicity, and bias evaluators.
  </Card>

  <Card title="Google Model Armor" href="/docs/ai-gateway/google-model-armor">
    Google Cloud Model Armor for prompt injection, harmful content, PII, and malicious URI detection.
  </Card>

  <Card title="GraySwan Cygnal" href="/docs/ai-gateway/grayswan-cygnal">
    Policy violation detection and content safety monitoring powered by GraySwan Cygnal.
  </Card>

  <Card title="Akto" href="/docs/ai-gateway/akto">
    LLM security, prompt injection detection, and policy violation monitoring with native streaming support.
  </Card>
</Columns>

### Bring Your Own Guardrail / Plugin

If the built-in and provider integrations don't cover your use case, you can write your own. Build a custom guardrail with Guardrails.AI, a plain Python function, or any framework you prefer, and plug it into the gateway.

<Card title="Custom Guardrails" href="/docs/ai-gateway/custom-guardrails" icon="plug">
  Build and integrate your own guardrail using Guardrails.AI, a Python function, or any custom logic.
</Card>

## FAQ

<AccordionGroup>
  <Accordion title="How do I control which messages guardrails evaluate?">
    By default, guardrails look at **all messages** in the conversation. If you only care about the latest message, set the `X-TFY-GUARDRAILS-SCOPE` header:

    * **`all`** (default): Checks the full conversation history
    * **`last`**: Checks only the most recent message

    Using `last` is faster when you don't need to scan the whole conversation.

    ```bash theme={"dark"}
    curl -X POST "{GATEWAY_BASE_URL}/chat/completions" \
      -H "Authorization: Bearer YOUR_API_KEY" \
      -H "Content-Type: application/json" \
      -H 'X-TFY-GUARDRAILS: {"llm_input_guardrails":["my-group/pii-redaction"]}' \
      -H "X-TFY-GUARDRAILS-SCOPE: last" \
      -d '{
        "model": "openai/gpt-4o",
        "messages": [
          {"role": "system", "content": "You are a helpful assistant."},
          {"role": "user", "content": "Hello, how can you help me today?"}
        ]
      }'
    ```
  </Accordion>

  <Accordion title="What happens if a guardrail service is down?">
    Depends on the enforcement strategy you picked:

    * **Enforce**: Request gets blocked.
    * **Enforce But Ignore On Error**: Request goes through anyway. This is the safest default — you stay protected when guardrails work, but a provider outage won't break your app.
    * **Audit**: Request always goes through.
  </Accordion>

  <Accordion title="How can I monitor guardrail execution?">
    Go to **AI Gateway → Monitor → Request Traces**. You'll see:

    * Which guardrails ran on each hook
    * Whether they passed or failed, and how long they took
    * What they found (secrets, SQL issues, unsafe patterns, etc.)
    * What mutations were applied

    Traces are logged for both successful and blocked requests.
  </Accordion>

  <Accordion title="How do output guardrails work with streaming responses?">
    Streaming and output guardrails are fundamentally at odds — guardrails need the complete response to evaluate it, but streaming sends tokens to the client as they're generated. The gateway handles this by **buffering the entire response**, running output guardrails on the full text, and only then streaming it back to the client if the guardrails pass. This means you lose the time-to-first-token benefit of streaming when output guardrails are active.

    There are alternative strategies like running guardrails on **batches of tokens** as they arrive, but these only work for guardrails that operate on local patterns — for example, regex matching, PII detection, or secrets scanning can catch issues in a chunk without seeing the full response. Guardrails that need the complete output as context — such as hallucination detection, factual consistency checks, or content moderation that evaluates the overall message — cannot work incrementally and still require the full response to be buffered.

    **In practice:**

    | Approach                     | Time to First Token               | Works For                                                        |
    | ---------------------------- | --------------------------------- | ---------------------------------------------------------------- |
    | **Full buffering** (default) | Delayed until guardrails complete | All guardrails — hallucination, moderation, factual checks, etc. |
    | **Chunked evaluation**       | Near-normal streaming             | Pattern-based guardrails only — regex, PII, secrets scanning     |

    If low latency matters more than full output coverage, consider using only input guardrails with streaming, and running output guardrails asynchronously for audit purposes.
  </Accordion>
</AccordionGroup>
