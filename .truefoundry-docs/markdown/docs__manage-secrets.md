> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Secret Management

> Store and manage secrets securely using secret groups with role-based access control in TrueFoundry.

Truefoundry makes it easy to store and manage the secrets for your deployed applications. Secrets are used to store sensitive information like API keys, database passwords, which you can then refer to in your code. Truefoundry integrates with different secret managers like AWS SSM, GCP Secret Manager, HashiCorp Vault, Azure Vault, etc and allows you to store the secrets in them.

<Note>
  When you save a secret using Truefoundry's UI, the actual secrets are stored
  in your SecretManager and Truefoundry never stores the secret with itself. You
  will then get a fqn (fully-qualified-name) for your secret which you can then
  use in your deployments.
</Note>

To save secrets in Truefoundry, we need to first [integrate a secret manager](/docs/integrations-secret-store) with Truefoundry.

## Creating and Managing Secrets

In a project, you likely have a set of secrets associated with it. Managing access to each individual secret can become cumbersome. Hence, we have the concept of **secret groups** which lets you organize and manage related secrets for a specific project.

Within a Secret Group, you can easily add, remove, and update secrets. You can configure access control on each secret group and grant users/teams read, write or admin access.

You can check the demo below on how to add a secret group and some secrets in the group.

<iframe provider="app.supademo.com" href="https://app.supademo.com/embed/v0pjkCehZV_HjnvLGLtGc" typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/v0pjkCehZV_HjnvLGLtGc" style={{ border: "none", display: "flex", margin: "auto" }} />

There are 3 roles in Secret groups:

* **Secret Group Admin**: Can create, edit and delete secrets in the group.
* **Secret Group Editor**: Can edit and delete secrets in the group and also see the secret values.
* **Secret Group Viewer**: Can only see the secret keys, but not the values.

<img src="https://mintcdn.com/truefoundry/s4Aj2_qGCrSP-zc8/images/7ee46337-683f51acead6a2ee94753905fc3de1263a8908c3a835b229e0e1e8ab53e8fcb9-Screenshot_2024-10-16_at_6.02.57_PM.png?fit=max&auto=format&n=s4Aj2_qGCrSP-zc8&q=85&s=d952459a993ab2b362c7c00fea657ba1" width="3572" height="582" data-path="images/7ee46337-683f51acead6a2ee94753905fc3de1263a8908c3a835b229e0e1e8ab53e8fcb9-Screenshot_2024-10-16_at_6.02.57_PM.png" />

<Note>
  By default, a tenant admin has access to all secret groups. For tenant members
  and teams you need to assign roles for each secret group.
</Note>

## Using the Secrets

Each secret will have a fully-qualified-name (FQN) next to the secret using which you can refer to and use the secret in the places mentioned below.

<img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/copy-secret-fqn.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=06401f815d7ade39a0b993dddaf3ffcf" width="3024" height="1712" data-path="images/docs/copy-secret-fqn.png" />

The secret FQN can be used for the following usecases:

<Accordion title="Environment variables in deployments">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/secrets-in-env-variables.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=64ed9bbb7d1a8ad613718c75e41a92ff" width="3022" height="1704" data-path="images/docs/secrets-in-env-variables.png" />
</Accordion>

<Accordion title="Volume Mounts in deployments">
  You can mount secrets as a volume in your deployments.

  <iframe provider="app.supademo.com" href="https://app.supademo.com/embed/0--BuOA6US4apaovMtx3l " typeofembed="iframe" height="475px" width="100%" src="https://app.supademo.com/embed/0--BuOA6US4apaovMtx3l " />
</Accordion>

<Accordion title="API keys / sensitive values in Integrations">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/secret-in-dockerhub-integration.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=026a5bac2f840a1d4e896e1593f527bb" width="3024" height="1714" data-path="images/docs/secret-in-dockerhub-integration.png" />
</Accordion>

<Accordion title="API keys for Models in AI Gateway">
  <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/secret-in-openai-integration.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=e5288f2d7233a9868cbe0b1bc20faa9a" width="1972" height="1644" data-path="images/docs/secret-in-openai-integration.png" />
</Accordion>

<Tip>
  You can use the secret FQN in YAMLs which helps enable Gitops for integrations
  and deployments.
</Tip>

<Tip>
  To manage secrets via API (create/update secret groups, add secrets, search), see [Secret Management API](/docs/apply-api-secret-management) in the API Reference.
</Tip>

## Use your Own Secret Manager

TrueFoundry also supports using your own Secret Manager directly. You can use following format to refer to a secret in your Secret Manager:

```
<secret-store-fqn>::<path-to-secret-in-secret-manager>
```

**Example:** If your Secret Store is registered as `example/aws/aws-ssm` and you wish to use a secret at path `/tfy-secret/example/key1`, the referred key would look like `example/aws/aws-ssm::/tfy-secret/example/key1`

You can find the Secret Store FQN from `Platform > Integrations` page, as shown below:

<img src="https://mintcdn.com/truefoundry/RZf3mUTs07au5PFx/images/secret-management-secret-store-fqn.png?fit=max&auto=format&n=RZf3mUTs07au5PFx&q=85&s=9b6b11101b9abe845ebd0e8d8d7e8df0" width="1852" height="378" data-path="images/secret-management-secret-store-fqn.png" />

## Generic Secret Service (use your own secret server)

To use your own secret store with TrueFoundry, add a Generic Secret Store integration. Your server stays in your infrastructure; TrueFoundry talks to it over HTTP. You can use Header Auth (e.g. API key or bearer token). The secret mount path is fixed at **tfy-secret**.

### Add the integration

1. Go to **Platform > Integration Providers** in the TrueFoundry UI.
2. Click **Add Integration Provider**.
3. Select **Custom** as the provider type.
4. Choose **Generic Secret Store** in the integrations section.
5. Fill out the form (see image below).

<Frame>
  <img src="https://mintcdn.com/truefoundry/kM9WBaDI7AQcMGl1/images/generic-secret-integration-form.png?fit=max&auto=format&n=kM9WBaDI7AQcMGl1&q=85&s=6c0f42ca15d5c3677e9915b73a881349" alt="Generic Secret Store Integration Form" width="1024" height="918" data-path="images/generic-secret-integration-form.png" />
</Frame>

**Fields:** Name, Display Name, Base URL (required). Optional: Custom Basic Auth, Header Auth, Headers, Access Control. The mount path is fixed at **tfy-secret**.

<Info>
  The root path for secrets is always **tfy-secret**.
</Info>

### API contract

Your server must expose these endpoints (relative to the Base URL). **Header Auth** headers are sent on every request.

| Method   | Path                                      | Request                                                                                                 | Response                                         |
| -------- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------- | ------------------------------------------------ |
| `PUT`    | `/secrets`                                | **Body (JSON):** `path` (string, required), `value` (string, required). Content-Type: application/json. | 2xx. Body must include `version`.                |
| `GET`    | `/secrets?path=<key>`                     | **Query:** `path` (required).                                                                           | 200 with `{ "value": "..." }`. 404 if not found. |
| `GET`    | `/secrets?path=<key>&version=<versionId>` | **Query:** `path` (required), `version` (required).                                                     | 200 with `{ "value": "..." }`. 404 if not found. |
| `DELETE` | `/secrets`                                | **Body (JSON):** `path` (required). Optionally `version` to delete one version.                         | 2xx. 400 if path/version invalid.                |

**Validation:** On save, TrueFoundry calls **GET /secrets?path=tfy-secret**. Your server must respond successfully (e.g. 200) for validation to pass. After saving, use this integration as a secret store when configuring applications.

***
