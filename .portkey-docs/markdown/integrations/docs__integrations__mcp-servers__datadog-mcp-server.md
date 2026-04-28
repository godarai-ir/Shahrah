> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Datadog MCP server

> The Datadog MCP server enables AI agents to interact with Datadog monitoring, dashboards, metrics, logs, and alerts through MCP. Built for conversational and automated observability workflows.

## When you should use this server

* Query and investigate **logs, traces, and spans** directly via natural language.
* Retrieve **metrics, timeseries data, and monitors** without manual dashboard navigation.
* Access **infrastructure insights** like host details for debugging or optimization.
* Manage and query **ongoing incidents** for faster triage and response.
* Discover and reference **dashboards** for context in troubleshooting or reporting.

## Key features

* Logs and traces analysis
* Metrics querying and visualization
* Monitor status and configuration
* Infrastructure monitoring
* Incident management
* Dashboard discovery

## Requirements

* **Hosting**: Remote MCP server (no local setup required)
* **Supported clients**: Works with Claude Code, Codex, Goose, Cursor, and any MCP-compatible tool

## Tools provided

### get\_logs

Retrieve a list of logs based on query filters.

### list\_spans

Investigate spans relevant to a query.

### get\_trace

Retrieve all spans from a specific trace.

### list\_metrics

Retrieve a list of available metrics in the environment.

### get\_metrics

Query timeseries metrics data.

### get\_monitors

Retrieve monitors and their configurations.

### list\_hosts

Provide detailed host information.

### list\_incidents

Retrieve a list of ongoing incidents.

### get\_incident

Retrieve details for a specific incident.

### list\_dashboards

Discover available dashboards and their context.
