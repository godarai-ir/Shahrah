> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Claude Enterprise Security Guide For Developers

> A guide to Claude enterprise security for admins to govern access, secure data, and monitor deployments across web and CLI interfaces.

This guide covers enterprise security for **Claude.ai (web)** and **Claude Code (CLI)** — the two interfaces where admins can enforce proxy routing via `ANTHROPIC_BASE_URL`. Each has a different attack surface and set of admin controls; this guide covers everything an enterprise admin needs to deploy them securely.

This Claude enterprise security guide focuses on practical controls teams can roll out quickly. For enterprise security for Claude programs that need centralized policy and auth governance, start with [Claude enterprise security](/docs/ai-gateway/mcp/mcp-gateway-auth-security) architecture patterns.

<Tip>
  Configure SSO and domain capture before rolling out any Claude interface to
  your organization. Identity controls are the foundation everything else builds
  on.
</Tip>

***

## 1. Configure SSO

Configure SSO in the **Claude Admin Console** to enforce identity controls across web and CLI from a single place. Claude supports **SAML 2.0** and **OIDC** with any compliant IdP — including Okta, Azure AD (Entra ID), Auth0, and Google Workspace. [Setup guide](https://support.claude.com/en/articles/13132885-set-up-single-sign-on-sso)

<Steps>
  <Step title="Connect your IdP">
    In the Claude Admin Console, configure your SAML 2.0 or OIDC connection. Enable `Require SSO for Console` and `Require SSO for Claude` to enforce SSO-based authentication and inherit MFA from your IdP.
  </Step>

  <Step title="Enable domain capture">
    Claim your corporate email domains so that any sign-in attempt with a corporate address is automatically routed to the enterprise workspace — preventing employees from using personal accounts on any interface.
  </Step>

  <Step title="Map IdP groups to roles">
    Use IdP groups to automatically assign Claude roles to users. Revoking a user in your IdP immediately blocks access across web and CLI.

    | Role          | Capabilities                                                                           |
    | ------------- | -------------------------------------------------------------------------------------- |
    | Primary Owner | Full admin access including billing, SSO config, and admin role assignment             |
    | Admin         | Manage users, configure policies, view audit logs — cannot change billing or ownership |
    | Member        | Use Claude within the policies set by admins                                           |

    <Frame>
      <img src="https://mintcdn.com/truefoundry/b-bXYCa3Uo5Jq0k5/images/enterprise-security-claude-idp-groups-assignment.png?fit=max&auto=format&n=b-bXYCa3Uo5Jq0k5&q=85&s=f111a551cdc3fd763459c4a90fb16875" alt="Claude admin console showing identity and access settings with role-based group mapping for owner, admin, and user" width="2678" height="910" data-path="images/enterprise-security-claude-idp-groups-assignment.png" />
    </Frame>
  </Step>
</Steps>

### API Key Management

Claude Code and programmatic integrations authenticate via API keys — they are required for CLI usage, CI/CD pipelines, and any automation that calls Claude's API outside of a browser session. SSO covers interactive web logins, but API keys are the identity layer for everything else.

* Issue API keys through the Admin Console — developers should never use personal keys in a corporate context
* Store keys in a centralized secrets manager (AWS Secrets Manager, HashiCorp Vault, Azure Key Vault)
* Rotate keys quarterly and revoke immediately upon suspected compromise or employee offboarding

***

## 2. Configure Models Access

Control which models are available to your organization and how Claude traffic is routed. A strong [Claude enterprise security](/docs/ai-gateway/intro-to-llm-gateway) posture also requires centralized model governance and routing controls.

### Model Restrictions

Prevent developers from switching to higher-cost or unapproved models.

```json theme={"dark"}
{
  "allowedModels": ["claude-sonnet-4-5", "claude-haiku-4-5"]
}
```

### Proxy Routing

Use your own LLM Gateway to inspect, log and control all Claude traffic at the network layer and governance layer. Set these environment variables on developer machines ([Claude docs](https://code.claude.com/docs/en/model-config#environment-variables)):

```bash theme={"dark"}
export ANTHROPIC_BASE_URL=https://your-gateway.internal.corp
export HTTPS_PROXY=https://proxy.your-company.com:8080
export ANTHROPIC_DEFAULT_OPUS_MODEL=your-opus-model-equivalent
export ANTHROPIC_DEFAULT_SONNET_MODEL=your-sonnet-model-equivalent
```

<Note>
  These environment variables apply to Claude Code (CLI) only. Model routing for
  Claude.ai (Web) is controlled at the organization level through the Admin
  Console.
</Note>

### Enforcing Model Routing Across Developer Machines

Setting environment variables on a single machine is straightforward — the challenge is enforcing them consistently across every developer in your organization. There are 3 approaches, each with different trade-offs.

<Tabs>
  <Tab title="MDM / Endpoint-Managed">
    Push a `managed-settings.json` file to every machine via your MDM solution (Jamf, Kandji, Mosyle, Intune) and lock it against modification at the OS level. This is Claude Code's **endpoint-managed settings** approach ([Claude docs](https://code.claude.com/docs/en/settings#settings-files)).

    ```json theme={"dark"}
    {
      "model": "sonnet",
      "availableModels": ["sonnet", "haiku"],
      "env": {
        "ANTHROPIC_BASE_URL": "https://your-gateway.internal.corp",
        "ANTHROPIC_DEFAULT_OPUS_MODEL": "your-opus-model-equivalent",
        "ANTHROPIC_DEFAULT_SONNET_MODEL": "your-sonnet-model-equivalent"
      }
    }
    ```

    **System-level paths:**

    * macOS: `/Library/Application Support/ClaudeCode/managed-settings.json`
    * Linux: `/etc/claude-code/managed-settings.json`

    Tamper-resistant and applies immediately at startup with no network dependency. Requires MDM infrastructure and only works on corporate-owned devices.
  </Tab>

  <Tab title="Server-Managed Settings (Beta)">
    Configure settings centrally via the **Claude Admin Console** (Admin Settings → Claude Code → Managed settings). Settings are delivered from Anthropic's servers when developers authenticate with their organization credentials — no file deployment needed ([Claude docs](https://code.claude.com/docs/en/server-managed-settings)).

    ```json theme={"dark"}
    {
      "model": "sonnet",
      "availableModels": ["sonnet", "haiku"],
      "permissions": {
        "deny": ["Bash(curl:*)", "Read(**/.env)"]
      },
      "disableBypassPermissionsMode": "disable"
    }
    ```

    Works on BYOD devices with no MDM needed — clients poll hourly for updates and cached settings persist through network failures. Currently in public beta with client-side enforcement only (sudo users can tamper).

    <Warning>
      Server-managed settings require a direct connection to `api.anthropic.com`. If
      you route Claude traffic through an LLM gateway via `ANTHROPIC_BASE_URL`,
      server-managed settings will be bypassed.
    </Warning>
  </Tab>

  <Tab title="Direct Cloud Provider (Bedrock / Vertex)">
    Route Claude Code directly to your cloud provider account. Model access is governed at the cloud IAM layer — developers can only use models you've enabled and provisioned in your AWS or GCP account ([Claude docs](https://code.claude.com/docs/en/third-party-integrations)).

    ```bash theme={"dark"}
    # AWS Bedrock
    export CLAUDE_CODE_USE_BEDROCK=1
    export AWS_REGION=us-east-1

    # Google Vertex AI
    export CLAUDE_CODE_USE_VERTEX=1
    export CLOUD_ML_REGION=us-east5
    export ANTHROPIC_VERTEX_PROJECT_ID=your-project-id
    ```

    Traffic stays within your VPC or GCP project with cloud-native audit logging and billing through your existing cloud contract. Locks you into a single provider with complex IAM setup, and you must [pin model versions](https://code.claude.com/docs/en/model-config#pin-models-for-third-party-deployments) manually.
  </Tab>
</Tabs>

<Tip>
  **Recommended approach:** Use **MDM / endpoint-managed settings** on
  corporate-managed devices for the strongest enforcement. For unmanaged or BYOD
  devices, use **server-managed settings** if you route through Anthropic's API
  directly, or **direct cloud provider** if you need network-level control.
</Tip>

### Route Through TrueFoundry AI Gateway

[TrueFoundry AI Gateway](/docs/ai-gateway/intro-to-llm-gateway) acts as a unified proxy layer between Claude Code and your model providers. Instead of configuring each provider directly, route all Claude traffic through the Gateway to get centralized access control, routing, rate limiting, and cost governance in one place.

Point Claude Code at TrueFoundry AI Gateway:

```bash theme={"dark"}
export ANTHROPIC_BASE_URL=https://<your-truefoundry-gateway-url>
```

Once routed through TrueFoundry, you can connect and access models from multiple providers — not just Anthropic's direct API. Add [provider accounts](/docs/ai-gateway/quick-start#add-provider-account-and-models) in the Gateway dashboard and Claude Code can reach all of them through a single endpoint.

<Tip>
  Use [Virtual Models](/docs/ai-gateway/virtual-model) to create a single model
  identifier that routes requests across multiple providers with weight-based,
  latency-based, or priority-based routing — and switch providers without
  changing any client configuration.
</Tip>

The Gateway enforces [access control](/docs/ai-gateway/gateway-access-control), [rate limits](/docs/ai-gateway/ratelimiting), [budget limits](/docs/ai-gateway/budgetlimiting), and [virtual model routing with failover](/docs/ai-gateway/virtual-model) on all requests before they reach the underlying provider.

### Summary: Model Access Approaches

| Approach                            | Enforcement                                  | Security Level | Key Trade-offs                                                                                                                                                                                   |
| ----------------------------------- | -------------------------------------------- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **MDM + TrueFoundry AI Gateway**    | OS-level settings lock + centralized gateway | Very High      | Route to any model across providers, configure fallback and load balancing, full tracing and guardrails. Requires MDM infrastructure.                                                            |
| **MDM + Anthropic Direct**          | OS-level settings lock + direct API          | High           | Simplest MDM setup — single provider, no gateway needed. Limited to Anthropic models only, no cross-provider fallback or centralized governance.                                                 |
| **MDM + AWS Bedrock**               | OS-level settings lock + cloud IAM           | Very High      | Traffic stays within your VPC, cloud-native audit logging via CloudTrail. Locked to a single cloud provider, complex IAM setup.                                                                  |
| **MDM + Google Vertex AI**          | OS-level settings lock + cloud IAM           | Very High      | Traffic stays within your GCP project, cloud-native audit logging. Locked to a single cloud provider, complex IAM setup.                                                                         |
| **Server-Managed Settings (Beta)**  | Admin Console (client-side)                  | Medium         | No MDM required — works on BYOD devices, centralized web UI. Client-side enforcement only (users with sudo can tamper), not available with third-party providers or custom `ANTHROPIC_BASE_URL`. |
| **Direct AWS Bedrock**              | Cloud IAM only                               | High           | No MDM or gateway needed, VPC-native traffic isolation. No OS-level enforcement — developers can reconfigure locally. Single provider only.                                                      |
| **Direct Google Vertex AI**         | Cloud IAM only                               | High           | No MDM or gateway needed, GCP-native traffic isolation. No OS-level enforcement — developers can reconfigure locally. Single provider only.                                                      |
| **TrueFoundry AI Gateway (no MDM)** | Gateway-level controls                       | High           | Multi-provider routing, fallback, tracing, and guardrails without MDM infrastructure. No OS-level enforcement — developers can bypass the gateway locally.                                       |

***

## 3. Configure Local Tool Access & Sandboxing

Claude Code operates directly in developers' terminals with the same permissions as the local user — reading files, executing shell commands, and connecting to external services.

<Tabs>
  <Tab title="Claude.ai (Web)">
    The Claude.ai web interface runs entirely in the browser — there is no local code execution. The primary risks are **data exfiltration through prompts** and **shadow IT** (employees using personal accounts to bypass corporate controls).

    **Domain capture** is the most effective web control. Once enabled, any sign-in attempt with a corporate email address is routed to the enterprise workspace — employees cannot fall back to personal accounts.

    From the Admin Console you can also configure:

    | Setting                  | Purpose                                                                                                  |
    | ------------------------ | -------------------------------------------------------------------------------------------------------- |
    | File upload restrictions | Disable document uploads if your data classification policies require it                                 |
    | Artifacts                | Disable the in-browser code execution environment if developers should not run untrusted code in-browser |
    | Conversation retention   | Control how long conversation history is stored and whether employees can export transcripts             |
    | Acceptable use notices   | Display policy notices at login to set expectations on what data may be used with Claude                 |

    **Restrict network egress**

    Allow Claude to execute code and create and edit docs, spreadsheets, presentations, PDFs, and data reports. Available on web and desktop. Control Claude network access to install packages and libraries in order to perform advanced data analysis, custom visualizations, and specialized file processing. Monitor chats closely as this comes with security risks.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/b-bXYCa3Uo5Jq0k5/images/enterprise-security-claude-network-egress.png?fit=max&auto=format&n=b-bXYCa3Uo5Jq0k5&q=85&s=91ee3251b7ba9d7939b4d97c33793713" alt="Claude admin settings for code execution with network access enabled and domain allowlist configuration" width="1900" height="844" data-path="images/enterprise-security-claude-network-egress.png" />
    </Frame>

    **Best practices for web:**

    * Enable domain capture and SSO enforcement before broad rollout
    * Communicate clear acceptable-use policies covering what data employees may paste into Claude
    * Classify Claude.ai as a third-party SaaS tool in your DLP tooling — apply the same controls you would to Google Docs or Notion
  </Tab>

  <Tab title="Claude Code CLI">
    Claude Code operates directly in developers' terminals with the same permissions as the local user — reading files, executing shell commands, and connecting to external services. This gives it the broadest attack surface of the three interfaces.

    | Attack Vector      | Risk                                                            |
    | ------------------ | --------------------------------------------------------------- |
    | File system access | Can read `.env` files, SSH keys, and credentials unless blocked |
    | Shell execution    | Runs arbitrary bash commands in the developer's user context    |
    | Data transmission  | Code and context are sent to Anthropic servers for processing   |

    **Recommended baseline configuration:**

    ```json theme={"dark"}
    {
      "permissions": {
        "disableBypassPermissionsMode": "disable",
        "deny": [
          "Bash(curl:*)",
          "Bash(wget:*)",
          "Read(**/.env)",
          "Read(**/.env.*)",
          "Read(**/secrets/**)",
          "Read(**/.ssh/**)",
          "Read(**/credentials/**)"
        ],
        "ask": ["Bash(git push:*)", "Write(**)"]
      },
      "allowManagedPermissionRulesOnly": true,
      "allowManagedHooksOnly": true,
      "transcriptRetentionDays": 14
    }
    ```

    | Setting                           | Purpose                                                                         |
    | --------------------------------- | ------------------------------------------------------------------------------- |
    | `disableBypassPermissionsMode`    | Prevents developers from using `--dangerously-skip-permissions` to bypass rules |
    | `allowManagedPermissionRulesOnly` | Locks out project- and user-level permission rules                              |
    | `allowManagedHooksOnly`           | Prevents rogue hook injection — hooks can run arbitrary code between sessions   |
    | `deny` rules                      | Block operations outright — no developer prompt, no override                    |
    | `ask` rules                       | Require explicit developer approval before execution                            |
    | `allow` rules                     | Auto-approve pre-vetted, safe operations                                        |
    | `transcriptRetentionDays`         | Short retention (7–14 days) limits sensitive data exposure from stored sessions |

    <Warning>
      Never run Claude Code as root. Ensure developers run it under their own user
      accounts with the minimum required filesystem permissions.
    </Warning>
  </Tab>
</Tabs>

### Sandboxing

Claude Code features native sandboxing that enforces filesystem and network isolation at the OS level (Seatbelt on macOS, bubblewrap on Linux). Instead of asking permission for each bash command, sandboxing creates defined boundaries upfront where Claude Code can work more freely with reduced risk. [Claude docs](https://code.claude.com/docs/en/sandboxing)

The recommended posture is **enabling sandbox for all developers** with restrictions on network egress and filesystem writes:

```json theme={"dark"}
{
  "sandbox": {
    "enabled": true,
    "network": {
      "httpProxyPort": 8080,
      "socksProxyPort": 8081
    }
  }
}
```

Key sandbox benefits:

* **Filesystem protection** — cannot modify system-level files in `/bin/`, cannot access `~/.bashrc`, and cannot read files denied in your permission settings
* **Network protection** — cannot contact domains not explicitly allowed, preventing data exfiltration to attacker-controlled servers
* **Prompt injection defense** — even if an attacker manipulates Claude's behavior, the sandbox ensures the system remains secure at the OS level

<Tip>
  Set `"allowUnsandboxedCommands": false` in sandbox settings to disable the
  escape hatch that allows commands to run outside the sandbox. This forces all
  commands to run sandboxed or be explicitly listed in `excludedCommands`.
</Tip>

***

## 4. Configure MCP Server Access

MCP (Model Context Protocol) servers connect Claude to external databases, APIs, SaaS tools, and internal services. Every MCP server a developer connects to expands the attack surface. This section covers how to govern which MCP servers your organization allows — a separate concern from local tool access covered above. Use [Claude enterprise security](/docs/ai-gateway/mcp/mcp-overview) governance boundaries so only trusted MCP endpoints are reachable.

### Risks of Unmanaged MCP Servers

Without centralized governance, MCP servers introduce prompt injection risks (malicious tool descriptions can alter Claude's behavior), credential sprawl across developer machines, and shadow IT from ad-hoc installations of unvetted public servers. There is no audit trail of which tools were called, by whom, or what data was returned — and content retrieved by MCP servers from external sources can be weaponized for indirect data exfiltration.

### Recommended Approach: MCP Gateway with Allowlisting

The recommended posture is to **route all MCP server access through a centralized gateway** and **allowlist only that gateway URL** in Claude's managed settings. This gives you a single control point — regardless of how many MCP servers your organization uses.

[TrueFoundry MCP Gateway](/docs/ai-gateway/mcp/mcp-overview) implements this pattern, purpose-built for enterprise MCP governance:

* **Centralized MCP registry** — register and manage all approved MCP servers in one place. Developers connect to the Gateway instead of managing individual server connections locally.
* **Unified authentication** — developers authenticate once with a TrueFoundry API key or external IdP token (Okta, Azure AD, Auth0). The Gateway handles outbound authentication to each downstream server via API keys, OAuth2, or token passthrough. [Learn more](/docs/ai-gateway/mcp/mcp-gateway-auth-security)
* **Role-based access control** — control which users and teams can access which MCP servers and tools, enforcing least-privilege from the dashboard.
* **Tool-level governance** — selectively disable individual tools within a server, or aggregate tools from multiple servers into a [virtual MCP server](/docs/ai-gateway/mcp/virtual-mcp-server) exposing only an approved subset per team.
* **Guardrails** — apply pre-execution checks, real-time blocking, and post-execution validation on tool calls. Configure user approval workflows for high-risk operations. [Learn more](/docs/ai-gateway/guardrails-overview)
* **Full audit trail** — every tool invocation is traced with user attribution, request/response payloads, and latency data — viewable in the dashboard and exportable via OpenTelemetry. [Learn more](/docs/ai-gateway/analytics-mcp-metrics)

Configure Claude's `managed-settings.json` to allowlist only the gateway URL:

```json theme={"dark"}
{
  "allowedMcpServers": [
    { "serverUrl": "https://truefoundry-mcp-gateway.your-company.com/*" }
  ],
  "strictKnownMarketplaces": []
}
```

<Warning>
  Set `strictKnownMarketplaces` to an empty array (`[]`) to block all
  marketplace-sourced MCP installations unless you have explicitly reviewed and
  approved them.
</Warning>

#### Developer Workflow

<Steps>
  <Step title="Admin registers MCP servers in TrueFoundry">
    Register approved MCP servers in the TrueFoundry Control Plane — configuring outbound authentication, access control policies, and guardrails for each server.
  </Step>

  <Step title="Developer accesses MCP servers from TrueFoundry UI">
    Developers log into the TrueFoundry UI to see available MCP servers and copy the ready-to-use connection URL for each.
  </Step>

  <Step title="Configure Claude to connect through the Gateway">
    Add each MCP server as a remote server in Claude's configuration, pointing to the TrueFoundry Gateway URL:

    ```json theme={"dark"}
    {
      "mcpServers": {
        "truefoundry-github": {
          "type": "url",
          "url": "https://<your-truefoundry-gateway-url>/mcp/v1/github/mcp"
        },
        "truefoundry-sentry": {
          "type": "url",
          "url": "https://<your-truefoundry-gateway-url>/mcp/v1/sentry/mcp"
        }
      }
    }
    ```
  </Step>
</Steps>

#### Automate Developer Setup with MDM

On managed devices, deploy a `managed-mcp.json` file via MDM to pre-seed every machine with approved gateway-backed MCP servers. When deployed, this file takes **exclusive control** — developers cannot add or use MCP servers beyond what is defined here.

```json theme={"dark"}
{
  "mcpServers": {
    "truefoundry-github": {
      "type": "http",
      "url": "https://<your-truefoundry-gateway-url>/mcp/v1/github/mcp"
    },
    "truefoundry-sentry": {
      "type": "http",
      "url": "https://<your-truefoundry-gateway-url>/mcp/v1/sentry/mcp"
    }
  }
}
```

**System-level paths:** macOS: `/Library/Application Support/ClaudeCode/managed-mcp.json` · Linux: `/etc/claude-code/managed-mcp.json`

Combined with `allowedMcpServers` in `managed-settings.json`, this creates a fully locked-down MCP configuration. Access control decisions happen at the gateway, so you only need to update the MDM-deployed file when adding or removing entire server integrations.

### Overall Managed Settings

Here is a recommended `managed-settings.json` that combines local tool governance, sandbox enforcement, MCP gateway allowlisting, and model routing:

```json expandable theme={"dark"}
{
  "permissions": {
    "disableBypassPermissionsMode": "disable",
    "deny": [
      "Bash(curl:*)",
      "Bash(wget:*)",
      "Read(**/.env)",
      "Read(**/.env.*)",
      "Read(**/secrets/**)",
      "Read(**/.ssh/**)",
      "Read(**/credentials/**)"
    ],
    "ask": ["Bash(git push:*)", "Write(**)"]
  },
  "allowManagedPermissionRulesOnly": true,
  "allowManagedHooksOnly": true,
  "transcriptRetentionDays": 14,
  "sandbox": {
    "enabled": true,
    "network": {
      "httpProxyPort": 8080,
      "socksProxyPort": 8081
    }
  },
  "allowedMcpServers": [
    { "serverUrl": "https://mcp-gateway.your-company.com/*" }
  ],
  "strictKnownMarketplaces": [],
  "env": {
    "ANTHROPIC_BASE_URL": "https://your-gateway.internal.corp",
    "HTTPS_PROXY": "https://proxy.your-company.com:8080",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "your-opus-model-equivalent",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "your-sonnet-model-equivalent"
  }
}
```

The primary governance mechanism for Claude Code is `managed-settings.json`. Policies placed at the system-level path apply to **all users on the machine** and cannot be overridden by developers.

**System-level paths:**

* macOS: `/Library/Application Support/ClaudeCode/managed-settings.json`
* Linux: `/etc/claude-code/managed-settings.json`

Claude Code reads settings in the following priority order:

```
System (managed-settings.json)  ← admin-controlled, highest priority
  └── Project (.claude/settings.json)
        └── User (~/.claude/settings.json)  ← developer-controlled, lowest priority
```

When `allowManagedPermissionRulesOnly` is set to `true`, system-level rules are the only rules that apply — project and user settings cannot add new permissions.

<Tabs>
  <Tab title="Apply via MDM (recommended)">
    For Claude Code on managed macOS fleets, push both `managed-settings.json` and `managed-mcp.json` via your MDM (Jamf, Mosyle, Kandji, or any MDM supporting script execution) and lock them against modification.

    ```bash theme={"dark"}
    #!/bin/bash
    # =============================================================================
    # MDM Deployment Script: Claude Code Managed Policy Files
    # Must be run as root (e.g., via MDM such as Jamf, Mosyle, or Kandji).
    #
    # Deploys:
    #   - managed-settings.json  → /Library/Application Support/ClaudeCode/
    #   - managed-mcp.json       → /Library/Application Support/ClaudeCode/
    #
    # Files are owned by root:wheel with 644 permissions so that:
    #   - root can read and write
    #   - all users (including the logged-in developer) can only read
    # =============================================================================

    set -euo pipefail

    # ---------------------------------------------------------------------------
    # Guard: must run as root
    # ---------------------------------------------------------------------------
    if [[ "$(id -u)" -ne 0 ]]; then
      echo "ERROR: This script must be run as root." >&2
      exit 1
    fi

    CURRENT_LOGGEDIN_USER="$(/usr/sbin/scutil <<<"show State:/Users/ConsoleUser" | /usr/bin/awk '/Name :/ && ! /loginwindow/ { print $3 }' | /usr/bin/awk -F '@' '{print $1}')"

    # ---------------------------------------------------------------------------
    # Config
    # ---------------------------------------------------------------------------
    DEST_DIR="/Library/Application Support/ClaudeCode"

    # File at which Users would mount their PAT
    TOKEN_FILE="/Users/${CURRENT_LOGGEDIN_USER}/.truefoundry/tfy-pat"

    # Your Gateway URL
    GATEWAY_URL="https://gateway.truefoundry.ai"

    # Your Control Plane URL
    CONTROL_PLANE_URL="https://<tenantName>.truefoundry.cloud"

    MANAGED_SETTINGS_DEST="${DEST_DIR}/managed-settings.json"
    MANAGED_MCP_DEST="${DEST_DIR}/managed-mcp.json"

    # ---------------------------------------------------------------------------
    # Generate policy content by fetching from API
    # ---------------------------------------------------------------------------
    generate_policy_files() {
      log "Checking for token file..."

      local token=""

      if [[ ! -f "$TOKEN_FILE" ]]; then
        log "WARNING: Token file not found at $TOKEN_FILE. Skipping MCP fetch and using empty token."
      else
        token=$(cat "$TOKEN_FILE")

        if [[ -z "$token" ]]; then
          log "WARNING: Token file is empty. Skipping MCP fetch and using empty token."
        fi
      fi

      # Only fetch MCP servers if token is available
      if [[ -n "$token" ]]; then
        log "Fetching MCP servers from API..."
        local api_response
        api_response=$(curl -s --request GET \
          --url "$CONTROL_PLANE_URL/api/svc/v1/mcp-servers?limit=100" \
          --header "Authorization: Bearer $token")

        if [[ $? -ne 0 ]]; then
          log "WARNING: Failed to fetch MCP servers. Using empty MCP config."
          MANAGED_MCP_CONTENT=$(python3 -c "import json; print(json.dumps({'mcpServers': {}}, indent=2))")
        else
          log "Generating managed-mcp.json..."
          MANAGED_MCP_CONTENT=$(echo "$api_response" | python3 -c "
    import json
    import sys

    try:
        response = json.load(sys.stdin)
        mcp_servers = {}

        for server in response.get('data', []):
            fqn_parts = server.get('fqn', '').split(':')
            server_name = fqn_parts[2] if len(fqn_parts) > 2 else server.get('id')

            proxy_url = server.get('proxyUrl', '')
            url = proxy_url.replace('{{mcpProxyBaseURL}}', '$GATEWAY_URL')

            mcp_servers[server_name] = {
                'type': 'http',
                'url': url
            }

        output = {'mcpServers': mcp_servers}
        print(json.dumps(output, indent=2))
    except Exception as e:
        print(f'ERROR: Failed to parse API response: {e}', file=sys.stderr)
        sys.exit(1)
    ")

          if [[ $? -ne 0 ]]; then
            log "WARNING: Failed to generate managed-mcp.json. Using empty MCP config."
            MANAGED_MCP_CONTENT=$(python3 -c "import json; print(json.dumps({'mcpServers': {}}, indent=2))")
          fi
        fi
      else
        log "Skipping MCP server fetch. Using empty MCP config."
        MANAGED_MCP_CONTENT=$(python3 -c "import json; print(json.dumps({'mcpServers': {}}, indent=2))")
      fi

      log "Generating managed-settings.json..."
      MANAGED_SETTINGS_CONTENT=$(python3 <<EOF
    import json

    settings = {
      "permissions": {
        "disableBypassPermissionsMode": "disable",
        "deny": [
          "Bash(curl:*)",
          "Bash(wget:*)",
          "Read(**/.env)",
          "Read(**/.env.*)",
          "Read(**/secrets/**)",
          "Read(**/.ssh/**)",
          "Read(**/credentials/**)"
        ],
        "ask": ["Bash(git push:*)", "Write(**)"]
      },
      "allowManagedPermissionRulesOnly": true,
      "allowManagedHooksOnly": true,
      "transcriptRetentionDays": 14,
      "sandbox": {
        "enabled": true,
        "network": {
          "httpProxyPort": 8080,
          "socksProxyPort": 8081
        }
      },
      "allowedMcpServers": [
        { "serverUrl": "$GATEWAY_URL/*" }
      ],
      "strictKnownMarketplaces": [],
      "env": {
          "ANTHROPIC_BASE_URL": "$GATEWAY_URL",
          "ANTHROPIC_CUSTOM_HEADERS": "Authorization: Bearer $token",
          "ANTHROPIC_MODEL": "claude-code/claude-sonnet",
          "ANTHROPIC_DEFAULT_OPUS_MODEL": "claude-code/claude-opus",
          "ANTHROPIC_DEFAULT_SONNET_MODEL": "claude-code/claude-sonnet",
          "ANTHROPIC_DEFAULT_HAIKU_MODEL": "claude-code/claude-haiku",
          "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS": "1"
      }
    }

    print(json.dumps(settings, indent=2))
    EOF
    )
    }

    # ---------------------------------------------------------------------------
    # Helpers
    # ---------------------------------------------------------------------------
    log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

    deploy_file() {
      local content="$1"
      local dest="$2"

      log "Deploying $(basename "$dest")"
      echo "$content" > "$dest"

      # root owns the file; group wheel; users can read but not write
      chown root:wheel "$dest"
      chmod 644 "$dest"

      # Lock with macOS immutable flag — requires root to explicitly unset before editing
      chflags schg "$dest"

      log "$(basename "$dest") deployed and locked (schg)."
    }

    # ---------------------------------------------------------------------------
    # Create destination directory
    # ---------------------------------------------------------------------------
    if [[ ! -d "$DEST_DIR" ]]; then
      log "Creating directory: $DEST_DIR"
      mkdir -p "$DEST_DIR"
    fi

    # Ensure the directory is owned by root and not writable by users
    chown root:wheel "$DEST_DIR"
    chmod 755 "$DEST_DIR"

    # ---------------------------------------------------------------------------
    # Generate and deploy policy files
    # ---------------------------------------------------------------------------
    generate_policy_files

    # Unlock first in case files already exist and are flagged immutable
    chflags noschg "$MANAGED_SETTINGS_DEST" 2>/dev/null || true
    chflags noschg "$MANAGED_MCP_DEST" 2>/dev/null || true

    deploy_file "$MANAGED_SETTINGS_CONTENT" "$MANAGED_SETTINGS_DEST"
    deploy_file "$MANAGED_MCP_CONTENT" "$MANAGED_MCP_DEST"

    # ---------------------------------------------------------------------------
    # Verify
    # ---------------------------------------------------------------------------
    log "--- Deployment complete. Verifying permissions ---"
    ls -lO "$MANAGED_SETTINGS_DEST"
    ls -lO "$MANAGED_MCP_DEST"

    log "Done. Claude Code will enforce these policies for all users on this machine."

    ```

    <Note>
      Developers need to configure their Gateway PAT at \~/.truefoundry/tfy-pat using command:

      ```bash theme={"dark"}
      echo <PAT> > ~/.truefoundry/tfy-pat
      ```

      Once the PAT is setup, the MDM can be configured to run every 15 min or as self-service and will update the MCP servers and model proxy for developers automatically.
    </Note>

    <Note>
      `chflags schg` sets the macOS system immutable flag. Even root cannot
      overwrite the file without first running `chflags noschg` — providing an
      explicit friction layer against accidental or malicious policy changes. The
      script handles this automatically on re-runs.
    </Note>

    | Path                                       | Owner        | Mode           | Effect                                                     |
    | ------------------------------------------ | ------------ | -------------- | ---------------------------------------------------------- |
    | `/Library/Application Support/ClaudeCode/` | `root:wheel` | `755`          | Users can read the directory but cannot write to it        |
    | `managed-settings.json`                    | `root:wheel` | `644` + `schg` | Readable by all; writable only after root removes the flag |
    | `managed-mcp.json`                         | `root:wheel` | `644` + `schg` | Exclusive MCP server control; locked against modification  |
  </Tab>

  <Tab title="Apply via Server-Side setting">
    For organizations that want to enforce settings without deploying files to individual machines, Claude's Admin Console provides server-side policy management. Settings configured here are applied to all members of your organization when they authenticate, regardless of which device they use.

    Navigate to the Admin Console → **Settings** → **Policies** to configure:

    * Permitted and denied tools and operations
    * MCP server allowlists and denylists
    * Model access restrictions
    * Feature flags (file uploads, Artifacts, export permissions)

    Server-side policies complement — and take precedence over — any local configuration files. This is the recommended approach for web governance where `managed-settings.json` does not apply.
  </Tab>
</Tabs>

***

## 5. Configure Data Retention

By default, Anthropic may retain prompts and outputs for safety and quality improvement. Audit log exports contain metadata only — not chat content or code. Review your enterprise agreement for the full scope of default retention.

<AccordionGroup>
  <Accordion title="Claude.ai (Web) Retention">
    Set retention to a maximum of 30 days at [Organization Settings > Data and Privacy](https://claude.ai/admin-settings/data-privacy-controls).

    <Frame>
      <img src="https://mintcdn.com/truefoundry/b-bXYCa3Uo5Jq0k5/images/enterprise-security-claude-data-security-retention.png?fit=max&auto=format&n=b-bXYCa3Uo5Jq0k5&q=85&s=bede6c2234974d0bfea291b9ffa5cfaa" width="2920" height="718" data-path="images/enterprise-security-claude-data-security-retention.png" />
    </Frame>
  </Accordion>

  <Accordion title="Claude Code (CLI) Transcript Retention">
    Local session transcripts are stored on the developer's machine. Use `transcriptRetentionDays` in `managed-settings.json` to auto-delete after a set period (recommended 7–14 days):

    ```json theme={"dark"}
    {
      "transcriptRetentionDays": 14
    }
    ```
  </Accordion>

  <Accordion title="Zero Data Retention (ZDR)">
    Enterprise add-on that prevents Anthropic from storing prompts and outputs beyond what is needed to serve the request. Prompts are not retained after the response, session data is not used for training, and it applies across all interfaces.

    **ZDR must be requested through your Claude account team** — it requires a contractual addendum. [Learn more](https://platform.claude.com/docs/en/build-with-claude/zero-data-retention)

    <Warning>
      ZDR must be in place before any Protected Health Information (PHI) or other
      regulated data is processed with Claude. Do not introduce sensitive regulated
      data to any Claude interface until the addendum is signed and confirmed active
      by your account team.
    </Warning>
  </Accordion>
</AccordionGroup>

***

## 6. Configure Audit Logging and Monitoring

### What to Capture

Enterprise audit logging should record every Claude interaction with full user attribution across all interfaces:

| Interface | What to Log                                                                     |
| --------- | ------------------------------------------------------------------------------- |
| Web       | Session start/end, file uploads, conversation metadata                          |
| CLI       | Tool invocations, file access patterns, shell command execution, denied actions |

Export logs to your SIEM on a regular schedule. Retain for a minimum of 90 days to meet SOC 2 requirements; longer for regulated industries.

### OpenTelemetry Settings

Administrators can configure OpenTelemetry settings for all users through the managed settings file. [Claude docs](https://code.claude.com/docs/en/monitoring-usage#administrator-configuration)

### Using TrueFoundry AI Gateway for Observability

When Claude traffic flows through [TrueFoundry AI Gateway](/docs/ai-gateway/intro-to-llm-gateway), you get built-in observability across both LLM and MCP requests — without additional instrumentation.

**Request tracing** — every request through the Gateway is traced with full attribution (user, model, MCP server, tool).

<Frame>
  <img src="https://mintcdn.com/truefoundry/Kp2QhbaDlupVHMTP/images/docs/ai-gateway/enterprise-security-claude-request-tracing.png?fit=max&auto=format&n=Kp2QhbaDlupVHMTP&q=85&s=b136ef35a23a0044d10ef95b16af7d12" alt="Claude request logs dashboard showing API calls, response times, and model usage metrics" width="1024" height="202" data-path="images/docs/ai-gateway/enterprise-security-claude-request-tracing.png" />
</Frame>

**Metrics dashboard** — a unified dashboard provides real-time visibility into model, MCP, and guardrail metrics.

<Frame>
  <img src="https://mintcdn.com/truefoundry/Kp2QhbaDlupVHMTP/images/docs/ai-gateway/enterprise-security-claude-metrics-dashboard.png?fit=max&auto=format&n=Kp2QhbaDlupVHMTP&q=85&s=741cf56c35870c05a1a8028486b35faa" alt="Claude monitoring dashboard displaying request metrics, token usage, failure rates, and performance trends" width="1024" height="544" data-path="images/docs/ai-gateway/enterprise-security-claude-metrics-dashboard.png" />
</Frame>

All traces can be exported to any OTEL-compatible platform (Grafana, Datadog, Splunk) for integration with your existing SIEM. [Learn more](/docs/ai-gateway/export-opentelemetry-data)

***

## 7. Configure Custom Usage

### Usage Limits

Set hard monthly spending limits per user and per team to prevent runaway costs. Admins configure this in the Admin Console under **Usage & Billing** → **Limits**. [Claude docs](https://code.claude.com/docs/en/costs#rate-limit-recommendations)

### Usage Reporting

The Admin Console provides per-user, per-team, and per-model usage breakdowns. Use these to:

* Identify top consumers for cost optimization conversations
* Attribute costs to teams for chargeback/showback reporting
* Detect anomalous usage spikes that may indicate misuse

<Frame>
  <img src="https://mintcdn.com/truefoundry/b-bXYCa3Uo5Jq0k5/images/exterprise-secrurity-claude-team-usage-tracking.png?fit=max&auto=format&n=b-bXYCa3Uo5Jq0k5&q=85&s=3838c722c0e0e8fc8bf4418ae76200b6" alt="Claude organization settings showing usage, spend limits, seat allocation, and billing controls" width="2566" height="1366" data-path="images/exterprise-secrurity-claude-team-usage-tracking.png" />
</Frame>

***

## Frequently Asked Questions

<AccordionGroup>
  <Accordion title="What is Claude enterprise security?">
    Claude enterprise security refers to the set of controls that help organizations manage access, protect data, and monitor usage across Claude deployments. It includes features like role-based access, audit logs, and data governance policies.
  </Accordion>

  <Accordion title="Is it safe to use Claude?">
    Claude can be safe to use in an enterprise, but only when proper controls are in place. On its own, it may have access to sensitive data, external tools, or systems without enough oversight. With TrueFoundry, organizations can secure Claude using SSO for controlled access, restrict network and tool usage, and monitor all activity through centralized logging, ensuring it operates within defined security boundaries.
  </Accordion>

  <Accordion title="Does Claude keep your data confidential?">
    Claude follows strong data privacy practices, but confidentiality ultimately depends on how it is used within an organization. Issues typically occur when employees use personal accounts or share sensitive data without governance. TrueFoundry helps enforce enterprise access, apply data policies, and control how data flows across systems, ensuring sensitive information remains protected and aligned with compliance requirements.
  </Accordion>

  <Accordion title="What are the biggest security risks when using Claude in an enterprise?">
    The biggest risks include uncontrolled access to sensitive data, use of personal accounts, unapproved tool integrations via MCP, and lack of audit logs. Without governance, it becomes difficult to track usage, enforce policies, or prevent data exposure.
  </Accordion>

  <Accordion title="What is the best way to secure Claude across web and CLI?">
    The most effective approach is to implement a centralized architecture: SSO for access control, an AI gateway for routing all traffic, an MCP gateway for tool governance, and audit logging for visibility across all interfaces.
  </Accordion>

  <Accordion title="Is it safe to allow Claude to execute code or access the internet?">
    It can be safe if properly controlled. Enterprise admins should restrict network access, define domain allowlists, and monitor activity closely to reduce the risk of data leaks or unauthorized actions.
  </Accordion>

  <Accordion title="Do you need an AI gateway to secure Claude usage?">
    Securing Claude usage across teams and interfaces becomes challenging without a centralized control layer. Usage can quickly become fragmented, making it difficult to enforce policies, monitor activity, or manage risk consistently. TrueFoundry's AI gateway brings all Claude traffic through a single layer, allowing organizations to enforce security policies, gain full visibility, and manage usage across web and CLI from one place.
  </Accordion>

  <Accordion title="What is an MCP gateway and how does it improve security?">
    An MCP gateway is a centralized layer that manages how Claude connects to external tools and services. Without it, tools can be integrated in an unstructured way, increasing the risk of unauthorized access and data exposure. TrueFoundry provides an MCP gateway that allows organizations to approve, restrict, and monitor all tool interactions, ensuring that Claude only connects to trusted systems while maintaining full visibility and control over data flow.
  </Accordion>
</AccordionGroup>

<script
  type="application/ld+json"
  dangerouslySetInnerHTML={{
__html: JSON.stringify({
  "@context": "https://schema.org",
  "@type": "TechArticle",
  headline: "Claude Enterprise Security: Governance & Access Controls",
  description:
    "A guide to Claude enterprise security for admins to govern access, secure data, and monitor deployments across web and CLI interfaces.",
  mainEntityOfPage: {
    "@type": "WebPage",
    "@id":
      "https://www.truefoundry.com/docs/ai-gateway/mcp/enterprise-security-claude"
  },
  author: {
    "@type": "Organization",
    name: "TrueFoundry"
  },
  publisher: {
    "@type": "Organization",
    name: "TrueFoundry",
    logo: {
      "@type": "ImageObject",
      url: "https://cdn.prod.website-files.com/6291b38507a5238373237679/6291e20016f0c749e47497d5_logo-header.avif"
    }
  },
  dateModified: "2026-03-26",
  keywords: ["Claude enterprise security", "Enterprise security for Claude"]
})
}}
/>

<script
  type="application/ld+json"
  dangerouslySetInnerHTML={{
__html: JSON.stringify({
  "@context": "https://schema.org",
  "@type": "FAQPage",
  mainEntity: [
    {
      "@type": "Question",
      name: "What is Claude enterprise security?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Claude enterprise security refers to the set of controls that help organizations manage access, protect data, and monitor usage across Claude deployments. It includes features like role-based access, audit logs, and data governance policies."
      }
    },
    {
      "@type": "Question",
      name: "Is it safe to use Claude?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Claude can be safe to use in an enterprise, but only when proper controls are in place. On its own, it may have access to sensitive data, external tools, or systems without enough oversight. With TrueFoundry, organizations can secure Claude using SSO for controlled access, restrict network and tool usage, and monitor all activity through centralized logging, ensuring it operates within defined security boundaries."
      }
    },
    {
      "@type": "Question",
      name: "Does Claude keep your data confidential?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Claude follows strong data privacy practices, but confidentiality ultimately depends on how it is used within an organization. Issues typically occur when employees use personal accounts or share sensitive data without governance. TrueFoundry helps enforce enterprise access, apply data policies, and control how data flows across systems, ensuring sensitive information remains protected and aligned with compliance requirements."
      }
    },
    {
      "@type": "Question",
      name: "What are the biggest security risks when using Claude in an enterprise?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "The biggest risks include uncontrolled access to sensitive data, use of personal accounts, unapproved tool integrations via MCP, and lack of audit logs. Without governance, it becomes difficult to track usage, enforce policies, or prevent data exposure."
      }
    },
    {
      "@type": "Question",
      name: "What is the best way to secure Claude across web and CLI?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "The most effective approach is to implement a centralized architecture: SSO for access control, an AI gateway for routing all traffic, an MCP gateway for tool governance, and audit logging for visibility across all interfaces."
      }
    },
    {
      "@type": "Question",
      name: "Is it safe to allow Claude to execute code or access the internet?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "It can be safe if properly controlled. Enterprise admins should restrict network access, define domain allowlists, and monitor activity closely to reduce the risk of data leaks or unauthorized actions."
      }
    },
    {
      "@type": "Question",
      name: "Do you need an AI gateway to secure Claude usage?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Securing Claude usage across teams and interfaces becomes challenging without a centralized control layer. Usage can quickly become fragmented, making it difficult to enforce policies, monitor activity, or manage risk consistently. TrueFoundry's AI gateway brings all Claude traffic through a single layer, allowing organizations to enforce security policies, gain full visibility, and manage usage across web and CLI from one place."
      }
    },
    {
      "@type": "Question",
      name: "What is an MCP gateway and how does it improve security?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "An MCP gateway is a centralized layer that manages how Claude connects to external tools and services. Without it, tools can be integrated in an unstructured way, increasing the risk of unauthorized access and data exposure. TrueFoundry provides an MCP gateway that allows organizations to approve, restrict, and monitor all tool interactions, ensuring that Claude only connects to trusted systems while maintaining full visibility and control over data flow."
      }
    }
  ]
})
}}
/>

<script
  type="application/ld+json"
  dangerouslySetInnerHTML={{
__html: JSON.stringify({
  "@context": "https://schema.org/",
  "@type": "BreadcrumbList",
  itemListElement: [
    {
      "@type": "ListItem",
      position: 1,
      name: "Home",
      item: "https://www.truefoundry.com"
    },
    {
      "@type": "ListItem",
      position: 2,
      name: "Truefoundry Docs",
      item: "https://www.truefoundry.com/docs/create-and-setup-your-account"
    },
    {
      "@type": "ListItem",
      position: 3,
      name: "AI/LLM Gateway Doc",
      item: "https://www.truefoundry.com/docs/ai-gateway/"
    },
    {
      "@type": "ListItem",
      position: 4,
      name: "MCP Gateway Doc",
      item: "https://www.truefoundry.com/docs/ai-gateway/mcp/mcp-overview"
    },
    {
      "@type": "ListItem",
      position: 5,
      name: "Claude Enterprise Security",
      item: "https://www.truefoundry.com/docs/ai-gateway/mcp/enterprise-security-claude"
    }
  ]
})
}}
/>

<script
  type="application/ld+json"
  dangerouslySetInnerHTML={{
__html: JSON.stringify({
  "@context": "https://schema.org",
  "@type": "WebSite",
  name: "TrueFoundry",
  url: "https://www.truefoundry.com/"
})
}}
/>

## 8. Compliance

<AccordionGroup>
  <Accordion title="SOC 2 Type II">
    Anthropic holds SOC 2 Type II certification, available under NDA. As an enterprise admin, your responsibilities include:

    * **User provisioning and deprovisioning** — document procedures and test them regularly
    * **Audit log retention** — retain logs for 90+ days and export them to your SIEM
    * **Vendor risk assessment** — maintain a current assessment of Anthropic's security posture and contractual obligations
  </Accordion>

  <Accordion title="HIPAA">
    <Warning>
      A Zero-Data-Retention (ZDR) addendum is required before processing Protected
      Health Information (PHI) with any Claude interface. Request this through your
      Claude account team — see [Zero Data
      Retention](https://platform.claude.com/docs/en/build-with-claude/zero-data-retention).
      Do not introduce PHI to any interface until the addendum is signed.
    </Warning>

    Additional HIPAA requirements:

    * All outputs involving patient data require mandatory human review before use
    * Maintain a complete audit trail of every PHI interaction across all interfaces
    * Document your AI processing use cases in your HIPAA risk analysis
  </Accordion>

  <Accordion title="GDPR">
    | Requirement        | Implementation                                                                             |
    | ------------------ | ------------------------------------------------------------------------------------------ |
    | Data residency     | Deploy via AWS EU regions or Google Vertex AI with Private Service Connect                 |
    | Right to erasure   | Use the Compliance API and documented deletion workflows; coordinate DSARs with Anthropic  |
    | Data minimization  | Configure `deny` rules (CLI) and upload restrictions (Web) to block PII-containing content |
    | Purpose limitation | Document approved AI processing use cases in your Records of Processing Activities         |
  </Accordion>
</AccordionGroup>

***
