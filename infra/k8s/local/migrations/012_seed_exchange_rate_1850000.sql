-- +goose Up
-- +goose StatementBegin
INSERT INTO public.exchange_rates (rate_id, usd_to_irr, source, fetched_at)
SELECT gen_random_uuid(), 1850000, 'seed', NOW()
WHERE NOT EXISTS (
    SELECT 1
    FROM public.exchange_rates
    WHERE source = 'seed'
      AND usd_to_irr = 1850000
      AND fetched_at::date = CURRENT_DATE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM public.exchange_rates
WHERE source = 'seed'
  AND usd_to_irr = 1850000
  AND fetched_at::date = CURRENT_DATE;
-- +goose StatementEnd
