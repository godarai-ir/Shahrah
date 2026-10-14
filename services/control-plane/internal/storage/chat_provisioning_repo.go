package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/chatprovisioning"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ChatProvisioningRepository persists which tenants are registered with GodarChat.
type ChatProvisioningRepository struct {
	pool *pgxpool.Pool
}

func newChatProvisioningRepository(pool *pgxpool.Pool) *ChatProvisioningRepository {
	return &ChatProvisioningRepository{pool: pool}
}

// claimDueChatTenantsQuery leases tenants that need work. A tenant needs work when it has no row
// yet, or its row is pending, or its slug or status changed since the last sync. The lease check
// sits in the upsert, so two replicas can never both claim the same tenant.
const claimDueChatTenantsQuery = `
WITH due AS (
	SELECT t.tenant_id
	FROM tenants t
	LEFT JOIN tenant_chat_provisioning s ON s.tenant_id = t.tenant_id
	WHERE s.tenant_id IS NULL
	   OR (s.next_attempt_at <= now()
	       AND (s.lease_until IS NULL OR s.lease_until < now())
	       AND (s.state = 'pending'
	            OR s.synced_slug <> t.slug
	            OR s.synced_status <> CASE WHEN t.status = 'active' THEN 'active' ELSE 'suspended' END))
	ORDER BY t.created_at
	LIMIT $1
	FOR UPDATE OF t SKIP LOCKED
), claimed AS (
	INSERT INTO tenant_chat_provisioning AS s (tenant_id, lease_until)
	SELECT tenant_id, now() + make_interval(secs => $2) FROM due
	ON CONFLICT (tenant_id) DO UPDATE
	SET lease_until = EXCLUDED.lease_until, updated_at = now()
	WHERE s.lease_until IS NULL OR s.lease_until < now()
	RETURNING tenant_id, credentials_delivered, oidc_client_id, attempts, synced_slug, synced_status
)
SELECT t.tenant_id::text, t.slug, t.status, c.credentials_delivered, c.oidc_client_id,
       c.attempts, c.synced_slug, c.synced_status
FROM claimed c
JOIN tenants t ON t.tenant_id = c.tenant_id
ORDER BY t.created_at`

// ClaimDue leases up to limit tenants that need chat provisioning work.
func (r *ChatProvisioningRepository) ClaimDue(ctx context.Context, limit int, lease time.Duration) ([]chatprovisioning.Work, error) {
	rows, err := r.pool.Query(ctx, claimDueChatTenantsQuery, limit, lease.Seconds())
	if err != nil {
		return nil, fmt.Errorf("claim chat provisioning work: %w", err)
	}
	defer rows.Close()

	work := make([]chatprovisioning.Work, 0, limit)

	for rows.Next() {
		var item chatprovisioning.Work

		if err := rows.Scan(
			&item.TenantID, &item.Slug, &item.Status, &item.CredentialsDelivered, &item.OIDCClientID,
			&item.Attempts, &item.SyncedSlug, &item.SyncedStatus,
		); err != nil {
			return nil, fmt.Errorf("scan chat provisioning work: %w", err)
		}

		work = append(work, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read chat provisioning work: %w", err)
	}

	return work, nil
}

// Complete records the outcome of one tenant and releases its lease. A failure raises the attempt
// count and schedules the next try; a success resets both.
func (r *ChatProvisioningRepository) Complete(ctx context.Context, outcome chatprovisioning.Outcome) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tenant_chat_provisioning
		SET state = $2,
		    synced_slug = $3,
		    synced_status = $4,
		    oidc_client_id = $5,
		    credentials_delivered = $6,
		    attempts = CASE WHEN $7 = '' THEN 0 ELSE attempts + 1 END,
		    last_error = $7,
		    next_attempt_at = now() + make_interval(secs => $8),
		    lease_until = NULL,
		    ready_at = CASE WHEN $2 = 'ready' THEN COALESCE(ready_at, now()) ELSE ready_at END,
		    updated_at = now()
		WHERE tenant_id = $1::uuid`,
		outcome.TenantID, outcome.State, outcome.SyncedSlug, outcome.SyncedStatus, outcome.OIDCClientID,
		outcome.CredentialsDelivered, truncateChatError(outcome.Err), outcome.RetryAfter.Seconds(),
	)
	if err != nil {
		return fmt.Errorf("record chat provisioning outcome: %w", err)
	}

	return nil
}

// maxChatErrorLength keeps a runaway error text from filling the table.
const maxChatErrorLength = 500

func truncateChatError(message string) string {
	runes := []rune(message)
	if len(runes) <= maxChatErrorLength {
		return message
	}

	return string(runes[:maxChatErrorLength])
}
