-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS uq_tenant_admin_invites_pending_tenant_email_ci
    ON tenant_admin_invitations (tenant_id, LOWER(email))
    WHERE status = 'pending' AND email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_tenant_admin_invites_pending_tenant_mobile
    ON tenant_admin_invitations (tenant_id, mobile)
    WHERE status = 'pending' AND mobile IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_tenant_admin_invites_pending_tenant_mobile;
DROP INDEX IF EXISTS uq_tenant_admin_invites_pending_tenant_email_ci;
-- +goose StatementEnd
