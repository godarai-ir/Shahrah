-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.exchange_rate_current (
    rate_id uuid DEFAULT gen_random_uuid() NOT NULL,
    source character varying(255) NOT NULL,
    symbol character varying(64),
    usd_to_irr double precision NOT NULL,
    fetched_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT exchange_rate_current_pkey PRIMARY KEY (rate_id),
    CONSTRAINT exchange_rate_current_source_key UNIQUE (source)
);

CREATE INDEX IF NOT EXISTS idx_exchange_rate_current_fetched_at
    ON public.exchange_rate_current USING btree (fetched_at DESC);

INSERT INTO public.exchange_rate_current (rate_id, source, symbol, usd_to_irr, fetched_at, updated_at)
SELECT rate_id,
       source,
       CASE
           WHEN POSITION(':' IN source) > 0 THEN UPPER(NULLIF(BTRIM(SPLIT_PART(source, ':', 2)), ''))
           ELSE NULL
       END AS symbol,
       usd_to_irr,
       fetched_at,
       NOW()
FROM (
    SELECT DISTINCT ON (source) rate_id, source, usd_to_irr, fetched_at
    FROM public.exchange_rates
    ORDER BY source, fetched_at DESC
) latest
ON CONFLICT (source) DO UPDATE SET
    symbol = EXCLUDED.symbol,
    usd_to_irr = EXCLUDED.usd_to_irr,
    fetched_at = EXCLUDED.fetched_at,
    updated_at = NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.exchange_rate_current;
-- +goose StatementEnd
