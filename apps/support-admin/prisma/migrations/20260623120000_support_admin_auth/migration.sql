CREATE TABLE "support_admin_users" (
    "id" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "passwordHash" TEXT NOT NULL,
    "role" TEXT NOT NULL DEFAULT 'support_agent',
    "status" TEXT NOT NULL DEFAULT 'active',
    "failedLoginCount" INTEGER NOT NULL DEFAULT 0,
    "lockedUntil" TIMESTAMP(3),
    "passwordChangedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,
    "lastLoginAt" TIMESTAMP(3),
    CONSTRAINT "support_admin_users_pkey" PRIMARY KEY ("id")
);

CREATE TABLE "support_admin_sessions" (
    "id" TEXT NOT NULL,
    "tokenHash" TEXT NOT NULL,
    "userId" TEXT NOT NULL,
    "expiresAt" TIMESTAMP(3) NOT NULL,
    "lastSeenAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "support_admin_sessions_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "support_admin_users_email_key" ON "support_admin_users"("email");
CREATE UNIQUE INDEX "support_admin_sessions_tokenHash_key" ON "support_admin_sessions"("tokenHash");
CREATE INDEX "support_admin_sessions_userId_idx" ON "support_admin_sessions"("userId");
CREATE INDEX "support_admin_sessions_expiresAt_idx" ON "support_admin_sessions"("expiresAt");

ALTER TABLE "support_admin_sessions"
ADD CONSTRAINT "support_admin_sessions_userId_fkey"
FOREIGN KEY ("userId") REFERENCES "support_admin_users"("id")
ON DELETE CASCADE ON UPDATE CASCADE;
