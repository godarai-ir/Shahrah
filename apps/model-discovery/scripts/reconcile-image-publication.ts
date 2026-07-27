import 'dotenv/config';
import { PrismaPg } from '@prisma/adapter-pg';
import { PrismaClient } from '../src/generated/prisma/client';
import { evaluatePublicationReadiness } from '../src/publication/publication-policy';
import { serializePricingProfiles } from '../src/pricing/pricing-response';

const connectionString =
  process.env.DATABASE_URL ??
  'postgresql://postgres:postgres@127.0.0.1:5432/model_discovery?schema=public';
const prisma = new PrismaClient({
  adapter: new PrismaPg({ connectionString }),
});

async function main() {
  const apply = process.argv.includes('--apply');
  const expectedArg = process.argv.indexOf('--expect-unresolved');
  const expected =
    expectedArg >= 0 ? Number(process.argv[expectedArg + 1]) : undefined;
  const rows = await prisma.model.findMany({
    where: {
      mode: { has: 'image' },
      status: { in: ['active', 'preview'] },
    },
    include: {
      provider: true,
      pricing: {
        include: { region: true, rates: true },
        orderBy: [{ region: { slug: 'asc' } }, { effectiveFrom: 'desc' }],
      },
    },
    orderBy: [{ provider: { slug: 'asc' } }, { canonicalId: 'asc' }],
  });

  const audited = rows.map((row) => ({
    row,
    readiness: evaluatePublicationReadiness({
      mode: row.mode,
      pricingStatus: row.pricingStatus,
      pricingProfiles: serializePricingProfiles(row.pricing),
    }),
  }));
  const unresolved = audited.filter((item) => !item.readiness.eligible);
  const providerCounts: Record<string, number> = {};
  const pricingCounts: Record<string, number> = {};
  for (const item of unresolved) {
    providerCounts[item.row.provider.name] =
      (providerCounts[item.row.provider.name] ?? 0) + 1;
    pricingCounts[item.row.pricingStatus] =
      (pricingCounts[item.row.pricingStatus] ?? 0) + 1;
  }

  const report = {
    apply,
    total: audited.length,
    resolved: audited.length - unresolved.length,
    unresolved: unresolved.length,
    pricingCounts,
    providerCounts,
    models: unresolved.map((item) => ({
      canonicalId: item.row.canonicalId,
      provider: item.row.provider.slug,
      pricingStatus: item.row.pricingStatus,
      lifecycleStatus: item.row.lifecycleStatus,
      blockers: item.readiness.blockers,
    })),
  };
  console.log(JSON.stringify(report, null, 2));

  if (Number.isFinite(expected) && unresolved.length !== expected) {
    throw new Error(
      `Expected ${expected} unresolved image models, found ${unresolved.length}.`,
    );
  }
  if (!apply) return;

  for (const item of unresolved) {
    if (item.row.lifecycleStatus !== 'published') continue;
    await prisma.$transaction([
      prisma.model.update({
        where: { id: item.row.id },
        data: { lifecycleStatus: 'draft' },
      }),
      prisma.auditLog.create({
        data: {
          entityType: 'model',
          entityId: item.row.id,
          action: 'publication_readiness_demote',
          beforeJson: {
            lifecycleStatus: item.row.lifecycleStatus,
          },
          afterJson: {
            lifecycleStatus: 'draft',
            blockers: item.readiness.blockers,
          },
          actorId: 'system-publication-reconciler',
          actorRole: 'system',
        },
      }),
      prisma.modelChangelog.create({
        data: {
          modelId: item.row.id,
          changeType: 'publication',
          title: 'Unpublished because image pricing is incomplete',
          createdBy: 'system-publication-reconciler',
        },
      }),
    ]);
  }
}

main()
  .catch((error) => {
    console.error(error);
    process.exitCode = 1;
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
