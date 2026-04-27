import { PrismaClient } from "@prisma/client";
import { PrismaPg } from "@prisma/adapter-pg";

const globalForPrisma = globalThis as typeof globalThis & {
  __tenantAdminPrisma?: PrismaClient;
};

const databaseURL = process.env.DATABASE_URL;
if (!databaseURL) {
  throw new Error("Missing required DATABASE_URL");
}

const adapter = new PrismaPg({
  connectionString: databaseURL,
});

export const prisma =
  globalForPrisma.__tenantAdminPrisma ??
  new PrismaClient({
    adapter,
    log: process.env.NODE_ENV === "development" ? ["warn", "error"] : ["error"],
  });

if (process.env.NODE_ENV !== "production") {
  globalForPrisma.__tenantAdminPrisma = prisma;
}
