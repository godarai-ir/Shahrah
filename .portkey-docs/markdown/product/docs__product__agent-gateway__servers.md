> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Agent Servers

> Host your agent servers

The Agent Servers are the virtual servers that act as a client to the upstream agent server and as a server to the downstream agent clients.

## Introduction

Instead of your clients (your users/your services) directly interacting with the upstream agent server,
portkey allows you to

* create virtual servers and provision access with virtual keys,
* control access to the agent itself and the specific skills and capabilities that the agent provides.

this way your clients can be restricted to the necessary scopes and capabilities with full observability and logging

## Usage

Simply replace your agent server url wherever you are invoking it with the virtual server url.

For example if your slug is `test-server` your url becomes `https://agents.portkey.ai/v1/agent/test-server`.

<Note>
  For self-hosted deployments, replace `agents.portkey.ai` with your gateway
  URL.
</Note>

## logging

Every request to the virtual server will be logged and can be viewed in the [logs](/product/observability/logs) section.

This forms the basis of the [observability](/product/observability) and [governance](/product/governance) features at portkey.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/H3BZmx0ZdvnzJ585/images/agent-gateway/log.png?fit=max&auto=format&n=H3BZmx0ZdvnzJ585&q=85&s=1b8933b4c2396f8db8276b160f78562f" width="2924" height="1600" data-path="images/agent-gateway/log.png" />
</Frame>
