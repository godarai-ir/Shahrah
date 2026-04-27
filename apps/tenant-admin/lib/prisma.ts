// @ts-expect-error - PrismaClient generated at build time
import { PrismaClient } from "@prisma/client";

declare global {
  // eslint-disable-next-line no-var
  var __tenantAdminPrisma: PrismaClient | undefined;
}

export const prisma =
  global.__tenantAdminPrisma ??
  new PrismaClient({
    log: process.env.NODE_ENV === "development" ? ["warn", "error"] : ["error"],
  });

if (process.env.NODE_ENV !== "production") {
  global.__tenantAdminPrisma = prisma;
}
