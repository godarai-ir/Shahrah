# Chat tenant provisioning

When a tenant exists, its GodarChat (LibreChat) workspace at `https://<slug>.chat.<domain>` must work
without anyone touching LibreChat. The control-plane does this with a reconciler: it creates the
tenant's OIDC client and registers the tenant in LibreChat through its internal provisioning API.

LibreChat side: `docs/multi-tenancy-host-mode.md` section 15 in the LibreChat repository.

## How it works

A worker (`internal/chatprovisioning`) runs every `CHAT_PROVISIONING_INTERVAL`. It claims tenants that
need work, so a tenant created by any path (onboarding, social sign-in, an operator) is picked up
without a hook in that path. For each tenant it:

1. Finds or creates the OIDC client named `GodarChat` (credit product `librechat`) whose redirect URIs are
   `https://<slug>.chat.<domain>/oauth/openid/callback` (chat sign-in) and
   `https://<slug>.chat.<domain>/api/admin/oauth/openid/callback` (admin panel sign-in). If the slug
   changed, or a client predates the admin callback, it updates the URIs; the secret is not resent.
2. Obtains the client secret. The secret is shown once, when a client is created or rotated, and is
   never stored here: a client the chat has not received credentials for gets a fresh secret.
3. Calls `PUT /tenants/<tenant_id>` on LibreChat with the slug, client id and secret. The tenant id is
   the tenant's UUID, which is the `tid` claim of its tokens.
4. Mirrors the tenant status (`active`, or `suspended` for anything else) with `PATCH`.
5. Marks the tenant `ready` only when LibreChat reports that sign-in works (`ready: true`).

Every step is idempotent, so any failure is simply retried. Retries back off from 15 seconds up to one
hour. Progress is stored in `tenant_chat_provisioning` (migration 064): state (`pending`, `ready`,
`skipped`), what was last synced, whether the chat holds the secret, the attempt count and the last
error. A lease on each claimed tenant keeps several replicas from working on the same tenant.

Tenants whose slug cannot be a chat host (reserved slugs such as `default`) are `skipped`.

## Admin panel sign-in

Each tenant's LibreChat admin panel (`https://<slug>.admin.<domain>`, LibreChat's
`ADMIN_HOST_BASE_DOMAIN`) signs in through the same `GodarChat` client: LibreChat starts the flow on the
chat origin and receives the code at `/api/admin/oauth/openid/callback`, then hands the panel a one-time
exchange code. LibreChat docs: `docs/multi-tenancy-host-mode.md` section 17.

What the panel needs from Godar is already issued to this client:

- `tid` (tenant UUID) and `tenant` (slug) in every ID token and userinfo, which LibreChat checks against
  the host's tenant.
- `roles: [<tenant role>]` with the `profile` scope. Tenant admins carry `tenant_admin`.
- A refresh token, which LibreChat uses to renew admin sessions. LibreChat's default tenant scopes are
  `openid profile email offline_access`.

On the LibreChat side, map the role to admin access:

```env
OPENID_ADMIN_ROLE=tenant_admin
OPENID_ADMIN_ROLE_PARAMETER_PATH=roles
OPENID_ADMIN_ROLE_TOKEN_KIND=id
```

Admin sign-in never creates users, so a tenant admin signs in to the chat once before using the panel.

## Settings

| Variable                             | Meaning                                                                        |
| ------------------------------------ | ------------------------------------------------------------------------------ |
| `CHAT_PROVISIONING_ENABLED`          | `true` turns the worker on. Requires `OIDC_ENABLED=true`. Default `false`.     |
| `CHAT_PROVISIONING_URL`              | Private base URL of LibreChat's provisioning API, for example `http://librechat:3081`. |
| `CHAT_PROVISIONING_TOKEN`            | Bearer token, at least 32 characters; LibreChat's `TENANT_PROVISIONING_TOKEN`. |
| `CHAT_URL_TEMPLATE`                  | Chat origin with `{slug}`, for example `https://{slug}.chat.godarai.ir`.       |
| `CHAT_PROVISIONING_INTERVAL`         | Pass interval. Default `15s`.                                                  |
| `CHAT_PROVISIONING_REQUEST_TIMEOUT`  | Timeout per API call. Default `15s`.                                           |
| `CHAT_PROVISIONING_BATCH_SIZE`       | Tenants per pass. Default `10`.                                                |

Invalid settings stop the control-plane at startup. The provisioning port must stay private: expose it
only to the control-plane (NetworkPolicy), never through an ingress.

## Turning it on for existing tenants

The first pass handles every existing tenant, so enabling it creates a `GodarChat` OIDC client for all
of them. To limit that, enable it in a staging environment first, or insert `skipped` rows for tenants
that must not get a chat:

```sql
INSERT INTO tenant_chat_provisioning (tenant_id, state, synced_slug, synced_status)
SELECT tenant_id, 'skipped', slug, CASE WHEN status = 'active' THEN 'active' ELSE 'suspended' END
FROM tenants WHERE slug IN ('tenant-to-leave-out');
```

## Operating it

```sql
-- What is waiting, and why
SELECT t.slug, p.state, p.attempts, p.last_error, p.next_attempt_at
FROM tenant_chat_provisioning p JOIN tenants t USING (tenant_id)
WHERE p.state <> 'ready' ORDER BY p.next_attempt_at;

-- Retry a tenant now
UPDATE tenant_chat_provisioning SET next_attempt_at = now(), attempts = 0 WHERE tenant_id = '<uuid>';

-- Send the chat a new secret (for example after LibreChat lost its database)
UPDATE tenant_chat_provisioning SET credentials_delivered = false, state = 'pending', next_attempt_at = now()
WHERE tenant_id = '<uuid>';
```

A `400` answer from LibreChat with `INVALID_OIDC` while no secret was sent means LibreChat no longer
knows the tenant; the worker then marks the credentials undelivered and the next pass rotates and
sends a new secret by itself.

## Limits

- A rotated or lost secret is replaced, never recovered: nothing sensitive is stored in this database.
- Deleting a tenant removes its row and its OIDC clients (cascade) but does not call LibreChat; a
  deleted tenant stays registered there until it is removed or suspended by hand.
- Changes to a tenant's name or plan are not mirrored; only its slug and active/suspended status are.
