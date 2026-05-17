import 'dotenv/config';
import { PrismaPg } from '@prisma/adapter-pg';
import { PrismaClient } from '../src/generated/prisma/client';

const connectionString =
  process.env.DATABASE_URL ??
  'postgresql://postgres:postgres@127.0.0.1:5432/model_discovery?schema=public';
const adapter = new PrismaPg({ connectionString });
const prisma = new PrismaClient({ adapter });

const providers = [
  'openai','anthropic','google','meta','mistral','deepseek','cohere','groq','xai','qwen'
];

const categories = ['programming','reasoning','vision','long-context','low-cost','multilingual'];
const capabilities = ['streaming','tool_calling','json_mode','vision','reasoning','web_search','prompt_caching'];
const params = ['temperature','top_p','max_tokens','tools','response_format','json_schema','stream'];
const modalities = ['text','image','audio','embeddings'];
const regions = ['global','us','eu','asia'];

async function main() {
  await prisma.modelTag.deleteMany();
  await prisma.modelBenchmark.deleteMany();
  await prisma.modelRegion.deleteMany();
  await prisma.modelSupportedParameter.deleteMany();
  await prisma.modelInputModality.deleteMany();
  await prisma.modelOutputModality.deleteMany();
  await prisma.modelCapability.deleteMany();
  await prisma.modelPricing.deleteMany();
  await prisma.modelVersion.deleteMany();
  await prisma.modelChangelog.deleteMany();
  await prisma.sourceReference.deleteMany();
  await prisma.model.deleteMany();
  await prisma.provider.deleteMany();

  for (const slug of categories) {
    await prisma.category.upsert({ where: { slug }, create: { slug, name: slug }, update: {} });
  }
  for (const slug of capabilities) {
    await prisma.capability.upsert({ where: { slug }, create: { slug, name: slug }, update: {} });
  }
  for (const slug of params) {
    await prisma.supportedParameter.upsert({ where: { slug }, create: { slug, name: slug }, update: {} });
  }
  for (const slug of modalities) {
    await prisma.modality.upsert({ where: { slug }, create: { slug, name: slug, directionType: 'both' }, update: {} });
  }
  for (const slug of regions) {
    await prisma.region.upsert({ where: { slug }, create: { slug, name: slug.toUpperCase() }, update: {} });
  }

  for (const p of providers) {
    await prisma.provider.create({
      data: {
        slug: p,
        name: p[0].toUpperCase() + p.slice(1),
        status: 'active',
        supportsByok: true,
        supportsRegionalRouting: true,
        description: `Sample provider ${p}`,
      },
    });
  }

  const providerRows = await prisma.provider.findMany();
  const modalityRows = await prisma.modality.findMany();
  const paramRows = await prisma.supportedParameter.findMany();
  const capabilityRows = await prisma.capability.findMany();
  const regionRows = await prisma.region.findMany();
  const categoryRows = await prisma.category.findMany();

  let i = 0;
  for (const provider of providerRows) {
    for (let k = 1; k <= 3; k++) {
      i++;
      const slug = `model-${k}`;
      const canonicalId = `${provider.slug}/${slug}`;
      const model = await prisma.model.create({
        data: {
          providerId: provider.id,
          canonicalId,
          slug,
          displayName: `${provider.name} Model ${k}`,
          shortDescription: 'Sample/demo model metadata for local development.',
          longDescription: 'Sample/demo model. Prices and benchmark values are illustrative and not guaranteed current.',
          status: k === 3 ? 'preview' : 'active',
          lifecycleStatus: 'published',
          contextWindow: 32000 * k,
          maxOutputTokens: 4096 * k,
          openSource: provider.slug === 'meta' || provider.slug === 'qwen',
          free: k === 1,
          categoryId: categoryRows[i % categoryRows.length].id,
          popularityScore: 100 - i,
          releaseDate: new Date(2025, (i % 12), 1),
        },
      });

      await prisma.modelPricing.create({
        data: {
          modelId: model.id,
          currency: 'USD',
          inputPricePerMillionTokens: k === 1 ? 0 : Number((0.2 * k).toFixed(2)),
          outputPricePerMillionTokens: k === 1 ? 0 : Number((0.8 * k).toFixed(2)),
          cachedInputPricePerMillionTokens: k === 1 ? 0 : Number((0.05 * k).toFixed(2)),
          regionId: regionRows[i % regionRows.length].id,
          effectiveFrom: new Date(2026, 0, 1),
          sourceUrl: 'https://example.com/pricing',
        },
      });

      for (const m of modalityRows.slice(0, 2)) {
        await prisma.modelInputModality.create({ data: { modelId: model.id, modalityId: m.id } });
      }
      await prisma.modelOutputModality.create({ data: { modelId: model.id, modalityId: modalityRows[0].id } });
      for (const c of capabilityRows.slice(0, 4)) {
        await prisma.modelCapability.create({ data: { modelId: model.id, capabilityId: c.id } });
      }
      for (const p of paramRows.slice(0, 5)) {
        await prisma.modelSupportedParameter.create({ data: { modelId: model.id, supportedParameterId: p.id } });
      }
      for (const r of regionRows.slice(0, 3)) {
        await prisma.modelRegion.create({ data: { modelId: model.id, regionId: r.id } });
      }
      await prisma.sourceReference.create({ data: { modelId: model.id, title: 'Sample source', url: 'https://example.com' } });
      await prisma.modelChangelog.create({ data: { modelId: model.id, changeType: 'seed', title: 'Seeded model data' } });
    }
  }
}

main().finally(async () => {
  await prisma.$disconnect();
});


