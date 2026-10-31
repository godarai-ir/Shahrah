import { existsSync, promises as fs } from 'node:fs';
import { basename, join, relative, resolve } from 'node:path';
import { parse } from 'yaml';
import { modelRuntimeProfileSchema } from '../../../src/catalog/runtime-profile.schema';
import type { ProviderPricingRate, ProviderRateMapping } from '../types';

export const XAI_PROVIDER_SLUG = 'xai';

export const XAI_SUPPORTED_MODEL_IDS = [
  'grok-4.3',
  'grok-4.5',
  'grok-imagine-image',
  'grok-imagine-image-quality',
  'grok-imagine-video',
  'grok-imagine-video-1.5',
  'grok-imagine-video-1.5-lite',
] as const;

export const XAI_RUNTIME_PROFILE_VERSIONS: Record<string, number> = {
  'grok-4.3': 4,
  'grok-4.5': 4,
  'grok-imagine-image': 4,
  'grok-imagine-image-quality': 4,
  'grok-imagine-video': 4,
  'grok-imagine-video-1.5': 4,
  'grok-imagine-video-1.5-lite': 4,
};

function isVideoModel(model: string): boolean {
  return model.startsWith('grok-imagine-video');
}

type StringRecord = Record<string, unknown>;
type XAIYamlDocument = {
  model?: unknown;
  aliases?: unknown;
  costs?: unknown;
  sources?: unknown;
  runtimeProfile?: unknown;
};

export type XAIProviderValidationResult = {
  providerDir: string;
  modelIds: string[];
  aliases: Record<string, string>;
  modelFiles: string[];
};

function asObject(value: unknown): StringRecord | undefined {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? (value as StringRecord)
    : undefined;
}

function asStrings(value: unknown): string[] {
  return Array.isArray(value)
    ? value.filter((item): item is string => typeof item === 'string')
    : [];
}

function asDecimalString(value: unknown): string | undefined {
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

function divideByThousand(value: unknown): string | undefined {
  const decimal = asDecimalString(value);
  return decimal === undefined ? undefined : String(Number(decimal) / 1000);
}

async function collectYamlFiles(dir: string): Promise<string[]> {
  const entries = await fs.readdir(dir, { withFileTypes: true });
  const files: string[] = [];
  for (const entry of entries) {
    const fullPath = join(dir, entry.name);
    if (entry.isDirectory()) files.push(...(await collectYamlFiles(fullPath)));
    else if (
      entry.isFile() &&
      entry.name.endsWith('.yaml') &&
      entry.name !== 'default.yaml'
    ) {
      files.push(fullPath);
    }
  }
  return files.sort();
}

function requiredCostKeys(model: string): string[] {
  if (model === 'grok-4.3') {
    return [
      'input_cost_per_token',
      'cache_read_input_token_cost',
      'output_cost_per_token',
      'xai_long_context_input_cost_per_token',
      'xai_long_context_cache_read_input_token_cost',
      'xai_long_context_output_cost_per_token',
      'input_cost_per_token_batches',
      'cache_read_input_token_cost_batches',
      'output_cost_per_token_batches',
      'xai_batch_long_context_input_cost_per_token',
      'xai_batch_long_context_cache_read_input_token_cost',
      'xai_batch_long_context_output_cost_per_token',
    ];
  }
  if (model === 'grok-4.5') {
    return [
      'input_cost_per_token',
      'cache_read_input_token_cost',
      'output_cost_per_token',
      'xai_long_context_input_cost_per_token',
      'xai_long_context_cache_read_input_token_cost',
      'xai_long_context_output_cost_per_token',
    ];
  }
  if (isVideoModel(model)) return ['output_cost_per_second'];
  return [
    'input_cost_per_image',
    'output_cost_per_image_1k',
    'output_cost_per_image_2k',
  ];
}

export async function validateXAIProviderDirectory(
  providersRoot: string,
): Promise<XAIProviderValidationResult> {
  const root = resolve(providersRoot);
  const providerDir = resolve(root, XAI_PROVIDER_SLUG);
  if (
    relative(root, providerDir).startsWith('..') ||
    basename(providerDir) !== XAI_PROVIDER_SLUG
  ) {
    throw new Error(`Unsafe xAI provider directory: ${providerDir}`);
  }
  const metadataPath = join(providerDir, 'default.yaml');
  if (!existsSync(providerDir) || !existsSync(metadataPath)) {
    throw new Error(`xAI provider metadata is missing: ${metadataPath}`);
  }
  const metadata = asObject(parse(await fs.readFile(metadataPath, 'utf8')));
  const documentation = asStrings(metadata?.documentation);
  if (
    documentation.length === 0 ||
    documentation.some((url) => !url.startsWith('https://docs.x.ai/'))
  ) {
    throw new Error(
      'xAI provider metadata must use official docs.x.ai sources',
    );
  }

  const files = await collectYamlFiles(providerDir);
  const modelIds: string[] = [];
  const aliases: Record<string, string> = {};
  const invalid: string[] = [];
  for (const file of files) {
    const relativeFile = relative(providerDir, file);
    let document: XAIYamlDocument | undefined;
    try {
      document = asObject(parse(await fs.readFile(file, 'utf8')));
    } catch {
      invalid.push(`${relativeFile}(malformed_yaml)`);
      continue;
    }
    const model =
      typeof document?.model === 'string' ? document.model.trim() : '';
    if (!model || basename(file, '.yaml') !== model) {
      invalid.push(`${relativeFile}(identity)`);
      continue;
    }
    modelIds.push(model);
    const profile = modelRuntimeProfileSchema.safeParse(
      document?.runtimeProfile,
    );
    if (!profile.success) {
      invalid.push(`${relativeFile}(runtime_profile)`);
      continue;
    }
    if (profile.data.version !== XAI_RUNTIME_PROFILE_VERSIONS[model]) {
      invalid.push(`${relativeFile}(runtime_profile_version)`);
      continue;
    }
    const declaredAliases = asStrings(document?.aliases);
    if (
      declaredAliases.length !== profile.data.aliases.length ||
      declaredAliases.some((alias) => !profile.data.aliases.includes(alias))
    ) {
      invalid.push(`${relativeFile}(alias_profile_mismatch)`);
    }
    for (const alias of declaredAliases) {
      if (
        !alias.trim() ||
        aliases[alias] ||
        alias === modelIds.find((id) => id === alias)
      ) {
        invalid.push(`${relativeFile}(duplicate_alias:${alias})`);
      } else {
        aliases[alias] = model;
      }
    }
    const sources = asStrings(document?.sources);
    if (
      sources.length === 0 ||
      sources.some((url) => !url.startsWith('https://docs.x.ai/'))
    ) {
      invalid.push(`${relativeFile}(sources)`);
    }
    const costs = Array.isArray(document?.costs) ? document.costs : [];
    const cost = asObject(costs[0]);
    const missingCosts = requiredCostKeys(model).filter(
      (key) => asDecimalString(cost?.[key]) === undefined,
    );
    if (costs.length !== 1 || missingCosts.length > 0) {
      invalid.push(`${relativeFile}(pricing:${missingCosts.join(',')})`);
    }
    // Video jobs are not offered through xAI batches.
    if (
      profile.success &&
      !isVideoModel(model) &&
      !profile.data.batch.eligible
    ) {
      invalid.push(`${relativeFile}(batch_required)`);
    }
    if (
      profile.success &&
      profile.data.batch.discounted !== (model === 'grok-4.3')
    ) {
      invalid.push(`${relativeFile}(batch_discount)`);
    }
  }

  const expected = new Set<string>(XAI_SUPPORTED_MODEL_IDS);
  const actual = new Set(modelIds);
  const missing = [...expected].filter((model) => !actual.has(model));
  const unexpected = [...actual].filter((model) => !expected.has(model));
  const duplicates = [...actual].filter(
    (model) => modelIds.filter((candidate) => candidate === model).length > 1,
  );
  for (const alias of Object.keys(aliases)) {
    if (actual.has(alias)) invalid.push(`alias_collision:${alias}`);
  }
  if (aliases['grok-imagine-image-pro'] !== 'grok-imagine-image-quality') {
    invalid.push('grok-imagine-image-pro(alias_target)');
  }
  if (
    missing.length ||
    unexpected.length ||
    duplicates.length ||
    invalid.length ||
    files.length !== XAI_SUPPORTED_MODEL_IDS.length
  ) {
    throw new Error(
      `xAI model allowlist validation failed for ${providerDir}: ` +
        [
          missing.length ? `missing=${missing.join(',')}` : '',
          unexpected.length ? `unexpected=${unexpected.join(',')}` : '',
          duplicates.length ? `duplicates=${duplicates.join(',')}` : '',
          invalid.length ? `invalid_files=${invalid.join(',')}` : '',
        ]
          .filter(Boolean)
          .join(' '),
    );
  }
  return {
    providerDir,
    modelIds: [...actual].sort(),
    aliases,
    modelFiles: files,
  };
}

const XAI_RATE_MAPPINGS: ProviderRateMapping[] = [
  ['xai_long_context_input_cost_per_token', 'long_context_input_tokens'],
  [
    'xai_long_context_cache_read_input_token_cost',
    'long_context_cache_read_tokens',
  ],
  ['xai_long_context_output_cost_per_token', 'long_context_output_tokens'],
].map(([key, dimension]) => ({
  key,
  dimension,
  unit: 'token',
  conditionJson: { min_prompt_tokens: 200000 },
}));

XAI_RATE_MAPPINGS.push(
  ...[
    [
      'xai_batch_long_context_input_cost_per_token',
      'batch_long_context_input_tokens',
    ],
    [
      'xai_batch_long_context_cache_read_input_token_cost',
      'batch_long_context_cache_read_tokens',
    ],
    [
      'xai_batch_long_context_output_cost_per_token',
      'batch_long_context_output_tokens',
    ],
  ].map(([key, dimension]) => ({
    key,
    dimension,
    unit: 'token',
    conditionJson: { api: 'batches', min_prompt_tokens: 200000 },
  })),
  {
    key: 'output_cost_per_image_1k',
    dimension: 'output_images',
    unit: 'image',
    conditionJson: { resolution: '1k' },
  },
  {
    key: 'output_cost_per_image_2k',
    dimension: 'output_images',
    unit: 'image',
    conditionJson: { resolution: '2k' },
  },
  ...[
    [
      'xai_priority_long_context_input_cost_per_token',
      'speed_long_context_input_tokens',
    ],
    [
      'xai_priority_long_context_cache_read_input_token_cost',
      'speed_long_context_cache_read_tokens',
    ],
    [
      'xai_priority_long_context_output_cost_per_token',
      'speed_long_context_output_tokens',
    ],
  ].map(([key, dimension]) => ({
    key,
    dimension,
    unit: 'token',
    conditionJson: { service_tier: 'priority', min_prompt_tokens: 200000 },
  })),
);

function appendToolPricingRates(input: {
  defaultToolPricing?: StringRecord;
  pricingRates: ProviderPricingRate[];
  sources: string[];
  documentedTools?: Set<string>;
}): void {
  const sourceUrl = input.sources.find((source) => source.includes('/pricing'));
  if (!sourceUrl) return;
  const mappings = [
    ['web_search', 'web_search_requests'],
    ['x_search', 'x_search_requests'],
    ['code_execution', 'code_execution_requests'],
    ['attachment_search', 'attachment_search_requests'],
    ['collections_search', 'collections_search_requests'],
  ] as const;
  for (const [key, dimension] of mappings) {
    if (!input.documentedTools?.has(key)) continue;
    const tool = asObject(input.defaultToolPricing?.[key]);
    const price = divideByThousand(tool?.per_thousand_requests);
    if (!price) continue;
    input.pricingRates.push({
      dimension,
      unit: 'request',
      unitPriceUsd: price,
      region: 'global',
      sourceUrl,
    });
  }
}

export const xaiProviderImporter = {
  slug: XAI_PROVIDER_SLUG,
  batchPricingCondition: {
    api: 'batches',
  } as ProviderRateMapping['conditionJson'],
  rateMappings: XAI_RATE_MAPPINGS,
  appendDerivedCostPricingRates: undefined,
  appendToolPricingRates,
};
