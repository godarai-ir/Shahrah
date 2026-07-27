-- This is an intentional one-shot pricing cutover. Model pricing is rebuilt
-- from the YAML source of truth after this migration.
CREATE TYPE "ModelPricingStatus" AS ENUM ('unknown', 'priced', 'free');
CREATE TYPE "PricingUsage" AS ENUM ('input', 'output', 'cache_read', 'cache_write');
CREATE TYPE "PricingServiceTier" AS ENUM ('standard', 'batch', 'priority');

TRUNCATE TABLE "ModelPricing";

ALTER TABLE "Model"
    ADD COLUMN "pricingStatus" "ModelPricingStatus" NOT NULL DEFAULT 'unknown';

UPDATE "Model"
SET "pricingStatus" = CASE
    WHEN "free" THEN 'free'::"ModelPricingStatus"
    ELSE 'unknown'::"ModelPricingStatus"
END;

ALTER TABLE "Model" DROP COLUMN "free";

ALTER TABLE "ModelPricing"
    DROP COLUMN "inputPricePerMillionTokens",
    DROP COLUMN "outputPricePerMillionTokens",
    DROP COLUMN "cachedInputPricePerMillionTokens",
    DROP COLUMN "imageInputPrice",
    DROP COLUMN "imageOutputPrice",
    DROP COLUMN "audioInputPrice",
    DROP COLUMN "audioOutputPrice",
    DROP COLUMN "requestPrice";

ALTER TABLE "ModelPricing"
    ALTER COLUMN "regionId" SET NOT NULL;

CREATE TABLE "ModelPricingRate" (
    "id" TEXT NOT NULL,
    "modelPricingId" TEXT NOT NULL,
    "rateKey" TEXT NOT NULL,
    "usage" "PricingUsage" NOT NULL,
    "unit" TEXT NOT NULL,
    "amount" DECIMAL(30,18) NOT NULL,
    "quantity" INTEGER NOT NULL,
    "serviceTier" "PricingServiceTier" NOT NULL DEFAULT 'standard',
    "resolution" TEXT,
    "tierFrom" INTEGER,
    "tierTo" INTEGER,
    "tierMode" TEXT,
    "qualifiers" JSONB,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "ModelPricingRate_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "ModelPricingRate_amount_nonnegative" CHECK ("amount" >= 0),
    CONSTRAINT "ModelPricingRate_quantity_positive" CHECK ("quantity" > 0),
    CONSTRAINT "ModelPricingRate_tier_bounds" CHECK (
        "tierTo" IS NULL OR "tierFrom" IS NULL OR "tierTo" > "tierFrom"
    )
);

CREATE UNIQUE INDEX "ModelPricingRate_modelPricingId_rateKey_key"
    ON "ModelPricingRate"("modelPricingId", "rateKey");
CREATE INDEX "ModelPricingRate_usage_unit_idx"
    ON "ModelPricingRate"("usage", "unit");
CREATE INDEX "ModelPricingRate_resolution_idx"
    ON "ModelPricingRate"("resolution");
CREATE INDEX "ModelPricingRate_amount_idx"
    ON "ModelPricingRate"("amount");
CREATE INDEX "ModelPricing_modelId_regionId_idx"
    ON "ModelPricing"("modelId", "regionId");
CREATE INDEX "ModelPricing_modelId_effectiveFrom_effectiveTo_idx"
    ON "ModelPricing"("modelId", "effectiveFrom", "effectiveTo");

ALTER TABLE "ModelPricingRate"
    ADD CONSTRAINT "ModelPricingRate_modelPricingId_fkey"
    FOREIGN KEY ("modelPricingId") REFERENCES "ModelPricing"("id")
    ON DELETE CASCADE ON UPDATE CASCADE;

DROP INDEX IF EXISTS "ModelPricing_modelId_idx";
DROP INDEX IF EXISTS "ModelPricing_effectiveFrom_effectiveTo_idx";
