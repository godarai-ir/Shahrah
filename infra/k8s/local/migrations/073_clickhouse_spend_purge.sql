-- +goose Up
-- Clears the IRR spend history that ClickHouse holds, as part of the credit
-- fleet cutover (072). ClickHouse migrations are re-applied on every deploy,
-- so the purge cannot live there; this row asks the control-plane to run it
-- once. It zeroes the cost columns of the analytics events, their 1-minute
-- aggregates, and the request logs written before the cutoff, and keeps the
-- rows, so request counts, tokens, and latency stay.
-- See plans/credit-pricing/phase-3-design.md, "ClickHouse spend purge".

-- +goose StatementBegin
CREATE TABLE public.clickhouse_spend_purges (
    purge_id bigserial PRIMARY KEY,
    reason character varying(64) NOT NULL,
    cutoff timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    completed_at timestamptz
);

INSERT INTO public.clickhouse_spend_purges (reason, cutoff) VALUES ('credit_cutover', NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.clickhouse_spend_purges;
-- +goose StatementEnd
