import { loadEnvFile } from 'node:process';
import { PrismaPg } from '@prisma/adapter-pg';
import { PrismaClient } from '../generated/prisma/client';
import { hashPassword } from '../lib/server/password';

try {
  loadEnvFile();
} catch (error) {
  const code = (error as NodeJS.ErrnoException).code;
  if (code !== 'ENOENT') throw error;
}

const allowedRoles = new Set(['super_admin', 'catalog_admin', 'support_agent']);

async function main() {
  const email = process.argv.find((value) => value.startsWith('--email='))?.slice('--email='.length).trim().toLowerCase();
  const role = process.argv.find((value) => value.startsWith('--role='))?.slice('--role='.length) ?? 'super_admin';
  const password = process.env.SUPPORT_ADMIN_INITIAL_PASSWORD;

  if (!email || !email.includes('@') || !password || password.length < 12 || !allowedRoles.has(role)) {
    throw new Error('Usage: set SUPPORT_ADMIN_INITIAL_PASSWORD (12+ chars), then pass --email=user@example.com [--role=super_admin|catalog_admin|support_agent]');
  }

  const connectionString = process.env.DATABASE_URL;
  if (!connectionString) throw new Error('DATABASE_URL is required');

  const prisma = new PrismaClient({ adapter: new PrismaPg({ connectionString }) });
  try {
    const passwordHash = await hashPassword(password);
    const user = await prisma.supportAdminUser.upsert({
      where: { email },
      create: { email, passwordHash, role, status: 'active' },
      update: {
        passwordHash,
        role,
        status: 'active',
        failedLoginCount: 0,
        lockedUntil: null,
        passwordChangedAt: new Date(),
        sessions: { deleteMany: {} },
      },
      select: { id: true, email: true, role: true },
    });
    process.stdout.write(`Support admin user ${user.email} (${user.role}) is ready. Existing sessions were revoked.\n`);
  } finally {
    await prisma.$disconnect();
  }
}

main().catch((error: unknown) => {
  console.error(error instanceof Error ? error.message : error);
  process.exitCode = 1;
});
