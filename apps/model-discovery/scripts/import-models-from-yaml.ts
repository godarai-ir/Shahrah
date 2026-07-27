import 'dotenv/config';
import { PrismaPg } from '@prisma/adapter-pg';
import { PrismaClient } from '../src/generated/prisma/client';
import { parse } from 'yaml';
import { existsSync } from 'node:fs';
import { promises as fs } from 'node:fs';
import { basename, join } from 'node:path';
import { parsePricingProfiles } from '../src/pricing/pricing-parser';
import { evaluatePublicationReadiness } from '../src/publication/publication-policy';
import {
  ParsedPricingProfile,
  PricingDiagnostic,
  PricingStatus,
} from '../src/pricing/pricing.types';

type StringRecord = Record<string, unknown>;

type ModelMode =
  | 'audio_transcription'
  | 'audio_translation'
  | 'chat'
  | 'completion'
  | 'embedding'
  | 'image'
  | 'moderation'
  | 'realtime'
  | 'rerank'
  | 'responses'
  | 'text_to_speech'
  | 'video'
  | 'unknown';

type ImportModelRow = {
  canonicalId: string;
  slug: string;
  displayName: string;
  status: 'active' | 'preview' | 'deprecated' | 'unavailable';
  lifecycleStatus: 'draft' | 'review' | 'published' | 'deprecated' | 'archived';
  mode: ModelMode[];
  shortDescription?: string;
  contextWindow?: number;
  maxOutputTokens?: number;
  releaseDate?: Date;
  deprecatedAt?: Date;
  officialDocsUrl?: string;
  pricingStatus: PricingStatus;
  openSource: boolean;
  capabilities: string[];
  inputModalities: string[];
  outputModalities: string[];
  supportedParameters: string[];
  regions: string[];
  sources: string[];
  pricingProfiles: ParsedPricingProfile[];
};

type ImportProviderBatch = {
  slug: string;
  name: string;
  docsUrl?: string;
  websiteUrl?: string;
  models: ImportModelRow[];
  diagnostics: Array<
    PricingDiagnostic & {
      filePath: string;
      modelId: string;
    }
  >;
};

const connectionString =
  process.env.DATABASE_URL ??
  'postgresql://postgres:postgres@127.0.0.1:5432/model_discovery?schema=public';

const prisma = new PrismaClient({
  adapter: new PrismaPg({ connectionString }),
});

const knownModes = new Set<ModelMode>([
  'audio_transcription',
  'audio_translation',
  'chat',
  'completion',
  'embedding',
  'image',
  'moderation',
  'realtime',
  'rerank',
  'responses',
  'text_to_speech',
  'video',
  'unknown',
]);

const supportedInputModalities = new Set([
  'audio',
  'code',
  'doc',
  'image',
  'pdf',
  'text',
  'video',
]);

const supportedOutputModalities = new Set([
  'audio',
  'code',
  'embedding',
  'image',
  'text',
  'video',
]);

const supportedCapabilities = new Set([
  'assistant_prefill',
  'cache_control',
  'code_execution',
  'function_calling',
  'json_output',
  'parallel_function_calling',
  'prompt_caching',
  'structured_output',
  'system_messages',
  'tool_choice',
]);

function getArgValue(flag: string): string | undefined {
  const idx = process.argv.findIndex((x) => x === flag);
  if (idx < 0) return undefined;
  return process.argv[idx + 1];
}

function hasFlag(flag: string): boolean {
  return process.argv.includes(flag);
}

function slugify(input: string): string {
  return input
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

function normalizeTaxonomyValue(input: string): string {
  return input
    .trim()
    .toLowerCase()
    .replace(/[\s-]+/g, '_')
    .replace(/[^a-z0-9_]+/g, '')
    .replace(/^_+|_+$/g, '');
}

function titleCaseFromSlug(input: string): string {
  return input
    .replace(/[_\-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .split(' ')
    .map((p) => p.charAt(0).toUpperCase() + p.slice(1))
    .join(' ');
}

function asObject(input: unknown): StringRecord | undefined {
  if (!input || typeof input !== 'object' || Array.isArray(input)) return undefined;
  return input as StringRecord;
}

function asStringArray(input: unknown): string[] {
  if (!Array.isArray(input)) return [];
  return input.filter((x): x is string => typeof x === 'string');
}

function asNumber(input: unknown): number | undefined {
  if (typeof input === 'number' && Number.isFinite(input)) return input;
  if (typeof input === 'string' && input.trim() !== '') {
    const n = Number(input);
    return Number.isFinite(n) ? n : undefined;
  }
  return undefined;
}

function toDate(input: unknown): Date | undefined {
  if (typeof input !== 'string' || input.trim() === '') return undefined;
  const dt = new Date(input);
  if (Number.isNaN(dt.getTime())) return undefined;
  return dt;
}

function inferDateFromName(name: string): Date | undefined {
  const m = name.match(/(20\d{2})-(\d{2})-(\d{2})/);
  if (!m) return undefined;
  const dt = new Date(`${m[1]}-${m[2]}-${m[3]}T00:00:00.000Z`);
  if (Number.isNaN(dt.getTime())) return undefined;
  return dt;
}

function mapModelStatus(rawStatus: unknown): 'active' | 'preview' | 'deprecated' | 'unavailable' {
  const normalized = typeof rawStatus === 'string' ? rawStatus.toLowerCase() : '';
  if (normalized === 'active') return 'active';
  if (normalized === 'preview') return 'preview';
  if (normalized === 'deprecated') return 'deprecated';
  if (normalized === 'unavailable') return 'unavailable';
  return 'active';
}

function normalizeModeValue(input: unknown): ModelMode {
  if (typeof input !== 'string' || input.trim() === '') return 'unknown';

  const normalized = input
    .trim()
    .toLowerCase()
    .replace(/[\s-]+/g, '_');

  switch (normalized) {
    case 'completion':
    case 'completions':
      return 'completion';
    case 'chat':
      return 'chat';
    case 'embedding':
    case 'embeddings':
      return 'embedding';
    case 'audio_transcription':
    case 'transcription':
    case 'speech_to_text':
    case 'stt':
      return 'audio_transcription';
    case 'audio_translation':
    case 'translation':
    case 'speech_translation':
      return 'audio_translation';
    case 'image':
    case 'image_generation':
    case 'text_to_image':
    case 'image_edit':
      return 'image';
    case 'video':
    case 'video_generation':
    case 'text_to_video':
    case 'image_to_video':
      return 'video';
    case 'rerank':
    case 'reranking':
      return 'rerank';
    case 'text_to_speech':
    case 'tts':
    case 'audio_generation':
    case 'text_to_audio':
      return 'text_to_speech';
    case 'realtime':
      return 'realtime';
    case 'responses':
    case 'response':
      return 'responses';
    case 'moderation':
    case 'moderations':
      return 'moderation';
    case 'unknown':
      return 'unknown';
    default:
      return 'unknown';
  }
}

function normalizeSupportedModes(...inputs: unknown[]): ModelMode[] {
  const modes = Array.from(
    new Set(
      inputs
        .flatMap((input) =>
          typeof input === 'string' ? [input] : asStringArray(input),
        )
        .map((value) => normalizeModeValue(value))
        .filter((mode) => knownModes.has(mode) && mode !== 'unknown'),
    ),
  );
  return modes.length > 0 ? modes : ['unknown'];
}

function filterSupportedValues(input: string[], allowed: Set<string>): string[] {
  return Array.from(
    new Set(
      input
        .map(normalizeTaxonomyValue)
        .filter((value) => allowed.has(value)),
    ),
  );
}

function mapLifecycle(status: 'active' | 'preview' | 'deprecated' | 'unavailable') {
  if (status === 'deprecated') return 'deprecated';
  if (status === 'unavailable') return 'archived';
  return 'published';
}

function paramsFromDoc(doc: StringRecord | undefined): string[] {
  if (!doc) return [];
  const params = doc['params'];
  if (!Array.isArray(params)) return [];
  const keys: string[] = [];
  for (const item of params) {
    const rec = asObject(item);
    const key = rec?.['key'];
    if (typeof key === 'string' && key.trim() !== '') keys.push(slugify(key));
  }
  return keys;
}

function inferOutputModalities(
  modes: ModelMode[],
  existingOutputModalities: string[],
): string[] {
  if (existingOutputModalities.length > 0) return existingOutputModalities;

  const inferred = new Set<string>();

  for (const signal of modes) {
    switch (signal) {
      case 'chat':
      case 'completion':
      case 'responses':
      case 'unknown':
        inferred.add('text');
        break;
      case 'image':
        inferred.add('image');
        break;
      case 'video':
        inferred.add('video');
        break;
      case 'embedding':
        inferred.add('embedding');
        break;
      case 'text_to_speech':
        inferred.add('audio');
        break;
      case 'audio_transcription':
      case 'audio_translation':
      case 'rerank':
      case 'moderation':
      case 'realtime':
        inferred.add('text');
        break;
      default:
        break;
    }
  }

  return Array.from(inferred);
}

async function parseProviderDir(providerDir: string): Promise<ImportProviderBatch> {
  const providerSlug = slugify(basename(providerDir));
  const defaultPath = join(providerDir, 'default.yaml');
  let defaultDoc: StringRecord | undefined;
  if (existsSync(defaultPath)) {
    const raw = await fs.readFile(defaultPath, 'utf8');
    defaultDoc = asObject(parse(raw));
  }

  const yamlFiles = await collectYamlFiles(providerDir);

  const providerName = titleCaseFromSlug(providerSlug);
  const docsUrl = asStringArray(defaultDoc?.['documentation'])[0];

  const models: ImportModelRow[] = [];
  const diagnostics: ImportProviderBatch['diagnostics'] = [];
  const defaultParams = paramsFromDoc(defaultDoc);

  for (const filePath of yamlFiles) {
    const raw = await fs.readFile(filePath, 'utf8');
    const doc = asObject(parse(raw));
    if (!doc) continue;

    const modelId = typeof doc['model'] === 'string' ? doc['model'] : undefined;
    if (!modelId) continue;
    if (doc['isDeprecated'] === true) continue;

    const status = mapModelStatus(doc['status']);
    let lifecycleStatus = mapLifecycle(status);
    const limits = asObject(doc['limits']);
    const modalities = asObject(doc['modalities']);
    const paramsCombined = new Set<string>([
      ...defaultParams,
      ...paramsFromDoc(doc),
    ]);
    const removeParams = asStringArray(doc['removeParams']).map(slugify);
    for (const rm of removeParams) paramsCombined.delete(rm);

    const features = filterSupportedValues(
      asStringArray(doc['features']),
      supportedCapabilities,
    );
    const inputModalities = filterSupportedValues(
      asStringArray(modalities?.['input']),
      supportedInputModalities,
    );
    const outputModalitiesRaw = filterSupportedValues(
      asStringArray(modalities?.['output']),
      supportedOutputModalities,
    );
    const sources = asStringArray(doc['sources']);

    const pricing = parsePricingProfiles(doc['costs'], {
      sourceUrl: sources[0],
      explicitFree: doc['free'] === true,
    });
    diagnostics.push(
      ...pricing.diagnostics.map((diagnostic) => ({
        ...diagnostic,
        filePath,
        modelId,
      })),
    );
    const regionSet = new Set(
      pricing.profiles.map((profile) => profile.regionSlug),
    );
    if (regionSet.size === 0) regionSet.add('global');

    const contextWindow = asNumber(limits?.['context_window']);
    const maxOutputTokens =
      asNumber(limits?.['max_output_tokens']) ?? asNumber(limits?.['max_tokens']);

    const deprecatedAt = toDate(doc['deprecationDate']);
    const releaseDate =
      inferDateFromName(modelId) ??
      inferDateFromName(basename(filePath, '.yaml'));

    const mode = normalizeSupportedModes(doc['mode'], doc['supportedModes']);
    const outputModalities = inferOutputModalities(mode, outputModalitiesRaw);
    if (mode.includes('image') && lifecycleStatus === 'published') {
      const readiness = evaluatePublicationReadiness({
        mode,
        pricingStatus: pricing.status,
        pricingProfiles: pricing.profiles.map((profile) => ({
          region: profile.regionSlug,
          currency: profile.currency,
          effectiveFrom: profile.effectiveFrom,
          effectiveTo: profile.effectiveTo,
          rates: profile.rates,
        })),
      });
      lifecycleStatus = readiness.eligible ? 'review' : 'draft';
    }

    models.push({
      canonicalId: `${providerSlug}/${modelId}`,
      slug: modelId,
      displayName: titleCaseFromSlug(modelId),
      status,
      lifecycleStatus,
      mode,
      contextWindow: contextWindow ? Math.floor(contextWindow) : undefined,
      maxOutputTokens: maxOutputTokens ? Math.floor(maxOutputTokens) : undefined,
      releaseDate,
      deprecatedAt,
      officialDocsUrl: sources[0],
      pricingStatus: pricing.status,
      openSource: false,
      capabilities: features,
      inputModalities,
      outputModalities,
      supportedParameters: Array.from(paramsCombined),
      regions: Array.from(regionSet),
      sources,
      pricingProfiles: pricing.profiles,
    });
  }

  return {
    slug: providerSlug,
    name: providerName,
    docsUrl,
    models,
    diagnostics,
  };
}

async function collectYamlFiles(dir: string): Promise<string[]> {
  const entries = await fs.readdir(dir, { withFileTypes: true });
  const files: string[] = [];

  for (const entry of entries) {
    const fullPath = join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await collectYamlFiles(fullPath)));
      continue;
    }
    if (!entry.isFile()) continue;
    if (!entry.name.endsWith('.yaml')) continue;
    if (entry.name === 'default.yaml') continue;
    files.push(fullPath);
  }

  return files;
}

async function ensureCapability(slug: string): Promise<string> {
  const row = await prisma.capability.upsert({
    where: { slug },
    create: { slug, name: titleCaseFromSlug(slug) },
    update: {},
  });
  return row.id;
}

async function ensureModality(slug: string): Promise<string> {
  const row = await prisma.modality.upsert({
    where: { slug },
    create: { slug, name: titleCaseFromSlug(slug) },
    update: {},
  });
  return row.id;
}

async function ensureParameter(slug: string): Promise<string> {
  const row = await prisma.supportedParameter.upsert({
    where: { slug },
    create: { slug, name: titleCaseFromSlug(slug) },
    update: {},
  });
  return row.id;
}

async function ensureRegion(slug: string): Promise<string> {
  const row = await prisma.region.upsert({
    where: { slug },
    create: { slug, name: titleCaseFromSlug(slug) },
    update: {},
  });
  return row.id;
}

async function importProviderBatch(
  batch: ImportProviderBatch,
  clearMissing: boolean,
  dryRun: boolean,
  importTimestamp: Date,
): Promise<{ modelsImported: number; modelsDeleted: number }> {
  if (dryRun) {
    return { modelsImported: batch.models.length, modelsDeleted: 0 };
  }

  const provider = await prisma.provider.upsert({
    where: { slug: batch.slug },
    create: {
      slug: batch.slug,
      name: batch.name,
      status: 'active',
      supportsByok: true,
      supportsRegionalRouting: true,
      docsUrl: batch.docsUrl,
      websiteUrl: batch.websiteUrl,
    },
    update: {
      name: batch.name,
      docsUrl: batch.docsUrl,
      websiteUrl: batch.websiteUrl,
      status: 'active',
      supportsByok: true,
      supportsRegionalRouting: true,
    },
  });

  if (clearMissing) {
    const keep = new Set(batch.models.map((m) => m.canonicalId));
    const existing = await prisma.model.findMany({
      where: { providerId: provider.id },
      select: { id: true, canonicalId: true },
    });
    const toDeleteIds = existing.filter((m) => !keep.has(m.canonicalId)).map((m) => m.id);
    if (toDeleteIds.length > 0) {
      await prisma.model.deleteMany({ where: { id: { in: toDeleteIds } } });
    }
  }

  for (const row of batch.models) {
    const existing = await prisma.model.findUnique({
      where: { canonicalId: row.canonicalId },
      include: {
        pricing: { include: { region: true, rates: true } },
      },
    });
    const isImage = row.mode.includes('image');
    let lifecycleStatus = row.lifecycleStatus;
    if (
      existing &&
      row.lifecycleStatus !== 'deprecated' &&
      row.lifecycleStatus !== 'archived'
    ) {
      if (existing.pricingAuthority === 'admin') {
        lifecycleStatus = existing.lifecycleStatus;
      } else if (isImage) {
        if (existing.lifecycleStatus === 'published') {
          lifecycleStatus =
            row.lifecycleStatus === 'review' ? 'published' : 'draft';
        } else {
          lifecycleStatus = existing.lifecycleStatus;
        }
      }
    }
    const preserveAdminPricing = existing?.pricingAuthority === 'admin';
    const model = await prisma.model.upsert({
      where: { canonicalId: row.canonicalId },
      create: {
        providerId: provider.id,
        canonicalId: row.canonicalId,
        slug: row.slug,
        displayName: row.displayName,
        mode: row.mode,
        status: row.status,
        lifecycleStatus,
        shortDescription: row.shortDescription,
        contextWindow: row.contextWindow,
        maxOutputTokens: row.maxOutputTokens,
        releaseDate: row.releaseDate,
        deprecatedAt: row.deprecatedAt,
        officialDocsUrl: row.officialDocsUrl,
        pricingStatus: row.pricingStatus,
        openSource: row.openSource,
      },
      update: {
        providerId: provider.id,
        slug: row.slug,
        displayName: row.displayName,
        mode: row.mode,
        status: row.status,
        lifecycleStatus,
        shortDescription: row.shortDescription,
        contextWindow: row.contextWindow,
        maxOutputTokens: row.maxOutputTokens,
        releaseDate: row.releaseDate,
        deprecatedAt: row.deprecatedAt,
        officialDocsUrl: row.officialDocsUrl,
        pricingStatus: preserveAdminPricing ? undefined : row.pricingStatus,
        openSource: row.openSource,
      },
    });

    await prisma.$transaction([
      ...(preserveAdminPricing
        ? []
        : [prisma.modelPricing.deleteMany({ where: { modelId: model.id } })]),
      prisma.modelCapability.deleteMany({ where: { modelId: model.id } }),
      prisma.modelInputModality.deleteMany({ where: { modelId: model.id } }),
      prisma.modelOutputModality.deleteMany({ where: { modelId: model.id } }),
      prisma.modelSupportedParameter.deleteMany({ where: { modelId: model.id } }),
      prisma.modelRegion.deleteMany({ where: { modelId: model.id } }),
      prisma.sourceReference.deleteMany({ where: { modelId: model.id } }),
    ]);

    for (const src of row.sources) {
      await prisma.sourceReference.create({
        data: {
          modelId: model.id,
          title: 'Source',
          url: src,
          sourceType: 'provider_yaml',
        },
      });
    }

    for (const pricing of preserveAdminPricing ? [] : row.pricingProfiles) {
      const regionId = await ensureRegion(pricing.regionSlug);
      await prisma.modelPricing.create({
        data: {
          modelId: model.id,
          regionId,
          currency: pricing.currency,
          sourceUrl: pricing.sourceUrl,
          effectiveFrom: importTimestamp,
          rates: {
            create: pricing.rates.map((rate) => ({
              rateKey: rate.rateKey,
              usage: rate.usage,
              unit: rate.unit,
              amount: rate.amount,
              quantity: rate.quantity,
              serviceTier: rate.serviceTier,
              resolution: rate.resolution,
              tierFrom: rate.tierFrom,
              tierTo: rate.tierTo,
              tierMode: rate.tierMode,
              qualifiers: rate.qualifiers,
            })),
          },
        },
      });
    }

    for (const capabilitySlug of row.capabilities) {
      const capabilityId = await ensureCapability(capabilitySlug);
      await prisma.modelCapability.create({
        data: { modelId: model.id, capabilityId },
      });
    }

    for (const modalitySlug of row.inputModalities) {
      const modalityId = await ensureModality(modalitySlug);
      await prisma.modelInputModality.create({
        data: { modelId: model.id, modalityId },
      });
    }

    for (const modalitySlug of row.outputModalities) {
      const modalityId = await ensureModality(modalitySlug);
      await prisma.modelOutputModality.create({
        data: { modelId: model.id, modalityId },
      });
    }

    for (const parameterSlug of row.supportedParameters) {
      const supportedParameterId = await ensureParameter(parameterSlug);
      await prisma.modelSupportedParameter.create({
        data: { modelId: model.id, supportedParameterId },
      });
    }

    for (const regionSlug of row.regions) {
      const regionId = await ensureRegion(regionSlug);
      await prisma.modelRegion.create({
        data: { modelId: model.id, regionId },
      });
    }

    await prisma.modelChangelog.create({
      data: {
        modelId: model.id,
        changeType: 'import',
        title: 'Imported from provider YAML',
        createdBy: 'system-importer',
      },
    });
  }

  return { modelsImported: batch.models.length, modelsDeleted: 0 };
}

async function main() {
  const providersRoot =
    getArgValue('--providers-root') ?? 'D:\\source\\models\\providers';
  const providersArg = getArgValue('--providers');
  const dryRun = hasFlag('--dry-run');
  const clearMissing = hasFlag('--clear-missing');
  const allowUnknownPricingKeys = hasFlag('--allow-unknown-pricing-keys');
  const importTimestamp = new Date();

  const selected = new Set(
    (providersArg ?? '')
      .split(',')
      .map((x) => slugify(x))
      .filter(Boolean),
  );

  const dirEntries = await fs.readdir(providersRoot, { withFileTypes: true });
  const providerDirs = dirEntries
    .filter((d) => d.isDirectory())
    .map((d) => join(providersRoot, d.name))
    .filter((dirPath) =>
      selected.size === 0 ? true : selected.has(slugify(basename(dirPath))),
    );

  if (providerDirs.length === 0) {
    throw new Error(`No provider directories found under: ${providersRoot}`);
  }

  let totalProviders = 0;
  let totalModels = 0;
  let totalProfiles = 0;
  let totalRates = 0;
  let totalWarnings = 0;
  const batches = await Promise.all(
    providerDirs.map((providerDir) => parseProviderDir(providerDir)),
  );

  for (const batch of batches) {
    const errors = batch.diagnostics.filter(
      (diagnostic) => diagnostic.severity === 'error',
    );
    const warnings = batch.diagnostics.filter(
      (diagnostic) => diagnostic.severity === 'warning',
    );
    totalWarnings += warnings.length;
    for (const diagnostic of batch.diagnostics) {
      console.error(
        `[pricing:${diagnostic.severity}] provider=${batch.slug} model=${diagnostic.modelId} region=${diagnostic.regionSlug} code=${diagnostic.code} key=${diagnostic.key ?? '-'} file=${diagnostic.filePath} message=${diagnostic.message}`,
      );
    }
    const blockingErrors = errors.filter(
      (diagnostic) =>
        diagnostic.code !== 'unknown_pricing_key' ||
        !allowUnknownPricingKeys,
    );
    if (blockingErrors.length > 0) {
      throw new Error(
        `Pricing validation failed for provider ${batch.slug} with ${blockingErrors.length} blocking error(s).`,
      );
    }
  }

  for (const batch of batches) {
    const warnings = batch.diagnostics.filter(
      (diagnostic) => diagnostic.severity === 'warning',
    );
    const result = await importProviderBatch(
      batch,
      clearMissing,
      dryRun,
      importTimestamp,
    );
    totalProviders += 1;
    totalModels += result.modelsImported;
    const profiles = batch.models.flatMap((model) => model.pricingProfiles);
    totalProfiles += profiles.length;
    totalRates += profiles.reduce(
      (count, profile) => count + profile.rates.length,
      0,
    );
    console.log(
      `[import] provider=${batch.slug} models=${result.modelsImported} profiles=${profiles.length} rates=${profiles.reduce((count, profile) => count + profile.rates.length, 0)} warnings=${warnings.length} dryRun=${dryRun}`,
    );
  }

  console.log(
    `[import] completed providers=${totalProviders} models=${totalModels} profiles=${totalProfiles} rates=${totalRates} warnings=${totalWarnings} dryRun=${dryRun}`,
  );
}

main()
  .catch((err) => {
    console.error(err);
    process.exitCode = 1;
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
