> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Use Secret Manager in Integrations

> Store API keys and other sensitive credentials in Secret manager and reference them in Model, MCP server, and Guardrail integrations.

Use TrueFoundry secret references or mounted Kubernetes secrets to supply sensitive values in AI Gateway integrations instead of pasting raw credentials. This works for **Model integrations**, **MCP servers**, and **Guardrail integrations**.

<Note>
  For Options A and B below, secrets are synced from your secret manager to
  TrueFoundry every hour by default. This lets secret rotation in your external
  secret manager get picked up automatically.
</Note>

<Steps>
  <Step title="Connect Your Secret Store to TrueFoundry for Options A and B">
    Integrate your secret store with TrueFoundry (for example AWS SSM, GCP
    Secret Manager, HashiCorp Vault, or Azure Key Vault) if you want to use
    Option A or Option B below.

    To connect your secret store, navigate to `Platform > Integrations` and add
    the integration for your provider. We support
    [AWS SSM](https://www.truefoundry.com/docs/integration-provider-aws#ssm-integration),
    [GCP Secrets Manager](https://www.truefoundry.com/docs/integration-provider-gcp#google-secrets-manager-integration),
    [Azure Key Vault](https://www.truefoundry.com/docs/integration-provider-azure#azure-key-vault-integration),
    and [HashiCorp Vault](https://www.truefoundry.com/docs/hashicorp).

    <Frame>
      <img src="https://mintcdn.com/truefoundry/ooNfi4AUJd2sj7M6/images/hashicorp.png?fit=max&auto=format&n=ooNfi4AUJd2sj7M6&q=85&s=fc8bb0b840cd6acc4485ce1ae38dd845" alt="Platform Integrations page showing a configured secret store integration" width="1502" height="1338" data-path="images/hashicorp.png" />
    </Frame>

    <Info>
      TrueFoundry does not store the secret values; they live in your secret
      manager. You get a reference to use in deployments and integrations. See
      [integrate a secret manager](/docs/integrations-secret-store) for setup.
    </Info>
  </Step>

  <Step title="Use the Secret in an Integration">
    When adding or editing an integration, such as a
    [Model provider](/docs/ai-gateway/openai), an
    [MCP server](/docs/ai-gateway/mcp/mcp-overview), or a
    [Guardrail](/docs/ai-gateway/guardrails-overview), use one of the three
    options below in any API key or other sensitive field instead of pasting
    the raw value.

    <img src="https://mintcdn.com/truefoundry/qZ3yGXZg_Nz17sVV/images/docs/secret-in-openai-integration.png?fit=max&auto=format&n=qZ3yGXZg_Nz17sVV&q=85&s=e5288f2d7233a9868cbe0b1bc20faa9a" alt="Using secret FQN in an integration API key field (example: OpenAI)" width="1972" height="1644" data-path="images/docs/secret-in-openai-integration.png" />

    <AccordionGroup>
      <Accordion title="Option A: Create the secret in TrueFoundry, then add the secret FQN">
        Create the secret in a [secret group](/docs/manage-secrets#creating-and-managing-secrets) in TrueFoundry, then use its **fully-qualified name (FQN)** in the integration.

        **Format:** `tfy-secret://<tenant>:<secret-group>:<secret-key>`

        **Example:** For tenant `my-tenant`, secret group `openai-keys`, and key `OPENAI_API_KEY`:

        ```
        tfy-secret://my-tenant:openai-keys:OPENAI_API_KEY
        ```

        <img src="https://mintcdn.com/truefoundry/-g83eZw0cKb4T5XU/images/docs/copy-secret-fqn.png?fit=max&auto=format&n=-g83eZw0cKb4T5XU&q=85&s=06401f815d7ade39a0b993dddaf3ffcf" alt="Copy secret FQN from the secret group in TrueFoundry" width="3024" height="1712" data-path="images/docs/copy-secret-fqn.png" />

        For full steps on creating secret groups and adding keys, see [Secret Management](/docs/manage-secrets). TrueFoundry resolves the FQN at runtime and injects the value.
      </Accordion>

      <Accordion title="Option B: Use any existing secret in your secret manager directly">
        Use a secret that **already exists** in your secret manager. You do not create the secret in TrueFoundry—reference it by your secret store FQN and the path in your store.

        **Format:** `<secret-store-fqn>::<path-to-secret-in-secret-manager>`

        **Example:** If your secret store FQN is `my-tenant:aws:my-aws-provider:secret-store:aws-ssm` and the secret path is `/tfy-secret/openai/api-key`:

        ```
        my-tenant:aws:my-aws-provider:secret-store:aws-ssm::/tfy-secret/openai/api-key
        ```

        You can find the Secret Store FQN on **Platform → Integrations**, as shown below:

        <img src="https://mintcdn.com/truefoundry/RZf3mUTs07au5PFx/images/secret-management-secret-store-fqn.png?fit=max&auto=format&n=RZf3mUTs07au5PFx&q=85&s=9b6b11101b9abe845ebd0e8d8d7e8df0" alt="Secret Store FQN on Platform Integrations page" width="1852" height="378" data-path="images/secret-management-secret-store-fqn.png" />
      </Accordion>

      <Accordion title="Option C: Use a Kubernetes secret for integrations auth data">
        Use a Kubernetes secret when you want the integration auth value to come
        from a secret mounted into the `servicefoundry-server` pod.

        <Info>
          **Self-hosted control plane only.** This option is available only when
          you host your own control plane. It is not available on the TrueFoundry
          managed control plane.
        </Info>

        **How it works:**

        1. Create a Kubernetes secret in the namespace where your TrueFoundry
           control plane is installed.
        2. Set `servicefoundryServer.tfyK8sSecretName` in your Helm values to
           that secret name.
        3. Upgrade your TrueFoundry Helm release.
        4. In the integration field, reference a key from that secret using
           `tfy-k8s-secret://<KEY_NAME>`.

        **Set the Helm value**

        This field is documented in the TrueFoundry chart's
        [`values.yaml`](https://github.com/truefoundry/infra-charts/blob/55553823ec157947b500095d6354a3bfbe408ee6/charts/truefoundry/values.yaml#L674).

        ```yaml theme={"dark"}
        servicefoundryServer:
          tfyK8sSecretName: provider-integration-secrets
        ```

        When this value is set, the secret is mounted into the
        `servicefoundry-server` pod at `/opt/truefoundry/tfy-k8s-secrets`.

        **Create the Kubernetes secret**

        ```yaml secrets.yaml theme={"dark"}
        apiVersion: v1
        kind: Secret
        metadata:
          name: provider-integration-secrets
          namespace: truefoundry
        type: Opaque
        stringData:
          HASHICORP: "replace-with-your-vault-token"
          GITHUB_APP_PRIVATE_KEY: |
            -----BEGIN PRIVATE KEY-----
            replace-with-your-private-key
            -----END PRIVATE KEY-----
        ```

        Replace `truefoundry` with the namespace where your TrueFoundry control
        plane is installed.

        Apply it with:

        ```bash theme={"dark"}
        kubectl apply -f secrets.yaml -n <namespace>
        ```

        Replace `<namespace>` with the namespace where your TrueFoundry control plane is installed.

        **Reference a secret key in the integration**
        **Format:** `tfy-k8s-secret://<KEY_NAME>`
        **Example:**

        ```text theme={"dark"}
        tfy-k8s-secret://HASHICORP
        ```

        The value after `tfy-k8s-secret://` must exactly match the key inside
        the Kubernetes secret.

        For example, if your secret contains:

        ```yaml theme={"dark"}
        stringData:
          HASHICORP: "replace-with-your-vault-token"
        ```

        then in the HashiCorp Vault integration form you can enter:

        * **Vault URL**: `https://vault.example.com`
        * **HashiCorp Vault Auth Token**: `tfy-k8s-secret://HASHICORP`

        <Frame>
          <img src="https://mintcdn.com/truefoundry/ybzQStqAy_Z96DBk/images/tfy_k8s_secret_ref.png?fit=max&auto=format&n=ybzQStqAy_Z96DBk&q=85&s=afd97d28a467042d6f0da46f4886d889" alt="HashiCorp Vault integration form using a Kubernetes secret reference" width="1628" height="1108" data-path="images/tfy_k8s_secret_ref.png" />
        </Frame>
      </Accordion>
    </AccordionGroup>

    <Tip>
      You can use the same secret reference formats in
      [YAML/GitOps](/docs/using-tfy-apply) when defining provider accounts and other
      integrations, so API keys stay out of version control.
    </Tip>
  </Step>
</Steps>
