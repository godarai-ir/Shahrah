> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Cedar Guardrails

> Learn how to use Cedar policy language to create fine-grained access control guardrails for MCP tool invocations

Cedar Guardrails enable you to implement fine-grained access control policies for MCP (Model Context Protocol) tool invocations using the [Cedar policy language](https://www.cedar-policy.org/). Cedar provides a powerful, declarative way to define who can execute which tools under what conditions, ensuring secure and compliant agentic workflows.

## Guardrail Overview

Cedar Guardrails use the Cedar policy language to evaluate authorization decisions for MCP tool executions. When an agent attempts to invoke an MCP tool, the Cedar guardrail evaluates the request against your defined policies to determine whether the action should be permitted or forbidden.

### Key Features

* **Fine-grained Access Control**: Control access at the level of individual tools, users, teams, and even specific tool arguments
* **Context-aware Policies**: Make authorization decisions based on request context, including tool names, arguments, and user attributes
* **Default Deny**: By default, all actions are denied unless explicitly permitted by a policy
* **Declarative Syntax**: Write policies in a clear, readable format that's easy to understand and maintain
* **Real-time Evaluation**: Policies are evaluated in real-time for each tool invocation

### When to Use Cedar Guardrails

Cedar Guardrails are ideal for:

* **MCP Tool Access Control**: Restrict which users or teams can invoke specific MCP tools
* **Argument-based Restrictions**: Allow or deny tool invocations based on specific argument values
* **Team-based Permissions**: Implement team-level access controls for tool usage
* **Compliance Requirements**: Enforce organizational policies and compliance rules

## Entities

Cedar policies operate on **entities**, which represent the actors and resources in your authorization model. In the context of MCP tool guardrails, the following entity types are used:

### Principal Entity

The **principal** represents the user, virtual account, or agent attempting to invoke the tool. Principal entities can be either:

**User** entities have the following structure:

```
User {
    id: String (optional),
    email: String (required),
    tenantName: String (required),
    userName: String (optional),
    teamNames: Set<String> (required)
}
```

**VirtualAccount** entities have the following structure:

```
VirtualAccount {
    id: String (required),
    name: String (required),
    tenantName: String (required)
}
```

### Resource Entity

The **resource** represents the MCP server or tool being accessed. Resource entities have the following structure:

```
MCPServer {
    name: String (optional)
}
```

### Context Entity

The **context** provides information about the current request, including:

```
Context {
    tool_name: String (required),
    tool_args: Set<{key: String, value: String}> (required)
}
```

## Schema

The Cedar schema defines the structure of entities, actions, and common types used in your policies. Here's the schema for MCP tool guardrails:

```json theme={"dark"}
{
  "": {
    "entityTypes": {
      "User": {
        "shape": {
          "type": "Record",
          "attributes": {
            "id": { "type": "String", "required": false },
            "email": { "type": "String", "required": true },
            "tenantName": { "type": "String", "required": true },
            "userName": { "type": "String", "required": false },
            "teamNames": { 
              "type": "Set", 
              "element": { "type": "String" }, 
              "required": true 
            }
          }
        }
      },
      "VirtualAccount": {
        "shape": {
          "type": "Record",
          "attributes": {
             "id": { "type": "String", "required": true },
             "name": { "type": "String", "required": true },
             "tenantName": { "type": "String", "required": true }
          }
        }
      },
      "MCPServer": {
        "shape": {
          "type": "Record",
          "attributes": { "name": { "type": "String", "required": false } }
        }
      }
    },
    "actions": {
      "execute_tool": {
        "appliesTo": {
          "principalTypes": ["User", "VirtualAccount"],
          "resourceTypes": ["MCPServer"],
          "context": {
            "type": "Record",
            "attributes": {
              "tool_name": { "type": "String", "required": true },
              "tool_args": { 
                "type": "Set", 
                "element": {
                  "type": "Record",
                  "attributes": {
                    "key": { "type": "String", "required": true },
                    "value": { "type": "String", "required": true }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
```

### Schema Components

* **Entity Types**: Define the structure of principals (User, VirtualAccount) and resources (MCPServer)
* **Actions**: Define the operations that can be performed (e.g., `execute_tool`) and their context requirements

## Example Policies

Here are practical examples of Cedar policies for MCP tool access control:

### Example 1: Forbid Specific User from Executing Tool with Specific Arguments

This policy prevents a specific user from executing a tool when certain conditions are met:

```cedar theme={"dark"}
forbid (
    principal, 
    action == Action::"execute_tool",
    resource == MCPServer::"test-mcps-deepwiki"
) when {
    principal is User &&
    principal.email == "person@truefoundry.com" &&
    context.tool_name == "ask_question" &&
    context.tool_args.contains({ key: "repoName", value: "truefoundry/CruiseKube" })
};
```

**What this policy does:**

* Blocks the user with email `person@truefoundry.com` from executing the `ask_question` tool
* Only applies when the tool is invoked on the `test-mcps-deepwiki` MCP server
* Specifically blocks when the `repoName` argument equals `"truefoundry/CruiseKube"`

### Example 2: Permit Team Members to Execute Tool

This policy allows members of a specific team to execute a tool:

```cedar theme={"dark"}
permit (
    principal, 
    action == Action::"execute_tool",
    resource == MCPServer::"test-mcps-deepwiki"
) when {
    principal is User &&
    principal.teamNames.contains("test-team") &&
    context.tool_name == "ask_question" &&
    context.tool_args.contains({ key: "repoName", value: "truefoundry/CruiseKube" })
};
```

**What this policy does:**

* Allows any user who is a member of the `"test-team"` team to execute the `ask_question` tool
* Only applies when the tool is invoked on the `test-mcps-deepwiki` MCP server
* Specifically permits when the `repoName` argument equals `"truefoundry/CruiseKube"`

### Additional Policy Examples

#### Permit All Users for Specific Tool

```cedar theme={"dark"}
permit (
    principal,
    action == Action::"execute_tool",
    resource == MCPServer::"test-mcps-deepwiki"
) when {
    principal is User &&
    context.tool_name == "get_weather"
};
```

#### Forbid Access to Sensitive Repositories

```cedar theme={"dark"}
forbid (
    principal,
    action == Action::"execute_tool",
    resource
) when {
    principal is User &&
    context.tool_name == "read_file" &&
    context.tool_args.contains({ key: "path", value: "/etc/passwd" })
};
```

#### Permit Based on Multiple Team Membership

```cedar theme={"dark"}
permit (
    principal,
    action == Action::"execute_tool",
    resource == MCPServer::"production-mcp"
) when {
    principal is User &&
    (principal.teamNames.contains("platform-team") || 
     principal.teamNames.contains("devops-team"))
};
```

## Safety and Enforcement

Cedar Guardrails implement a **default deny** security model, ensuring that only explicitly permitted actions are allowed.

### Default Deny Model

**By default, all tool invocations are denied** unless a policy explicitly permits them. This means:

1. **No Policy = Denied**: If no policy matches a request, the action is automatically denied
2. **Explicit Permit Required**: An action is only allowed if at least one `permit` policy matches
3. **Forbid Overrides Permit**: If any `forbid` policy matches, the action is denied even if a `permit` policy also matches

### Enforcement Points

Cedar Guardrails are enforced at the **MCP Pre Tool** hook, which means:

* Policies are evaluated **before** the tool is actually invoked
* If a policy denies the request, the tool invocation is blocked
* The user receives an authorization error without the tool being executed
* This prevents unauthorized access

### Configuration

Users can configure Cedar guardrail rules on the **MCP Pre Tool** hook from **AI Gateway → Controls → Guardrails**. Navigate to the Guardrails section and add your Cedar guardrail to the `mcp_tool_pre_invoke_guardrails` field in your guardrail configuration rules.

### Best Practices for Safety

1. **Principle of Least Privilege**: Start with deny-all and add specific permits only for necessary access
2. **Validate Tool Arguments**: Always validate tool arguments in your policies to prevent unauthorized access to specific resources
3. **Team-based Access**: Prefer team-based permissions over individual user permissions for easier management
4. **Regular Policy Review**: Regularly review and audit your policies to ensure they match your security requirements

### Error Handling

When a Cedar guardrail denies a request:

* The tool invocation is blocked before execution
* An authorization error is returned to the caller
* The error message indicates that access was denied by a Cedar policy
* Request traces will show which policy (if any) matched and caused the denial

<Tip>
  **Security First**: The default deny model ensures that your system is secure by default. Always test your policies thoroughly in a development environment before deploying to production.
</Tip>

<Note>
  Cedar policies are evaluated in real-time for each tool invocation. This ensures that changes to user attributes (like team membership) are immediately reflected in authorization decisions without requiring policy updates.
</Note>
