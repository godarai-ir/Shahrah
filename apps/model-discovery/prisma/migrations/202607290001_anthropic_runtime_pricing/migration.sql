-- Lifecycle dates are separate: deprecation warns callers while retirement makes
-- a model unavailable. deprecated_at remains for backward-compatible reads.
ALTER TABLE "Model"
  ADD COLUMN "deprecationDate" TIMESTAMP(3),
  ADD COLUMN "retirementDate" TIMESTAMP(3),
  ADD COLUMN "replacementCanonicalId" TEXT;

CREATE TABLE "ModelRuntimeProfile" (
  "id" TEXT NOT NULL,
  "modelId" TEXT NOT NULL,
  "version" INTEGER NOT NULL,
  "profileJson" JSONB NOT NULL,
  "sourceUrl" TEXT,
  "effectiveFrom" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "effectiveTo" TIMESTAMP(3),
  "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT "ModelRuntimeProfile_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "ModelPricingRate" (
  "id" TEXT NOT NULL,
  "modelId" TEXT NOT NULL,
  "dimension" TEXT NOT NULL,
  "unit" TEXT NOT NULL,
  "unitPriceUsd" DECIMAL(30,12) NOT NULL,
  "currency" TEXT NOT NULL DEFAULT 'USD',
  "conditionJson" JSONB,
  "region" TEXT NOT NULL DEFAULT 'global',
  "effectiveFrom" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "effectiveTo" TIMESTAMP(3),
  "sourceUrl" TEXT NOT NULL,
  "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updatedAt" TIMESTAMP(3) NOT NULL,
  CONSTRAINT "ModelPricingRate_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "ModelRuntimeProfile_modelId_version_key"
  ON "ModelRuntimeProfile"("modelId", "version");
CREATE INDEX "ModelRuntimeProfile_modelId_effectiveFrom_effectiveTo_idx"
  ON "ModelRuntimeProfile"("modelId", "effectiveFrom", "effectiveTo");
CREATE INDEX "ModelPricingRate_modelId_dimension_region_idx"
  ON "ModelPricingRate"("modelId", "dimension", "region");
CREATE INDEX "ModelPricingRate_effectiveFrom_effectiveTo_idx"
  ON "ModelPricingRate"("effectiveFrom", "effectiveTo");

ALTER TABLE "ModelRuntimeProfile"
  ADD CONSTRAINT "ModelRuntimeProfile_modelId_fkey"
  FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "ModelPricingRate"
  ADD CONSTRAINT "ModelPricingRate_modelId_fkey"
  FOREIGN KEY ("modelId") REFERENCES "Model"("id") ON DELETE CASCADE ON UPDATE CASCADE;
