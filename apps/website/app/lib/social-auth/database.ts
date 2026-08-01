import { Pool } from "pg";

declare global {
  var __godaraiSocialAuthPool: Pool | undefined;
}

function createPool() {
  return new Pool({
    connectionString: process.env.DATABASE_URL,
    max: Number(process.env.SOCIAL_AUTH_DATABASE_POOL_SIZE ?? 5),
    idleTimeoutMillis: 30_000,
    connectionTimeoutMillis: 5_000,
    application_name: "godarai-social-auth",
  });
}

export const socialAuthPool =
  globalThis.__godaraiSocialAuthPool ?? createPool();

if (process.env.NODE_ENV !== "production") {
  globalThis.__godaraiSocialAuthPool = socialAuthPool;
}

