# Support Admin Authentication Runbook

## Trust boundaries

- `support-admin` authenticates humans, owns RBAC, and stores opaque sessions.
- `model-discovery` authenticates workloads using hashed service-account credentials and endpoint scopes.
- Browser credentials and sessions must never be forwarded to another service.

## Deployment order

1. Create the separate `support_admin` database and its least-privilege database role.
2. Apply `apps/support-admin/prisma/migrations` only to `support_admin`.
3. Apply main control-plane migrations, including `023_platform_service_accounts.sql`, to `sharah_db`.
4. Apply the latest `apps/model-discovery/prisma/migrations` migration.
5. Provision a support-admin user using `admin:user:create`.
6. Provision the model-discovery service key using `service-account:create`.
7. Provision the control-plane service key using `go run ./services/control-plane/cmd/platform-service-account`.
8. Provision the platform and tenant provider credential encryption keys described below.
9. Run the `Sync Production Kubernetes Secrets` workflow.
10. Deploy the GitOps values that enable `PLATFORM_PROVIDER_CREDENTIAL_MODE=platform_store`.
11. Verify `/api/auth/me` and one read-only admin request on the dedicated support-admin origin before enabling mutations.

Provisioning never accepts a password or generated service secret on the command line. The user password is read from `SUPPORT_ADMIN_INITIAL_PASSWORD`; workload secrets are randomly generated and displayed once.

Create the control-plane workload credential from the repository root:

```powershell
$env:DATABASE_URL = "postgresql://.../sharah_db"
go -C services/control-plane run ./cmd/platform-service-account --profile support-admin
```

The profile owns the complete support-admin scope set in code, so deployment commands cannot silently omit newly required scopes. The command prints `CONTROL_PLANE_SERVICE_TOKEN` once. Rotate by creating a second credential, deploying it, then revoke the old key ID:

```powershell
go -C services/control-plane run ./cmd/platform-service-account --revoke-key-id <key-id>
```

## Platform provider credential encryption

Provider API credentials stored by control-plane use envelope encryption and a keyed fingerprint. Production requires two independent, persistent GitHub Actions secrets:

- `PLATFORM_PROVIDER_CREDENTIAL_KEKS_JSON`: a JSON object mapping KEK versions to base64-encoded 32-byte AES keys.
- `PLATFORM_PROVIDER_CREDENTIAL_FINGERPRINT_KEY`: an independent base64-encoded key of at least 32 bytes.

Generate the initial `prod-v1` keys from PowerShell without writing plaintext secrets to disk or command history:

```powershell
$kek = [byte[]]::new(32)
[Security.Cryptography.RandomNumberGenerator]::Fill($kek)
$fingerprintKey = [byte[]]::new(32)
[Security.Cryptography.RandomNumberGenerator]::Fill($fingerprintKey)

@{ 'prod-v1' = [Convert]::ToBase64String($kek) } |
  ConvertTo-Json -Compress |
  gh secret set PLATFORM_PROVIDER_CREDENTIAL_KEKS_JSON --repo Delpak/Shahrah

[Convert]::ToBase64String($fingerprintKey) |
  gh secret set PLATFORM_PROVIDER_CREDENTIAL_FINGERPRINT_KEY --repo Delpak/Shahrah

[Array]::Clear($kek, 0, $kek.Length)
[Array]::Clear($fingerprintKey, 0, $fingerprintKey.Length)
```

The secret-sync workflow validates that both values exist and writes them to `shahrah-secret`. It reaches the private Kubernetes API through the production bastion, using the PuTTY private key stored in the `PRODUCTION_BASTION_SSH_PRIVATE_KEY` GitHub Actions secret and a host-key fingerprint pinned in the workflow. GitOps injects the provider keys into control-plane and sets the active version to `prod-v1`. Keep old KEK versions in `PLATFORM_PROVIDER_CREDENTIAL_KEKS_JSON` during rotation until every stored DEK has been rewrapped; removing an in-use version makes existing provider credentials undecryptable.

## Tenant provider credential encryption

Tenant-owned provider API credentials use an independent envelope-encryption key set. Production requires these persistent GitHub Actions secrets:

- `TENANT_PROVIDER_CREDENTIAL_ACTIVE_KEK_VERSION`: the KEK version used for new credential writes; start with `prod-v1`.
- `TENANT_PROVIDER_CREDENTIAL_KEKS_JSON`: a JSON object mapping every readable KEK version to a base64-encoded 32-byte AES key.
- `TENANT_PROVIDER_CREDENTIAL_FINGERPRINT_KEY`: an independent base64-encoded key of at least 32 bytes.

Generate the initial values into three separate files outside the repository. The script refuses an existing output directory, never prints secret values, and restricts filesystem access to the current user where the operating system supports it:

```powershell
.\scripts\generate-tenant-provider-credential-secrets.ps1 `
  -OutputDirectory 'D:\secure-transfer\shahrah-tenant-provider-prod-v1'
```

The output directory contains:

- `TENANT_PROVIDER_CREDENTIAL_ACTIVE_KEK_VERSION.txt`
- `TENANT_PROVIDER_CREDENTIAL_KEKS_JSON.json`
- `TENANT_PROVIDER_CREDENTIAL_FINGERPRINT_KEY.txt`

After copying the files into an encrypted secret manager and verifying the backup, register them as GitHub Actions secrets without placing their values on the command line:

```powershell
$secretDirectory = 'D:\secure-transfer\shahrah-tenant-provider-prod-v1'

Get-Content -Raw (Join-Path $secretDirectory 'TENANT_PROVIDER_CREDENTIAL_ACTIVE_KEK_VERSION.txt') |
  gh secret set TENANT_PROVIDER_CREDENTIAL_ACTIVE_KEK_VERSION --repo Delpak/Shahrah

Get-Content -Raw (Join-Path $secretDirectory 'TENANT_PROVIDER_CREDENTIAL_KEKS_JSON.json') |
  gh secret set TENANT_PROVIDER_CREDENTIAL_KEKS_JSON --repo Delpak/Shahrah

Get-Content -Raw (Join-Path $secretDirectory 'TENANT_PROVIDER_CREDENTIAL_FINGERPRINT_KEY.txt') |
  gh secret set TENANT_PROVIDER_CREDENTIAL_FINGERPRINT_KEY --repo Delpak/Shahrah
```

Run the `Sync Production Kubernetes Secrets` workflow after setting the three values. The workflow validates the JSON, ensures the active version is present, verifies each KEK decodes to exactly 32 bytes, verifies the fingerprint key decodes to at least 32 bytes, and writes the values to `shahrah-secret` without printing them.

During rotation, add the new version to `TENANT_PROVIDER_CREDENTIAL_KEKS_JSON` before changing `TENANT_PROVIDER_CREDENTIAL_ACTIVE_KEK_VERSION`. Keep every old KEK version until all credential records using it have been re-encrypted or rewrapped. Keep the fingerprint key stable unless a separate fingerprint migration has been designed.

## Database ownership and tenant data

`support_admin` owns only operator identities/sessions, support workflows, notes, audit data, and denormalized tenant projections. It must not connect directly to tenant databases.

- Live tenant reads and every tenant mutation go through scoped control-plane APIs.
- Cross-tenant search/dashboard data is projected into `support_admin` through an outbox/event/CDC pipeline.
- Projection rows carry the source tenant ID and update/version metadata, but no cross-database foreign key.
- Provider secrets, raw prompts, and unnecessary financial or personal data are not copied into the projection.

For an existing installation where `support_admin_users` was temporarily created in `sharah_db`, export/import those rows into `support_admin` before deployment, verify login, and only then remove the obsolete main-database tables. The repository does not drop them automatically to avoid operator lockout.

## Rotation and incident response

1. Generate a second credential for the existing service account.
2. Update the support-admin secret and restart/roll out all replicas.
3. Confirm successful protected requests and `lastUsedAt` on the new credential.
4. Revoke the old credential with `service-account:revoke`.

To revoke a human session, rerun `admin:user:create` for that user (which resets the password and deletes all sessions), or disable the user by setting `status` to `disabled`. Disabled users fail session validation immediately.

## Roles and service scopes

| Human role | Permissions |
| --- | --- |
| `super_admin` | All support-admin and platform operations |
| `catalog_admin` | Catalog, provider, import, indexing, and display settings |
| `support_agent` | Read-only catalog administration |

Service scopes are independently enforced by model-discovery and control-plane. Model-discovery scopes include `catalog:read`, `catalog:write`, `provider:write`, `import:execute`, `indexing:execute`, and `display-settings:write`. Control-plane scopes include `platform:read`, `platform:catalog:write`, `platform:tenant:write`, and the provider credential scopes `platform:provider-credentials:read-metadata`, `platform:provider-credentials:create-version`, `platform:provider-credentials:test`, `platform:provider-credentials:activate`, and `platform:provider-credentials:disable`.
