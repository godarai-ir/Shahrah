import 'dotenv/config';
import { PrismaPg } from '@prisma/adapter-pg';
import { PrismaClient } from '../src/generated/prisma/client';

async function main() {
  const credentialId = process.argv.find((value) => value.startsWith('--key-id='))?.slice('--key-id='.length);
  if (!credentialId || !/^[A-Za-z0-9_-]{8,64}$/.test(credentialId)) {
    throw new Error('Usage: --key-id=<credential id from the sa_<key-id>.<secret> token>');
  }
  const connectionString = process.env.DATABASE_URL;
  if (!connectionString) throw new Error('DATABASE_URL is required');

  const prisma = new PrismaClient({ adapter: new PrismaPg({ connectionString }) });
  try {
    const result = await prisma.serviceAccountCredential.updateMany({
      where: { id: credentialId, revokedAt: null },
      data: { revokedAt: new Date() },
    });
    if (result.count !== 1) throw new Error('Active credential was not found');
    process.stdout.write(`Service credential ${credentialId} was revoked.\n`);
  } finally {
    await prisma.$disconnect();
  }
}

main().catch((error: unknown) => {
  console.error(error instanceof Error ? error.message : error);
  process.exitCode = 1;
});
