> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Access Control

> Control access of models among teams, users and applications

TrueFoundry Gateway allows you to configure access control to models for developers, teams, or applications. This can be configured at the provider account level on the UI.

<iframe typeofembed="iframe" href="https://app.supademo.com/embed/cmcakxm0r0001vi0isudcq99t?embed_v=2" height="500px" width="100%" src="https://app.supademo.com/embed/cmcakxm0r0001vi0isudcq99t?embed_v=2" style={{ border:"none",display:"flex",margin:"auto" }} />

There are two permission levels when granting access to provider accounts:

1. **Provider Account Manager**
   * Can modify provider account settings
   * Can add or remove models
   * Can manage access permissions for others
2. **Provider Account User**
   * Can use all models within the provider account
   * Cannot change provider account settings
   * Cannot modify access permissions

<Info>
  When you assign these permissions to teams or individual users, everyone in that team (or the specific user) will receive the corresponding level of access.
</Info>

<Info>
  **Tenant Admin Access**: Tenant admins automatically have access to all models across the platform without needing to be added as collaborators. This allows tenant admins to manage, configure, and use any model in the AI Gateway.
</Info>

## Authenticating with the Gateway

To access models through the Gateway, you use TrueFoundry API keys — not the original provider keys (OpenAI, Anthropic, etc.). TrueFoundry supports two types of tokens:

* **Personal Access Tokens (PATs)** — tied to a user, recommended for development
* **Virtual Account Tokens (VATs)** — tied to a virtual identity, recommended for production applications

For detailed instructions on creating and using these tokens, see the [Authentication](/docs/ai-gateway/authentication) documentation.
