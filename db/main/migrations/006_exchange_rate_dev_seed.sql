-- Ensure a stable development exchange-rate seed exists.
-- 1 USD = 1,500,000 IRR
WITH updated AS (
    UPDATE exchange_rates
    SET
        usd_to_irr = 1500000,
        fetched_at = NOW()
    WHERE source = 'manual-dev-seed'
    RETURNING 1
)
INSERT INTO exchange_rates (usd_to_irr, source, fetched_at)
SELECT 1500000, 'manual-dev-seed', NOW()
WHERE NOT EXISTS (SELECT 1 FROM updated);
