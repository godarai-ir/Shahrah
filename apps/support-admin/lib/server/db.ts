import 'server-only';

import { PrismaPg } from '@prisma/adapter-pg';
import { PrismaClient } from '../../generated/prisma/client';

const globalForPrisma = globalThis as unknown as { supportAdminPrisma?: PrismaClient };

function createClient() {
  const connectionString = process.env.DATABASE_URL;
  if (!connectionString) throw new Error('DATABASE_URL is required for support-admin authentication');
  return new PrismaClient({ adapter: new PrismaPg({ connectionString }) });
}

export const supportAdminDb = globalForPrisma.supportAdminPrisma ?? createClient();

if (process.env.NODE_ENV !== 'production') globalForPrisma.supportAdminPrisma = supportAdminDb;
