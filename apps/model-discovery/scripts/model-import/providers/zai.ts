import { existsSync, promises as fs } from 'node:fs';
import { basename, join, relative, resolve } from 'node:path';
import { parse } from 'yaml';
import { modelRuntimeProfileSchema } from '../../../src/catalog/runtime-profile.schema';

export const ZAI_PROVIDER_SLUG = 'zai';
export const ZAI_SUPPORTED_MODEL_IDS = [
  'glm-5.2',
  'glm-4.7-flashx',
  'glm-5v-turbo',
  'glm-ocr',
  'glm-5.3-flash',
] as const;

const ZAI_MODEL_EXPECTATIONS: Readonly<
  Record<
    string,
    {
      endpoint: 'chat_completions' | 'layout_parsing';
      input: number;
      cached?: number;
      output: number;
      contextWindow: number;
      maxOutputTokens: number;
      verifiedAt: string;
    }
  >
> = {
  'glm-5.2': {
    endpoint: 'chat_completions',
    input: 0.0000014,
    cached: 0.00000026,
    output: 0.0000044,
    contextWindow: 1_000_000,
    maxOutputTokens: 128_000,
    verifiedAt: '2026-08-17',
  },
  'glm-4.7-flashx': {
    endpoint: 'chat_completions',
    input: 0.00000007,
    cached: 0.00000001,
    output: 0.0000004,
    contextWindow: 200_000,
    maxOutputTokens: 128_000,
    verifiedAt: '2026-08-17',
  },
  'glm-5v-turbo': {
    endpoint: 'chat_completions',
    input: 0.0000012,
    cached: 0.00000024,
    output: 0.000004,
    contextWindow: 200_000,
    maxOutputTokens: 128_000,
    verifiedAt: '2026-08-17',
  },
  'glm-ocr': {
    endpoint: 'layout_parsing',
    input: 0.00000003,
    output: 0.00000003,
    contextWindow: 0,
    maxOutputTokens: 0,
    verifiedAt: '2026-08-17',
  },
  'glm-5.3-flash': {
    endpoint: 'chat_completions',
    input: 0.00000015,
    cached: 0.00000003,
    output: 0.0000005,
    contextWindow: 1_000_000,
    maxOutputTokens: 128_000,
    verifiedAt: '2026-09-11',
  },
};

type StringRecord = Record<string, unknown>;

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

function isOfficialSource(value: string): boolean {
  try {
    return new URL(value).hostname.toLowerCase() === 'docs.z.ai';
  } catch {
    return false;
  }
}

export async function validateZAIProviderDirectory(
  providersRoot: string,
): Promise<{ providerDir: string; modelIds: string[]; modelFiles: string[] }> {
  const root = resolve(providersRoot);
  const providerDir = resolve(root, ZAI_PROVIDER_SLUG);
  if (
    relative(root, providerDir).startsWith('..') ||
    basename(providerDir) !== ZAI_PROVIDER_SLUG
  ) {
    throw new Error(`Unsafe Z.ai provider directory: ${providerDir}`);
  }

  const defaultPath = join(providerDir, 'default.yaml');
  if (!existsSync(providerDir) || !existsSync(defaultPath)) {
    throw new Error(`Z.ai provider metadata is missing: ${defaultPath}`);
  }

  const defaults = asObject(parse(await fs.readFile(defaultPath, 'utf8')));
  const documentation = asStrings(defaults?.documentation);
  if (
    documentation.length === 0 ||
    documentation.some((source) => !isOfficialSource(source))
  ) {
    throw new Error('Z.ai provider metadata must use official Z.ai sources');
  }
  const manifest = asObject(defaults?.source_manifest);
  if (
    manifest?.openapi_sha256 !==
      '76f83c8525f993e33c79c5483e4cdde7db0924b5c5e5d7480b9b574ce909f4da' ||
    manifest?.llms_sha256 !==
      '60b34406d17fe6804b7c719e96886a8ad30082ada8b2187f85dcbc739cc7c25b'
  ) {
    throw new Error('Z.ai provider source manifest is missing or unreviewed');
  }

  const entries = await fs.readdir(providerDir, { withFileTypes: true });
  const files = entries
    .filter(
      (entry) =>
        entry.isFile() &&
        entry.name.endsWith('.yaml') &&
        entry.name !== 'default.yaml',
    )
    .map((entry) => join(providerDir, entry.name))
    .sort();
  const modelIds: string[] = [];
  const invalid: string[] = [];

  for (const file of files) {
    const relativeFile = relative(providerDir, file);
    const document = asObject(parse(await fs.readFile(file, 'utf8')));
    const model =
      typeof document?.model === 'string' ? document.model.trim() : '';
    if (!model || basename(file, '.yaml') !== model || !document) {
      invalid.push(`${relativeFile}(identity)`);
      continue;
    }

    modelIds.push(model);
    const expectation = ZAI_MODEL_EXPECTATIONS[model];
    if (!expectation || document.verified_at !== expectation.verifiedAt) {
      invalid.push(`${relativeFile}(verified_at)`);
    }

    const sources = asStrings(document.sources);
    if (
      sources.length === 0 ||
      sources.some((source) => !isOfficialSource(source))
    ) {
      invalid.push(`${relativeFile}(sources)`);
    }

    const profile = modelRuntimeProfileSchema.safeParse(
      document.runtimeProfile,
    );
    if (!profile.success || profile.data.version !== 1) {
      invalid.push(`${relativeFile}(runtime_profile)`);
    } else if (
      !expectation ||
      profile.data.endpoints.length !== 1 ||
      profile.data.endpoints[0] !== expectation.endpoint ||
      profile.data.limits.contextWindow !== expectation.contextWindow ||
      profile.data.limits.maxOutputTokens !== expectation.maxOutputTokens
    ) {
      invalid.push(`${relativeFile}(runtime_profile_facts)`);
    }

    const costs = Array.isArray(document.costs) ? document.costs : [];
    const cost = asObject(costs[0]);
    if (
      !expectation ||
      cost?.input_cost_per_token !== expectation.input ||
      cost?.output_cost_per_token !== expectation.output ||
      (expectation.cached !== undefined &&
        cost?.cache_read_input_token_cost !== expectation.cached)
    ) {
      invalid.push(`${relativeFile}(pricing_facts)`);
    }
    for (const key of ['input_cost_per_token', 'output_cost_per_token']) {
      if (
        typeof cost?.[key] !== 'number' ||
        !Number.isFinite(cost[key]) ||
        cost[key] <= 0
      ) {
        invalid.push(`${relativeFile}(pricing:${key})`);
      }
    }
    if (model !== 'glm-ocr') {
      const cached = cost?.cache_read_input_token_cost;
      if (
        typeof cached !== 'number' ||
        !Number.isFinite(cached) ||
        cached <= 0
      ) {
        invalid.push(`${relativeFile}(pricing:cache_read_input_token_cost)`);
      }
    }
    if (costs.length !== 1) invalid.push(`${relativeFile}(pricing_rows)`);
  }

  const expected = new Set<string>(ZAI_SUPPORTED_MODEL_IDS);
  const actual = new Set(modelIds);
  const missing = [...expected].filter((model) => !actual.has(model));
  const unexpected = [...actual].filter((model) => !expected.has(model));
  const duplicates = [...actual].filter(
    (model) => modelIds.filter((candidate) => candidate === model).length > 1,
  );
  if (
    files.length !== ZAI_SUPPORTED_MODEL_IDS.length ||
    missing.length > 0 ||
    unexpected.length > 0 ||
    duplicates.length > 0 ||
    invalid.length > 0
  ) {
    throw new Error(
      `Z.ai model allowlist validation failed: ${[
        missing.length ? `missing=${missing.join(',')}` : '',
        unexpected.length ? `unexpected=${unexpected.join(',')}` : '',
        duplicates.length ? `duplicates=${duplicates.join(',')}` : '',
        invalid.length ? `invalid=${invalid.join(',')}` : '',
      ]
        .filter(Boolean)
        .join(' ')}`,
    );
  }

  return { providerDir, modelIds: [...actual].sort(), modelFiles: files };
}

export const zaiProviderImporter = {
  slug: ZAI_PROVIDER_SLUG,
  batchPricingCondition: undefined,
  rateMappings: [],
  appendDerivedCostPricingRates: undefined,
  appendToolPricingRates: undefined,
};
