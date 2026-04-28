> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Overview

> The single place to build, register, discover, and govern every AI agent in your organization — whether built on TrueFoundry or running anywhere else.

Teams across your organization are already building AI agents. The problem isn't building them — it's everything that comes after. These show up the moment you go from two agents to twenty:

* **Discoverability** — No shared catalog means duplicated work and word-of-mouth discovery.
* **Access Control** — No consistent way to control who can invoke which agent.
* **Metrics** — No centralized view of which agents are active, how much traffic they handle, or what to invest in.
* **Traces & Logs** — No unified record of what an agent did, when, and who invoked it.
* **Governance** — No central layer to enforce content policies, rate limits, or guardrails.

Agent Registry solves this. It is a single platform where every agent in your organization is registered, discovered, accessed, and governed — regardless of where it runs.

<Frame caption="Agent Registry page listing all registered agents">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/agent_registry.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=f8892c67055260c6168fecea516615d7" alt="Agent Registry page listing all registered agents" width="3020" height="1656" data-path="images/agent_registry.png" />
</Frame>

## Types of Agents

There are two ways to add an agent to the registry. Click **Create New Agent** on the Agent Registry page to get started.

<Frame caption="Create New Agent button on the Agent Registry page">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/create_new_agent_button.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=6ee5dd1bf485d3138417d81a3dbfe558" alt="Create New Agent button on the Agent Registry page" width="2671" height="511" data-path="images/create_new_agent_button.png" />
</Frame>

<Frame caption="Create New Agent dialog with TrueFoundry and Remote options">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/new_agent.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=521d53ec8db619738c8f70c491263e1e" alt="Create New Agent dialog showing TrueFoundry and Remote options" width="1688" height="693" data-path="images/new_agent.png" />
</Frame>

<CardGroup cols={2}>
  <Card title="TrueFoundry Agent" icon="robot" href="/docs/ai-gateway/agents/truefoundry-agents">
    Build agents natively on TrueFoundry. The platform handles infrastructure, model routing, tool orchestration, and access control. You focus on the prompt and the tools.

    <br />

    <br />

    <Note>Available on **SaaS** only.</Note>
  </Card>

  <Card title="Remote Agent" icon="globe" href="/docs/ai-gateway/agents/remote-agents">
    Register an agent running elsewhere — Bedrock, LangGraph, a custom HTTP service, A2A agents, anything. Point the registry at it, and it gets the same catalog, controls, and observability.

    <br />

    <br />

    <Info>Available on both **SaaS** and **Self-hosted**.</Info>
  </Card>
</CardGroup>

## Unified Observability

Every registered agent — whether built on TrueFoundry or running remotely — gets centralized metrics and traces out of the box. Request counts, latency percentiles, error rates, and usage data are all available from a single dashboard.

<Frame caption="Unified Agent Metrics dashboard showing request volume, latency, and failure rates">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/agent_metrics.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=133c371f9ebf4518811b3c4b9684db08" alt="Agent Metrics dashboard" width="3024" height="1708" data-path="images/agent_metrics.png" />
</Frame>
