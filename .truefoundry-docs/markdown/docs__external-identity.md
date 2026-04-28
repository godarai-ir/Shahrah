> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Manage External Identity

> Bring your own Auth by setting up external identity for secure authentication and access control.

External Identity allows you to bring your own Auth. It authenticate and authorize external, third-party identity providers (Okta, Azure AD) to securely access a TrueFoundry's APIs (e.g., AI Gateway, MCP Gateway) instead of using TrueFoundry's built-in authentication. Users/Applications authenticate with corporate credentials without TrueFoundry user accounts.

## Overview

External Identity enables you to:

* Authenticate using existing identity provider (Okta, Azure AD)
* Control RBAC access to TrueFoundry resources (e.g., Models, MCP Servers, Agents, etc.)
* No need to create TrueFoundry user accounts

<Info>External Identity requires an identity provider to be configured.</Info>

## When to Use External Identity

Organizations with existing identity providers (Okta, Azure AD) can use External Identity.

Common scenarios:

* **B2B applications**: Your customers authenticate through their own Identity Provider (IdP)
* **No TrueFoundry accounts**: Users who aren’t in TrueFoundry need Model, MCP, Agent access

## How External Identity Works

1. User/application authenticate and get a JWT token from the identity provider.
2. User/application adds the JWT token as Bearer authorization header in TrueFoundry API's
3. TrueFoundry validates the token against the configured identity provider.
4. TrueFoundry extracts claims from the token and maps it to external identity based on the claims configured.
5. The selected external identity is used for RBAC when accessing TrueFoundry resources
6. TrueFoundry uses the `sub` claim to uniquely identify each user and manage their authentication for MCP Servers for personalization.

<Note>
  TrueFoundry never handles user credentials. Your identity provider remains the
  source of truth for identity.
</Note>

## Prerequisites

Before creating an external identity, you must configure Identity Provider.

### Creating an Identity Provider

<Steps>
  <Step title="Navigate to Identity Provider tab">
    Navigate to **Access > External Auth > Identity Provider**.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/3H1wGCsRFedJpwgv/images/access-ext-auth-idp-tab.png?fit=max&auto=format&n=3H1wGCsRFedJpwgv&q=85&s=1ea40b035d8bb52a4e4b5307310ea93d" width="3020" height="990" data-path="images/access-ext-auth-idp-tab.png" />
    </Frame>
  </Step>

  <Step title="Add a new External Identity Provider">
    Click on `New Identity Provider` button on the top right.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/3H1wGCsRFedJpwgv/images/access-ext-auth-idp-new.png?fit=max&auto=format&n=3H1wGCsRFedJpwgv&q=85&s=30b59f7b422203ef8251dd872aaf8864" width="1794" height="384" data-path="images/access-ext-auth-idp-new.png" />
    </Frame>
  </Step>

  <Step title="Fill the form">
    Configure the identity provider with config for JWT token validation. The form contains following fields:

    <ParamField body="Name" type="string" required>
      A unique name to identify this external identity provider. Should follow the
      format as Alphanumeric characters and hyphens.

      <Note>
        The name should be descriptive and help you identify the purpose of this
        external identity.
      </Note>
    </ParamField>

    <ParamField body="Config" type="object" required>
      Contains JWKS URI that exposed the JWKS for fetching the public key to validate the JWT token.

      <Note>
        (Optional) You can also validate the token against the issuer and audience.
      </Note>
    </ParamField>

    <ParamField body="Advanced Options">
      You can configure claim keys for `unique id` and `email`. Defaults are `sub` and `email` respectively.

      <Note>
        `sub` claim is **mandatory** for the token to be validated. You can change the
        claim `key` under `Advanced Settings`.
      </Note>

      <Tip>
        (Optional) You can also configure `email` claim to allow tracking of the
        user's requests in traces.
      </Tip>
    </ParamField>

    <Frame>
      <img src="https://mintcdn.com/truefoundry/3H1wGCsRFedJpwgv/images/access-ext-auth-idp-form-2.png?fit=max&auto=format&n=3H1wGCsRFedJpwgv&q=85&s=2d560d941d3bcd756760fbbe0ad9dad2" width="3018" height="1720" data-path="images/access-ext-auth-idp-form-2.png" />
    </Frame>
  </Step>

  <Step title="Save the configuration">
    Click **Submit** to create the external identity provider. You’ll see it
    listed in the Identities providers tab.
  </Step>
</Steps>

### Examples

#### Okta

Note `<okta-organization>` is the organization name of the Okta.

```yaml theme={"dark"}
config:
  jwks_uri: https://<okta-organization>.okta.com/oauth2/v1/keys
  issuer: https://<okta-organization>.okta.com
```

#### Azure AD

Note `<tenant-id>` and `<client-id>` are the tenant ID and client ID of the Azure AD application.

```yaml theme={"dark"}
config:
  jwks_uri: https://login.microsoftonline.com/<tenant-id>/discovery/v2.0/keys
  issuer: https://login.microsoftonline.com/<tenant-id>/v2.0
  audience: <client-id>
```

## Creating an External Identity

<Steps>
  <Step title="Navigate to External Identities">
    Go to **Access > External Auth > Idenities**.

    <Frame>
      <img src="https://mintcdn.com/truefoundry/XnXVxVsk1dnqxsr8/images/access-ext-auth-tab.png?fit=max&auto=format&n=XnXVxVsk1dnqxsr8&q=85&s=fbba9df4656b71d9a10227fc63a0be1f" alt="Navigate to External Auth > Identities" data-og-width="3022" width="3022" data-og-height="650" height="650" data-path="images/access-ext-auth-tab.png" data-optimize="true" data-opv="3" srcset="https://mintcdn.com/truefoundry/XnXVxVsk1dnqxsr8/images/access-ext-auth-tab.png?w=280&fit=max&auto=format&n=XnXVxVsk1dnqxsr8&q=85&s=40cc0ec49877cfb359010f6bfa4a12f3 280w, https://mintcdn.com/truefoundry/XnXVxVsk1dnqxsr8/images/access-ext-auth-tab.png?w=560&fit=max&auto=format&n=XnXVxVsk1dnqxsr8&q=85&s=dab3cdf2d4a5ce7eb0904d3e40e10f6b 560w, https://mintcdn.com/truefoundry/XnXVxVsk1dnqxsr8/images/access-ext-auth-tab.png?w=840&fit=max&auto=format&n=XnXVxVsk1dnqxsr8&q=85&s=e1b4fd64ff80159ea610e6ff94c38914 840w, https://mintcdn.com/truefoundry/XnXVxVsk1dnqxsr8/images/access-ext-auth-tab.png?w=1100&fit=max&auto=format&n=XnXVxVsk1dnqxsr8&q=85&s=7f338f2fe9db43784d6afad0e76efe9b 1100w, https://mintcdn.com/truefoundry/XnXVxVsk1dnqxsr8/images/access-ext-auth-tab.png?w=1650&fit=max&auto=format&n=XnXVxVsk1dnqxsr8&q=85&s=f45df050487d7af15d03827360fd7078 1650w, https://mintcdn.com/truefoundry/XnXVxVsk1dnqxsr8/images/access-ext-auth-tab.png?w=2500&fit=max&auto=format&n=XnXVxVsk1dnqxsr8&q=85&s=c439fe0031d739617cde1690d2220396 2500w" />
    </Frame>
  </Step>

  <Step title="Add an external identity">
    Click the **Add External Identity** button to open the configuration form.
  </Step>

  <Step title="Fill in the form">
    The form contains the following fields:

    <ParamField body="Name" type="string" required>
      A unique name to identify this external identity. Should follow the format as alphanumeric characters and hyphens. Examples: `api-service`, `external-app`, `partner-integration`

      <Note>
        The name should be descriptive and help you identify the purpose of this
        external identity.
      </Note>
    </ParamField>

    <ParamField body="Identity Provider" type="string" required>
      Select the identity provider from the list of configured identity providers.

      <Info>
        You can create multiple external identities using the same identity provider.
        When Truefoundry gets a token from the external identity, it will validate the
        token against the identity provider and assign it to all the matching external
        identity.
      </Info>
    </ParamField>

    <Frame>
      <img src="https://mintcdn.com/truefoundry/3H1wGCsRFedJpwgv/images/access-ext-auth-id-form.png?fit=max&auto=format&n=3H1wGCsRFedJpwgv&q=85&s=4529cb6e7d9f76f3b9ebf9a51b2474ff" width="3022" height="1722" data-path="images/access-ext-auth-id-form.png" />
    </Frame>
  </Step>

  <Step title="Submit the configuration">
    Click **Submit** to create the external identity. You'll see it listed in the External Identities section.
  </Step>
</Steps>

<Tip>
  **Best Practice**: Create more specific external identities with detailed
  claims for precise access control, and use general external identities (with
  fewer or no claims) as fallbacks.
</Tip>

## Using External Identity as Collaborator

Once created, external identities can be added as collaborators to TrueFoundry resources, granting them access based on the assigned role, similar to how you add users and teams as collaborators.

### Adding External Identity as Collaborator

<Steps>
  <Step title="Navigate to the Resource">
    Navigate to the resource you want to grant access to (e.g., Model, MCP Server,
    Workspace, Cluster, etc.)
  </Step>

  <Step title="Edit the resource">
    Edit the resource you want to grant access to
  </Step>

  <Step title="Add Collaborator">
    Click **Add Collaborator** in the collaborator selection and search for your
    \**External Identity*. Choose the appropriate **Role** (e.g., Viewer, Editor,
    Manager).

    <Frame>
      <img src="https://mintcdn.com/truefoundry/1ouOhp1bbFOFfgP_/images/external_identity_in_collaborators.png?fit=max&auto=format&n=1ouOhp1bbFOFfgP_&q=85&s=ca3cea8e8680cd16281e3e1b5a7a0721" alt="Adding external identity as collaborator showing the selection interface" width="1532" height="438" data-path="images/external_identity_in_collaborators.png" />
    </Frame>
  </Step>

  <Step title="Save the Configuration">
    Click **Submit** to save the collaborators in your resource. Your External
    Identity can now access the selected resource.
  </Step>
</Steps>

## Troubleshooting

### External Identity Not Matching

**Problem**: Token authentication fails or wrong external identity is selected.

**Solutions**:

* verify the JWKS URI is correct and the JWKS is accessible by TrueFoundry.
* verify the token issuer and audience in the token are correct if configured.
* verify the `sub` claim is present in the token and is correct.
* verify the identity provider is correct and matches your identity provider configuration name

***
