import 'dotenv/config';
import { PrismaPg } from '@prisma/adapter-pg';
import { PrismaClient } from '../src/generated/prisma/client';

const suspiciousNames = new Set([
  'apikey',
  'clientsecret',
  'secretaccesskey',
  'sessiontoken',
  'serviceaccountjson',
  'credential',
  'credentials',
  'token',
  'password',
]);
const secretReferencePattern = /(?:tfy-(?:k8s-)?secret:\/\/|(?:vault|secret):\/\/|::\/)/i;
const normalize = (value: string) => value.toLowerCase().replace(/[^a-z0-9]/g, '');

function suspiciousPaths(value: unknown, path = '$'): string[] {
  if (Array.isArray(value)) {
    return value.flatMap((item, index) => suspiciousPaths(item, `${path}[${index}]`));
  }
  if (!value || typeof value !== 'object') {
    return typeof value === 'string' && secretReferencePattern.test(value) ? [path] : [];
  }
  return Object.entries(value).flatMap(([key, nested]) => {
    const child = `${path}.${key}`;
    return [
      ...(suspiciousNames.has(normalize(key)) ? [child] : []),
      ...suspiciousPaths(nested, child),
    ];
  });
}

async function main() {
  const connectionString = process.env.DATABASE_URL;
  if (!connectionString) throw new Error('DATABASE_URL is required');
  const prisma = new PrismaClient({ adapter: new PrismaPg({ connectionString }) });
  try {
    const rows = await prisma.providerConnectionProfile.findMany({
      select: {
        id: true,
        configJson: true,
        provider: { select: { slug: true } },
      },
    });
    const affected = rows
      .map((row) => ({
        rowId: row.id,
        providerSlug: row.provider.slug,
        suspiciousKeys: [...new Set(suspiciousPaths(row.configJson))].sort(),
      }))
      .filter((row) => row.suspiciousKeys.length > 0);
    process.stdout.write(`${JSON.stringify({ affected }, null, 2)}\n`);
    process.stderr.write(
      'Audit only: rotate every affected credential through its provider and store the replacement in control-plane before retiring the legacy row. No data was changed.\n',
    );
  } finally {
    await prisma.$disconnect();
  }
}

main().catch((error: unknown) => {
  process.stderr.write(`${error instanceof Error ? error.message : 'audit_failed'}\n`);
  process.exitCode = 1;
});
