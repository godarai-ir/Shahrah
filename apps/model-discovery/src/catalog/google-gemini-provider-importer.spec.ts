import { cp, mkdtemp, mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { parse } from 'yaml';
import { modelRuntimeProfileSchema } from './runtime-profile.schema';
import {
  GOOGLE_GEMINI_SUPPORTED_MODEL_IDS,
  googleGeminiProviderImporter,
  parseGoogleGeminiImportArgs,
  validateGoogleGeminiProviderDirectory,
} from '../../scripts/model-import/providers/google-gemini';

describe('Google Gemini provider importer', () => {
  it('validates the curated directory and exact allowlist', async () => {
    const result = await validateGoogleGeminiProviderDirectory(
      join(__dirname, '..', '..', 'providers'),
    );
    expect(result.modelIds).toEqual(
      [...GOOGLE_GEMINI_SUPPORTED_MODEL_IDS].sort(),
    );
    expect(result.modelFiles).toHaveLength(
      GOOGLE_GEMINI_SUPPORTED_MODEL_IDS.length,
    );
  });

  it('keeps official endpoint, limit, sampling, tier and quota evidence in profiles', async () => {
    const providerDir = join(
      __dirname,
      '..',
      '..',
      'providers',
      'google-gemini',
    );
    const flashImage = parse(
      await readFile(join(providerDir, 'gemini-3.1-flash-image.yaml'), 'utf8'),
    ) as GeminiProfileFixture;
    expect(flashImage.limits).toMatchObject({
      context_window: 131072,
      max_input_tokens: 131072,
      max_output_tokens: 32768,
    });
    expect(flashImage.modalities?.input).toContain('video');
    expect(flashImage.endpoints).not.toContain('live');

    for (const model of [
      'gemini-3.7-flash',
      'gemini-3.6-flash',
      'gemini-3.5-flash-lite',
    ]) {
      const profile = parse(
        await readFile(join(providerDir, `${model}.yaml`), 'utf8'),
      ) as GeminiProfileFixture;
      expect(profile.removeParams).toEqual(
        expect.arrayContaining(['temperature', 'top_p', 'top_k']),
      );
      expect(profile.serviceTiers).toEqual([
        'standard',
        'batch',
        'flex',
        'priority',
      ]);
      expect(profile.rateLimits).toHaveLength(4);
      expect(profile.rateLimits?.every((limit) => limit.dynamic)).toBe(true);
      expect(
        profile.rateLimits?.every((limit) => limit.scope === 'project'),
      ).toBe(true);
    }
  });

  it('imports the Gemini 3.7 Flash contract and scheduled pricing', async () => {
    const document = parse(
      await readFile(
        join(
          __dirname,
          '..',
          '..',
          'providers',
          'google-gemini',
          'gemini-3.7-flash.yaml',
        ),
        'utf8',
      ),
    ) as GeminiProfileFixture;

    expect(document.status).toBe('active');
    expect(document.limits).toMatchObject({
      context_window: 1048576,
      max_input_tokens: 1048576,
      max_output_tokens: 65536,
    });
    expect(document.thinkingProfile?.effortValues).toEqual([
      'low',
      'medium',
      'high',
    ]);
    expect(document.removeParams).toEqual(['temperature', 'top_p', 'top_k']);
    expect(document.costs).toHaveLength(2);
    expect(document.costs?.[0]).toMatchObject({
      effective_from: '2026-08-13T00:00:00Z',
      effective_to: '2027-01-01T00:00:00Z',
      input_cost_per_token: 0.00000075,
      output_cost_per_token: 0.00000375,
      cache_read_input_token_cost: 0.000000075,
    });
    expect(document.costs?.[1]).toMatchObject({
      effective_from: '2027-01-01T00:00:00Z',
      input_cost_per_token: 0.0000015,
      output_cost_per_token: 0.0000075,
      cache_read_input_token_cost: 0.00000015,
    });
  });

  it('imports the stable Gemini 2.5 contracts from official Google sources', async () => {
    const providerDir = join(
      __dirname,
      '..',
      '..',
      'providers',
      'google-gemini',
    );
    const profiles = await Promise.all(
      ['gemini-2.5-flash-lite', 'gemini-2.5-flash', 'gemini-2.5-pro'].map(
        async (model) =>
          parse(
            await readFile(join(providerDir, `${model}.yaml`), 'utf8'),
          ) as GeminiProfileFixture,
      ),
    );

    for (const profile of profiles) {
      expect(profile.status).toBe('active');
      expect(profile.limits).toMatchObject({
        context_window: 1048576,
        max_input_tokens: 1048576,
        max_output_tokens: 65536,
      });
      expect(profile.endpoints).toEqual(
        expect.arrayContaining(['interactions', 'generate_content', 'batches']),
      );
      expect(profile.serviceTiers).toEqual([
        'standard',
        'batch',
        'flex',
        'priority',
      ]);
      expect(profile.sources?.length).toBeGreaterThan(0);
      expect(
        profile.sources?.every((source) =>
          source.startsWith('https://ai.google.dev/'),
        ),
      ).toBe(true);
    }

    expect(profiles[0].thinkingProfile).toMatchObject({
      defaultMode: 'disabled',
      allowedModes: ['disabled', 'adaptive', 'manual'],
      effortValues: ['low', 'medium', 'high'],
    });
    expect(profiles[1].thinkingProfile).toMatchObject({
      defaultMode: 'adaptive',
      allowedModes: ['disabled', 'adaptive', 'manual'],
    });
    expect(profiles[2].thinkingProfile).toMatchObject({
      defaultMode: 'adaptive',
      allowedModes: ['adaptive', 'manual'],
    });
    expect(profiles[0].costs?.[0]).toMatchObject({
      input_cost_per_token: 0.0000001,
      input_cost_per_audio_token: 0.0000003,
      output_cost_per_token: 0.0000004,
    });
    expect(profiles[1].costs?.[0]).toMatchObject({
      input_cost_per_token: 0.0000003,
      input_cost_per_audio_token: 0.000001,
      output_cost_per_token: 0.0000025,
    });
    expect(profiles[2].costs?.[0]).toMatchObject({
      input_cost_per_token: 0.00000125,
      output_cost_per_token: 0.00001,
    });
  });

  it('maps Gemini modality-specific audio rates for every consumption tier', () => {
    expect(googleGeminiProviderImporter.rateMappings).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          key: 'cache_read_input_audio_token_cost',
          dimension: 'audio_cache_read_tokens',
        }),
        expect.objectContaining({
          key: 'input_cost_per_audio_token_batches',
          dimension: 'batch_audio_input_tokens',
        }),
        expect.objectContaining({
          key: 'input_cost_per_audio_token_flex',
          dimension: 'flex_audio_input_tokens',
        }),
        expect.objectContaining({
          key: 'input_cost_per_audio_token_priority',
          dimension: 'priority_audio_input_tokens',
        }),
      ]),
    );
  });

  it('requires an explicit versioned runtime profile matching every YAML contract', async () => {
    const providerDir = join(
      __dirname,
      '..',
      '..',
      'providers',
      'google-gemini',
    );
    for (const model of GOOGLE_GEMINI_SUPPORTED_MODEL_IDS) {
      const document = parse(
        await readFile(join(providerDir, `${model}.yaml`), 'utf8'),
      ) as GeminiProfileFixture;
      const profile = modelRuntimeProfileSchema.parse(document.runtimeProfile);
      expect(profile.version).toBe(2);
      expect(profile.endpoints).toEqual(document.endpoints);
      expect(profile.transports).toEqual(document.transports);
      expect(profile.lifecycle).toBe(document.status);
      expect(profile.modalities).toEqual(document.modalities);
      expect(profile.tools.supported).toEqual(document.tools);
      expect(profile.requestControls.serviceTierValues).toEqual(
        document.serviceTiers,
      );
      expect(profile.referenceRateLimits).toEqual(document.rateLimits);
      expect(profile.limits).toMatchObject({
        contextWindow: document.limits?.context_window,
        maxInputTokens: document.limits?.max_input_tokens,
        maxOutputTokens: document.limits?.max_output_tokens ?? 0,
      });
    }
  });

  it('rejects a curated model whose explicit runtime profile is missing', async () => {
    const root = await mkdtemp(join(tmpdir(), 'gemini-import-'));
    const source = join(__dirname, '..', '..', 'providers', 'google-gemini');
    const providerDir = join(root, 'google-gemini');
    await cp(source, providerDir, { recursive: true });
    const modelPath = join(providerDir, 'gemini-embedding-2.yaml');
    const yaml = await readFile(modelPath, 'utf8');
    await writeFile(modelPath, yaml.replace(/\nruntimeProfile:[\s\S]*$/, '\n'));
    await expect(validateGoogleGeminiProviderDirectory(root)).rejects.toThrow(
      'invalid_runtime_profile',
    );
    await rm(root, { recursive: true, force: true });
  });

  it.each([
    [
      'missing model',
      async (dir: string) => rm(join(dir, 'gemini-embedding-2.yaml')),
    ],
    [
      'unexpected model',
      async (dir: string) =>
        writeFile(join(dir, 'gemini-unknown.yaml'), 'model: gemini-unknown\n'),
    ],
    [
      'malformed YAML',
      async (dir: string) =>
        writeFile(join(dir, 'gemini-embedding-2.yaml'), 'model: [\n'),
    ],
    [
      'filename mismatch',
      async (dir: string) =>
        writeFile(
          join(dir, 'gemini-embedding-2.yaml'),
          'model: gemini-3.6-flash\n',
        ),
    ],
  ])('rejects %s', async (_name, mutate) => {
    const root = await mkdtemp(join(tmpdir(), 'gemini-import-'));
    const providerDir = join(root, 'google-gemini');
    await mkdir(providerDir);
    await writeFile(join(providerDir, 'default.yaml'), 'documentation: []\n');
    for (const model of GOOGLE_GEMINI_SUPPORTED_MODEL_IDS) {
      await writeFile(join(providerDir, `${model}.yaml`), `model: ${model}\n`);
    }
    await mutate(providerDir);
    await expect(validateGoogleGeminiProviderDirectory(root)).rejects.toThrow(
      'Google Gemini model allowlist validation failed',
    );
    await rm(root, { recursive: true, force: true });
  });

  it('rejects a missing metadata document', async () => {
    const root = await mkdtemp(join(tmpdir(), 'gemini-import-'));
    await mkdir(join(root, 'google-gemini'));
    await expect(validateGoogleGeminiProviderDirectory(root)).rejects.toThrow(
      'provider metadata is missing',
    );
    await rm(root, { recursive: true, force: true });
  });

  it('rejects duplicate model IDs across nested YAML files', async () => {
    const root = await createValidFixture();
    const nested = join(root, 'google-gemini', 'nested');
    await mkdir(nested);
    await writeFile(
      join(nested, 'gemini-embedding-2.yaml'),
      'model: gemini-embedding-2\n',
    );
    await expect(validateGoogleGeminiProviderDirectory(root)).rejects.toThrow(
      'duplicates=gemini-embedding-2',
    );
    await rm(root, { recursive: true, force: true });
  });

  it('rejects an empty model ID', async () => {
    const root = await createValidFixture();
    await writeFile(
      join(root, 'google-gemini', 'gemini-embedding-2.yaml'),
      'model: ""\n',
    );
    await expect(validateGoogleGeminiProviderDirectory(root)).rejects.toThrow(
      'invalid_files=',
    );
    await rm(root, { recursive: true, force: true });
  });

  it('parses dry-run and clear-missing without accepting provider overrides', () => {
    expect(
      parseGoogleGeminiImportArgs([
        '--providers-root',
        'D:/catalog/providers',
        '--dry-run',
        '--clear-missing',
      ]),
    ).toEqual({
      providersRoot: 'D:/catalog/providers',
      dryRun: true,
      clearMissing: true,
    });
    expect(() =>
      parseGoogleGeminiImportArgs(['--providers', 'openai']),
    ).toThrow('locked to provider=google-gemini');
    expect(() => parseGoogleGeminiImportArgs(['--providers-root'])).toThrow(
      'requires a path value',
    );
  });
});

async function createValidFixture(): Promise<string> {
  const root = await mkdtemp(join(tmpdir(), 'gemini-import-'));
  const providerDir = join(root, 'google-gemini');
  await mkdir(providerDir);
  await writeFile(join(providerDir, 'default.yaml'), 'documentation: []\n');
  for (const model of GOOGLE_GEMINI_SUPPORTED_MODEL_IDS) {
    await writeFile(join(providerDir, `${model}.yaml`), `model: ${model}\n`);
  }
  return root;
}

type GeminiProfileFixture = {
  status?: string;
  limits?: Record<string, number>;
  modalities?: { input?: string[]; output?: string[] };
  endpoints?: string[];
  transports?: string[];
  tools?: string[];
  removeParams?: string[];
  serviceTiers?: string[];
  rateLimits?: Array<{ dynamic?: boolean; scope?: string }>;
  thinkingProfile?: {
    defaultMode?: string;
    allowedModes?: string[];
    effortValues?: string[];
  };
  costs?: Array<Record<string, unknown>>;
  sources?: string[];
  runtimeProfile?: unknown;
};
