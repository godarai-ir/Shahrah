import { existsSync, promises as fs } from 'node:fs';
import { basename, join, relative, resolve } from 'node:path';
import { parse } from 'yaml';
import { modelRuntimeProfileSchema } from '../../../src/catalog/runtime-profile.schema';
import type { ProviderRateMapping } from '../types';

export const OPENAI_PROVIDER_SLUG = 'openai';

export const OPENAI_SUPPORTED_MODEL_IDS = [
  'gpt-6-sol',
  'gpt-6-luna',
  'gpt-5.6-sol',
  'gpt-5.6-terra',
  'gpt-5.6-luna',
  'gpt-5.5',
  'gpt-5.5-pro',
  'gpt-5.4',
  'gpt-5.4-pro',
  'gpt-5.4-mini',
  'gpt-5.4-nano',
  'gpt-4.1',
  'gpt-4.1-mini',
  'gpt-4o-mini',
  'text-embedding-3-small',
  'text-embedding-3-large',
  'omni-moderation-latest',
  'gpt-image-2',
  'gpt-transcribe',
  'gpt-4o-mini-transcribe',
  'gpt-4o-transcribe-diarize',
  'tts-1',
  'tts-1-hd',
] as const;

// Fast mode (service_tier "priority" or "fast") long-context rates. OpenAI
// prices prompts with more than 272K input tokens at long-context rates for
// the full request, including under Fast mode for models that publish them.
export const OPENAI_PRIORITY_LONG_CONTEXT_RATE_MAPPINGS: ProviderRateMapping[] =
  [
    [
      'input_cost_per_token_priority_above_272k_tokens',
      'speed_long_context_input_tokens',
    ],
    [
      'output_cost_per_token_priority_above_272k_tokens',
      'speed_long_context_output_tokens',
    ],
    [
      'cache_read_input_token_cost_priority_above_272k_tokens',
      'speed_long_context_cache_read_tokens',
    ],
    [
      'cache_creation_input_token_cost_priority_above_272k_tokens',
      'speed_long_context_cache_write_tokens',
    ],
  ].map(([key, dimension]) => ({
    key,
    dimension,
    unit: 'token',
    conditionJson: { service_tier: 'priority', min_prompt_tokens: 272001 },
  }));

type OpenAIYamlDocument = {
  model?: unknown;
  runtimeProfile?: unknown;
  serviceTiers?: unknown;
  costs?: unknown;
};

export type OpenAIProviderValidationResult = {
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
      continue;
    }
    if (
      entry.isFile() &&
      entry.name.endsWith('.yaml') &&
      entry.name !== 'default.yaml'
    ) {
      files.push(fullPath);
    }
  }

  return files.sort();
}

function formatValidationError(
  providerDir: string,
  missing: string[],
  unexpected: string[],
  duplicates: string[],
  invalidFiles: string[],
): string {
  const details = [
    missing.length > 0 ? `missing=${missing.join(',')}` : undefined,
    unexpected.length > 0 ? `unexpected=${unexpected.join(',')}` : undefined,
    duplicates.length > 0 ? `duplicates=${duplicates.join(',')}` : undefined,
    invalidFiles.length > 0
      ? `invalid_files=${invalidFiles.join(',')}`
      : undefined,
  ].filter(Boolean);

  return `OpenAI model allowlist validation failed for ${providerDir}: ${details.join(
    ' ',
  )}`;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function hasNonNegativePrice(
  records: Array<Record<string, unknown>>,
  key: string,
): boolean {
  return records.some((record) => {
    const value = record[key];
    return typeof value === 'number' && Number.isFinite(value) && value >= 0;
  });
}

function validateTierPricing(document: OpenAIYamlDocument): string[] {
  const tiers = Array.isArray(document.serviceTiers)
    ? document.serviceTiers.filter(
        (value): value is string => typeof value === 'string',
      )
    : [];
  if (tiers.length === 0) return [];

  const allowedTiers = new Set([
    'standard',
    'batch',
    'flex',
    'priority',
    'fast',
  ]);
  const invalidTiers = tiers.filter((tier) => !allowedTiers.has(tier));
  const costs = Array.isArray(document.costs)
    ? document.costs.filter(isRecord)
    : [];
  const missing: string[] = invalidTiers.map(
    (tier) => `unsupported_service_tier:${tier}`,
  );
  const requirePrices = (tier: string, keys: string[]) => {
    if (!tiers.includes(tier)) return;
    for (const key of keys) {
      if (!hasNonNegativePrice(costs, key)) {
        missing.push(`${tier}:${key}`);
      }
    }
  };

  requirePrices('standard', ['input_cost_per_token', 'output_cost_per_token']);
  requirePrices('batch', [
    'input_cost_per_token_batches',
    'output_cost_per_token_batches',
  ]);
  requirePrices('flex', [
    'input_cost_per_token_flex',
    'output_cost_per_token_flex',
  ]);

  const priorityEnabled = tiers.includes('priority');
  const fastEnabled = tiers.includes('fast');
  if (priorityEnabled !== fastEnabled) {
    missing.push('priority_fast_alias_pair');
  }
  if (priorityEnabled || fastEnabled) {
    const priorityPricing = costs
      .map((cost) => cost.priority_pricing)
      .filter(isRecord);
    for (const key of ['input_cost_per_token', 'output_cost_per_token']) {
      if (!hasNonNegativePrice(priorityPricing, key)) {
        missing.push(`priority:${key}`);
      }
    }
  }

  return missing;
}

export async function validateOpenAIProviderDirectory(
  providersRoot: string,
): Promise<OpenAIProviderValidationResult> {
  const resolvedRoot = resolve(providersRoot);
  const providerDir = resolve(resolvedRoot, OPENAI_PROVIDER_SLUG);
  const relativeProviderDir = relative(resolvedRoot, providerDir);

  if (
    relativeProviderDir.startsWith('..') ||
    relativeProviderDir === '' ||
    basename(providerDir) !== OPENAI_PROVIDER_SLUG
  ) {
    throw new Error(`Unsafe OpenAI provider directory: ${providerDir}`);
  }
  if (!existsSync(providerDir)) {
    throw new Error(`OpenAI provider directory does not exist: ${providerDir}`);
  }
  if (!existsSync(join(providerDir, 'default.yaml'))) {
    throw new Error(
      `OpenAI provider metadata is missing: ${join(providerDir, 'default.yaml')}`,
    );
  }

  const modelFiles = await collectYamlFiles(providerDir);
  const modelIds: string[] = [];
  const invalidFiles: string[] = [];

  for (const filePath of modelFiles) {
    const relativeFile = relative(providerDir, filePath);
    let document: OpenAIYamlDocument;
    try {
      document = parse(
        await fs.readFile(filePath, 'utf8'),
      ) as OpenAIYamlDocument;
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
    } else if (runtimeProfile.data.version < 3) {
      invalidFiles.push(`${relativeFile}(runtime_profile_version)`);
    } else {
      const declaredTiers = Array.isArray(document.serviceTiers)
        ? document.serviceTiers
        : [];
      if (
        JSON.stringify(
          runtimeProfile.data.requestControls.serviceTierValues,
        ) !== JSON.stringify(declaredTiers)
      ) {
        invalidFiles.push(`${relativeFile}(service_tier_profile_mismatch)`);
      }
    }

    const missingTierPricing = validateTierPricing(document);
    if (missingTierPricing.length > 0) {
      invalidFiles.push(
        `${relativeFile}(missing_tier_pricing:${missingTierPricing.join('|')})`,
      );
    }
  }

  const expected = new Set<string>(OPENAI_SUPPORTED_MODEL_IDS);
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
    modelFiles.length !== OPENAI_SUPPORTED_MODEL_IDS.length
  ) {
    throw new Error(
      formatValidationError(
        providerDir,
        missing,
        unexpected,
        duplicates,
        invalidFiles,
      ),
    );
  }

  return {
    providerDir,
    modelIds: [...actual].sort(),
    modelFiles,
  };
}
