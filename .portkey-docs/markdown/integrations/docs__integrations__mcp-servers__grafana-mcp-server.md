> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Grafana MCP server

> The Grafana MCP server enables AI agents to interact with Grafana dashboards, metrics, logs, incidents, alerts, and monitoring tools through MCP. Built for observability and monitoring workflows.

## When should you use this server

* Search, retrieve, and update **dashboards** programmatically
* Query **Prometheus metrics** or **Loki logs** directly from MCP clients
* Access and manage **alert rules, contact points, and incidents**
* Integrate **OnCall schedules and shifts** into automated workflows
* Perform **Sift investigations, error detection, and performance profiling** via Pyroscope
* Generate deep links for Grafana resources that can be shared in AI-driven reports

## Key features

* Dashboard creation and management
* Metrics querying and visualization
* Log search and analysis
* Alerting and incident management
* OnCall schedule management
* Performance profiling with Pyroscope
* Deep links generation for sharing resources

## Authentication

* **Method:** RBAC permissions and API scopes depending on the tool
* **Notes:** Different tools require different permission levels

## Endpoints

* **Grafana version**: 9.0 or later (required for full functionality)
* **Supported environments**:
  * Local Grafana (`http://localhost:3000`)
  * Grafana Cloud (use your instance URL, e.g. `https://myinstance.grafana.net`)

## Tools provided

### list\_teams

List all teams in the Grafana organization.

### list\_users\_by\_org

List all users in an organization with their roles and permissions.

### search\_dashboards

Search for dashboards by name, tag, or other criteria.

### get\_dashboard\_by\_uid

Get a dashboard by its unique identifier (UID).

### update\_dashboard

Create a new dashboard or update an existing one.

### get\_dashboard\_panel\_queries

Get queries and datasource information for dashboard panels.

### get\_dashboard\_property

Extract dashboard data using JSONPath expressions.

### get\_dashboard\_summary

Retrieve a compact summary of a dashboard's content and structure.

### list\_datasources

List all configured datasources in the Grafana instance.

### get\_datasource\_by\_uid

Get a datasource by its unique identifier (UID).

### get\_datasource\_by\_name

Get a datasource by its name.

### query\_prometheus

Execute a PromQL query against Prometheus datasources.

### list\_prometheus\_metric\_metadata

List metadata about available Prometheus metrics.

### list\_prometheus\_metric\_names

List available metric names in Prometheus datasources.

### list\_prometheus\_label\_names

List available label names in Prometheus datasources.

### list\_prometheus\_label\_values

List possible values for a specific Prometheus label.

### query\_loki\_logs

Run LogQL queries against Loki log datasources.

### list\_loki\_label\_names

List available label names in Loki logs.

### list\_loki\_label\_values

List possible values for a specific Loki log label.

### query\_loki\_stats

Retrieve statistics about log streams and volume.

### list\_incidents

List incidents in Grafana Incident management.

### get\_incident

Get detailed information about a specific incident.

### create\_incident

Create a new incident in Grafana Incident management.

### add\_activity\_to\_incident

Add an activity entry to an existing incident.

### list\_alert\_rules

List alert rules configured in Grafana.

### get\_alert\_rule\_by\_uid

Get detailed information about a specific alert rule.

### list\_contact\_points

List alert notification contact points.

### list\_oncall\_schedules

List OnCall rotation schedules.

### get\_oncall\_shift

Get details about a specific OnCall shift.

### get\_current\_oncall\_users

Retrieve information about users currently on-call.

### list\_oncall\_teams

List teams configured in OnCall.

### list\_oncall\_users

List users participating in OnCall rotations.

### get\_sift\_investigation

Retrieve a Sift investigation by its UUID.

### get\_sift\_analysis

Get a specific analysis from a Sift investigation.

### list\_sift\_investigations

List available Sift investigations.

### find\_error\_pattern\_logs

Detect error patterns in Loki logs using Sift.

### find\_slow\_requests

Find slow requests from Tempo datasources.

### list\_pyroscope\_label\_names

List available label names in Pyroscope profiles.

### list\_pyroscope\_label\_values

List possible values for a specific Pyroscope label.

### list\_pyroscope\_profile\_types

List available profile types in Pyroscope.

### fetch\_pyroscope\_profile

Fetch a profile in DOT format for analysis.

### get\_assertions

Retrieve assertion summaries for monitored entities.

### generate\_deeplink

Generate shareable deep links to Grafana resources.

## Notes

* Tools respect **RBAC permissions** and require the correct **API scopes**
* Datasource-related operations may not work on **Grafana versions older than 9.0**
* Works with both **Grafana Cloud** and **self-managed instances**
