CREATE TABLE "ImportSourceState" (
    "id" TEXT NOT NULL,
    "sourceKey" TEXT NOT NULL,
    "repoUrl" TEXT NOT NULL,
    "ref" TEXT NOT NULL,
    "lastCommitSha" TEXT,
    "lastImportedAt" TIMESTAMP(3),
    "lastCheckedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "ImportSourceState_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "ImportSourceState_sourceKey_key" ON "ImportSourceState"("sourceKey");
CREATE INDEX "ImportSourceState_repoUrl_ref_idx" ON "ImportSourceState"("repoUrl", "ref");
