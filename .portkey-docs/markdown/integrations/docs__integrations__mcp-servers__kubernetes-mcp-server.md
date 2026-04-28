> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Kubernetes MCP server

> The Kubernetes MCP server provides a flexible Model Context Protocol (MCP) interface for managing and interacting with Kubernetes and OpenShift clusters.

## When should you use it

Use the Kubernetes MCP server when you want to:

* Manage Kubernetes and OpenShift clusters through **AI assistants**.
* Perform **CRUD operations** on Kubernetes resources (pods, deployments, services, etc.).
* Work with **Helm charts** (install, list, uninstall).
* Query logs, metrics, and events from the cluster.
* Automate operational workflows like **resource scaling, monitoring, and troubleshooting**.
* Extend support to OpenShift with **projects** and additional resource types.

## Requirements

* **Supported platforms**: Kubernetes and OpenShift.
* **Configuration**:
  * Uses `~/.kube/config` or in-cluster config.
  * Automatically detects changes and reloads.
* **Dependencies**: Kubernetes cluster access with sufficient RBAC permissions.

## Tools

### Configuration

**configuration\_view** — View the current Kubernetes configuration (full or minified).

### Events

**events\_list** — List Kubernetes events across namespaces (or filtered by namespace).

### Helm

**helm\_install** — Install a Helm chart.

**helm\_list** — List Helm releases (all or specific namespaces).

**helm\_uninstall** — Uninstall a Helm release.

### Namespaces & Projects

**namespaces\_list** — List all namespaces in the cluster.

**projects\_list** — List all OpenShift projects.

### Pods

**pods\_list** — List all pods in the cluster.

**pods\_list\_in\_namespace** — List pods in a specific namespace.

**pods\_get** — Get details of a specific pod.

**pods\_log** — Retrieve logs from a pod (with container and previous options).

**pods\_exec** — Execute commands inside a pod container.

**pods\_delete** — Delete a pod by name.

**pods\_run** — Run a new pod with a specified image.

**pods\_top** — Retrieve pod CPU/memory usage (via metrics server).

### Resources (Generic CRUD)

**resources\_create\_or\_update** — Create or update a Kubernetes resource from YAML/JSON.

**resources\_get** — Retrieve a specific resource.

**resources\_list** — List resources (by API version, kind, namespace, label selectors).

**resources\_delete** — Delete a resource by kind and name.

### Dashboards & Monitoring

**pods\_top** — Get resource usage for pods (cluster-wide or namespace-specific).

## Notes

* Works with both **self-managed Kubernetes clusters** and **OpenShift**.
* Supports **Helm operations** natively for release management.
* RBAC permissions must match the requested operations (e.g., creating pods, deleting resources).
* Ideal for use cases where AI agents need to act as **cluster copilots** for devops, troubleshooting, and automation.
