CREATE TABLE "support_tenant_projections" (
    "tenantId" TEXT NOT NULL,
    "slug" TEXT NOT NULL,
    "displayName" TEXT,
    "status" TEXT NOT NULL,
    "plan" TEXT,
    "walletStatus" TEXT,
    "lastActivityAt" TIMESTAMP(3),
    "sourceUpdatedAt" TIMESTAMP(3) NOT NULL,
    "projectionVersion" BIGINT NOT NULL,
    "projectedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deletedAt" TIMESTAMP(3),
    CONSTRAINT "support_tenant_projections_pkey" PRIMARY KEY ("tenantId")
);

CREATE UNIQUE INDEX "support_tenant_projections_slug_key" ON "support_tenant_projections"("slug");
CREATE INDEX "support_tenant_projections_status_idx" ON "support_tenant_projections"("status");
CREATE INDEX "support_tenant_projections_lastActivityAt_idx" ON "support_tenant_projections"("lastActivityAt");
