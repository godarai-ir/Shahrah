> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# TrueFoundry Agents

> Build, publish, and share AI agents natively on TrueFoundry with model selection, MCP tools, sandboxed execution, and built-in observability.

<Note>TrueFoundry Agents are available on **SaaS** only.</Note>

Build agents natively on TrueFoundry. You bring the intent — TrueFoundry handles everything else.

<CardGroup cols={2}>
  <Card title="Bring Your Own Model" icon="brain">
    Connect any LLM available through TrueFoundry's AI Gateway. Switch models without rewriting your agent.
  </Card>

  <Card title="MCP Server Integration" icon="plug">
    Plug in any MCP server to give your agent tools — file access, web search, APIs, databases, and more.
  </Card>

  <Card title="Sandbox Access" icon="box">
    Agents run in a secure execution environment. Code execution, file operations, and external calls are sandboxed by default.
  </Card>

  <Card title="ReAct Agent Framework" icon="arrows-spin">
    Agents follow a reason-then-act loop out of the box. A built-in chat interface lets you interact with them immediately.
  </Card>
</CardGroup>

TrueFoundry manages the platform problems so you don't have to: context window management, tool call orchestration, conversation history, system prompt handling, and multi-turn state. You focus on building and shipping.

***

## Creating a New Agent

On the Agent Registry page, click **Create New Agent** and select **Build Agent on TrueFoundry**.

<Frame caption="Create New Agent button on the Agent Registry page">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/create_new_agent_button.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=6ee5dd1bf485d3138417d81a3dbfe558" alt="Create New Agent button on the Agent Registry page" width="2671" height="511" data-path="images/create_new_agent_button.png" />
</Frame>

<Frame caption="Select Build Agent on TrueFoundry">
  <img src="https://mintcdn.com/truefoundry/CKXkvA2y5OE3Joq_/images/new_agent_truefoundry.png?fit=max&auto=format&n=CKXkvA2y5OE3Joq_&q=85&s=85fe518be8bce60ad294b4a43867b049" alt="Create New Agent dialog with Build Agent on TrueFoundry highlighted" width="1683" height="629" data-path="images/new_agent_truefoundry.png" />
</Frame>

This opens the Agent Playground where you build and test your agent.

<Steps>
  <Step title="Select a model">
    Choose any model available through AI Gateway.
  </Step>

  <Step title="Write your instructions">
    Add a system prompt that defines the agent's behavior, persona, and constraints.
  </Step>

  <Step title="Connect MCP Servers">
    Click the **+** next to MCP Servers to open the tool selection dialog. Search for MCP servers, browse available tools, and select the ones your agent needs.
  </Step>

  <Step title="Run and iterate">
    Use the chat interface to test the agent. Adjust the prompt, switch models, or add more tools until the agent behaves as expected.
  </Step>
</Steps>

<Frame caption="Agent Playground with model selection, system prompt, and MCP servers">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/agent_playground.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=88d902c12e3c572975e574d9f1fe7621" alt="Agent Playground with model selection, system prompt, and MCP servers" width="3010" height="1692" data-path="images/agent_playground.png" />
</Frame>

<Frame caption="Tools and MCP Servers selection dialog">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/agent_playground_mcp_server.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=bf26b6acf5f27006cec5848ae44086c6" alt="Tools and MCP Servers selection dialog" width="2662" height="1704" data-path="images/agent_playground_mcp_server.png" />
</Frame>

### Saving the Agent

Once you're satisfied with the agent's behavior, click **Save Agent** in the top-right corner. The dropdown gives you two options:

* **Save New Version** — Update an existing agent with your changes.
* **Save New Agent** — Create a brand-new agent from the current configuration.

<Frame caption="Save dropdown with Save New Version and Save New Agent options">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/agent_playground_save_agent.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=130496b685c54f92b55930e2009b55d8" alt="Save dropdown with Save New Version and Save New Agent options" width="684" height="368" data-path="images/agent_playground_save_agent.png" />
</Frame>

### Agent Configuration

To edit an agent's configuration after creation, click **Manage** on the Agent Registry page.

<Frame caption="Agent Registry showing the Manage button for a TrueFoundry agent">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/agent_registry_manage.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=ec8adf90b54480618b9a9ba163c47c3a" alt="Agent Registry with Manage button highlighted on a TrueFoundry agent" width="3020" height="1656" data-path="images/agent_registry_manage.png" />
</Frame>

The agent form contains the following fields:

| Field             | Description                                                                                 |
| ----------------- | ------------------------------------------------------------------------------------------- |
| **Name**          | A unique name for the agent within your organization.                                       |
| **Description**   | A short description of what the agent does.                                                 |
| **Tags**          | Optional tags for categorization and search.                                                |
| **Collaborators** | Team members or groups with **Agent Manager** (edit) or **Agent Access** (use) permissions. |

<Frame caption="Agent edit form showing name, description, and collaborator access control">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/truefoundry_agent_collaborators.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=0826d0598d582d7c2f29e3e1543e8dc4" alt="Agent edit form showing name, description, and collaborator access control" width="1650" height="1714" data-path="images/truefoundry_agent_collaborators.png" />
</Frame>

For details on prompt management and versioning, see [Prompt Management](/docs/ai-gateway/prompt-management). For MCP server configuration, refer to [Prompt Playground — MCP Servers](/docs/ai-gateway/prompt-playground#mcp-servers-add-tools-and-mcp-servers).

***

## Sharing and Using Agents

Once published, every agent appears on the Agent Registry page. Agents you have access to show **Try** and **Manage** actions.

### Try an Agent

Click **Try** on any agent in the registry to open a standalone chat page.

<Frame caption="Agent Registry showing the Try button for a TrueFoundry agent">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/agent_registry_try_now.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=60b893c57fc5263dcd866ec3e2f7bf30" alt="Agent Registry with Try button highlighted on a TrueFoundry agent" width="2653" height="935" data-path="images/agent_registry_try_now.png" />
</Frame>

This is a full conversational interface with chat history — ready for end-users to interact with immediately.

<Frame caption="Standalone agent chat page for interacting with a published agent">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/saved_agent_try.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=5ccdacddea497ae096937cef5db12723" alt="Standalone agent chat page for interacting with a published agent" width="3020" height="1094" data-path="images/saved_agent_try.png" />
</Frame>

### Integrate with Code

Every agent on TrueFoundry is API-accessible. The API contract is currently under active development and will be documented in the coming weeks.

***

## Monitoring

Every TrueFoundry agent execution is traced end-to-end. Traces capture each step of the agent's reasoning loop — LLM calls, tool invocations, sandbox executions — along with cost, token usage, and latency.

Navigate to **AI Monitoring > Request Traces** and filter by span type **AgentResponse** to inspect individual agent runs.

<Frame caption="Request Traces showing agent execution spans, cost, and token usage">
  <img src="https://mintcdn.com/truefoundry/Q_Txgy7BrCvpBj-e/images/truefoundry_agent_trace.png?fit=max&auto=format&n=Q_Txgy7BrCvpBj-e&q=85&s=f73673f5bfc9a5ac523c5082ea710f45" alt="Request Traces showing agent execution spans, cost, and token usage" width="3024" height="1710" data-path="images/truefoundry_agent_trace.png" />
</Frame>
