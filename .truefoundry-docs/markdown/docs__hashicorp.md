> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# HashiCorp Vault

> Integrate HashiCorp tools for secrets and infrastructure management.

To integrate with HashiCorp Vault for secret management in TrueFoundry, you need to set up a HashiCorp Vault integration. This guide walks you through adding HashiCorp Vault integration, describing each field in the forms, and providing helpful links for setup.

## HashiCorp Vault Integration

To add a HashiCorp Vault integration:

1. **Go to the Platform > Integration Providers page** in the TrueFoundry UI.
2. **Click on 'Add Integration Provider'.**
3. **Select 'HashiCorp' as the provider type.**
4. **Choose 'HashiCorp Vault Integration'** in the integrations section.
5. **Fill out the form fields as shown below:**

<Frame>
  <img src="https://mintcdn.com/truefoundry/ooNfi4AUJd2sj7M6/images/hashicorp.png?fit=max&auto=format&n=ooNfi4AUJd2sj7M6&q=85&s=fc8bb0b840cd6acc4485ce1ae38dd845" alt="HashiCorp Vault Integration Form" width="1502" height="1338" data-path="images/hashicorp.png" />
</Frame>

### Field Descriptions

* **Name**: The name of the HashiCorp Vault provider account. Must be 3 to 32 lowercase alphanumeric characters, may contain hyphens, and cannot start with a number.
* **Display Name**: The name of the integration that will be displayed in the TrueFoundry UI. Again, it must be 3 to 32 lowercase alphanumeric characters, may contain hyphens, and cannot start with a number.
* **Vault URL**: The URL of your HashiCorp Vault server (e.g., `https://vault.example.com:8200`).
* **Auth Token**: The Vault authentication token used for authentication. This is required for the integration to access your Vault instance. You can either paste the token directly or reference a mounted Kubernetes secret key using the format `tfy-k8s-secret://HASHICORP`. Here, `HASHICORP` is the key inside the Kubernetes secret, not the secret name.

The user whose token you are going to use should have the following permissions for the kv path, lets say the kv path is `tfy-secrets`, then the user should have the following permissions:

```
path "tfy-secrets/data/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "tfy-secrets/metadata/*" {
  capabilities = ["read"]
}
```

You can refer to the [HashiCorp Vault Permissions documentation](https://developer.hashicorp.com/vault/docs/secrets/kv/kv-v2/setup#step-2-create-an-acl-policy-file) for more details.

* **Access Control (Optional)**: List of users or groups fqn in format `<user_type>:<username>` allowed to use this integration. By default, set to `everyone`.

> **How to Get a HashiCorp Vault Auth Token:**
>
> 1. Access your HashiCorp Vault instance through the CLI or UI.
> 2. Authenticate using your preferred method (e.g., `vault auth-method=userpass username=myuser`).
> 3. You can copy the token by clicking on the profile on UI and click on copy token.
> 4. For more details, see the [HashiCorp Vault Authentication documentation](https://developer.hashicorp.com/vault/docs/auth).

## Final Steps

* After filling all required fields, click **Add Integration Provider** to save your integration.
* You can now use this integration as a secret store when configuring your applications to securely manage secrets through HashiCorp Vault.

***

With this integration in place, you can securely manage and access secrets stored in your HashiCorp Vault instance directly from TrueFoundry!
