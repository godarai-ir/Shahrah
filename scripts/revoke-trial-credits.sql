-- Removes the remaining trial credits from every tenant. Trial credits are
-- retired in favor of the free seat grant (seat_free).
--
-- The credit ledger is append-only, so lots are never deleted. For each trial
-- lot with credits left this writes one audited 'expire' entry that zeroes it,
-- the same way the expiry job does, and updates both balance projections.
-- Spent credits stay in history. Re-running is safe: revoke:{lot_id} keys make
-- each lot revoke at most once.
--
-- Usage (psql, main control-plane database):
--   1. Dry run (default): psql "$DATABASE_URL" -v actor=<your-name> -f scripts/revoke-trial-credits.sql
--      It prints what would change and ends with ROLLBACK.
--   2. Apply:             psql "$DATABASE_URL" -v actor=<your-name> -v apply=true -f scripts/revoke-trial-credits.sql
--
-- Optional: -v require_seat_free=false also revokes tenants that hold no
-- seat_free lot (they would be left without credits). Default is true.
--
-- Take a backup first. After applying, pool windows cached in Redis are not
-- re-seeded; they refresh at the next window rollover.

\set ON_ERROR_STOP on

\if :{?actor}
\else
  \echo 'error: pass -v actor=<your-name>'
  \quit 1
\endif

\if :{?apply}
\else
  \set apply false
\endif

\if :{?require_seat_free}
\else
  \set require_seat_free true
\endif

BEGIN;

SET LOCAL lock_timeout = '10s';

SELECT public.credit_ledger_ensure_partitions(now(), 1);

-- Take the tenant locks first, like the application does, so writers never deadlock.
INSERT INTO public.credit_tenant_balances (tenant_id)
SELECT DISTINCT l.tenant_id
FROM public.credit_lots l
JOIN public.credit_lot_balances b ON b.lot_id = l.lot_id
WHERE l.source = 'trial' AND b.remaining_mc > 0
ON CONFLICT (tenant_id) DO NOTHING;

DO $lock$
BEGIN
    PERFORM 1
    FROM public.credit_tenant_balances t
    WHERE t.tenant_id IN (
        SELECT l.tenant_id
        FROM public.credit_lots l
        JOIN public.credit_lot_balances b ON b.lot_id = l.lot_id
        WHERE l.source = 'trial' AND b.remaining_mc > 0
    )
    ORDER BY t.tenant_id
    FOR UPDATE OF t;
END
$lock$;

CREATE TEMP TABLE revoke_candidates ON COMMIT DROP AS
SELECT b.lot_id,
       b.tenant_id,
       b.remaining_mc,
       gen_random_uuid() AS entry_id,
       EXISTS (SELECT 1 FROM public.credit_lots f
               WHERE f.tenant_id = b.tenant_id AND f.source = 'seat_free') AS has_seat_free
FROM public.credit_lot_balances b
JOIN public.credit_lots l ON l.lot_id = b.lot_id
WHERE l.source = 'trial'
  AND b.remaining_mc > 0
  AND NOT EXISTS (SELECT 1 FROM public.credit_ledger_keys k WHERE k.idempotency_key = 'revoke:' || b.lot_id::text)
ORDER BY b.tenant_id, b.lot_id
FOR UPDATE OF b;

DELETE FROM revoke_candidates WHERE :require_seat_free AND NOT has_seat_free;

\echo '--- To revoke ---'
SELECT count(*) AS lots, count(DISTINCT tenant_id) AS tenants,
       COALESCE(sum(remaining_mc), 0) / 1000.0 AS credits
FROM revoke_candidates;

\echo '--- Skipped: tenants with trial credits but no seat_free lot ---'
SELECT l.tenant_id, sum(b.remaining_mc) / 1000.0 AS credits
FROM public.credit_lots l
JOIN public.credit_lot_balances b ON b.lot_id = l.lot_id
WHERE l.source = 'trial' AND b.remaining_mc > 0
  AND NOT EXISTS (SELECT 1 FROM revoke_candidates c WHERE c.lot_id = l.lot_id)
  AND NOT EXISTS (SELECT 1 FROM public.credit_ledger_keys k WHERE k.idempotency_key = 'revoke:' || l.lot_id::text)
GROUP BY l.tenant_id
ORDER BY l.tenant_id;

INSERT INTO public.credit_ledger_keys (idempotency_key, entry_id, tenant_id, occurred_at)
SELECT 'revoke:' || lot_id::text, entry_id, tenant_id, now() FROM revoke_candidates;

INSERT INTO public.credit_ledger (entry_id, tenant_id, kind, amount_mc, idempotency_key, occurred_at, lot_id, reason, actor)
SELECT entry_id, tenant_id, 'expire', -remaining_mc, 'revoke:' || lot_id::text, now(), lot_id,
       'trial credits retired in favor of the free seat grant', :'actor'
FROM revoke_candidates;

INSERT INTO public.credit_ledger_allocations (allocation_id, entry_id, tenant_id, lot_id, amount_mc, occurred_at)
SELECT gen_random_uuid(), entry_id, tenant_id, lot_id, -remaining_mc, now() FROM revoke_candidates;

UPDATE public.credit_lot_balances b
SET remaining_mc = 0, updated_at = now()
FROM revoke_candidates c
WHERE b.lot_id = c.lot_id;

UPDATE public.credit_tenant_balances t
SET available_mc = t.available_mc - s.total,
    ledger_sum_mc = t.ledger_sum_mc - s.total,
    updated_at = now()
FROM (SELECT tenant_id, sum(remaining_mc) AS total FROM revoke_candidates GROUP BY tenant_id) s
WHERE t.tenant_id = s.tenant_id;

-- Invariants: both must return zero rows, or nothing should be committed.
\echo '--- Check: tenant projections vs ledger and lots (expect 0 rows) ---'
SELECT t.tenant_id
FROM public.credit_tenant_balances t
WHERE t.tenant_id IN (SELECT tenant_id FROM revoke_candidates)
  AND (t.available_mc <> COALESCE((SELECT sum(remaining_mc) FROM public.credit_lot_balances b WHERE b.tenant_id = t.tenant_id), 0)
       OR t.available_mc - t.debt_mc <> t.ledger_sum_mc
       OR t.ledger_sum_mc <> COALESCE((SELECT sum(amount_mc) FROM public.credit_ledger g WHERE g.tenant_id = t.tenant_id), 0));

\echo '--- Check: revoked lots are zero (expect 0 rows) ---'
SELECT b.lot_id FROM public.credit_lot_balances b
JOIN revoke_candidates c ON c.lot_id = b.lot_id
WHERE b.remaining_mc <> 0;

\if :apply
  \echo 'COMMIT (apply=true)'
  COMMIT;
\else
  \echo 'ROLLBACK (dry run; pass -v apply=true to commit)'
  ROLLBACK;
\endif
