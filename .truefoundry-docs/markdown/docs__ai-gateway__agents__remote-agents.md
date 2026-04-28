> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Remote Agents

> Register agents running on any platform — Bedrock, Vertex AI, LangGraph, or custom infrastructure — into a unified control plane with governance.

<Info>Remote Agents are available on both **SaaS** and **Self-hosted** deployments.</Info>

Already have agents running elsewhere? Register them here.

Your organization likely has agents running on Bedrock, Vertex AI, Azure AI, LangGraph, or custom infrastructure. Agent Registry lets you connect all of them into a single control plane — without moving or rewriting anything.

Once a remote agent is registered, it gets the same treatment as any TrueFoundry-native agent:

<CardGroup cols={2}>
  <Card title="Single Discovery Surface" icon="compass">
    All agents — regardless of where they run — are visible in one place. Teams stop asking "where's that agent?"
  </Card>

  <Card title="RBAC" icon="lock">
    Control which teams and users can discover and invoke each agent.
  </Card>

  <Card title="Policies" icon="shield">
    Apply guardrails, content filters, and governance policies uniformly across all remote agents.
  </Card>

  <Card title="Unified Metrics" icon="chart-line">
    Get request counts, latency, error rates, and usage data across every agent from a single pane of glass.
  </Card>
</CardGroup>

<Frame caption="Agent Registry showing registered Remote agents">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/agent_registry.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=f8892c67055260c6168fecea516615d7" alt="Agent Registry showing registered Remote agents" width="3020" height="1656" data-path="images/agent_registry.png" />
</Frame>

***

## Supported Frameworks

Remote agents are registered with a **framework** type that determines how TrueFoundry communicates with the agent. Two framework types are supported:

| Framework  | Protocol                                              | Description                                                                                                                                                                                  |
| ---------- | ----------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **A2A**    | [Agent-to-Agent (A2A)](https://google.github.io/A2A/) | Google's open interoperability standard. Agents communicate via JSONRPC, expose an agent card at a well-known path, and support structured message exchange. **Recommended for new agents.** |
| **Custom** | HTTP                                                  | Proxy any HTTP-compatible agent endpoint. No protocol constraints — TrueFoundry forwards requests to your endpoint as-is.                                                                    |

***

## Registering a Remote Agent

On the Agent Registry page, click **Create New Agent** and select **Integrate a Remote Agent**.

<Frame caption="Create New Agent button on the Agent Registry page">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/create_new_agent_button.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=6ee5dd1bf485d3138417d81a3dbfe558" alt="Create New Agent button on the Agent Registry page" width="2671" height="511" data-path="images/create_new_agent_button.png" />
</Frame>

<Frame caption="Select Integrate a Remote Agent">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/new_agent_remote.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=dcec9d763798693e894368fe653858d6" alt="Create New Agent dialog with Integrate a Remote Agent highlighted" width="1691" height="748" data-path="images/new_agent_remote.png" />
</Frame>

Fill in the registration form:

<Steps>
  <Step title="Basic information">
    Provide the agent's **Name**, **Description**, and optional **Tags**.
  </Step>

  <Step title="Collaborators">
    Add team members or groups with access. Each collaborator gets either **Agent Manager** (can edit the agent) or **Agent Access** (can invoke the agent) permissions.
  </Step>

  <Step title="Endpoint URL">
    Enter the URL where the agent is reachable. TrueFoundry will proxy all requests to this endpoint.
  </Step>

  <Step title="Framework">
    Select **A2A** or **Custom**. For A2A agents, specify the **Agent Card Path** (defaults to `/.well-known/agent-card.json`).
  </Step>

  <Step title="Authentication">
    Optionally configure auth for requests to the remote agent:

    * **Header Auth** — Static API key or token sent via request headers (e.g., `Authorization: Bearer <token>`).
    * **Token Passthrough** — Forward the caller's token to the remote agent as-is.
  </Step>
</Steps>

<Frame caption="Remote agent registration form: name, description, collaborators, and endpoint URL">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/remote_agent_form_1.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=77c812e3e66d7e22ea2cbfdfed95695e" alt="Remote agent registration form showing name, description, collaborators, and endpoint URL" width="1604" height="1706" data-path="images/remote_agent_form_1.png" />
</Frame>

<Frame caption="Remote agent registration form: framework selection (A2A or Custom), agent card path, and authentication">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/remote_agent_form_2.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=3a84035058e53bc039e38b69da86e5c6" alt="Remote agent registration form showing framework, agent card path, and authentication options" width="1584" height="1712" data-path="images/remote_agent_form_2.png" />
</Frame>

### Registration Fields Reference

| Field               | Required | Description                                                            |
| ------------------- | -------- | ---------------------------------------------------------------------- |
| **Name**            | Yes      | Unique identifier for the agent within your organization.              |
| **Description**     | Yes      | Short description of the agent's purpose.                              |
| **URL**             | Yes      | The endpoint URL where the agent is reachable.                         |
| **Framework**       | Yes      | `A2A` or `Custom`.                                                     |
| **Agent Card Path** | A2A only | Path to the agent card JSON (default: `/.well-known/agent-card.json`). |
| **Collaborators**   | Yes      | Users or teams with `agent-access` or `agent-manager` roles.           |
| **Auth Data**       | No       | Header Auth (static headers) or Token Passthrough.                     |

***

## A2A Agents

The [Agent-to-Agent (A2A) protocol](https://google.github.io/A2A/) is an open standard for agent interoperability — a common language that lets agents from different frameworks and vendors communicate with each other.

Agent Registry natively supports registering A2A-compatible agents. If your agent speaks A2A, register it and it immediately becomes part of your organization's agent ecosystem — discoverable, governable, and composable with other agents.

This matters as the agent ecosystem matures. Agents built by different teams, on different platforms, using different frameworks will need to interoperate. A2A is how that happens. And Agent Registry is where all of them come together.

### Invoking an A2A Agent

Once registered, the agent appears on the Agent Registry page. Click the **code snippet** button to get a ready-to-use snippet for invoking the agent via the A2A Python SDK.

<Frame caption="Agent Registry showing the code snippet button on a Remote agent">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/agent_registry_code_snippet.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=80c483c80d6f61b69c7c833356ddac56" alt="Agent Registry with code snippet button highlighted on a Remote agent" width="3020" height="1656" data-path="images/agent_registry_code_snippet.png" />
</Frame>

The generated snippet handles agent card resolution, client setup, and message exchange — all routed through TrueFoundry's proxy with your configured authentication.

<Frame caption="Python code snippet using A2A SDK to invoke a registered remote agent">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/remote_agent_a2a_code_snippet.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=551b4c3d6919e94a7e5454e518666ba9" alt="Python code snippet using A2A SDK to invoke a registered remote agent" width="2108" height="1718" data-path="images/remote_agent_a2a_code_snippet.png" />
</Frame>

<Tip>
  If you're building new agents — internal or external — consider building them A2A-compatible. It makes them registerable, composable, and interoperable by default.
</Tip>

***

## Monitoring

### Traces

Every request to a remote agent is traced through TrueFoundry. Navigate to **AI Monitoring > Request Traces** to inspect individual requests, including the full JSONRPC request/response payload for A2A agents.

<Frame caption="Request Traces for a remote A2A agent showing JSONRPC request and response">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/remote_agent_trace.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=f48de93d8cd9ab5e93d01e7f63c105d3" alt="Request Traces for a remote A2A agent showing JSONRPC request and response" width="3024" height="1710" data-path="images/remote_agent_trace.png" />
</Frame>

### Metrics

Agent-level metrics — request counts, latency percentiles, error rates, and failure breakdowns — are available under **AI Monitoring > Metrics > Agent Metrics**. Filter by agent name to drill into a specific remote agent's performance.

<Frame caption="Agent Metrics dashboard showing request volume, latency, and failure rates">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/agent_metrics.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=133c371f9ebf4518811b3c4b9684db08" alt="Agent Metrics dashboard showing request volume, latency, and failure rates" width="3024" height="1708" data-path="images/agent_metrics.png" />
</Frame>
