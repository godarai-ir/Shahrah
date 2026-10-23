-- +goose Up
-- Trial credits are retired: the free seat grant (seat_free) now plays the
-- trial's role. This zeroes the credits that are left in existing trial lots.
--
-- The credit ledger is append-only, so lots are never deleted. Each trial lot
-- with credits left gets one audited 'expire' entry, the same way the expiry
-- job zeroes a lot, and both balance projections are updated. Spent credits
-- stay in history. Tenants without a seat_free lot keep their trial lot so
-- nobody is left with no credits; it expires on its own within the trial window.
-- The revoke:{lot_id} key makes each lot revoke at most once.

-- +goose StatementBegin
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

CREATE TEMP TABLE revoke_trial_candidates ON COMMIT DROP AS
SELECT b.lot_id, b.tenant_id, b.remaining_mc, gen_random_uuid() AS entry_id
FROM public.credit_lot_balances b
JOIN public.credit_lots l ON l.lot_id = b.lot_id
WHERE l.source = 'trial'
  AND b.remaining_mc > 0
  AND EXISTS (SELECT 1 FROM public.credit_lots f WHERE f.tenant_id = b.tenant_id AND f.source = 'seat_free')
  AND NOT EXISTS (SELECT 1 FROM public.credit_ledger_keys k WHERE k.idempotency_key = 'revoke:' || b.lot_id::text)
ORDER BY b.tenant_id, b.lot_id
FOR UPDATE OF b;

INSERT INTO public.credit_ledger_keys (idempotency_key, entry_id, tenant_id, occurred_at)
SELECT 'revoke:' || lot_id::text, entry_id, tenant_id, now() FROM revoke_trial_candidates;

INSERT INTO public.credit_ledger (entry_id, tenant_id, kind, amount_mc, idempotency_key, occurred_at, lot_id, reason, actor)
SELECT entry_id, tenant_id, 'expire', -remaining_mc, 'revoke:' || lot_id::text, now(), lot_id,
       'trial credits retired in favor of the free seat grant', 'migration:078'
FROM revoke_trial_candidates;

INSERT INTO public.credit_ledger_allocations (allocation_id, entry_id, tenant_id, lot_id, amount_mc, occurred_at)
SELECT gen_random_uuid(), entry_id, tenant_id, lot_id, -remaining_mc, now() FROM revoke_trial_candidates;

UPDATE public.credit_lot_balances b
SET remaining_mc = 0, updated_at = now()
FROM revoke_trial_candidates c
WHERE b.lot_id = c.lot_id;

UPDATE public.credit_tenant_balances t
SET available_mc = t.available_mc - s.total,
    ledger_sum_mc = t.ledger_sum_mc - s.total,
    updated_at = now()
FROM (SELECT tenant_id, sum(remaining_mc) AS total FROM revoke_trial_candidates GROUP BY tenant_id) s
WHERE t.tenant_id = s.tenant_id;

-- A projection that disagrees with the ledger aborts the migration.
DO $check$
DECLARE
    broken integer;
BEGIN
    SELECT count(*) INTO broken
    FROM public.credit_tenant_balances t
    WHERE t.tenant_id IN (SELECT tenant_id FROM revoke_trial_candidates)
      AND (t.available_mc <> COALESCE((SELECT sum(remaining_mc) FROM public.credit_lot_balances b WHERE b.tenant_id = t.tenant_id), 0)
           OR t.available_mc - t.debt_mc <> t.ledger_sum_mc
           OR t.ledger_sum_mc <> COALESCE((SELECT sum(amount_mc) FROM public.credit_ledger g WHERE g.tenant_id = t.tenant_id), 0));

    IF broken > 0 THEN
        RAISE EXCEPTION 'trial revocation left % tenant credit projections inconsistent', broken;
    END IF;
END
$check$;
-- +goose StatementEnd

-- +goose Down
-- Revoked credits are not restored: the ledger is append-only and the trial grant is retired.
