// @ts-expect-error - PrismaClient generated at build time
import { PrismaClient } from "@prisma/client";
import { PrismaPg } from "@prisma/adapter-pg";

declare global {
  // eslint-disable-next-line no-var
  var __tenantAdminPrisma: PrismaClient | undefined;
}

const databaseURL = process.env.DATABASE_URL;
if (!databaseURL) {
  throw new Error("Missing required DATABASE_URL");
}

const adapter = new PrismaPg({
  connectionString: databaseURL,
});

export const prisma =
  global.__tenantAdminPrisma ??
  new PrismaClient({
    adapter,
    log: process.env.NODE_ENV === "development" ? ["warn", "error"] : ["error"],
  });

if (process.env.NODE_ENV !== "production") {
  global.__tenantAdminPrisma = prisma;
}
