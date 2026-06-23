ALTER TABLE "AuditLog"
ADD COLUMN "actorRole" TEXT,
ADD COLUMN "serviceAccountId" TEXT,
ADD COLUMN "requestId" TEXT;

CREATE TABLE "ServiceAccount" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "scopes" TEXT[],
    "active" BOOLEAN NOT NULL DEFAULT true,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    CONSTRAINT "ServiceAccount_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "ServiceAccountCredential" (
    "id" TEXT NOT NULL,
    "serviceAccountId" TEXT NOT NULL,
    "secretHash" TEXT NOT NULL,
    "expiresAt" TIMESTAMP(3),
    "revokedAt" TIMESTAMP(3),
    "lastUsedAt" TIMESTAMP(3),
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "ServiceAccountCredential_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "ServiceAccount_name_key" ON "ServiceAccount"("name");
CREATE INDEX "ServiceAccountCredential_serviceAccountId_idx" ON "ServiceAccountCredential"("serviceAccountId");
CREATE INDEX "ServiceAccountCredential_expiresAt_idx" ON "ServiceAccountCredential"("expiresAt");

ALTER TABLE "ServiceAccountCredential"
ADD CONSTRAINT "ServiceAccountCredential_serviceAccountId_fkey"
FOREIGN KEY ("serviceAccountId") REFERENCES "ServiceAccount"("id")
ON DELETE CASCADE ON UPDATE CASCADE;
