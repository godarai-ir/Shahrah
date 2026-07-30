import { existsSync, promises as fs } from 'node:fs';
import { basename, join, relative, resolve } from 'node:path';
import { parse } from 'yaml';
import type { ModelRuntimeProfile } from '../../../src/catalog/runtime-profile.schema';
import type { ProviderPricingRate, ProviderRateMapping } from '../types';

export const ANTHROPIC_PROVIDER_SLUG = 'anthropic';

export const ANTHROPIC_SUPPORTED_MODEL_IDS = [
  'claude-fable-5',
  'claude-haiku-4-5',
  'claude-opus-5',
  'claude-sonnet-5',
] as const;

type StringRecord = Record<string, unknown>;

type AnthropicYamlDocument = {
  model?: unknown;
};

export type AnthropicProviderValidationResult = {
  providerDir: string;
  modelIds: string[];
  modelFiles: string[];
};

type AppendCostPricingRatesInput = {
  cost: StringRecord;
  explicitRuntimeProfile?: ModelRuntimeProfile;
  pricingRates: ProviderPricingRate[];
  region: string;
  sourceUrl: string;
};

type AppendToolPricingRatesInput = {
  defaultToolPricing?: StringRecord;
  pricingRates: ProviderPricingRate[];
  sources: string[];
};

const ANTHROPIC_RATE_MAPPINGS: ProviderRateMapping[] = [
  {
    key: 'data_residency_input_cost_per_token',
    dimension: 'data_residency_input_tokens',
    unit: 'token',
    conditionJson: { inference_geo: 'configured' },
  },
  {
    key: 'data_residency_output_cost_per_token',
    dimension: 'data_residency_output_tokens',
    unit: 'token',
    conditionJson: { inference_geo: 'configured' },
  },
  {
    key: 'data_residency_cache_creation_input_token_cost',
    dimension: 'data_residency_cache_write_5m_tokens',
    unit: 'token',
    conditionJson: { inference_geo: 'configured' },
  },
  {
    key: 'data_residency_cache_creation_input_token_cost_per_hour',
    dimension: 'data_residency_cache_write_1h_tokens',
    unit: 'token',
    conditionJson: { inference_geo: 'configured' },
  },
  {
    key: 'data_residency_cache_read_input_token_cost',
    dimension: 'data_residency_cache_read_tokens',
    unit: 'token',
    conditionJson: { inference_geo: 'configured' },
  },
  {
    key: 'data_residency_batch_input_cost_per_token',
    dimension: 'data_residency_batch_input_tokens',
    unit: 'token',
    conditionJson: {
      api: 'message_batches',
      inference_geo: 'configured',
    },
  },
  {
    key: 'data_residency_batch_output_cost_per_token',
    dimension: 'data_residency_batch_output_tokens',
    unit: 'token',
    conditionJson: {
      api: 'message_batches',
      inference_geo: 'configured',
    },
  },
  {
    key: 'data_residency_batch_cache_creation_input_token_cost',
    dimension: 'data_residency_batch_cache_write_5m_tokens',
    unit: 'token',
    conditionJson: {
      api: 'message_batches',
      inference_geo: 'configured',
    },
  },
  {
    key: 'data_residency_batch_cache_creation_input_token_cost_per_hour',
    dimension: 'data_residency_batch_cache_write_1h_tokens',
    unit: 'token',
    conditionJson: {
      api: 'message_batches',
      inference_geo: 'configured',
    },
  },
  {
    key: 'data_residency_batch_cache_read_input_token_cost',
    dimension: 'data_residency_batch_cache_read_tokens',
    unit: 'token',
    conditionJson: {
      api: 'message_batches',
      inference_geo: 'configured',
    },
  },
  {
    key: 'fast_mode_input_cost_per_token',
    dimension: 'speed_input_tokens',
    unit: 'token',
    conditionJson: { speed: 'fast' },
  },
  {
    key: 'fast_mode_output_cost_per_token',
    dimension: 'speed_output_tokens',
    unit: 'token',
    conditionJson: { speed: 'fast' },
  },
  {
    key: 'fast_mode_cache_creation_input_token_cost',
    dimension: 'speed_cache_write_5m_tokens',
    unit: 'token',
    conditionJson: { speed: 'fast' },
  },
  {
    key: 'fast_mode_cache_creation_input_token_cost_per_hour',
    dimension: 'speed_cache_write_1h_tokens',
    unit: 'token',
    conditionJson: { speed: 'fast' },
  },
  {
    key: 'fast_mode_cache_read_input_token_cost',
    dimension: 'speed_cache_read_tokens',
    unit: 'token',
    conditionJson: { speed: 'fast' },
  },
  {
    key: 'data_residency_fast_mode_input_cost_per_token',
    dimension: 'data_residency_speed_input_tokens',
    unit: 'token',
    conditionJson: { speed: 'fast', inference_geo: 'configured' },
  },
  {
    key: 'data_residency_fast_mode_output_cost_per_token',
    dimension: 'data_residency_speed_output_tokens',
    unit: 'token',
    conditionJson: { speed: 'fast', inference_geo: 'configured' },
  },
  {
    key: 'data_residency_fast_mode_cache_creation_input_token_cost',
    dimension: 'data_residency_speed_cache_write_5m_tokens',
    unit: 'token',
    conditionJson: { speed: 'fast', inference_geo: 'configured' },
  },
  {
    key: 'data_residency_fast_mode_cache_creation_input_token_cost_per_hour',
    dimension: 'data_residency_speed_cache_write_1h_tokens',
    unit: 'token',
    conditionJson: { speed: 'fast', inference_geo: 'configured' },
  },
  {
    key: 'data_residency_fast_mode_cache_read_input_token_cost',
    dimension: 'data_residency_speed_cache_read_tokens',
    unit: 'token',
    conditionJson: { speed: 'fast', inference_geo: 'configured' },
  },
];

function asObject(input: unknown): StringRecord | undefined {
  if (!input || typeof input !== 'object' || Array.isArray(input)) {
    return undefined;
  }
  return input as StringRecord;
}

function asNumber(input: unknown): number | undefined {
  if (typeof input === 'number' && Number.isFinite(input)) return input;
  if (typeof input === 'string' && input.trim() !== '') {
    const value = Number(input);
    return Number.isFinite(value) ? value : undefined;
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

  return `Anthropic model allowlist validation failed for ${providerDir}: ${details.join(
    ' ',
  )}`;
}

export async function validateAnthropicProviderDirectory(
  providersRoot: string,
): Promise<AnthropicProviderValidationResult> {
  const resolvedRoot = resolve(providersRoot);
  const providerDir = resolve(resolvedRoot, ANTHROPIC_PROVIDER_SLUG);
  const relativeProviderDir = relative(resolvedRoot, providerDir);

  if (
    relativeProviderDir.startsWith('..') ||
    relativeProviderDir === '' ||
    basename(providerDir) !== ANTHROPIC_PROVIDER_SLUG
  ) {
    throw new Error(`Unsafe Anthropic provider directory: ${providerDir}`);
  }
  if (!existsSync(providerDir)) {
    throw new Error(
      `Anthropic provider directory does not exist: ${providerDir}`,
    );
  }
  if (!existsSync(join(providerDir, 'default.yaml'))) {
    throw new Error(
      `Anthropic provider metadata is missing: ${join(
        providerDir,
        'default.yaml',
      )}`,
    );
  }

  const modelFiles = await collectYamlFiles(providerDir);
  const modelIds: string[] = [];
  const invalidFiles: string[] = [];

  for (const filePath of modelFiles) {
    const document = parse(
      await fs.readFile(filePath, 'utf8'),
    ) as AnthropicYamlDocument;
    if (typeof document?.model !== 'string' || document.model.trim() === '') {
      invalidFiles.push(relative(providerDir, filePath));
      continue;
    }
    modelIds.push(document.model.trim());
  }

  const expected = new Set<string>(ANTHROPIC_SUPPORTED_MODEL_IDS);
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
    modelFiles.length !== ANTHROPIC_SUPPORTED_MODEL_IDS.length
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

function appendDerivedCostPricingRates({
  cost,
  explicitRuntimeProfile,
  pricingRates,
  region,
  sourceUrl,
}: AppendCostPricingRatesInput): void {
  const hasBatchPricing =
    asDecimalString(cost['input_cost_per_token_batches']) !== undefined &&
    asDecimalString(cost['output_cost_per_token_batches']) !== undefined;

  if (hasBatchPricing) {
    const batchCacheMappings = [
      ['cache_creation_input_token_cost', 'batch_cache_write_5m_tokens'],
      [
        'cache_creation_input_token_cost_per_hour',
        'batch_cache_write_1h_tokens',
      ],
      ['cache_read_input_token_cost', 'batch_cache_read_tokens'],
    ] as const;
    for (const [key, dimension] of batchCacheMappings) {
      const standard = asNumber(cost[key]);
      if (standard === undefined) continue;
      pricingRates.push({
        dimension,
        unit: 'token',
        unitPriceUsd: String(standard * 0.5),
        conditionJson: { api: 'message_batches' },
        region,
        sourceUrl,
      });
    }
  }

  if (
    !explicitRuntimeProfile?.requestControls.inferenceGeoValues.includes('us')
  ) {
    return;
  }

  const residencyMappings = [
    ['input_cost_per_token', 'data_residency_input_tokens'],
    ['output_cost_per_token', 'data_residency_output_tokens'],
    ['cache_creation_input_token_cost', 'data_residency_cache_write_5m_tokens'],
    [
      'cache_creation_input_token_cost_per_hour',
      'data_residency_cache_write_1h_tokens',
    ],
    ['cache_read_input_token_cost', 'data_residency_cache_read_tokens'],
    ['input_cost_per_token_batches', 'data_residency_batch_input_tokens'],
    ['output_cost_per_token_batches', 'data_residency_batch_output_tokens'],
  ] as const;
  for (const [key, dimension] of residencyMappings) {
    const standard = asNumber(cost[key]);
    if (standard === undefined) continue;
    pricingRates.push({
      dimension,
      unit: 'token',
      unitPriceUsd: String(standard * 1.1),
      conditionJson: { inference_geo: 'us' },
      region,
      sourceUrl,
    });
  }

  if (!hasBatchPricing) return;

  const residencyBatchCacheMappings = [
    [
      'cache_creation_input_token_cost',
      'data_residency_batch_cache_write_5m_tokens',
    ],
    [
      'cache_creation_input_token_cost_per_hour',
      'data_residency_batch_cache_write_1h_tokens',
    ],
    ['cache_read_input_token_cost', 'data_residency_batch_cache_read_tokens'],
  ] as const;
  for (const [key, dimension] of residencyBatchCacheMappings) {
    const standard = asNumber(cost[key]);
    if (standard === undefined) continue;
    pricingRates.push({
      dimension,
      unit: 'token',
      unitPriceUsd: String(standard * 0.5 * 1.1),
      conditionJson: {
        api: 'message_batches',
        inference_geo: 'us',
      },
      region,
      sourceUrl,
    });
  }
}

function appendToolPricingRates({
  defaultToolPricing,
  pricingRates,
  sources,
}: AppendToolPricingRatesInput): void {
  const codeExecution = asObject(defaultToolPricing?.['code_execution']);
  const unitPriceUsd = asDecimalString(codeExecution?.['per_container_hour']);
  const sourceUrl = sources.find((source) =>
    source.includes('code-execution-tool'),
  );
  if (!unitPriceUsd || !sourceUrl) return;

  pricingRates.push({
    dimension: 'code_execution_container_hours',
    unit: 'container_hour',
    unitPriceUsd,
    conditionJson: {
      free_organization_hours_per_month:
        asNumber(codeExecution?.['free_organization_hours_per_month']) ?? 1550,
      minimum_minutes: asNumber(codeExecution?.['minimum_minutes']) ?? 5,
      free_with_web_tool_min_version:
        asNumber(codeExecution?.['free_with_web_tool_min_version']) ?? 20260209,
      billable_gateway_mode:
        typeof codeExecution?.['billable_gateway_mode'] === 'string'
          ? codeExecution['billable_gateway_mode']
          : 'disabled',
      reconciliation:
        typeof codeExecution?.['reconciliation'] === 'string'
          ? codeExecution['reconciliation']
          : 'anthropic_cost_api',
      request_attribution: 'unavailable',
    },
    region: 'global',
    sourceUrl,
  });
}

export const anthropicProviderImporter = {
  slug: ANTHROPIC_PROVIDER_SLUG,
  batchPricingCondition: {
    api: 'message_batches',
  } as ProviderRateMapping['conditionJson'],
  rateMappings: ANTHROPIC_RATE_MAPPINGS,
  appendDerivedCostPricingRates,
  appendToolPricingRates,
};
