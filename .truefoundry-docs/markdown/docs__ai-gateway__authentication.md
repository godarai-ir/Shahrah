> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Authentication

> Authenticate with the TrueFoundry AI Gateway using Personal Access Tokens (PATs) or Virtual Account Tokens (VATs).

To access models through the AI Gateway, you authenticate with a **TrueFoundry API key** — not the original provider keys (OpenAI, Anthropic, etc.). The Gateway manages provider credentials centrally, so developers only need a single TrueFoundry token.

Pass your API key as a Bearer token in the `Authorization` header:

```
Authorization: Bearer your-truefoundry-api-key
```

Or set it as an environment variable for the OpenAI SDK:

```bash theme={"dark"}
export OPENAI_BASE_URL="https://gateway.truefoundry.ai"
export OPENAI_API_KEY="your-truefoundry-api-key"
```

TrueFoundry supports two types of API keys:

| Token type                      | Tied to            | Best for                | Who can create  |
| ------------------------------- | ------------------ | ----------------------- | --------------- |
| **Personal Access Token (PAT)** | A user             | Development and testing | Any user        |
| **Virtual Account Token (VAT)** | A virtual identity | Production applications | Platform admins |

## Personal Access Token (PAT)

A PAT is tied to your user account and inherits all model permissions assigned to you. Use PATs during development to test and iterate quickly.

To create a PAT:

1. Go to the **Access** section in the TrueFoundry platform
2. Open **Personal Access Tokens**
3. Click **New Personal Access Token**

<img src="https://mintcdn.com/truefoundry/iMid4yIOHvzf4Z4V/images/pat-create.png?fit=max&auto=format&n=iMid4yIOHvzf4Z4V&q=85&s=4469739497adc7c64581265d50608912" alt="Personal Access Token creation interface in TrueFoundry dashboard" width="2854" height="1600" data-path="images/pat-create.png" />

```python highlight={4} lines theme={"dark"}
from openai import OpenAI

client = OpenAI(
    api_key="your-personal-access-token",
    base_url="https://gateway.truefoundry.ai"
)

response = client.chat.completions.create(
    model="openai-main/gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

To learn more about managing PATs, see the [User Management](/docs/platform/user-management#personal-access-tokens-pats) documentation.

## Virtual Account Token (VAT)

Virtual accounts are not tied to a user — they represent a virtual identity with explicitly defined model access. Use VATs in production applications so that access is not dependent on any individual user.

<Tip>
  Use separate virtual accounts for different applications or services. This gives you per-application usage tracking and independent access control.

  Virtual accounts can only be created and revoked by a platform admin.
</Tip>

To create a VAT:

1. Go to the **Access** section in the TrueFoundry platform
2. Open **Virtual Accounts**
3. Click **New Virtual Account**
4. Name the account and select which models it can access

<iframe typeofembed="iframe" href="https://app.supademo.com/embed/cmcak8iln0001u70ivpvldgrn?embed_v=2" height="500px" width="100%" src="https://app.supademo.com/embed/cmcak8iln0001u70ivpvldgrn?embed_v=2" style={{ border:"none",display:"flex",margin:"auto" }} />

```python highlight={4} lines theme={"dark"}
from openai import OpenAI

client = OpenAI(
    api_key="your-virtual-account-token",
    base_url="https://gateway.truefoundry.ai"
)

response = client.chat.completions.create(
    model="openai-main/gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

To learn more about virtual accounts, see the [Virtual Account Management](/docs/platform/virtual-account-management) documentation.

## Choosing Between PAT and VAT

| Scenario                      | Recommended token |
| ----------------------------- | ----------------- |
| Local development and testing | PAT               |
| CI/CD pipelines               | VAT               |
| Production services and APIs  | VAT               |
| Shared team applications      | VAT               |
| Quick prototyping             | PAT               |

<Warning>
  Avoid using PATs in production applications. If the user who created the PAT leaves the organization, the token becomes invalid and your application will lose access.
</Warning>

## Model Access Control

Access to models is managed at the provider account level. Users and virtual accounts must be granted access to a provider account before they can call its models. For details on managing permissions, see [Access Control](/docs/ai-gateway/gateway-access-control).
