-- +goose Up
-- +goose NO TRANSACTION
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tenant_subscriptions_lifecycle_due
    ON public.tenant_subscriptions (current_period_end ASC, subscription_id)
    WHERE status IN ('trialing', 'active', 'past_due');

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS public.idx_tenant_subscriptions_lifecycle_due;
