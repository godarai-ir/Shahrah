import { existsSync, promises as fs } from 'node:fs';
import { basename, join, relative, resolve } from 'node:path';
import { parse } from 'yaml';
import { modelRuntimeProfileSchema } from '../../../src/catalog/runtime-profile.schema';
import type { ProviderRateMapping } from '../types';

export const GOOGLE_GEMINI_PROVIDER_SLUG = 'google-gemini';

export const GOOGLE_GEMINI_SUPPORTED_MODEL_IDS = [
  'gemini-3.1-pro-preview',
  'gemini-3.5-flash',
  'gemini-3.5-flash-lite',
  'gemini-3.6-flash',
  'gemini-3.7-flash',
  'gemini-3-pro-image',
  'gemini-3.1-flash-image',
  'gemini-3.1-flash-lite-image',
  'gemini-3.1-flash-tts-preview',
  'gemini-embedding-2',
] as const;

type GeminiYamlDocument = { model?: unknown; runtimeProfile?: unknown };

export type GoogleGeminiProviderValidationResult = {
  providerDir: string;
  modelIds: string[];
  modelFiles: string[];
};

export type GoogleGeminiImportOptions = {
  providersRoot?: string;
  dryRun: boolean;
  clearMissing: boolean;
};

export function parseGoogleGeminiImportArgs(
  args: string[],
): GoogleGeminiImportOptions {
  if (args.includes('--providers')) {
    throw new Error(
      'The Google Gemini importer is locked to provider=google-gemini; remove --providers.',
    );
  }
  const rootIndex = args.indexOf('--providers-root');
  if (
    rootIndex >= 0 &&
    (rootIndex + 1 >= args.length || args[rootIndex + 1].startsWith('--'))
  ) {
    throw new Error('--providers-root requires a path value');
  }
  const supported = new Set([
    '--providers-root',
    '--dry-run',
    '--clear-missing',
  ]);
  for (let index = 0; index < args.length; index += 1) {
    const value = args[index];
    if (index > 0 && args[index - 1] === '--providers-root') continue;
    if (value.startsWith('--') && !supported.has(value)) {
      throw new Error(`Unsupported Google Gemini importer option: ${value}`);
    }
  }
  return {
    providersRoot: rootIndex >= 0 ? args[rootIndex + 1] : undefined,
    dryRun: args.includes('--dry-run'),
    clearMissing: args.includes('--clear-missing'),
  };
}

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

export async function validateGoogleGeminiProviderDirectory(
  providersRoot: string,
): Promise<GoogleGeminiProviderValidationResult> {
  const resolvedRoot = resolve(providersRoot);
  const providerDir = resolve(resolvedRoot, GOOGLE_GEMINI_PROVIDER_SLUG);
  const relativeProviderDir = relative(resolvedRoot, providerDir);
  if (
    relativeProviderDir.startsWith('..') ||
    relativeProviderDir === '' ||
    basename(providerDir) !== GOOGLE_GEMINI_PROVIDER_SLUG
  ) {
    throw new Error(`Unsafe Google Gemini provider directory: ${providerDir}`);
  }
  if (!existsSync(providerDir)) {
    throw new Error(
      `Google Gemini provider directory does not exist: ${providerDir}`,
    );
  }
  const metadataPath = join(providerDir, 'default.yaml');
  if (!existsSync(metadataPath)) {
    throw new Error(
      `Google Gemini provider metadata is missing: ${metadataPath}`,
    );
  }

  const modelFiles = await collectYamlFiles(providerDir);
  const modelIds: string[] = [];
  const invalidFiles: string[] = [];
  for (const filePath of modelFiles) {
    const relativeFile = relative(providerDir, filePath);
    let document: GeminiYamlDocument;
    try {
      document = parse(
        await fs.readFile(filePath, 'utf8'),
      ) as GeminiYamlDocument;
    } catch {
      invalidFiles.push(`${relativeFile}(malformed_yaml)`);
      continue;
    }
    if (typeof document?.model !== 'string' || document.model.trim() === '') {
      invalidFiles.push(relativeFile);
      continue;
    }
    const modelId = document.model.trim();
    if (basename(filePath, '.yaml') !== modelId) {
      invalidFiles.push(`${relativeFile}(filename_mismatch)`);
      continue;
    }
    modelIds.push(modelId);
    const runtimeProfile = modelRuntimeProfileSchema.safeParse(
      document.runtimeProfile,
    );
    if (!runtimeProfile.success) {
      invalidFiles.push(`${relativeFile}(invalid_runtime_profile)`);
    } else if (runtimeProfile.data.version < 2) {
      invalidFiles.push(`${relativeFile}(runtime_profile_version)`);
    }
  }

  const expected = new Set<string>(GOOGLE_GEMINI_SUPPORTED_MODEL_IDS);
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
    modelFiles.length !== GOOGLE_GEMINI_SUPPORTED_MODEL_IDS.length
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
      `Google Gemini model allowlist validation failed for ${providerDir}: ${details.join(' ')}`,
    );
  }

  return { providerDir, modelIds: [...actual].sort(), modelFiles };
}

type GoogleGeminiRateMappingEntry = readonly [
  key: string,
  dimension: string,
  unit: string,
  serviceTier: string,
  minPromptTokens?: number,
];

const GOOGLE_GEMINI_RATE_MAPPING_ENTRIES = [
  [
    'cache_storage_cost_per_token_per_hour_batches',
    'batch_cache_storage_token_hours',
    'token_hour',
    'batch',
  ],
  [
    'cache_storage_cost_per_token_per_hour_flex',
    'flex_cache_storage_token_hours',
    'token_hour',
    'flex',
  ],
  [
    'input_cost_per_token_priority',
    'priority_input_tokens',
    'token',
    'priority',
  ],
  [
    'output_cost_per_token_priority',
    'priority_output_tokens',
    'token',
    'priority',
  ],
  [
    'cache_read_input_token_cost_priority',
    'priority_cache_read_tokens',
    'token',
    'priority',
  ],
  [
    'cache_storage_cost_per_token_per_hour_priority',
    'priority_cache_storage_token_hours',
    'token_hour',
    'priority',
  ],
  [
    'input_cost_per_token_batches_above_200k_tokens',
    'batch_long_context_input_tokens',
    'token',
    'batch',
    200001,
  ],
  [
    'output_cost_per_token_batches_above_200k_tokens',
    'batch_long_context_output_tokens',
    'token',
    'batch',
    200001,
  ],
  [
    'cache_read_input_token_cost_batches_above_200k_tokens',
    'batch_long_context_cache_read_tokens',
    'token',
    'batch',
    200001,
  ],
  [
    'input_cost_per_token_flex_above_200k_tokens',
    'flex_long_context_input_tokens',
    'token',
    'flex',
    200001,
  ],
  [
    'output_cost_per_token_flex_above_200k_tokens',
    'flex_long_context_output_tokens',
    'token',
    'flex',
    200001,
  ],
  [
    'cache_read_input_token_cost_flex_above_200k_tokens',
    'flex_long_context_cache_read_tokens',
    'token',
    'flex',
    200001,
  ],
  [
    'input_cost_per_token_priority_above_200k_tokens',
    'priority_long_context_input_tokens',
    'token',
    'priority',
    200001,
  ],
  [
    'output_cost_per_token_priority_above_200k_tokens',
    'priority_long_context_output_tokens',
    'token',
    'priority',
    200001,
  ],
  [
    'cache_read_input_token_cost_priority_above_200k_tokens',
    'priority_long_context_cache_read_tokens',
    'token',
    'priority',
    200001,
  ],
  ['input_cost_per_image_batches', 'batch_input_images', 'image', 'batch'],
  ['input_cost_per_image_flex', 'flex_input_images', 'image', 'flex'],
  [
    'output_cost_per_audio_token_batches',
    'batch_audio_output_tokens',
    'token',
    'batch',
  ],
  [
    'input_cost_per_audio_token_batches',
    'batch_audio_input_tokens',
    'token',
    'batch',
  ],
  ['input_cost_per_video_token', 'video_input_tokens', 'token', 'standard'],
  [
    'input_cost_per_video_token_batches',
    'batch_video_input_tokens',
    'token',
    'batch',
  ],
  ['output_cost_per_image_0_5k', 'output_images_0_5k', 'image', 'standard'],
  ['output_cost_per_image_1k', 'output_images_1k', 'image', 'standard'],
  ['output_cost_per_image_2k', 'output_images_2k', 'image', 'standard'],
  ['output_cost_per_image_4k', 'output_images_4k', 'image', 'standard'],
  [
    'output_cost_per_image_0_5k_batches',
    'batch_output_images_0_5k',
    'image',
    'batch',
  ],
  [
    'output_cost_per_image_1k_batches',
    'batch_output_images_1k',
    'image',
    'batch',
  ],
  [
    'output_cost_per_image_2k_batches',
    'batch_output_images_2k',
    'image',
    'batch',
  ],
  [
    'output_cost_per_image_4k_batches',
    'batch_output_images_4k',
    'image',
    'batch',
  ],
  [
    'input_cost_per_image_priority',
    'priority_input_images',
    'image',
    'priority',
  ],
  [
    'output_cost_per_image_token_priority',
    'priority_image_output_tokens',
    'token',
    'priority',
  ],
  [
    'output_cost_per_image_token_flex',
    'flex_image_output_tokens',
    'token',
    'flex',
  ],
] as const satisfies readonly GoogleGeminiRateMappingEntry[];

const GOOGLE_GEMINI_RATE_MAPPINGS: ProviderRateMapping[] =
  GOOGLE_GEMINI_RATE_MAPPING_ENTRIES.map(
    ([key, dimension, unit, serviceTier, minPromptTokens]) => ({
      key,
      dimension,
      unit,
      conditionJson: {
        service_tier: serviceTier,
        ...(typeof minPromptTokens === 'number'
          ? { min_prompt_tokens: minPromptTokens }
          : {}),
      },
    }),
  );

export const googleGeminiProviderImporter = {
  slug: GOOGLE_GEMINI_PROVIDER_SLUG,
  batchPricingCondition: {
    service_tier: 'batch',
  } as ProviderRateMapping['conditionJson'],
  rateMappings: GOOGLE_GEMINI_RATE_MAPPINGS,
  appendDerivedCostPricingRates: undefined,
  appendToolPricingRates: undefined,
};
