> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Auth Overrides

> Replace a server's default credentials with user-specific tokens for MCP servers in TrueFoundry AI Gateway

Auth Overrides let you replace an MCP server's default outbound credentials with user-specific tokens. When an override exists for the caller, the Gateway uses that value instead of the server's configured defaults for upstream requests.

## When to use Auth Overrides

Auth Overrides are useful when:

* The upstream MCP server requires **user-specific API keys** (e.g., each developer has their own key for a third-party service)
* You need to **attribute API usage** to individual users at the downstream service
* The upstream service enforces **per-user rate limits or permissions**

## Managing Auth Overrides

Navigate to the MCP server detail page and click on the **Auth Overrides** tab to view all existing overrides.

<img src="https://mintcdn.com/truefoundry/yR_clVDeJDlQkXKY/images/2026-04-06_23.20.41.png?fit=max&auto=format&n=yR_clVDeJDlQkXKY&q=85&s=353dcd3611f2402eb0ccc155290d037b" alt="Auth Overrides Tab" width="1739" height="911" data-path="images/2026-04-06_23.20.41.png" />

The Auth Overrides table shows the following columns:

| Column          | Description                                         |
| --------------- | --------------------------------------------------- |
| **Entity**      | The user or virtual account the override applies to |
| **Type**        | Whether the entity is a user or virtual account     |
| **Auth**        | The authentication type being overridden            |
| **Updated by**  | Who last modified this override                     |
| **Credentials** | The masked credential value                         |

### Adding an Auth Override

<Steps>
  <Step title="Open the Auth Overrides tab">
    Navigate to the MCP server detail page and click the **Auth Overrides** tab.
  </Step>

  <Step title="Click Add Override">
    Click the **+ Add Override** button in the top right corner.
  </Step>

  <Step title="Configure the override">
    * **Subject** (managers only): Select the user or virtual account that will use this override
    * **Credential**: The API key, secret, or full header value as required by the server's auth mode (for non-template configs, provide the complete value used for `Authorization` header)

          <img src="https://mintcdn.com/truefoundry/yR_clVDeJDlQkXKY/images/auth-overide-token.png?fit=max&auto=format&n=yR_clVDeJDlQkXKY&q=85&s=5c388ab170330da5c0c0c2339b4c10e7" alt="Add Auth Override" width="1753" height="994" data-path="images/auth-overide-token.png" />
  </Step>

  <Step title="Submit">
    Click **Submit** to save the override. The Gateway will now use this credential for all requests made by the selected subject.
  </Step>
</Steps>

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cmnncinco4zj6abur0879bjq4?embed_v=2" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/cmnncinco4zj6abur0879bjq4?embed_v=2" style={{ border:"none",display:"flex",margin:"auto" }} />

### Editing or Deleting an Auth Override

**Update** - click the edit icon next to the override in the Auth Overrides table.

**Delete** - click the delete icon, the user will revert to the server's default credentials (if any).

<Note>
  Changing outbound auth (type, Shared vs Individual, or related settings) can **delete** overrides on some transitions.
</Note>

## Permissions

### MCP Server Manager

* Open the **Auth Overrides** tab on supported servers.
* **List** all overrides. The table columns are **Entity**, **Type**, **Auth**, **Updated by**, and **Credentials** (masked values).
* **Delete any** override.

### User (non-manager)

* **Add** overrides only for **API Key** with **Individual Credentials**
* **View** overrides **they** added.
* **Delete** only **their** override.

## Override support

### By outbound auth

| Outbound auth                        | Who can add overrides                    |
| ------------------------------------ | ---------------------------------------- |
| **API Key** (Individual Credentials) | MCP Server Managers and MCP Server Users |
| **API Key** (Shared Credentials)     | *Not supported* - coming soon            |
| **OAuth2** (Authorization Code)      | MCP Server Managers only                 |
| **OAuth2** (Client Credentials)      | *Not supported* - coming soon            |
| **Token Passthrough**                | *Not supported*                          |

<Tip>
  For details on configuring API Key authentication with Shared or Individual Credentials, see the [Getting Started](/docs/ai-gateway/mcp/mcp-server-getting-started) guide.
</Tip>

### By Subject

| Subject               | Who can add overrides                    |
| --------------------- | ---------------------------------------- |
| **User**              | MCP Server Managers and MCP Server Users |
| **Virtual Account**   | MCP Server Managers                      |
| **External Identity** | *Not supported* - coming soon            |

### By Auth override type

| Auth override          | Who can add overrides                                       |
| ---------------------- | ----------------------------------------------------------- |
| **Header**             | MCP Server Managers and MCP Server Users (where applicable) |
| **Client Credentials** | *Not supported* - coming soon                               |

## How Auth Overrides work at runtime

When a user makes a request through the Gateway to an MCP server with Auth Overrides configured:

1. The Gateway authenticates the user (inbound authentication).
2. The Gateway checks if an Auth Override exists for this user on this MCP server.
3. **If an override exists**: The Gateway injects the user's override credentials into the request (template substitution for Individual, or full `Authorization` value when applicable).
4. **If no override exists**: The Gateway uses the server's default credentials (for Shared Credentials) or returns an error (for Individual Credentials).

<Note>
  For MCP servers configured with **Individual Credentials**, users without an Auth Override will not be able to use the server until an override is created for them.
</Note>
