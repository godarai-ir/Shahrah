package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/chatprovisioning"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// newChatProvisioningTestRepo returns a repository bound to a throwaway schema that holds a minimal
// tenants table and the real migration. It skips unless CHAT_PROVISIONING_TEST_DATABASE_URL is set.
func newChatProvisioningTestRepo(t *testing.T) (*ChatProvisioningRepository, *pgxpool.Pool) {
	t.Helper()

	databaseURL := os.Getenv("CHAT_PROVISIONING_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set CHAT_PROVISIONING_TEST_DATABASE_URL to run the chat provisioning repository tests")
	}

	ctx := context.Background()
	schema := "chatprov_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]

	admin, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)

	_, err = admin.Exec(ctx, "CREATE SCHEMA "+schema)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")

		admin.Close()
	})

	config, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(t, err)

	config.ConnConfig.RuntimeParams["search_path"] = schema + ",public"

	pool, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, fmt.Sprintf(`CREATE TABLE %s.tenants (
		tenant_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
		slug text NOT NULL,
		status text NOT NULL DEFAULT 'active',
		created_at timestamptz NOT NULL DEFAULT now())`, schema))
	require.NoError(t, err)

	migration, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "db", "main", "migrations", "064_tenant_chat_provisioning.sql"))
	require.NoError(t, err)

	up, _, _ := strings.Cut(string(migration), "-- +goose Down")
	up = strings.NewReplacer("-- +goose Up", "", "-- +goose StatementBegin", "", "-- +goose StatementEnd", "", "public.", schema+".").Replace(up)

	_, err = pool.Exec(ctx, up)
	require.NoError(t, err)

	return newChatProvisioningRepository(pool), pool
}

func insertChatTestTenant(t *testing.T, pool *pgxpool.Pool, slug, status string) string {
	t.Helper()

	var id string

	require.NoError(t, pool.QueryRow(context.Background(),
		`INSERT INTO tenants (slug, status) VALUES ($1, $2) RETURNING tenant_id::text`, slug, status).Scan(&id))

	return id
}

func TestChatProvisioningClaimsATenantOnceUntilItsLeaseEnds(t *testing.T) {
	repo, pool := newChatProvisioningTestRepo(t)
	ctx := context.Background()
	id := insertChatTestTenant(t, pool, "acme", "active")

	first, err := repo.ClaimDue(ctx, 10, time.Minute)
	require.NoError(t, err)
	require.Len(t, first, 1)
	require.Equal(t, id, first[0].TenantID)
	require.Equal(t, "acme", first[0].Slug)

	second, err := repo.ClaimDue(ctx, 10, time.Minute)
	require.NoError(t, err)
	require.Empty(t, second, "a leased tenant must not be handed out twice")
}

func TestChatProvisioningStopsOnceReadyAndResumesOnAChange(t *testing.T) {
	repo, pool := newChatProvisioningTestRepo(t)
	ctx := context.Background()
	id := insertChatTestTenant(t, pool, "acme", "active")

	_, err := repo.ClaimDue(ctx, 10, time.Minute)
	require.NoError(t, err)
	require.NoError(t, repo.Complete(ctx, chatprovisioning.Outcome{
		TenantID: id, State: chatprovisioning.StateReady, SyncedSlug: "acme", SyncedStatus: "active",
		OIDCClientID: "gdc_1", CredentialsDelivered: true,
	}))

	idle, err := repo.ClaimDue(ctx, 10, time.Minute)
	require.NoError(t, err)
	require.Empty(t, idle, "a synced tenant needs no work")

	_, err = pool.Exec(ctx, `UPDATE tenants SET status = 'suspended' WHERE tenant_id = $1::uuid`, id)
	require.NoError(t, err)

	suspended, err := repo.ClaimDue(ctx, 10, time.Minute)
	require.NoError(t, err)
	require.Len(t, suspended, 1)
	require.True(t, suspended[0].CredentialsDelivered)
	require.Equal(t, "gdc_1", suspended[0].OIDCClientID)
	require.NoError(t, repo.Complete(ctx, chatprovisioning.Outcome{
		TenantID: id, State: chatprovisioning.StateReady, SyncedSlug: "acme", SyncedStatus: "suspended",
		OIDCClientID: "gdc_1", CredentialsDelivered: true,
	}))

	_, err = pool.Exec(ctx, `UPDATE tenants SET slug = 'acme2', status = 'active' WHERE tenant_id = $1::uuid`, id)
	require.NoError(t, err)

	renamed, err := repo.ClaimDue(ctx, 10, time.Minute)
	require.NoError(t, err)
	require.Len(t, renamed, 1)
	require.Equal(t, "acme2", renamed[0].Slug)
}

func TestChatProvisioningBacksOffAFailedTenantAndCountsAttempts(t *testing.T) {
	repo, pool := newChatProvisioningTestRepo(t)
	ctx := context.Background()
	id := insertChatTestTenant(t, pool, "acme", "active")

	_, err := repo.ClaimDue(ctx, 10, time.Minute)
	require.NoError(t, err)
	require.NoError(t, repo.Complete(ctx, chatprovisioning.Outcome{
		TenantID: id, State: chatprovisioning.StatePending, Err: "chat is down", RetryAfter: time.Hour,
	}))

	waiting, err := repo.ClaimDue(ctx, 10, time.Minute)
	require.NoError(t, err)
	require.Empty(t, waiting, "the retry is an hour away")

	_, err = pool.Exec(ctx, `UPDATE tenant_chat_provisioning SET next_attempt_at = now() - interval '1 second'`)
	require.NoError(t, err)

	retry, err := repo.ClaimDue(ctx, 10, time.Minute)
	require.NoError(t, err)
	require.Len(t, retry, 1)
	require.Equal(t, 1, retry[0].Attempts)

	require.NoError(t, repo.Complete(ctx, chatprovisioning.Outcome{
		TenantID: id, State: chatprovisioning.StateReady, SyncedSlug: "acme", SyncedStatus: "active",
	}))

	var attempts int

	require.NoError(t, pool.QueryRow(ctx, `SELECT attempts FROM tenant_chat_provisioning`).Scan(&attempts))
	require.Zero(t, attempts, "a success resets the attempts")
}

func TestChatProvisioningTakesAnExpiredLeaseOverAndSkipsSkippedTenants(t *testing.T) {
	repo, pool := newChatProvisioningTestRepo(t)
	ctx := context.Background()
	id := insertChatTestTenant(t, pool, "acme", "active")
	skipped := insertChatTestTenant(t, pool, "default", "active")

	_, err := repo.ClaimDue(ctx, 10, time.Minute)
	require.NoError(t, err)
	require.NoError(t, repo.Complete(ctx, chatprovisioning.Outcome{
		TenantID: skipped, State: chatprovisioning.StateSkipped, SyncedSlug: "default", SyncedStatus: "active",
	}))

	_, err = pool.Exec(ctx, `UPDATE tenant_chat_provisioning SET lease_until = now() - interval '1 second' WHERE tenant_id = $1::uuid`, id)
	require.NoError(t, err)

	again, err := repo.ClaimDue(ctx, 10, time.Minute)
	require.NoError(t, err)
	require.Len(t, again, 1, "a crashed replica's lease expires")
	require.Equal(t, id, again[0].TenantID)
}

func TestChatProvisioningNeverHandsOneTenantToTwoReplicas(t *testing.T) {
	repo, pool := newChatProvisioningTestRepo(t)
	ctx := context.Background()

	const tenants = 24

	for i := range tenants {
		insertChatTestTenant(t, pool, fmt.Sprintf("tenant-%d", i), "active")
	}

	var (
		mutex   sync.Mutex
		claimed = map[string]int{}
		group   sync.WaitGroup
	)

	for range 8 {
		group.Add(1)

		go func() {
			defer group.Done()

			for {
				work, err := repo.ClaimDue(ctx, 3, time.Minute)
				if err != nil || len(work) == 0 {
					return
				}

				mutex.Lock()
				for _, item := range work {
					claimed[item.TenantID]++
				}
				mutex.Unlock()
			}
		}()
	}

	group.Wait()

	require.Len(t, claimed, tenants)

	for id, times := range claimed {
		require.Equal(t, 1, times, "tenant %s was claimed %d times", id, times)
	}
}
