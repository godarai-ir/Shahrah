CREATE TABLE "ProviderLogoAsset" (
    "id" TEXT NOT NULL,
    "providerId" TEXT NOT NULL,
    "fileName" TEXT NOT NULL,
    "contentHash" TEXT NOT NULL,
    "contentType" TEXT NOT NULL DEFAULT 'image/png',
    "width" INTEGER NOT NULL DEFAULT 64,
    "height" INTEGER NOT NULL DEFAULT 64,
    "sizeBytes" INTEGER NOT NULL,
    "data" BYTEA NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "ProviderLogoAsset_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "ProviderLogoAsset_fileName_key" ON "ProviderLogoAsset"("fileName");
CREATE INDEX "ProviderLogoAsset_providerId_idx" ON "ProviderLogoAsset"("providerId");
CREATE INDEX "ProviderLogoAsset_contentHash_idx" ON "ProviderLogoAsset"("contentHash");

ALTER TABLE "ProviderLogoAsset"
    ADD CONSTRAINT "ProviderLogoAsset_providerId_fkey"
    FOREIGN KEY ("providerId") REFERENCES "Provider"("id") ON DELETE CASCADE ON UPDATE CASCADE;
