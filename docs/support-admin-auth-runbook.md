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
8. Provision the platform provider credential encryption keys described below.
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

The secret-sync workflow validates that both values exist and writes them to `shahrah-secret`. GitOps injects them into control-plane and sets the active version to `prod-v1`. Keep old KEK versions in `PLATFORM_PROVIDER_CREDENTIAL_KEKS_JSON` during rotation until every stored DEK has been rewrapped; removing an in-use version makes existing provider credentials undecryptable.

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
