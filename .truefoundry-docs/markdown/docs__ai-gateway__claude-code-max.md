> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Claude Code Max

> How to connect Claude Code Max with the TrueFoundry AI Gateway.

This guide explains how to integrate [Claude Code](https://docs.anthropic.com/en/docs/claude-code) with an Anthropic Max subscription through the TrueFoundry AI Gateway.

## Integrate Truefoundry with Claude Code Max

Claude Code Max can be integrated with TrueFoundry AI Gateway using a dedicated configuration pattern.

### Prerequisites for Claude Code Max integration

Before integrating Claude Code Max with TrueFoundry, ensure you have:

1. **Claude Max Subscription**: An active Anthropic Claude Max subscription (required for this integration flow)
2. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/gateway/quick-start)

<Note>
  When creating the Anthropic provider account in TrueFoundry for this setup, leave the API key field empty. The actual Anthropic `Authorization` header is sent by Claude Code (from your Max account session) and forwarded by the gateway at request time. See [Anthropic model integration](/docs/ai-gateway/anthropic) for account setup steps.
</Note>

3. **Claude Code Installation**: Install Claude Code following the [official documentation](https://docs.anthropic.com/en/docs/claude-code)
4. Configure models in TrueFoundry (leave the Anthropic account API key empty)

First, get the base URL and model name from your TrueFoundry AI Gateway playground using the unified code snippet:

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

Claude Code can be configured either globally or per project by editing the settings file:

* Global: `~/.claude/settings.json`
* Project-specific: `.claude/settings.json` in your project directory

**Why this is good:**

* Anthropic Max includes everything in Pro, plus significantly higher usage capacity (5x or 20x vs Pro), higher output limits, priority access during high traffic, and early access to advanced features.
* You keep using your Anthropic account/session for Claude Code, while TrueFoundry handles gateway authentication independently via `x-tfy-api-key`.
* You get centralized governance from TrueFoundry (usage visibility, quotas/rate limits, RBAC, logs, and guardrails) without changing your day-to-day Claude Code workflow.
* You keep one consistent gateway endpoint and model abstraction, so teams can evolve routing/fallback/cost controls centrally with minimal client-side config changes.

If you use Claude Code Max, Claude Code reserves the `Authorization` header for Anthropic account auth.\
Use `x-tfy-api-key` in `ANTHROPIC_CUSTOM_HEADERS` so the gateway can authenticate separately.

1. Set `ANTHROPIC_BASE_URL` to your TrueFoundry Gateway URL.
2. Set `ANTHROPIC_CUSTOM_HEADERS` with `x-tfy-api-key`.
3. Start Claude Code and sign in with your Anthropic subscription account when prompted.

Use this `settings.json` variant for Max subscription:

```json lines theme={"dark"}
{
  "env": {
    "ANTHROPIC_BASE_URL": "{GATEWAY_BASE_URL}",
    "ANTHROPIC_CUSTOM_HEADERS": "x-tfy-api-key: your-truefoundry-api-key\nX-TFY-LOGGING-CONFIG: {\"enabled\": true}",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "anthropic/claude-4-opus-20250514",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "anthropic/claude-4-sonnet-20250514",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "anthropic/claude-3-5-haiku-20241022"
  }
}
```

```bash theme={"dark"}
export ANTHROPIC_BASE_URL="https://<your-gateway-base-url>"
export ANTHROPIC_CUSTOM_HEADERS="x-tfy-api-key: <your-truefoundry-api-key>"
export ANTHROPIC_DEFAULT_OPUS_MODEL="anthropic/claude-4-opus-20250514"
export ANTHROPIC_DEFAULT_SONNET_MODEL="anthropic/claude-4-sonnet-20250514"
export ANTHROPIC_DEFAULT_HAIKU_MODEL="anthropic/claude-3-5-haiku-20241022"
claude
```

<Info>
  `x-tfy-api-key` authenticates to the gateway. The `Authorization` header from Claude Code is forwarded to Anthropic only when the provider account has no API key.
</Info>

Replace:

* `your-truefoundry-api-key` with your TrueFoundry API key
* `GATEWAY_BASE_URL` with your TrueFoundry AI Gateway Base URL ([how to find it](/docs/ai-gateway/quick-start#gateway-base-url))
* Update the model names to match your TrueFoundry configured models in `provider-name/model-name` format (supports Anthropic, Bedrock and Vertex AI providers)

<Note>
  For Claude Code CLI and VS Code extension usage patterns, refer to the main [Claude Code integration guide](/docs/ai-gateway/claude-code).
</Note>
