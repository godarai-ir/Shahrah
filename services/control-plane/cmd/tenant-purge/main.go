package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type config struct {
	tenant             string
	confirm            string
	execute            bool
	allowActive        bool
	skipClickHouse     bool
	skipSupportDB      bool
	databaseURL        string
	clickHouseDSN      string
	supportDatabaseURL string
	timeout            time.Duration
}

type tenantRecord struct {
	ID     string
	Slug   string
	Domain string
	Status string
}

type tableCount struct {
	Name  string
	Count int64
}

func main() {
	cfg := parseFlags()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()

	if err := run(ctx, cfg); err != nil {
		fatal("%v", err)
	}
}

func parseFlags() config {
	var cfg config

	flag.StringVar(&cfg.tenant, "tenant", "", "exact tenant UUID, slug, or domain")
	flag.StringVar(&cfg.confirm, "confirm", "", "tenant slug; required together with --execute")
	flag.BoolVar(&cfg.execute, "execute", false, "permanently delete the tenant data (default is dry-run)")
	flag.BoolVar(&cfg.allowActive, "allow-active", false, "allow deleting an active tenant; stop all tenant writers first")
	flag.BoolVar(&cfg.skipClickHouse, "skip-clickhouse", false, "intentionally leave ClickHouse tenant data untouched")
	flag.BoolVar(&cfg.skipSupportDB, "skip-support-db", false, "intentionally leave the support-admin projection untouched")
	flag.DurationVar(&cfg.timeout, "timeout", 30*time.Minute, "overall operation timeout")
	flag.Parse()

	cfg.tenant = strings.TrimSpace(cfg.tenant)
	cfg.confirm = strings.TrimSpace(cfg.confirm)
	cfg.databaseURL = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	cfg.clickHouseDSN = strings.TrimSpace(os.Getenv("CLICKHOUSE_DSN"))
	cfg.supportDatabaseURL = strings.TrimSpace(os.Getenv("SUPPORT_ADMIN_DATABASE_URL"))

	return cfg
}

func run(ctx context.Context, cfg config) error {
	if cfg.tenant == "" {
		return errors.New("--tenant is required")
	}
	if cfg.databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if cfg.timeout <= 0 {
		return errors.New("--timeout must be greater than zero")
	}

	mainDB, err := pgxpool.New(ctx, cfg.databaseURL)
	if err != nil {
		return fmt.Errorf("connect to main PostgreSQL database: %w", err)
	}
	defer mainDB.Close()

	if err := mainDB.Ping(ctx); err != nil {
		return fmt.Errorf("ping main PostgreSQL database: %w", err)
	}

	tenant, err := resolveTenant(ctx, mainDB, cfg.tenant)
	if err != nil {
		return err
	}

	fmt.Printf("Tenant: slug=%q domain=%q id=%s status=%q\n", tenant.Slug, tenant.Domain, tenant.ID, tenant.Status)

	mainCounts, err := inventoryPostgres(ctx, mainDB, tenant.ID)
	if err != nil {
		return fmt.Errorf("inventory main PostgreSQL database: %w", err)
	}
	printInventory("Main PostgreSQL", mainCounts)

	var clickHouseDB *sql.DB
	if !cfg.skipClickHouse {
		if cfg.clickHouseDSN == "" {
			if cfg.execute {
				return errors.New("CLICKHOUSE_DSN is required for a complete purge; set it or explicitly pass --skip-clickhouse")
			}
			fmt.Println("ClickHouse: skipped in dry-run because CLICKHOUSE_DSN is not set")
		} else {
			clickHouseDB, err = sql.Open("clickhouse", cfg.clickHouseDSN)
			if err != nil {
				return fmt.Errorf("open ClickHouse database: %w", err)
			}
			defer clickHouseDB.Close()

			if err := clickHouseDB.PingContext(ctx); err != nil {
				return fmt.Errorf("ping ClickHouse database: %w", err)
			}

			clickHouseCounts, inventoryErr := inventoryClickHouse(ctx, clickHouseDB, tenant.ID)
			if inventoryErr != nil {
				return fmt.Errorf("inventory ClickHouse database: %w", inventoryErr)
			}
			printInventory("ClickHouse", clickHouseCounts)
		}
	} else {
		fmt.Println("ClickHouse: intentionally skipped")
	}

	var supportDB *pgxpool.Pool
	if !cfg.skipSupportDB {
		if cfg.supportDatabaseURL == "" {
			if cfg.execute {
				return errors.New("SUPPORT_ADMIN_DATABASE_URL is required for a complete purge; set it or explicitly pass --skip-support-db")
			}
			fmt.Println("Support-admin PostgreSQL: skipped in dry-run because SUPPORT_ADMIN_DATABASE_URL is not set")
		} else {
			supportDB, err = pgxpool.New(ctx, cfg.supportDatabaseURL)
			if err != nil {
				return fmt.Errorf("connect to support-admin PostgreSQL database: %w", err)
			}
			defer supportDB.Close()

			if err := supportDB.Ping(ctx); err != nil {
				return fmt.Errorf("ping support-admin PostgreSQL database: %w", err)
			}

			count, inventoryErr := supportProjectionCount(ctx, supportDB, tenant.ID)
			if inventoryErr != nil {
				return fmt.Errorf("inventory support-admin PostgreSQL database: %w", inventoryErr)
			}
			printInventory("Support-admin PostgreSQL", []tableCount{{Name: "support_tenant_projections", Count: count}})
		}
	} else {
		fmt.Println("Support-admin PostgreSQL: intentionally skipped")
	}

	if !cfg.execute {
		fmt.Printf("\nDry-run only. To permanently delete this tenant, rerun with --execute --confirm %s\n", tenant.Slug)
		return nil
	}

	if cfg.confirm != tenant.Slug {
		return fmt.Errorf("--confirm must exactly equal the resolved tenant slug %q", tenant.Slug)
	}
	if strings.EqualFold(tenant.Status, "active") && !cfg.allowActive {
		return errors.New("tenant is active; suspend it first, or stop all tenant writers and explicitly pass --allow-active")
	}

	fmt.Println("\nPermanent purge started.")

	// Delete auxiliary copies first. If the final transactional main-database
	// purge fails, the tenant still exists and the command can be rerun safely.
	if clickHouseDB != nil {
		if err := purgeClickHouse(ctx, clickHouseDB, tenant.ID); err != nil {
			return fmt.Errorf("purge ClickHouse database: %w", err)
		}
		fmt.Println("ClickHouse tenant rows deleted and mutations synchronized.")
	}

	if supportDB != nil {
		deleted, purgeErr := purgeSupportProjection(ctx, supportDB, tenant.ID)
		if purgeErr != nil {
			return fmt.Errorf("purge support-admin PostgreSQL database: %w", purgeErr)
		}
		fmt.Printf("Support-admin projection rows deleted: %d\n", deleted)
	}

	deletedCounts, err := purgeMainPostgres(ctx, mainDB, tenant)
	if err != nil {
		return fmt.Errorf("purge main PostgreSQL database: %w", err)
	}
	printInventory("Deleted from main PostgreSQL", deletedCounts)

	fmt.Printf("\nTenant %q (%s) was permanently deleted.\n", tenant.Slug, tenant.ID)
	if cfg.skipClickHouse || cfg.skipSupportDB {
		fmt.Println("Warning: one or more auxiliary databases were explicitly skipped; the purge is not complete across all Shahrah databases.")
	}

	return nil
}

func resolveTenant(ctx context.Context, db *pgxpool.Pool, identity string) (tenantRecord, error) {
	rows, err := db.Query(ctx, `
		SELECT tenant_id::text, slug, domain, status
		FROM public.tenants
		WHERE tenant_id::text = $1 OR slug = $1 OR domain = $1
		ORDER BY tenant_id
	`, identity)
	if err != nil {
		return tenantRecord{}, fmt.Errorf("look up tenant: %w", err)
	}
	defer rows.Close()

	matches := make([]tenantRecord, 0, 2)
	for rows.Next() {
		var item tenantRecord
		if err := rows.Scan(&item.ID, &item.Slug, &item.Domain, &item.Status); err != nil {
			return tenantRecord{}, fmt.Errorf("scan tenant: %w", err)
		}
		matches = append(matches, item)
	}
	if err := rows.Err(); err != nil {
		return tenantRecord{}, fmt.Errorf("read tenant matches: %w", err)
	}

	switch len(matches) {
	case 0:
		return tenantRecord{}, fmt.Errorf("no tenant exactly matches UUID, slug, or domain %q", identity)
	case 1:
		return matches[0], nil
	default:
		return tenantRecord{}, fmt.Errorf("identity %q matches more than one tenant; use the tenant UUID instead", identity)
	}
}

func postgresTenantTables(ctx context.Context, q interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}) ([]string, error) {
	rows, err := q.Query(ctx, `
		SELECT c.table_name
		FROM information_schema.columns c
		JOIN information_schema.tables t
		  ON t.table_schema = c.table_schema
		 AND t.table_name = c.table_name
		WHERE c.table_schema = 'public'
		  AND c.column_name = 'tenant_id'
		  AND t.table_type = 'BASE TABLE'
		  AND c.table_name <> 'tenants'
		ORDER BY c.table_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

func inventoryPostgres(ctx context.Context, db *pgxpool.Pool, tenantID string) ([]tableCount, error) {
	tables, err := postgresTenantTables(ctx, db)
	if err != nil {
		return nil, err
	}

	counts := make([]tableCount, 0, len(tables)+1)
	for _, table := range tables {
		var count int64
		query := "SELECT count(*) FROM " + pgx.Identifier{"public", table}.Sanitize() + " WHERE tenant_id = $1"
		if err := db.QueryRow(ctx, query, tenantID).Scan(&count); err != nil {
			return nil, fmt.Errorf("count public.%s: %w", table, err)
		}
		if count > 0 {
			counts = append(counts, tableCount{Name: table, Count: count})
		}
	}
	counts = append(counts, tableCount{Name: "tenants", Count: 1})

	return counts, nil
}

func purgeMainPostgres(ctx context.Context, db *pgxpool.Pool, tenant tenantRecord) ([]tableCount, error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var lockedID, slug string
	if err := tx.QueryRow(ctx, `
		SELECT tenant_id::text, slug
		FROM public.tenants
		WHERE tenant_id = $1
		FOR UPDATE
	`, tenant.ID).Scan(&lockedID, &slug); err != nil {
		return nil, fmt.Errorf("lock tenant: %w", err)
	}
	if lockedID != tenant.ID || slug != tenant.Slug {
		return nil, errors.New("tenant identity changed after preflight; aborting")
	}

	tables, err := postgresTenantTables(ctx, tx)
	if err != nil {
		return nil, err
	}

	remaining := append([]string(nil), tables...)
	deleted := make([]tableCount, 0, len(tables)+1)
	attempt := 0

	for len(remaining) > 0 {
		progress := false
		next := make([]string, 0, len(remaining))
		blockers := make(map[string]string)

		for _, table := range remaining {
			attempt++
			savepoint := fmt.Sprintf("tenant_purge_%d", attempt)
			if _, err := tx.Exec(ctx, "SAVEPOINT "+savepoint); err != nil {
				return nil, err
			}

			query := "DELETE FROM " + pgx.Identifier{"public", table}.Sanitize() + " WHERE tenant_id = $1"
			tag, deleteErr := tx.Exec(ctx, query, tenant.ID)
			if deleteErr == nil {
				if _, err := tx.Exec(ctx, "RELEASE SAVEPOINT "+savepoint); err != nil {
					return nil, err
				}
				if tag.RowsAffected() > 0 {
					deleted = append(deleted, tableCount{Name: table, Count: tag.RowsAffected()})
				}
				progress = true
				continue
			}

			if _, err := tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+savepoint); err != nil {
				return nil, err
			}
			if _, err := tx.Exec(ctx, "RELEASE SAVEPOINT "+savepoint); err != nil {
				return nil, err
			}

			var pgErr *pgconn.PgError
			if !errors.As(deleteErr, &pgErr) || pgErr.Code != "23503" {
				return nil, fmt.Errorf("delete public.%s: %w", table, deleteErr)
			}

			next = append(next, table)
			blockers[table] = pgErr.ConstraintName
		}

		if !progress {
			return nil, fmt.Errorf("foreign-key cycle or non-tenant child rows block deletion: %s", formatBlockers(blockers))
		}

		remaining = next
	}

	tag, err := tx.Exec(ctx, "DELETE FROM public.tenants WHERE tenant_id = $1", tenant.ID)
	if err != nil {
		return nil, fmt.Errorf("delete public.tenants: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return nil, fmt.Errorf("expected to delete one tenant row, deleted %d", tag.RowsAffected())
	}
	deleted = append(deleted, tableCount{Name: "tenants", Count: 1})

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	sort.Slice(deleted, func(i, j int) bool {
		return deleted[i].Name < deleted[j].Name
	})

	return deleted, nil
}

func clickHouseTenantTables(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT c.table
		FROM system.columns c
		INNER JOIN system.tables t
		  ON t.database = c.database
		 AND t.name = c.table
		WHERE c.database = currentDatabase()
		  AND c.name = 'tenant_id'
		  AND t.engine NOT IN ('View', 'MaterializedView', 'LiveView', 'WindowView')
		ORDER BY c.table
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

func inventoryClickHouse(ctx context.Context, db *sql.DB, tenantID string) ([]tableCount, error) {
	tables, err := clickHouseTenantTables(ctx, db)
	if err != nil {
		return nil, err
	}

	counts := make([]tableCount, 0, len(tables))
	for _, table := range tables {
		var count uint64
		query := "SELECT count() FROM " + quoteClickHouseIdentifier(table) + " WHERE tenant_id = ?"
		if err := db.QueryRowContext(ctx, query, tenantID).Scan(&count); err != nil {
			return nil, fmt.Errorf("count %s: %w", table, err)
		}
		if count > 0 {
			counts = append(counts, tableCount{Name: table, Count: int64(count)})
		}
	}

	return counts, nil
}

func purgeClickHouse(ctx context.Context, db *sql.DB, tenantID string) error {
	tables, err := clickHouseTenantTables(ctx, db)
	if err != nil {
		return err
	}

	for _, table := range tables {
		query := "ALTER TABLE " + quoteClickHouseIdentifier(table) +
			" DELETE WHERE tenant_id = ? SETTINGS mutations_sync = 2"
		if _, err := db.ExecContext(ctx, query, tenantID); err != nil {
			return fmt.Errorf("delete from %s: %w", table, err)
		}

		var remaining uint64
		verifyQuery := "SELECT count() FROM " + quoteClickHouseIdentifier(table) + " WHERE tenant_id = ?"
		if err := db.QueryRowContext(ctx, verifyQuery, tenantID).Scan(&remaining); err != nil {
			return fmt.Errorf("verify %s: %w", table, err)
		}
		if remaining != 0 {
			return fmt.Errorf("verify %s: %d tenant rows remain", table, remaining)
		}
	}

	return nil
}

func supportProjectionCount(ctx context.Context, db *pgxpool.Pool, tenantID string) (int64, error) {
	var exists bool
	if err := db.QueryRow(ctx, `SELECT to_regclass('public.support_tenant_projections') IS NOT NULL`).Scan(&exists); err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	var count int64
	err := db.QueryRow(ctx, `
		SELECT count(*)
		FROM public.support_tenant_projections
		WHERE "tenantId" = $1
	`, tenantID).Scan(&count)

	return count, err
}

func purgeSupportProjection(ctx context.Context, db *pgxpool.Pool, tenantID string) (int64, error) {
	var exists bool
	if err := db.QueryRow(ctx, `SELECT to_regclass('public.support_tenant_projections') IS NOT NULL`).Scan(&exists); err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	tag, err := db.Exec(ctx, `
		DELETE FROM public.support_tenant_projections
		WHERE "tenantId" = $1
	`, tenantID)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

func quoteClickHouseIdentifier(value string) string {
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}

func printInventory(label string, counts []tableCount) {
	var total int64
	for _, item := range counts {
		total += item.Count
	}

	fmt.Printf("\n%s (%d rows):\n", label, total)
	if len(counts) == 0 {
		fmt.Println("  no tenant rows found")
		return
	}

	for _, item := range counts {
		fmt.Printf("  %-48s %d\n", item.Name, item.Count)
	}
}

func formatBlockers(blockers map[string]string) string {
	tables := make([]string, 0, len(blockers))
	for table := range blockers {
		tables = append(tables, table)
	}
	sort.Strings(tables)

	parts := make([]string, 0, len(tables))
	for _, table := range tables {
		parts = append(parts, fmt.Sprintf("%s (%s)", table, blockers[table]))
	}

	return strings.Join(parts, ", ")
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "tenant-purge: "+format+"\n", args...)
	os.Exit(1)
}
