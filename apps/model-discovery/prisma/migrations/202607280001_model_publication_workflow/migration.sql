CREATE TYPE "PricingAuthority" AS ENUM ('import', 'admin');

ALTER TABLE "Model"
  ADD COLUMN "pricingAuthority" "PricingAuthority" NOT NULL DEFAULT 'import';

CREATE INDEX "Model_lifecycleStatus_idx" ON "Model"("lifecycleStatus");
CREATE INDEX "Model_pricingAuthority_idx" ON "Model"("pricingAuthority");
