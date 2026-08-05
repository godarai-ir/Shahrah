import { existsSync, promises as fs } from 'node:fs';
import { basename, join, relative, resolve } from 'node:path';
import { parse } from 'yaml';
import { modelRuntimeProfileSchema } from '../../../src/catalog/runtime-profile.schema';

export const DEEPSEEK_PROVIDER_SLUG = 'deepseek';
export const DEEPSEEK_SUPPORTED_MODEL_IDS = [
  'deepseek-v4-flash',
  'deepseek-v4-pro',
] as const;

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
    return new URL(value).hostname.toLowerCase() === 'api-docs.deepseek.com';
  } catch {
    return false;
  }
}

export async function validateDeepSeekProviderDirectory(
  providersRoot: string,
): Promise<{ providerDir: string; modelIds: string[]; modelFiles: string[] }> {
  const root = resolve(providersRoot);
  const providerDir = resolve(root, DEEPSEEK_PROVIDER_SLUG);
  if (
    relative(root, providerDir).startsWith('..') ||
    basename(providerDir) !== DEEPSEEK_PROVIDER_SLUG
  ) {
    throw new Error(`Unsafe DeepSeek provider directory: ${providerDir}`);
  }

  const defaultPath = join(providerDir, 'default.yaml');
  if (!existsSync(providerDir) || !existsSync(defaultPath)) {
    throw new Error(`DeepSeek provider metadata is missing: ${defaultPath}`);
  }

  const defaults = asObject(parse(await fs.readFile(defaultPath, 'utf8')));
  const documentation = asStrings(defaults?.documentation);
  if (
    documentation.length === 0 ||
    documentation.some((source) => !isOfficialSource(source))
  ) {
    throw new Error(
      'DeepSeek provider metadata must use official DeepSeek sources',
    );
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
    if (!profile.success || profile.data.version < 1) {
      invalid.push(`${relativeFile}(runtime_profile)`);
    } else {
      const modes = asStrings(document.supportedModes);
      const responsesEnabled =
        modes.includes('responses') &&
        profile.data.endpoints.includes('responses') &&
        !profile.data.responseResources.store &&
        !profile.data.responseResources.retrieve &&
        !profile.data.responseResources.delete &&
        !profile.data.responseResources.compact &&
        !profile.data.responseResources.countInputTokens;

      if (
        (model === 'deepseek-v4-flash' && !responsesEnabled) ||
        (model === 'deepseek-v4-pro' &&
          (modes.includes('responses') ||
            profile.data.endpoints.includes('responses')))
      ) {
        invalid.push(`${relativeFile}(responses_scope)`);
      }
    }

    const costs = Array.isArray(document.costs) ? document.costs : [];
    const cost = asObject(costs[0]);
    for (const key of [
      'input_cost_per_token',
      'cache_read_input_token_cost',
      'output_cost_per_token',
    ]) {
      if (
        typeof cost?.[key] !== 'number' ||
        !Number.isFinite(cost[key]) ||
        cost[key] <= 0
      ) {
        invalid.push(`${relativeFile}(pricing:${key})`);
      }
    }
    if (costs.length !== 1) invalid.push(`${relativeFile}(pricing_rows)`);
  }

  const expected = new Set<string>(DEEPSEEK_SUPPORTED_MODEL_IDS);
  const actual = new Set(modelIds);
  const missing = [...expected].filter((model) => !actual.has(model));
  const unexpected = [...actual].filter((model) => !expected.has(model));
  const duplicates = [...actual].filter(
    (model) => modelIds.filter((candidate) => candidate === model).length > 1,
  );
  if (
    files.length !== DEEPSEEK_SUPPORTED_MODEL_IDS.length ||
    missing.length > 0 ||
    unexpected.length > 0 ||
    duplicates.length > 0 ||
    invalid.length > 0
  ) {
    throw new Error(
      `DeepSeek model allowlist validation failed: ${[
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
