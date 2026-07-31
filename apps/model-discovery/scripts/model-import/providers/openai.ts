import { existsSync, promises as fs } from 'node:fs';
import { basename, join, relative, resolve } from 'node:path';
import { parse } from 'yaml';
import { modelRuntimeProfileSchema } from '../../../src/catalog/runtime-profile.schema';

export const OPENAI_PROVIDER_SLUG = 'openai';

export const OPENAI_SUPPORTED_MODEL_IDS = [
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

type OpenAIYamlDocument = {
  model?: unknown;
  runtimeProfile?: unknown;
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
