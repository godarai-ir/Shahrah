import { existsSync, promises as fs } from 'node:fs';
import { basename, join, relative, resolve } from 'node:path';
import { parse } from 'yaml';
import { modelRuntimeProfileSchema } from '../../../src/catalog/runtime-profile.schema';
import type { ProviderPricingRate, ProviderRateMapping } from '../types';

export const MISTRAL_AI_PROVIDER_SLUG = 'mistral-ai';

export const MISTRAL_AI_SUPPORTED_MODEL_IDS = [
  'codestral-2508',
  'codestral-embed-2505',
  'mistral-embed-2312',
  'mistral-large-2512',
  'mistral-medium-3-5',
  'mistral-moderation-2603',
  'mistral-ocr-4-0',
  'mistral-small-2603',
  'voxtral-mini-2602',
  'voxtral-mini-tts-2603',
] as const;

type MistralYamlDocument = {
  model?: unknown;
  status?: unknown;
  runtimeProfile?: unknown;
  sources?: unknown;
  costs?: unknown;
};

const EXPECTED_ENDPOINTS: Record<string, string[]> = {
  'codestral-2508': [
    'chat_completions',
    'fim_completions',
    'conversations',
    'batches',
  ],
  'codestral-embed-2505': ['embeddings', 'batches'],
  'mistral-embed-2312': ['embeddings', 'batches'],
  'mistral-large-2512': [
    'chat_completions',
    'conversations',
    'agents',
    'files',
    'batches',
  ],
  'mistral-medium-3-5': [
    'chat_completions',
    'conversations',
    'agents',
    'files',
    'batches',
  ],
  'mistral-moderation-2603': ['moderations', 'chat_moderations', 'batches'],
  'mistral-ocr-4-0': ['ocr', 'files', 'batches'],
  'mistral-small-2603': [
    'chat_completions',
    'conversations',
    'agents',
    'files',
    'batches',
  ],
  'voxtral-mini-2602': ['audio_transcriptions', 'batches'],
  'voxtral-mini-tts-2603': ['audio_speech'],
};

const EXPECTED_BATCH_ENDPOINTS: Record<string, string[]> = {
  'codestral-2508': [
    'chat_completions',
    'fim_completions',
    'conversations',
  ],
  'codestral-embed-2505': ['embeddings'],
  'mistral-embed-2312': ['embeddings'],
  'mistral-large-2512': ['chat_completions', 'conversations'],
  'mistral-medium-3-5': ['chat_completions', 'conversations'],
  'mistral-moderation-2603': ['moderations', 'chat_moderations'],
  'mistral-ocr-4-0': ['ocr'],
  'mistral-small-2603': ['chat_completions', 'conversations'],
  'voxtral-mini-2602': ['audio_transcriptions'],
};

const REQUIRED_COST_KEYS: Record<string, string[]> = {
  'codestral-2508': ['input_cost_per_token', 'output_cost_per_token'],
  'codestral-embed-2505': ['input_cost_per_token'],
  'mistral-embed-2312': ['input_cost_per_token'],
  'mistral-large-2512': ['input_cost_per_token', 'output_cost_per_token'],
  'mistral-medium-3-5': ['input_cost_per_token', 'output_cost_per_token'],
  'mistral-moderation-2603': ['input_cost_per_token'],
  'mistral-ocr-4-0': ['input_cost_per_page', 'input_cost_per_annotated_page'],
  'mistral-small-2603': ['input_cost_per_token', 'output_cost_per_token'],
  'voxtral-mini-2602': ['input_cost_per_second'],
  'voxtral-mini-tts-2603': ['input_cost_per_character'],
};

export type MistralAIProviderValidationResult = {
  providerDir: string;
  modelIds: string[];
  modelFiles: string[];
};

async function collectYamlFiles(dir: string): Promise<string[]> {
  const entries = await fs.readdir(dir, { withFileTypes: true });
  const files: string[] = [];
  for (const entry of entries) {
    const fullPath = join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await collectYamlFiles(fullPath)));
    } else if (
      entry.isFile() &&
      entry.name.endsWith('.yaml') &&
      entry.name !== 'default.yaml'
    ) {
      files.push(fullPath);
    }
  }
  return files.sort();
}

function stringArray(value: unknown): string[] {
  return Array.isArray(value)
    ? value.filter((item): item is string => typeof item === 'string')
    : [];
}

export async function validateMistralAIProviderDirectory(
  providersRoot: string,
): Promise<MistralAIProviderValidationResult> {
  const resolvedRoot = resolve(providersRoot);
  const providerDir = resolve(resolvedRoot, MISTRAL_AI_PROVIDER_SLUG);
  const relativeProviderDir = relative(resolvedRoot, providerDir);
  if (
    relativeProviderDir.startsWith('..') ||
    relativeProviderDir === '' ||
    basename(providerDir) !== MISTRAL_AI_PROVIDER_SLUG
  ) {
    throw new Error(`Unsafe Mistral AI provider directory: ${providerDir}`);
  }
  const metadataPath = join(providerDir, 'default.yaml');
  if (!existsSync(providerDir) || !existsSync(metadataPath)) {
    throw new Error(`Mistral AI provider metadata is missing: ${metadataPath}`);
  }

  const metadata = parse(await fs.readFile(metadataPath, 'utf8')) as {
    documentation?: unknown;
  };
  if (
    !stringArray(metadata.documentation).includes(
      'https://mistral.ai/pricing/api/',
    )
  ) {
    throw new Error(
      'Mistral AI metadata must use the official API pricing URL',
    );
  }

  const modelFiles = await collectYamlFiles(providerDir);
  const modelIds: string[] = [];
  const invalidFiles: string[] = [];
  for (const filePath of modelFiles) {
    const relativeFile = relative(providerDir, filePath);
    let document: MistralYamlDocument;
    try {
      document = parse(
        await fs.readFile(filePath, 'utf8'),
      ) as MistralYamlDocument;
    } catch {
      invalidFiles.push(`${relativeFile}(malformed_yaml)`);
      continue;
    }
    if (typeof document.model !== 'string' || document.model.trim() === '') {
      invalidFiles.push(`${relativeFile}(missing_model)`);
      continue;
    }
    const modelId = document.model.trim();
    modelIds.push(modelId);
    if (document.status !== 'active') {
      invalidFiles.push(`${relativeFile}(non_active_model)`);
    }
    if (basename(filePath, '.yaml') !== modelId) {
      invalidFiles.push(`${relativeFile}(filename_mismatch)`);
    }
    const profile = modelRuntimeProfileSchema.safeParse(
      document.runtimeProfile,
    );
    if (!profile.success || profile.data.version < 2) {
      invalidFiles.push(`${relativeFile}(invalid_runtime_profile)`);
    } else {
      const expectedEndpoints = EXPECTED_ENDPOINTS[modelId] ?? [];
      const actualEndpoints = [...profile.data.endpoints].sort();
      if (
        JSON.stringify(actualEndpoints) !==
        JSON.stringify([...expectedEndpoints].sort())
      ) {
        invalidFiles.push(`${relativeFile}(endpoint_matrix_mismatch)`);
      }

      const expectedBatchEndpoints = EXPECTED_BATCH_ENDPOINTS[modelId];
      if (expectedBatchEndpoints) {
        const batch = profile.data.batch;
        const actualBatchEndpoints = [...batch.supportedEndpoints].sort();
        if (
          !batch.eligible ||
          !batch.discounted ||
          JSON.stringify(actualBatchEndpoints) !==
            JSON.stringify([...expectedBatchEndpoints].sort())
        ) {
          invalidFiles.push(`${relativeFile}(batch_profile_mismatch)`);
        }
      }
    }
    const sources = stringArray(document.sources);
    if (
      !sources.some((source) => source.startsWith('https://docs.mistral.ai/'))
    ) {
      invalidFiles.push(`${relativeFile}(missing_official_source)`);
    }
    if (!Array.isArray(document.costs) || document.costs.length === 0) {
      invalidFiles.push(`${relativeFile}(missing_pricing)`);
    } else {
      const presentCostKeys = new Set(
        document.costs.flatMap((cost) => Object.keys(asObject(cost) ?? {})),
      );
      if (
        (REQUIRED_COST_KEYS[modelId] ?? []).some(
          (key) => !presentCostKeys.has(key),
        )
      ) {
        invalidFiles.push(`${relativeFile}(incomplete_pricing_dimensions)`);
      }
    }
  }

  const expected = new Set<string>(MISTRAL_AI_SUPPORTED_MODEL_IDS);
  const actual = new Set(modelIds);
  const missing = [...expected].filter((modelId) => !actual.has(modelId));
  const unexpected = [...actual].filter((modelId) => !expected.has(modelId));
  const duplicates = [...actual].filter(
    (modelId) =>
      modelIds.filter((candidate) => candidate === modelId).length > 1,
  );
  if (
    missing.length > 0 ||
    unexpected.length > 0 ||
    duplicates.length > 0 ||
    invalidFiles.length > 0 ||
    modelFiles.length !== MISTRAL_AI_SUPPORTED_MODEL_IDS.length
  ) {
    const details = [
      missing.length ? `missing=${missing.join(',')}` : undefined,
      unexpected.length ? `unexpected=${unexpected.join(',')}` : undefined,
      duplicates.length ? `duplicates=${duplicates.join(',')}` : undefined,
      invalidFiles.length
        ? `invalid_files=${invalidFiles.join(',')}`
        : undefined,
    ].filter(Boolean);
    throw new Error(
      `Mistral AI model allowlist validation failed for ${providerDir}: ${details.join(' ')}`,
    );
  }

  return { providerDir, modelIds: [...actual].sort(), modelFiles };
}

const MISTRAL_AI_RATE_MAPPINGS: ProviderRateMapping[] = [
  {
    key: 'input_cost_per_second',
    dimension: 'audio_input_seconds',
    unit: 'second',
  },
  {
    key: 'input_cost_per_page',
    dimension: 'ocr_input_pages',
    unit: 'page',
  },
  {
    key: 'input_cost_per_annotated_page',
    dimension: 'ocr_annotated_pages',
    unit: 'page',
    conditionJson: { document_ai: true },
  },
];

function asObject(value: unknown): Record<string, unknown> | undefined {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : undefined;
}

function decimal(value: unknown): string | undefined {
  if (typeof value === 'number' && Number.isFinite(value)) return String(value);
  if (
    typeof value === 'string' &&
    value.trim() &&
    Number.isFinite(Number(value))
  ) {
    return value.trim();
  }
  return undefined;
}

function appendToolRate(
  pricingRates: ProviderPricingRate[],
  sourceUrl: string,
  value: unknown,
  dimension: string,
  unit: string,
): void {
  const unitPriceUsd = decimal(value);
  if (!unitPriceUsd) return;
  pricingRates.push({
    dimension,
    unit,
    unitPriceUsd,
    region: 'global',
    sourceUrl,
  });
}

function appendToolPricingRates(input: {
  defaultToolPricing?: Record<string, unknown>;
  pricingRates: ProviderPricingRate[];
  sources: string[];
}): void {
  const sourceUrl = input.sources.find(
    (source) => source === 'https://mistral.ai/pricing/api/',
  );
  if (!sourceUrl) return;
  const codeExecution = asObject(input.defaultToolPricing?.code_execution);
  const imageGeneration = asObject(input.defaultToolPricing?.image_generation);
  const libraries = asObject(input.defaultToolPricing?.libraries);
  const premiumWebSearch = asObject(
    input.defaultToolPricing?.web_search_premium,
  );
  appendToolRate(
    input.pricingRates,
    sourceUrl,
    premiumWebSearch?.per_call,
    'web_search_premium_requests',
    'request',
  );
  appendToolRate(
    input.pricingRates,
    sourceUrl,
    codeExecution?.per_call,
    'code_execution_calls',
    'call',
  );
  appendToolRate(
    input.pricingRates,
    sourceUrl,
    imageGeneration?.per_image,
    'image_generation_outputs',
    'image',
  );
  appendToolRate(
    input.pricingRates,
    sourceUrl,
    libraries?.per_call,
    'library_calls',
    'call',
  );
  appendToolRate(
    input.pricingRates,
    sourceUrl,
    libraries?.ocr_per_page,
    'library_ocr_pages',
    'page',
  );
  appendToolRate(
    input.pricingRates,
    sourceUrl,
    libraries?.indexing_per_token,
    'library_indexing_tokens',
    'token',
  );
}

export const mistralAIProviderImporter = {
  slug: MISTRAL_AI_PROVIDER_SLUG,
  batchPricingCondition: {
    api: 'batches',
  } as ProviderRateMapping['conditionJson'],
  rateMappings: MISTRAL_AI_RATE_MAPPINGS,
  appendDerivedCostPricingRates: undefined,
  appendToolPricingRates,
};
