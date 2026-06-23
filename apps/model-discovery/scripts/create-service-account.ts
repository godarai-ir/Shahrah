import 'dotenv/config';
import { PrismaPg } from '@prisma/adapter-pg';
import { PrismaClient } from '../src/generated/prisma/client';
import { createHash, randomBytes } from 'crypto';

const allowedScopes = new Set([
  'catalog:read', 'catalog:write', 'provider:write', 'import:execute',
  'indexing:execute', 'display-settings:write',
]);

async function main() {
  const name = process.argv.find((value) => value.startsWith('--name='))?.slice('--name='.length);
  const scopesArg = process.argv.find((value) => value.startsWith('--scopes='))?.slice('--scopes='.length);
  const expiresAtArg = process.argv.find((value) => value.startsWith('--expires-at='))?.slice('--expires-at='.length);
  const scopes = scopesArg?.split(',').map((value) => value.trim()).filter(Boolean) ?? [];

  if (!name || !/^[a-z0-9-]{3,64}$/.test(name) || !scopes.length || scopes.some((scope) => !allowedScopes.has(scope))) {
    throw new Error('Usage: --name=support-admin --scopes=catalog:read,catalog:write,provider:write[,import:execute,indexing:execute,display-settings:write] [--expires-at=ISO_DATE]');
  }

  const expiresAt = expiresAtArg ? new Date(expiresAtArg) : null;
  if (expiresAt && (!Number.isFinite(expiresAt.getTime()) || expiresAt <= new Date())) {
    throw new Error('--expires-at must be a valid future ISO date');
  }
  const connectionString = process.env.DATABASE_URL;
  if (!connectionString) throw new Error('DATABASE_URL is required');

  const prisma = new PrismaClient({ adapter: new PrismaPg({ connectionString }) });
  try {
    const secret = randomBytes(32).toString('base64url');
    const credentialId = randomBytes(12).toString('base64url');
    const account = await prisma.serviceAccount.upsert({
      where: { name },
      create: { name, scopes },
      update: { scopes, active: true },
    });
    await prisma.serviceAccountCredential.create({
      data: {
        id: credentialId,
        serviceAccountId: account.id,
        secretHash: createHash('sha256').update(secret).digest('hex'),
        expiresAt,
      },
    });
    process.stdout.write(`MODEL_DISCOVERY_SERVICE_TOKEN=sa_${credentialId}.${secret}\n`);
    process.stderr.write('Store this credential now; only its hash is persisted. Re-running creates a rotation credential.\n');
  } finally {
    await prisma.$disconnect();
  }
}

main().catch((error: unknown) => {
  console.error(error instanceof Error ? error.message : error);
  process.exitCode = 1;
});
