import 'dotenv/config';
import { PrismaPg } from '@prisma/adapter-pg';
import { Prisma, PrismaClient } from '../src/generated/prisma/client';
import { parse } from 'yaml';
import { existsSync } from 'node:fs';
import { promises as fs } from 'node:fs';
import { basename, dirname, join } from 'node:path';
import {
  modelRuntimeProfileSchema,
  type ModelRuntimeProfile,
} from '../src/catalog/runtime-profile.schema';
import { findMissingProviderSlugs } from '../src/catalog/provider-catalog-reconciliation';
import {
  assertRuntimeProfileVersionIsImmutable,
  stableJson,
} from '../src/catalog/runtime-profile-immutability';
import { anthropicProviderImporter } from './model-import/providers/anthropic';
import { googleGeminiProviderImporter } from './model-import/providers/google-gemini';
import {
  OPENAI_PRIORITY_LONG_CONTEXT_RATE_MAPPINGS,
  OPENAI_PROVIDER_SLUG,
  validateOpenAIProviderDirectory,
} from './model-import/providers/openai';
import {
  mistralAIProviderImporter,
  validateMistralAIProviderDirectory,
} from './model-import/providers/mistral-ai';
import {
  validateXAIProviderDirectory,
  xaiProviderImporter,
} from './model-import/providers/xai';
import {
  MOONSHOT_PROVIDER_SLUG,
  validateMoonshotProviderDirectory,
} from './model-import/providers/moonshot';
import {
  DEEPSEEK_PROVIDER_SLUG,
  deepSeekProviderImporter,
  validateDeepSeekProviderDirectory,
} from './model-import/providers/deepseek';
import {
  ZAI_PROVIDER_SLUG,
  validateZAIProviderDirectory,
  zaiProviderImporter,
} from './model-import/providers/zai';
import type {
  ProviderPricingRate,
  ProviderRateMapping,
} from './model-import/types';

export const DEFAULT_PROVIDERS_ROOT = join(__dirname, '..', 'providers');

type StringRecord = Record<string, unknown>;

type ModelMode =
  | 'audio_transcription'
  | 'audio_translation'
  | 'chat'
  | 'completion'
  | 'embedding'
  | 'generate_content'
  | 'image'
  | 'moderation'
  | 'ocr'
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
  deprecationDate?: Date;
  retirementDate?: Date;
  replacementCanonicalId?: string;
  officialDocsUrl?: string;
  free: boolean;
  openSource: boolean;
  capabilities: string[];
  inputModalities: string[];
  outputModalities: string[];
  supportedParameters: string[];
  regions: string[];
  sources: string[];
  runtimeProfile?: ModelRuntimeProfile;
  pricingRates: ProviderPricingRate[];
  pricingRows: Array<{
    regionSlug: string;
    sourceUrl?: string;
    effectiveFrom?: Date;
    effectiveTo?: Date;
    inputPricePerMillionTokens?: number;
    outputPricePerMillionTokens?: number;
    cachedInputPricePerMillionTokens?: number;
    imageInputPrice?: number;
    imageOutputPrice?: number;
    audioInputPrice?: number;
    audioOutputPrice?: number;
    requestPrice?: number;
  }>;
};

type ImportProviderBatch = {
  slug: string;
  name: string;
  docsUrl?: string;
  websiteUrl?: string;
  models: ImportModelRow[];
};

export type ProviderYamlImportOptions = {
  providersRoot: string;
  providerSlugs?: string[];
  dryRun?: boolean;
  clearMissing?: boolean;
  clearMissingProviders?: boolean;
  log?: (message: string) => void;
};

export type ProviderYamlImportSummary = {
  providersImported: number;
  providersDeleted: number;
  modelsImported: number;
  modelsDeleted: number;
  dryRun: boolean;
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
  'generate_content',
  'image',
  'moderation',
  'realtime',
  'rerank',
  'responses',
  'text_to_speech',
  'video',
  'unknown',
  'ocr',
]);

const supportedInputModalities = new Set([
  'audio',
  'code',
  'doc',
  'image',
  'pdf',
  'text',
  'video',
  'document',
  'moderation_result',
]);

const supportedOutputModalities = new Set([
  'audio',
  'code',
  'embedding',
  'image',
  'text',
  'video',
  'document',
]);

const supportedCapabilities = new Set([
  'assistant_prefill',
  'cache_control',
  'code_execution',
  'fine_tuning',
  'function_calling',
  'json_output',
  'parallel_function_calling',
  'predicted_outputs',
  'prompt_caching',
  'structured_output',
  'system_messages',
  'tool_choice',
  'agents',
  'annotations',
  'bbox_extraction',
  'built_in_tools',
  'document_qna',
  'timestamps',
  'voice_cloning',
]);

function getArgValue(
  flag: string,
  args: string[] = process.argv.slice(2),
): string | undefined {
  const idx = args.findIndex((x) => x === flag);
  if (idx < 0) return undefined;
  return args[idx + 1];
}

function hasFlag(
  flag: string,
  args: string[] = process.argv.slice(2),
): boolean {
  return args.includes(flag);
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
  if (!input || typeof input !== 'object' || Array.isArray(input))
    return undefined;
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

function asDecimalString(input: unknown): string | undefined {
  if (typeof input === 'number' && Number.isFinite(input)) return String(input);
  if (typeof input === 'string' && input.trim() !== '') {
    const value = Number(input);
    return Number.isFinite(value) ? input.trim() : undefined;
  }
  return undefined;
}

function divideDecimalString(
  input: unknown,
  divisor: number,
): string | undefined {
  const value = asNumber(input);
  return value === undefined ? undefined : String(value / divisor);
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

function mapModelStatus(
  rawStatus: unknown,
): 'active' | 'preview' | 'deprecated' | 'unavailable' {
  const normalized =
    typeof rawStatus === 'string' ? rawStatus.toLowerCase() : '';
  if (normalized === 'active') return 'active';
  if (normalized === 'preview') return 'preview';
  if (normalized === 'deprecated') return 'deprecated';
  if (normalized === 'retired') return 'unavailable';
  if (normalized === 'unavailable') return 'unavailable';
  return 'unavailable';
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
    case 'generate_content':
    case 'generatecontent':
      return 'generate_content';
    case 'moderation':
    case 'moderations':
      return 'moderation';
    case 'ocr':
    case 'document_ocr':
      return 'ocr';
    case 'unknown':
      return 'unknown';
    default:
      return 'unknown';
  }
}

function normalizeSupportedModes(input: unknown): ModelMode[] {
  const modes = Array.from(
    new Set(
      asStringArray(input)
        .map((value) => normalizeModeValue(value))
        .filter((mode) => knownModes.has(mode) && mode !== 'unknown'),
    ),
  );
  return modes.length > 0 ? modes : ['unknown'];
}

function filterSupportedValues(
  input: string[],
  allowed: Set<string>,
): string[] {
  return Array.from(
    new Set(
      input.map(normalizeTaxonomyValue).filter((value) => allowed.has(value)),
    ),
  );
}

function mapLifecycle(
  status: 'active' | 'preview' | 'deprecated' | 'unavailable',
) {
  if (status === 'deprecated') return 'deprecated';
  if (status === 'unavailable') return 'archived';
  return 'published';
}

function enforceOpenAIReleaseStatus(
  provider: string,
  modelId: string,
  status: 'active' | 'preview' | 'deprecated' | 'unavailable',
): 'active' | 'preview' | 'deprecated' | 'unavailable' {
  if (provider !== 'openai') return status;
  const id = modelId.toLowerCase().replace(/^openai\//, '');
  if (id === 'chat-latest' || id === 'chatgpt-image-latest') {
    return 'unavailable';
  }
  if (id.includes('preview')) return 'preview';
  return status;
}

// Modes that name a gateway surface must be backed by the runtime endpoint
// that serves it, so the catalog never advertises a route the model cannot
// take (for example, Gemini models have no OpenAI Responses endpoint).
const modeRequiredRuntimeEndpoints: Partial<
  Record<ModelMode, ModelRuntimeProfile['endpoints']>
> = {
  responses: ['responses'],
  generate_content: ['generate_content'],
};

function assertModesMatchRuntimeEndpoints(
  modelId: string,
  modes: ModelMode[],
  runtimeProfile: ModelRuntimeProfile | undefined,
): void {
  if (!runtimeProfile) return;
  for (const mode of modes) {
    const missing = (modeRequiredRuntimeEndpoints[mode] ?? []).filter(
      (endpoint) => !runtimeProfile.endpoints.includes(endpoint),
    );
    if (missing.length > 0) {
      throw new Error(
        `${modelId}: supportedModes includes ${mode}, but runtimeProfile.endpoints is missing ${missing.join(', ')}`,
      );
    }
  }
}

function enforceOpenAIModes(
  provider: string,
  modelId: string,
  modes: ModelMode[],
): ModelMode[] {
  if (provider !== 'openai' || !modes.includes('unknown')) return modes;
  const id = modelId.toLowerCase().replace(/^openai\//, '');
  if (id === 'gpt-3.5-turbo-instruct') return ['completion'];
  if (id.startsWith('sora-')) return ['video'];
  return modes;
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

function supportedValuesForParam(
  doc: StringRecord | undefined,
  parameter: string,
): string[] {
  if (!doc || !Array.isArray(doc['params'])) return [];
  for (const item of doc['params']) {
    const record = asObject(item);
    if (
      typeof record?.['key'] === 'string' &&
      normalizeTaxonomyValue(record['key']) ===
        normalizeTaxonomyValue(parameter)
    ) {
      return asStringArray(record['supportedValues']).map(
        normalizeTaxonomyValue,
      );
    }
  }
  return [];
}

function normalizeRegion(raw: string): string {
  const val = raw.trim();
  if (val === '*' || val === 'global') return 'global';
  return slugify(val);
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
      case 'generate_content':
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
      case 'ocr':
        inferred.add('document');
        break;
      default:
        break;
    }
  }

  return Array.from(inferred);
}

function perTokenToPerMillion(input: number | undefined): number | undefined {
  if (input === undefined) return undefined;
  return input * 1_000_000;
}

async function parseProviderDir(
  providerDir: string,
): Promise<ImportProviderBatch> {
  const providerSlug = slugify(basename(providerDir));
  if (providerSlug === OPENAI_PROVIDER_SLUG) {
    await validateOpenAIProviderDirectory(dirname(providerDir));
  }
  if (providerSlug === 'mistral-ai') {
    await validateMistralAIProviderDirectory(join(providerDir, '..'));
  }
  const providerImporter =
    providerSlug === anthropicProviderImporter.slug
      ? anthropicProviderImporter
      : providerSlug === googleGeminiProviderImporter.slug
        ? googleGeminiProviderImporter
        : providerSlug === xaiProviderImporter.slug
          ? xaiProviderImporter
          : providerSlug === mistralAIProviderImporter.slug
            ? mistralAIProviderImporter
            : providerSlug === deepSeekProviderImporter.slug
              ? deepSeekProviderImporter
              : providerSlug === zaiProviderImporter.slug
                ? zaiProviderImporter
                : undefined;
  if (providerSlug === xaiProviderImporter.slug) {
    await validateXAIProviderDirectory(dirname(providerDir));
  }
  if (providerSlug === MOONSHOT_PROVIDER_SLUG) {
    await validateMoonshotProviderDirectory(dirname(providerDir));
  }
  if (providerSlug === DEEPSEEK_PROVIDER_SLUG) {
    await validateDeepSeekProviderDirectory(dirname(providerDir));
  }
  if (providerSlug === ZAI_PROVIDER_SLUG) {
    await validateZAIProviderDirectory(dirname(providerDir));
  }
  const defaultPath = join(providerDir, 'default.yaml');
  let defaultDoc: StringRecord | undefined;
  if (existsSync(defaultPath)) {
    const raw = await fs.readFile(defaultPath, 'utf8');
    defaultDoc = asObject(parse(raw));
  }

  const yamlFiles = await collectYamlFiles(providerDir);

  const providerName =
    typeof defaultDoc?.['name'] === 'string' && defaultDoc['name'].trim() !== ''
      ? defaultDoc['name'].trim()
      : titleCaseFromSlug(providerSlug);
  const docsUrl = asStringArray(defaultDoc?.['documentation'])[0];
  const websiteUrl =
    typeof defaultDoc?.['website'] === 'string'
      ? defaultDoc['website'].trim()
      : undefined;
  const providerSources = asStringArray(defaultDoc?.['documentation']);
  const defaultToolPricing = asObject(defaultDoc?.['tool_pricing']);

  const models: ImportModelRow[] = [];
  const defaultParams = paramsFromDoc(defaultDoc);

  for (const filePath of yamlFiles) {
    const raw = await fs.readFile(filePath, 'utf8');
    const doc = asObject(parse(raw));
    if (!doc) continue;

    const modelId = typeof doc['model'] === 'string' ? doc['model'] : undefined;
    if (!modelId) continue;
    const deprecationDate = toDate(doc['deprecationDate']);
    const retirementDate = toDate(doc['retirementDate']);
    const now = new Date();
    const importedStatus =
      retirementDate && retirementDate <= now
        ? 'unavailable'
        : deprecationDate && deprecationDate <= now
          ? 'deprecated'
          : mapModelStatus(doc['status']);
    const status = enforceOpenAIReleaseStatus(
      providerSlug,
      modelId,
      importedStatus,
    );
    const lifecycleStatus = mapLifecycle(status);
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
    const sources = Array.from(
      new Set([...asStringArray(doc['sources']), ...providerSources]),
    );
    const runtimeProfileValue = doc['runtimeProfile'];
    const explicitRuntimeProfile =
      runtimeProfileValue === undefined
        ? undefined
        : modelRuntimeProfileSchema.parse(runtimeProfileValue);

    const costs = Array.isArray(doc['costs']) ? doc['costs'] : [];
    const pricingRows: ImportModelRow['pricingRows'] = [];
    const pricingRates: ImportModelRow['pricingRates'] = [];
    const regionSet = new Set<string>();
    for (const item of costs) {
      const cost = asObject(item);
      if (!cost) continue;
      const region =
        typeof cost['region'] === 'string'
          ? normalizeRegion(cost['region'])
          : 'global';
      const effectiveFrom = toDate(cost['effective_from']);
      const effectiveTo = toDate(cost['effective_to']);
      if (cost['effective_from'] !== undefined && !effectiveFrom) {
        throw new Error(`${modelId}: invalid pricing effective_from`);
      }
      if (cost['effective_to'] !== undefined && !effectiveTo) {
        throw new Error(`${modelId}: invalid pricing effective_to`);
      }
      if (
        effectiveFrom &&
        effectiveTo &&
        effectiveFrom.getTime() >= effectiveTo.getTime()
      ) {
        throw new Error(
          `${modelId}: pricing effective_from must precede effective_to`,
        );
      }
      regionSet.add(region);
      const sourceUrl = sources.find((source) => source.includes('pricing'));
      if (!sourceUrl) {
        throw new Error(
          `${modelId}: pricing data requires an official pricing source`,
        );
      }
      pricingRows.push({
        regionSlug: region,
        sourceUrl,
        effectiveFrom,
        effectiveTo,
        inputPricePerMillionTokens: perTokenToPerMillion(
          asNumber(cost['input_cost_per_token']),
        ),
        outputPricePerMillionTokens: perTokenToPerMillion(
          asNumber(cost['output_cost_per_token']),
        ),
        cachedInputPricePerMillionTokens: perTokenToPerMillion(
          asNumber(cost['cache_read_input_token_cost']),
        ),
        imageInputPrice: asNumber(cost['input_cost_per_image']),
        imageOutputPrice: asNumber(cost['output_cost_per_image']),
        audioInputPrice: asNumber(cost['input_cost_per_audio_token']),
        audioOutputPrice: asNumber(cost['output_cost_per_audio_token']),
        requestPrice: asNumber(cost['request_cost']),
      });

      const rateMappings: ProviderRateMapping[] = [
        {
          key: 'input_cost_per_token',
          dimension: 'input_tokens',
          unit: 'token',
        },
        {
          key: 'output_cost_per_token',
          dimension: 'output_tokens',
          unit: 'token',
        },
        {
          key: 'cache_creation_input_token_cost',
          dimension:
            providerSlug === 'openai'
              ? 'cache_write_tokens'
              : 'cache_write_5m_tokens',
          unit: 'token',
        },
        {
          key: 'cache_creation_input_token_cost_per_hour',
          dimension: 'cache_write_1h_tokens',
          unit: 'token',
        },
        {
          key: 'cache_read_input_token_cost',
          dimension: 'cache_read_tokens',
          unit: 'token',
        },
        {
          key: 'cache_read_input_token_cost_batches',
          dimension: 'batch_cache_read_tokens',
          unit: 'token',
          conditionJson: providerImporter?.batchPricingCondition ?? {
            service_tier: 'batch',
          },
        },
        {
          key: 'cache_creation_input_token_cost_batches',
          dimension: 'batch_cache_write_5m_tokens',
          unit: 'token',
          conditionJson: providerImporter?.batchPricingCondition ?? {
            service_tier: 'batch',
          },
        },
        {
          key: 'cache_creation_input_token_cost_per_hour_batches',
          dimension: 'batch_cache_write_1h_tokens',
          unit: 'token',
          conditionJson: providerImporter?.batchPricingCondition ?? {
            service_tier: 'batch',
          },
        },
        {
          key: 'input_cost_per_token_batches',
          dimension: 'batch_input_tokens',
          unit: 'token',
          conditionJson:
            providerImporter?.batchPricingCondition ??
            (providerSlug === 'google-gemini'
              ? { api: 'batches' }
              : { service_tier: 'batch' }),
        },
        {
          key: 'output_cost_per_token_batches',
          dimension: 'batch_output_tokens',
          unit: 'token',
          conditionJson:
            providerImporter?.batchPricingCondition ??
            (providerSlug === 'google-gemini'
              ? { api: 'batches' }
              : { service_tier: 'batch' }),
        },
        {
          key: 'input_cost_per_image',
          dimension: 'input_images',
          unit: 'image',
        },
        {
          key: 'output_cost_per_image',
          dimension: 'output_images',
          unit: 'image',
        },
        {
          key: 'input_cost_per_audio_token',
          dimension: 'audio_input_tokens',
          unit: 'token',
        },
        {
          key: 'output_cost_per_audio_token',
          dimension: 'audio_output_tokens',
          unit: 'token',
        },
        {
          key: 'input_cost_per_character',
          dimension: 'speech_input_characters',
          unit: 'character',
        },
        {
          key: 'input_cost_per_minute',
          dimension: 'audio_input_minutes',
          unit: 'minute',
        },
        {
          key: 'input_cost_per_image_token',
          dimension: 'image_input_tokens',
          unit: 'token',
        },
        {
          key: 'cache_read_input_image_token_cost',
          dimension: 'image_cache_read_tokens',
          unit: 'token',
        },
        {
          key: 'output_cost_per_image_token',
          dimension: 'image_output_tokens',
          unit: 'token',
        },
        {
          key: 'input_cost_per_image_token_batches',
          dimension: 'batch_image_input_tokens',
          unit: 'token',
          conditionJson: { service_tier: 'batch' },
        },
        {
          key: 'cache_read_input_image_token_cost_batches',
          dimension: 'batch_image_cache_read_tokens',
          unit: 'token',
          conditionJson: { service_tier: 'batch' },
        },
        {
          key: 'output_cost_per_image_token_batches',
          dimension: 'batch_image_output_tokens',
          unit: 'token',
          conditionJson: { service_tier: 'batch' },
        },
        {
          key: 'request_cost',
          dimension: 'requests',
          unit: 'request',
        },
        {
          key: 'output_cost_per_second',
          dimension: 'output_video_seconds',
          unit: 'second',
        },
        {
          key: 'cache_storage_cost_per_token_per_hour',
          dimension: 'cache_storage_token_hours',
          unit: 'token_hour',
        },
        {
          key: 'input_cost_per_token_above_200k_tokens',
          dimension: 'long_context_input_tokens',
          unit: 'token',
          conditionJson: {
            min_prompt_tokens:
              providerSlug === 'google-gemini' ? 200001 : 272001,
          },
        },
        {
          key: 'output_cost_per_token_above_200k_tokens',
          dimension: 'long_context_output_tokens',
          unit: 'token',
          conditionJson: {
            min_prompt_tokens:
              providerSlug === 'google-gemini' ? 200001 : 272001,
          },
        },
        {
          key: 'cache_read_input_token_cost_above_200k_tokens',
          dimension: 'long_context_cache_read_tokens',
          unit: 'token',
          conditionJson: {
            min_prompt_tokens:
              providerSlug === 'google-gemini' ? 200001 : 272001,
          },
        },
        {
          key: 'input_cost_per_token_flex',
          dimension: 'flex_input_tokens',
          unit: 'token',
          conditionJson: { service_tier: 'flex' },
        },
        {
          key: 'output_cost_per_token_flex',
          dimension: 'flex_output_tokens',
          unit: 'token',
          conditionJson: { service_tier: 'flex' },
        },
        {
          key: 'cache_read_input_token_cost_flex',
          dimension: 'flex_cache_read_tokens',
          unit: 'token',
          conditionJson: { service_tier: 'flex' },
        },
        {
          key: 'cache_creation_input_token_cost_flex',
          dimension: 'flex_cache_write_tokens',
          unit: 'token',
          conditionJson: { service_tier: 'flex' },
        },
        {
          key: 'input_cost_per_token_batches_above_272k_tokens',
          dimension: 'batch_long_context_input_tokens',
          unit: 'token',
          conditionJson: {
            service_tier: 'batch',
            min_prompt_tokens: 272001,
          },
        },
        {
          key: 'output_cost_per_token_batches_above_272k_tokens',
          dimension: 'batch_long_context_output_tokens',
          unit: 'token',
          conditionJson: {
            service_tier: 'batch',
            min_prompt_tokens: 272001,
          },
        },
        {
          key: 'cache_read_input_token_cost_batches_above_272k_tokens',
          dimension: 'batch_long_context_cache_read_tokens',
          unit: 'token',
          conditionJson: {
            service_tier: 'batch',
            min_prompt_tokens: 272001,
          },
        },
        {
          key: 'cache_creation_input_token_cost_batches_above_272k_tokens',
          dimension: 'batch_long_context_cache_write_tokens',
          unit: 'token',
          conditionJson: {
            service_tier: 'batch',
            min_prompt_tokens: 272001,
          },
        },
        {
          key: 'input_cost_per_token_flex_above_272k_tokens',
          dimension: 'flex_long_context_input_tokens',
          unit: 'token',
          conditionJson: {
            service_tier: 'flex',
            min_prompt_tokens: 272001,
          },
        },
        {
          key: 'output_cost_per_token_flex_above_272k_tokens',
          dimension: 'flex_long_context_output_tokens',
          unit: 'token',
          conditionJson: {
            service_tier: 'flex',
            min_prompt_tokens: 272001,
          },
        },
        {
          key: 'cache_read_input_token_cost_flex_above_272k_tokens',
          dimension: 'flex_long_context_cache_read_tokens',
          unit: 'token',
          conditionJson: {
            service_tier: 'flex',
            min_prompt_tokens: 272001,
          },
        },
        {
          key: 'cache_creation_input_token_cost_flex_above_272k_tokens',
          dimension: 'flex_long_context_cache_write_tokens',
          unit: 'token',
          conditionJson: {
            service_tier: 'flex',
            min_prompt_tokens: 272001,
          },
        },
        ...OPENAI_PRIORITY_LONG_CONTEXT_RATE_MAPPINGS,
        ...(providerImporter?.rateMappings ?? []),
      ];
      for (const mapping of rateMappings) {
        const unitPriceUsd = asDecimalString(cost[mapping.key]);
        if (unitPriceUsd === undefined) continue;
        pricingRates.push({
          dimension: mapping.dimension,
          unit: mapping.unit,
          unitPriceUsd,
          conditionJson: mapping.conditionJson,
          region,
          sourceUrl,
          effectiveFrom,
          effectiveTo,
        });
      }
      const priorityPricing = asObject(cost['priority_pricing']);
      for (const [key, dimension] of [
        ['input_cost_per_token', 'speed_input_tokens'],
        ['output_cost_per_token', 'speed_output_tokens'],
        ['cache_read_input_token_cost', 'speed_cache_read_tokens'],
        ['cache_creation_input_token_cost', 'speed_cache_write_tokens'],
      ] as const) {
        const unitPriceUsd = asDecimalString(priorityPricing?.[key]);
        if (unitPriceUsd === undefined) continue;
        pricingRates.push({
          dimension,
          unit: 'token',
          unitPriceUsd,
          conditionJson: { service_tier: 'priority' },
          region,
          sourceUrl,
        });
      }
      const tieredPricing = asObject(cost['tiered_pricing']);
      for (const [key, dimension] of [
        ['input', 'long_context_input_tokens'],
        ['output', 'long_context_output_tokens'],
        ['cache_read', 'long_context_cache_read_tokens'],
        ['cache_write', 'long_context_cache_write_tokens'],
      ] as const) {
        const tiers = tieredPricing?.[key];
        if (!Array.isArray(tiers)) continue;
        const tier = asObject(tiers[0]);
        const threshold = asNumber(tier?.['from']);
        const unitPriceUsd = asDecimalString(tier?.['cost_per_token']);
        if (threshold === undefined || unitPriceUsd === undefined) {
          continue;
        }
        pricingRates.push({
          dimension,
          unit: 'token',
          unitPriceUsd,
          conditionJson: { min_prompt_tokens: threshold + 1 },
          region,
          sourceUrl,
        });
      }
      if (asDecimalString(cost['output_cost_per_second']) === undefined) {
        const unitPriceUsd = asDecimalString(cost['cost_per_second']);
        if (unitPriceUsd !== undefined) {
          pricingRates.push({
            dimension: 'output_video_seconds',
            unit: 'second',
            unitPriceUsd,
            region,
            sourceUrl,
          });
        }
      }
      providerImporter?.appendDerivedCostPricingRates?.({
        cost,
        explicitRuntimeProfile,
        pricingRates,
        region,
        sourceUrl,
      });
    }

    const webSearch = asObject(defaultToolPricing?.['web_search']);
    const webSearchPrice = divideDecimalString(
      webSearch?.['per_thousand_requests'],
      1000,
    );
    const pricingSource = sources.find((source) => source.includes('pricing'));
    const documentedTools = new Set(
      asStringArray(doc['tools']).map(normalizeTaxonomyValue),
    );
    if (
      webSearchPrice &&
      pricingSource &&
      (providerSlug !== 'google-gemini' ||
        documentedTools.has('google_search')) &&
      (providerSlug !== 'xai' || documentedTools.has('web_search'))
    ) {
      pricingRates.push({
        dimension:
          providerSlug === 'google-gemini'
            ? 'google_search_queries'
            : 'web_search_requests',
        unit: providerSlug === 'google-gemini' ? 'query' : 'request',
        unitPriceUsd: webSearchPrice,
        conditionJson:
          providerSlug === 'mistral-ai'
            ? undefined
            : {
                free_monthly_queries:
                  asNumber(webSearch?.['free_monthly_queries']) ?? 0,
                billing_basis:
                  providerSlug === 'google-gemini'
                    ? 'actual_queries'
                    : 'upstream_query',
              },
        region: 'global',
        sourceUrl: pricingSource,
      });
    }
    const googleMaps = asObject(defaultToolPricing?.['google_maps']);
    const googleMapsPrice = divideDecimalString(
      googleMaps?.['per_thousand_requests'],
      1000,
    );
    if (
      googleMapsPrice &&
      pricingSource &&
      (providerSlug !== 'google-gemini' || documentedTools.has('google_maps'))
    ) {
      pricingRates.push({
        dimension: 'google_maps_queries',
        unit: 'query',
        unitPriceUsd: googleMapsPrice,
        conditionJson: {
          free_monthly_queries:
            asNumber(googleMaps?.['free_monthly_queries']) ?? 0,
          billing_basis: 'actual_queries',
        },
        region: 'global',
        sourceUrl: pricingSource,
      });
    }
    const fileSearch = asObject(defaultToolPricing?.['file_search']);
    const fileSearchPrice = divideDecimalString(
      fileSearch?.['per_thousand_requests'],
      1000,
    );
    if (fileSearchPrice && pricingSource) {
      pricingRates.push({
        dimension: 'file_search_requests',
        unit: 'request',
        unitPriceUsd: fileSearchPrice,
        region: 'global',
        sourceUrl: pricingSource,
      });
    }
    const fileSearchStoragePrice = asDecimalString(
      fileSearch?.['storage_per_gb_day'],
    );
    if (fileSearchStoragePrice && pricingSource) {
      pricingRates.push({
        dimension: 'file_search_storage_gb_days',
        unit: 'gb_day',
        unitPriceUsd: fileSearchStoragePrice,
        conditionJson: {
          free_storage_gb: asNumber(fileSearch?.['free_storage_gb']) ?? 1,
        },
        region: 'global',
        sourceUrl: pricingSource,
      });
    }
    const containers = asObject(defaultToolPricing?.['containers']);
    const containerSessions = asObject(containers?.['per_session']);
    for (const memory of ['1gb', '4gb', '16gb', '64gb'] as const) {
      const price = asDecimalString(containerSessions?.[memory]);
      if (!price || !pricingSource) continue;
      pricingRates.push({
        dimension: `container_sessions_${memory}`,
        unit: 'session',
        unitPriceUsd: price,
        conditionJson: {
          memory,
          session_minutes: asNumber(containers?.['session_minutes']) ?? 20,
          minimum_minutes: asNumber(containers?.['minimum_minutes']) ?? 5,
        },
        region: 'global',
        sourceUrl: pricingSource,
      });
    }
    providerImporter?.appendToolPricingRates?.({
      defaultToolPricing,
      pricingRates,
      sources,
      documentedTools,
    });
    if (regionSet.size === 0) regionSet.add('global');

    const contextWindow = asNumber(limits?.['context_window']);
    const maxInputTokens = asNumber(limits?.['max_input_tokens']);
    const maxInputCharacters = asNumber(limits?.['max_input_characters']);
    const maxInputBytes = asNumber(limits?.['max_input_bytes']);
    const maxOutputTokens =
      asNumber(limits?.['max_output_tokens']) ??
      asNumber(limits?.['max_tokens']);
    const mode = enforceOpenAIModes(
      providerSlug,
      modelId,
      normalizeSupportedModes(doc['supportedModes']),
    );
    const outputModalities = inferOutputModalities(mode, outputModalitiesRaw);
    if (
      (providerSlug === 'google-gemini' ||
        providerSlug === 'openai' ||
        providerSlug === 'xai' ||
        providerSlug === 'mistral-ai' ||
        providerSlug === ZAI_PROVIDER_SLUG) &&
      !explicitRuntimeProfile
    ) {
      throw new Error(`${modelId}: explicit runtimeProfile is required`);
    }
    const minimumRuntimeProfileVersion =
      providerSlug === 'openai' || providerSlug === 'xai'
        ? 3
        : providerSlug === 'google-gemini' || providerSlug === 'mistral-ai'
          ? 2
          : 1;
    if (
      explicitRuntimeProfile &&
      explicitRuntimeProfile.version < minimumRuntimeProfileVersion
    ) {
      throw new Error(
        `${modelId}: runtimeProfile.version must be at least ${minimumRuntimeProfileVersion}`,
      );
    }
    const runtimeProfile = explicitRuntimeProfile;
    assertModesMatchRuntimeEndpoints(modelId, mode, runtimeProfile);

    const deprecatedAt = deprecationDate;
    const releaseDate =
      toDate(doc['releaseDate']) ??
      inferDateFromName(modelId) ??
      inferDateFromName(basename(filePath, '.yaml'));

    const free =
      pricingRows.length > 0
        ? pricingRows.every(
            (p) =>
              (p.inputPricePerMillionTokens ?? 0) === 0 &&
              (p.outputPricePerMillionTokens ?? 0) === 0 &&
              (p.cachedInputPricePerMillionTokens ?? 0) === 0 &&
              (p.imageInputPrice ?? 0) === 0 &&
              (p.imageOutputPrice ?? 0) === 0 &&
              (p.audioInputPrice ?? 0) === 0 &&
              (p.audioOutputPrice ?? 0) === 0 &&
              (p.requestPrice ?? 0) === 0,
          )
        : false;

    models.push({
      canonicalId: `${providerSlug}/${modelId}`,
      slug: modelId,
      displayName:
        typeof doc['displayName'] === 'string' &&
        doc['displayName'].trim() !== ''
          ? doc['displayName'].trim()
          : titleCaseFromSlug(modelId),
      status,
      lifecycleStatus,
      mode,
      contextWindow: contextWindow ? Math.floor(contextWindow) : undefined,
      maxOutputTokens: maxOutputTokens
        ? Math.floor(maxOutputTokens)
        : undefined,
      releaseDate,
      deprecatedAt,
      deprecationDate,
      retirementDate,
      replacementCanonicalId:
        typeof doc['replacementModel'] === 'string'
          ? `${providerSlug}/${doc['replacementModel']}`
          : undefined,
      officialDocsUrl: sources[0],
      free,
      openSource: false,
      capabilities: features,
      inputModalities,
      outputModalities,
      supportedParameters: Array.from(paramsCombined),
      regions: Array.from(regionSet),
      sources,
      runtimeProfile,
      pricingRates,
      pricingRows,
    });
  }

  return {
    slug: providerSlug,
    name: providerName,
    docsUrl,
    websiteUrl,
    models,
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
      supportsRegionalRouting:
        batch.slug !== MOONSHOT_PROVIDER_SLUG &&
        batch.slug !== DEEPSEEK_PROVIDER_SLUG &&
        batch.slug !== ZAI_PROVIDER_SLUG,
      docsUrl: batch.docsUrl,
      websiteUrl: batch.websiteUrl,
    },
    update: {
      name: batch.name,
      docsUrl: batch.docsUrl,
      websiteUrl: batch.websiteUrl,
      status: 'active',
      supportsByok: true,
      supportsRegionalRouting:
        batch.slug !== MOONSHOT_PROVIDER_SLUG &&
        batch.slug !== DEEPSEEK_PROVIDER_SLUG &&
        batch.slug !== ZAI_PROVIDER_SLUG,
    },
  });

  let modelsDeleted = 0;
  if (clearMissing) {
    const keep = new Set(batch.models.map((m) => m.canonicalId));
    const existing = await prisma.model.findMany({
      where: { providerId: provider.id },
      select: { id: true, canonicalId: true },
    });
    const toDeleteIds = existing
      .filter((m) => !keep.has(m.canonicalId))
      .map((m) => m.id);
    if (toDeleteIds.length > 0) {
      const result = await prisma.model.deleteMany({
        where: { id: { in: toDeleteIds } },
      });
      modelsDeleted = result.count;
    }
  }

  for (const row of batch.models) {
    const model = await prisma.model.upsert({
      where: { canonicalId: row.canonicalId },
      create: {
        providerId: provider.id,
        canonicalId: row.canonicalId,
        slug: row.slug,
        displayName: row.displayName,
        mode: row.mode,
        status: row.status,
        lifecycleStatus: row.lifecycleStatus,
        shortDescription: row.shortDescription,
        contextWindow: row.contextWindow,
        maxOutputTokens: row.maxOutputTokens,
        releaseDate: row.releaseDate,
        deprecatedAt: row.deprecatedAt,
        deprecationDate: row.deprecationDate,
        retirementDate: row.retirementDate,
        replacementCanonicalId: row.replacementCanonicalId,
        officialDocsUrl: row.officialDocsUrl,
        free: row.free,
        openSource: row.openSource,
      },
      update: {
        providerId: provider.id,
        slug: row.slug,
        displayName: row.displayName,
        mode: row.mode,
        status: row.status,
        lifecycleStatus: row.lifecycleStatus,
        shortDescription: row.shortDescription,
        contextWindow: row.contextWindow,
        maxOutputTokens: row.maxOutputTokens,
        releaseDate: row.releaseDate,
        deprecatedAt: row.deprecatedAt,
        deprecationDate: row.deprecationDate,
        retirementDate: row.retirementDate,
        replacementCanonicalId: row.replacementCanonicalId,
        officialDocsUrl: row.officialDocsUrl,
        free: row.free,
        openSource: row.openSource,
      },
    });

    const replacementDeletes = [
      prisma.modelPricing.deleteMany({ where: { modelId: model.id } }),
      prisma.modelPricingRate.deleteMany({ where: { modelId: model.id } }),
      prisma.modelCapability.deleteMany({ where: { modelId: model.id } }),
      prisma.modelInputModality.deleteMany({ where: { modelId: model.id } }),
      prisma.modelOutputModality.deleteMany({ where: { modelId: model.id } }),
      prisma.modelSupportedParameter.deleteMany({
        where: { modelId: model.id },
      }),
      prisma.modelRegion.deleteMany({ where: { modelId: model.id } }),
      prisma.sourceReference.deleteMany({ where: { modelId: model.id } }),
    ];
    await prisma.$transaction(replacementDeletes);

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

    for (const pricing of row.pricingRows) {
      const regionId = await ensureRegion(pricing.regionSlug);
      await prisma.modelPricing.create({
        data: {
          modelId: model.id,
          regionId,
          currency: 'USD',
          inputPricePerMillionTokens: pricing.inputPricePerMillionTokens,
          outputPricePerMillionTokens: pricing.outputPricePerMillionTokens,
          cachedInputPricePerMillionTokens:
            pricing.cachedInputPricePerMillionTokens,
          imageInputPrice: pricing.imageInputPrice,
          imageOutputPrice: pricing.imageOutputPrice,
          audioInputPrice: pricing.audioInputPrice,
          audioOutputPrice: pricing.audioOutputPrice,
          requestPrice: pricing.requestPrice,
          sourceUrl: pricing.sourceUrl,
          effectiveFrom: pricing.effectiveFrom ?? new Date(),
          effectiveTo: pricing.effectiveTo,
        },
      });
    }

    for (const rate of row.pricingRates) {
      await prisma.modelPricingRate.create({
        data: {
          modelId: model.id,
          dimension: rate.dimension,
          unit: rate.unit,
          unitPriceUsd: rate.unitPriceUsd,
          currency: 'USD',
          conditionJson: rate.conditionJson,
          region: rate.region,
          sourceUrl: rate.sourceUrl,
          effectiveFrom: rate.effectiveFrom ?? new Date(),
          effectiveTo: rate.effectiveTo,
        },
      });
    }

    if (row.runtimeProfile) {
      const existingRuntimeProfile =
        await prisma.modelRuntimeProfile.findUnique({
          where: {
            modelId_version: {
              modelId: model.id,
              version: row.runtimeProfile.version,
            },
          },
        });
      assertRuntimeProfileVersionIsImmutable(
        row.canonicalId,
        existingRuntimeProfile,
        row.runtimeProfile,
      );
      await prisma.modelRuntimeProfile.upsert({
        where: {
          modelId_version: {
            modelId: model.id,
            version: row.runtimeProfile.version,
          },
        },
        create: {
          modelId: model.id,
          version: row.runtimeProfile.version,
          profileJson: row.runtimeProfile,
          sourceUrl: row.sources[0],
          effectiveFrom: new Date(),
        },
        update: {},
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

  return { modelsImported: batch.models.length, modelsDeleted };
}

export async function importProviderYaml(
  options: ProviderYamlImportOptions,
): Promise<ProviderYamlImportSummary> {
  const {
    providersRoot,
    providerSlugs = [],
    dryRun = false,
    clearMissing = false,
    clearMissingProviders = false,
    log = console.log,
  } = options;
  const selected = new Set(
    providerSlugs.map((x) => slugify(x)).filter(Boolean),
  );
  if (clearMissingProviders && selected.size > 0) {
    throw new Error(
      '--clear-missing-providers cannot be combined with --providers; run a full catalog sync',
    );
  }

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
  let totalProvidersDeleted = 0;
  let totalModels = 0;
  let totalDeleted = 0;
  const catalogProviderSlugs: string[] = [];
  for (const providerDir of providerDirs) {
    const batch = await parseProviderDir(providerDir);
    const result = await importProviderBatch(batch, clearMissing, dryRun);
    catalogProviderSlugs.push(batch.slug);
    totalProviders += 1;
    totalModels += result.modelsImported;
    totalDeleted += result.modelsDeleted;
    log(
      `[import] provider=${batch.slug} models=${result.modelsImported} dryRun=${dryRun}`,
    );
  }

  if (clearMissingProviders && !dryRun) {
    const databaseProviders = await prisma.provider.findMany({
      select: { slug: true },
    });
    const missingProviderSlugs = findMissingProviderSlugs(
      databaseProviders.map((provider) => provider.slug),
      catalogProviderSlugs,
    );
    if (missingProviderSlugs.length > 0) {
      const result = await prisma.provider.deleteMany({
        where: { slug: { in: missingProviderSlugs } },
      });
      totalProvidersDeleted = result.count;
      log(
        `[import] deleted providers=${result.count} slugs=${missingProviderSlugs.join(',')}`,
      );
    }
  }

  log(
    `[import] completed providers=${totalProviders} providersDeleted=${totalProvidersDeleted} models=${totalModels} modelsDeleted=${totalDeleted} dryRun=${dryRun}`,
  );

  return {
    providersImported: totalProviders,
    providersDeleted: totalProvidersDeleted,
    modelsImported: totalModels,
    modelsDeleted: totalDeleted,
    dryRun,
  };
}

export async function disconnectProviderYamlImporter(): Promise<void> {
  await prisma.$disconnect();
}

export async function verifyActiveProviderModelInvariant(
  providerSlug: string,
  expectedModelIds: readonly string[],
): Promise<void> {
  const provider = await prisma.provider.findUnique({
    where: { slug: providerSlug },
    select: {
      models: {
        where: { enabled: true, lifecycleStatus: 'published' },
        select: { slug: true },
        orderBy: { slug: 'asc' },
      },
    },
  });
  const actual = provider?.models.map((model) => model.slug) ?? [];
  const expected = [...expectedModelIds].sort();
  if (stableJson(actual) !== stableJson(expected)) {
    throw new Error(
      `${providerSlug} active model invariant failed: expected=${expected.join(',')} actual=${actual.join(',')}`,
    );
  }
}

async function main(args: string[] = process.argv.slice(2)): Promise<void> {
  const providersRoot =
    getArgValue('--providers-root', args) ?? DEFAULT_PROVIDERS_ROOT;
  const providersArg = getArgValue('--providers', args);

  await importProviderYaml({
    providersRoot,
    providerSlugs: (providersArg ?? '')
      .split(',')
      .map((provider) => provider.trim())
      .filter(Boolean),
    dryRun: hasFlag('--dry-run', args),
    clearMissing: hasFlag('--clear-missing', args),
    clearMissingProviders: hasFlag('--clear-missing-providers', args),
  });
}

if (require.main === module) {
  void main()
    .catch((err) => {
      console.error(err);
      process.exitCode = 1;
    })
    .finally(disconnectProviderYamlImporter);
}
