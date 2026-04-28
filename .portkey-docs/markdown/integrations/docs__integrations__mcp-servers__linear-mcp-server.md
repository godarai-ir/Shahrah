> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Linear MCP server

> The Linear MCP server provides a standardized interface for AI models and agents to securely access and manage data from your Linear workspace.

Linear's MCP server is **centrally hosted and managed**, following the authenticated remote MCP specification. It supports creating, updating, and retrieving objects across Linear in a way that respects your existing permissions.

## **When should you use it?**

Use the Linear MCP server when you want to:

* Search and retrieve issues, projects, cycles, and documents from Linear.

* Create and update issues, projects, or comments via natural language commands.

* Integrate Linear data directly into AI-powered workflows (e.g., task planning, meeting follow-ups).

* Provide secure, real-time access to Linear data without custom integrations.

## **Endpoints**

* **Transport options**:

  * HTTP: `https://mcp.linear.app/mcp`

  * SSE: `https://mcp.linear.app/sse`

* **Authentication**: OAuth 2.1 with dynamic client registration.

* **Recommendation**: Use the **streamable HTTP endpoint** where supported for improved reliability.

* **Compatibility**: Works natively with Claude, Cursor, and other MCP clients. For older clients, the `mcp-remote` module ensures backwards compatibility.

## **Tools**

### **list\_comments** — Retrieve a list of comments.

### **create\_comment** — Create a new comment.

### **list\_cycles** — Retrieve all cycles.

### **get\_document** — Retrieve a specific document.

### **list\_documents** — List all documents.

### **get\_issue** — Get details of a specific issue.

### **list\_issues** — List issues in the workspace.

### **create\_issue** — Create a new issue.

### **update\_issue** — Update an existing issue.

### **list\_issue\_statuses** — Retrieve all available issue statuses.

### **get\_issue\_status** — Retrieve details of a specific issue status.

### **list\_my\_issues** — List issues assigned to the authenticated user.

### **list\_issue\_labels** — Retrieve all available issue labels.

### **list\_projects** — List all projects.

### **get\_project** — Retrieve details of a specific project.

### **create\_project** — Create a new project.

### **update\_project** — Update an existing project.

### **list\_teams** — List all teams.

## **Notes**

* The server is **centrally hosted** by Linear; no local deployment is required.

* Supports both **SSE** and **HTTP** transports, though HTTP is recommended.
