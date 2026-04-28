> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Getting Started

> Learn how to create MCP Server Groups, add MCP servers, and use them in the TrueFoundry AI Gateway playground.

This guide will walk you through the steps to add a MCP server to the TrueFoundry AI Gateway and then use it in the playground or through
Cursor/VSCode/Claude or in your code agents.

<Steps>
  <Step title="Add an MCP Server">
    Navigate to **MCP Gateway** in the left sidebar and click the **Add MCP Server** button. You will see the following options:

    <img src="https://mintcdn.com/truefoundry/yR_clVDeJDlQkXKY/images/docs/ai-gateway/add-mcp-server-options.png?fit=max&auto=format&n=yR_clVDeJDlQkXKY&q=85&s=6157985cfcc7ff4262eb4fb540ec67b5" alt="Add MCP Server Options" width="1024" height="495" data-path="images/docs/ai-gateway/add-mcp-server-options.png" />

    | Option                                  | Description                                                                                                                          |
    | --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
    | **Connect Official Remote MCP Servers** | Connect to official MCP servers like GitHub, Sentry, Atlassian, Figma, etc. with your own auth                                       |
    | **Connect any Remote MCP Server**       | Provide the URI and auth credentials to connect any remote MCP server                                                                |
    | **Create a Virtual MCP Server**         | Combine multiple MCPs into a bundle so each user or team gets access to exactly the tools they need                                  |
    | **Import from OpenAPI Spec**            | Convert your existing APIs to MCP server by importing OpenAPI spec                                                                   |
    | **Hosted Stdio-based MCP Server**       | Register a CLI-style MCP server by command and arguments; see [Hosted Stdio-based MCP Server](/docs/ai-gateway/mcp/stdio-mcp-server) |

    If you want to build an MCP server from scratch, you can follow the [Create Calculator MCP Server](/docs/ai-gateway/mcp/create-calculator-mcp-server) tutorial.

    To add a custom Remote MCP server, click **Connect any Remote MCP Server** and provide the following details:

    | Field             | Description                                         |
    | ----------------- | --------------------------------------------------- |
    | **URL**           | The URL of the MCP server                           |
    | **Collaborators** | Users and teams that have access to this MCP server |

    #### Collaborators

    The Collaborators section allows you to control who can access and manage this MCP server. You can add individual users or entire teams and assign them specific roles:

    <img src="https://mintcdn.com/truefoundry/02OfhC3cmaN0Ncb8/images/docs/ai-gateway/mcp-server-collaborators.png?fit=max&auto=format&n=02OfhC3cmaN0Ncb8&q=85&s=48b20a8d9e99249dd186879a5b7f27da" alt="MCP Server Collaborators" width="3840" height="1854" data-path="images/docs/ai-gateway/mcp-server-collaborators.png" />

    | Role                   | Description                                                                                                            |
    | ---------------------- | ---------------------------------------------------------------------------------------------------------------------- |
    | **MCP Server Manager** | Full access to manage the MCP server, including editing configuration, managing collaborators, and deleting the server |
    | **MCP Server User**    | Can use the MCP server's tools in the playground and IDEs, but cannot modify server settings                           |

    Click **+ Add Collaborators** to add more users or teams. You can also click **View Permission Details** to see the complete list of permissions for each role.

    #### Auth Data

    When registering an MCP server, choose the authentication type that determines how requests are authenticated with the upstream MCP server. You can select from: `API Key`, `OAuth2`, or `Token Passthrough`.

    <AccordionGroup>
      <Accordion title="API Key">
        Authenticate using a static API key or token sent as a request header. When you select **API Key**, you need to choose between two credential modes:

        ##### Shared Credentials

        One key is used by everyone. All users share the same downstream API key. Configure the header name (e.g., `Authorization`) and the corresponding value.

        <img src="https://mintcdn.com/truefoundry/yR_clVDeJDlQkXKY/images/docs/ai-gateway/add-mcp-server-api-key-shared.png?fit=max&auto=format&n=yR_clVDeJDlQkXKY&q=85&s=bdd8c03ce7241753703264147b1e6f41" alt="API Key with Shared Credentials" width="1024" height="494" data-path="images/docs/ai-gateway/add-mcp-server-api-key-shared.png" />

        <Note>
          Use shared credentials for MCP servers where all users should have the same level of access, such as shared knowledge bases or read-only APIs.
        </Note>

        ##### Individual Credentials

        Each user provides their own key. When **Individual Credentials** is selected, configure the header name and use a placeholder like `Bearer {{API_KEY}}` in the value field. Each user will need to provide their own API key before they can use the MCP server.

        <img src="https://mintcdn.com/truefoundry/yR_clVDeJDlQkXKY/images/docs/ai-gateway/add-mcp-server-api-key-individual.png?fit=max&auto=format&n=yR_clVDeJDlQkXKY&q=85&s=e6d65900ec50a61da0e9db7d4d885498" alt="API Key with Individual Credentials" width="1024" height="494" data-path="images/docs/ai-gateway/add-mcp-server-api-key-individual.png" />

        When individual credentials are configured, users can supply their API key through the **Auth Overrides** tab on the MCP server detail page. See [Auth Overrides](/docs/ai-gateway/mcp/mcp-server-auth-overrides) for details.

        <Note>
          Use individual credentials when the upstream MCP server requires user-specific API keys, for example, when each user has their own API key for a third-party service.
        </Note>
      </Accordion>

      <Accordion title="OAuth2">
        Users sign in through a third-party login flow to get temporary access tokens. OAuth2 is supported by popular services like Slack, GitHub, Atlassian, Google, and more.

        <Tip>
          **Recommended for production.** OAuth2 allows you to configure scopes (e.g., read-only), users can revoke their own authorization, and access is limited to resources each user is permitted to use.
        </Tip>

        When you select **OAuth2**, you need to choose a **Grant Type**:

        <img src="https://mintcdn.com/truefoundry/yR_clVDeJDlQkXKY/images/docs/ai-gateway/mcp-server-oauth2-config.png?fit=max&auto=format&n=yR_clVDeJDlQkXKY&q=85&s=e810314f881eb955da194e66e8fde5d9" alt="OAuth2 Configuration" width="1024" height="494" data-path="images/docs/ai-gateway/mcp-server-oauth2-config.png" />

        ##### Authorization Code

        Standard flow where the user is redirected to the provider to authorize. This is the most common OAuth2 flow for user-facing applications.

        | Field                                | Description                                                         |
        | ------------------------------------ | ------------------------------------------------------------------- |
        | **Authorization URL**                | The URL where users are redirected to authorize (optional)          |
        | **Token URL**                        | The URL used to exchange the authorization code for an access token |
        | **Client ID**                        | The client ID from your OAuth2 app (optional)                       |
        | **Client Secret**                    | The client secret from your OAuth2 app (optional)                   |
        | **Registration URL**                 | Used for Dynamic Client Registration (optional)                     |
        | **Code Challenge Methods Supported** | List of supported PKCE code challenge methods, S256 only (optional) |
        | **JWT Source**                       | Where to extract the JWT from — Access Token or ID Token            |

        ##### Client Credentials

        Server-to-server flow with shared secret. No user interaction is required. This is suitable for backend services that need to authenticate without user involvement.

        | Field             | Description                             |
        | ----------------- | --------------------------------------- |
        | **Token URL**     | The URL used to obtain the access token |
        | **Client ID**     | The client ID for the service           |
        | **Client Secret** | The client secret for the service       |

        <Steps>
          <Step title="Create an OAuth2 app in your provider's developer portal">
            Set the redirect URI to:

            ```bash theme={"dark"}
            https://<tfy-control-plane-base-url>/api/svc/v1/llm-gateway/mcp-servers/oauth2/callback
            ```

            Replace `<tfy-control-plane-base-url>` with your TrueFoundry control plane URL. Note your OAuth2 App ID, Secret, and required scopes.
          </Step>

          <Step title="Register the MCP Server in the AI Gateway">
            1. Navigate to **MCP Gateway** in the left sidebar and click **Add MCP Server**
            2. Select **Connect any Remote MCP Server**
            3. Provide the MCP server URL and select **OAuth2** as the authentication type
            4. Choose the appropriate grant type (**Authorization Code** or **Client Credentials**)
            5. Fill in the required OAuth2 fields

            <Tip>
              Store OAuth2 credentials in the TrueFoundry secrets store and reference their FQN for enhanced security.
            </Tip>
          </Step>

          <Step title="Connect and authorize">
            1. Click **Add Tool/MCP Servers** in the AI Gateway UI
            2. For OAuth2 MCP Servers using Authorization Code flow, click **Connect Now** to authorize

            <Frame caption="Authorize your MCP Server in AI Gateway">
              <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/ai-gateway-add-tool-mcp-server-authorise-view.png?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=50549d1e22bca79b488d4277dcc70ee4" alt="OAuth2 authorization interface for MCP Server in AI Gateway" width="3596" height="1934" data-path="images/ai-gateway-add-tool-mcp-server-authorise-view.png" />
            </Frame>

            3. Once authorized, the MCP Server's tools appear in the list

            <Frame caption="MCP Server tools available after authentication">
              <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/ai-gateway-add-tool-mcp-server-list-tools.png?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=844d32f8c8beb65e1b978da5d805910f" alt="MCP Server tools available after successful authentication" width="3598" height="1940" data-path="images/ai-gateway-add-tool-mcp-server-list-tools.png" />
            </Frame>

            4. You can revoke authorization at any time:

            <Frame caption="Revoke your OAuth2 authorization">
              <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/ai-gateway-add-tool-mcp-server-authorise-revoke.png?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=5a4614f4161bee8581ec44805bf28c26" alt="TrueFoundry Gateway Dashboard showing how to revoke OAuth2 authorization for an MCP Server" width="2568" height="776" data-path="images/ai-gateway-add-tool-mcp-server-authorise-revoke.png" />
            </Frame>
          </Step>
        </Steps>
      </Accordion>

      <Accordion title="Token Passthrough">
        The Gateway passes the user's JWT directly to the MCP server without transformation. The MCP server validates the token itself.

        **Configuration:**

        1. Register the MCP server with **Token Passthrough** authentication type
        2. Configure an External Identity to allow your IdP tokens
        3. The user's JWT is automatically forwarded to the MCP server

        <Note>
          With Token Passthrough, the MCP server is responsible for validating the token. Make sure the MCP server is configured to trust your IdP.
        </Note>
      </Accordion>
    </AccordionGroup>
  </Step>

  <Step title="Use MCP Server in Your IDE">
    Once you've added an MCP server, you can easily integrate it with your favorite IDE or AI coding assistant. Navigate to the MCP server details page and click on the **How To Use** tab.

    <img src="https://mintcdn.com/truefoundry/02OfhC3cmaN0Ncb8/images/docs/ai-gateway/use-mcp-server-ide.png?fit=max&auto=format&n=02OfhC3cmaN0Ncb8&q=85&s=291ea5d6d5f29a0dd1ff84d1701db6ca" alt="Use MCP Server in IDE" width="3840" height="1856" data-path="images/docs/ai-gateway/use-mcp-server-ide.png" />

    <Tip>
      Use the **Add MCP to Cursor** button to automatically add the MCP server configuration to your Cursor IDE. You can also click **Show API Key** to reveal the full authorization token, or **Copy** to copy the configuration to your clipboard.
    </Tip>

    <Tip>
      If no API key is shown in your code snippet, your version of the platform supports OAuth flow for obtaining the key. Cursor will automatically handle the OAuth flow when you use the MCP server.
    </Tip>
  </Step>

  <Step title="Use MCP Servers in Playground">
    You can select the MCP servers from the playground, select the tools and send your prompt to see which tools are being called.

    <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/cmh8a1iqy0txl6133h9xkswh7" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/cmh8a1iqy0txl6133h9xkswh7" style={{ border:"none",display:"flex",margin:"auto" }} />
  </Step>
</Steps>
