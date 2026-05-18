-- AlterTable
ALTER TABLE "Provider" ADD COLUMN "enabled" BOOLEAN NOT NULL DEFAULT true;

-- AlterTable
ALTER TABLE "Model" ADD COLUMN "enabled" BOOLEAN NOT NULL DEFAULT true,
ADD COLUMN "modelSeriesId" TEXT;

-- CreateTable
CREATE TABLE "ModelSeries" (
  "id" TEXT NOT NULL,
  "slug" TEXT NOT NULL,
  "name" TEXT NOT NULL,
  "logoUrl" TEXT,
  "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updatedAt" TIMESTAMP(3) NOT NULL,
  CONSTRAINT "ModelSeries_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ProviderConnectionProfile" (
  "id" TEXT NOT NULL,
  "providerId" TEXT NOT NULL,
  "providerType" TEXT NOT NULL,
  "authMode" TEXT,
  "configJson" JSONB NOT NULL,
  "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updatedAt" TIMESTAMP(3) NOT NULL,
  CONSTRAINT "ProviderConnectionProfile_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "ModelSeries_slug_key" ON "ModelSeries"("slug");
CREATE INDEX "Model_enabled_idx" ON "Model"("enabled");
CREATE UNIQUE INDEX "ProviderConnectionProfile_providerId_key" ON "ProviderConnectionProfile"("providerId");

-- AddForeignKey
ALTER TABLE "Model" ADD CONSTRAINT "Model_modelSeriesId_fkey" FOREIGN KEY ("modelSeriesId") REFERENCES "ModelSeries"("id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "ProviderConnectionProfile" ADD CONSTRAINT "ProviderConnectionProfile_providerId_fkey" FOREIGN KEY ("providerId") REFERENCES "Provider"("id") ON DELETE CASCADE ON UPDATE CASCADE;
