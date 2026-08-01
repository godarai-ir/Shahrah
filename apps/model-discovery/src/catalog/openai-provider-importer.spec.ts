import { cp, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { parse } from 'yaml';
import {
  OPENAI_SUPPORTED_MODEL_IDS,
  validateOpenAIProviderDirectory,
} from '../../scripts/model-import/providers/openai';
import { modelRuntimeProfileSchema } from './runtime-profile.schema';

describe('OpenAI provider importer', () => {
  const providerDir = join(__dirname, '..', '..', 'providers', 'openai');

  it('validates the exact curated model set', async () => {
    const result = await validateOpenAIProviderDirectory(
      join(__dirname, '..', '..', 'providers'),
    );
    expect(result.modelIds).toEqual([...OPENAI_SUPPORTED_MODEL_IDS].sort());
    expect(result.modelFiles).toHaveLength(OPENAI_SUPPORTED_MODEL_IDS.length);
  });

  it('requires an explicit immutable v3 profile matching each YAML contract', async () => {
    for (const model of OPENAI_SUPPORTED_MODEL_IDS) {
      const document = parse(
        await readFile(join(providerDir, `${model}.yaml`), 'utf8'),
      ) as OpenAIProfileFixture;
      const profile = modelRuntimeProfileSchema.parse(document.runtimeProfile);
      expect(profile.version).toBe(3);
      expect(profile.lifecycle).toBe(document.status);
      expect(profile.modalities).toEqual(document.modalities);
      expect(profile.tools.supported).toEqual(document.tools ?? []);
      expect(profile.requestControls.serviceTierValues).toEqual(
        document.serviceTiers ?? [],
      );
      const referenceRateLimits = document.rateLimits ?? [];
      expect(profile.referenceRateLimits).toHaveLength(
        referenceRateLimits.length,
      );
      referenceRateLimits.forEach((limits, index) => {
        expect(profile.referenceRateLimits[index]).toMatchObject(limits);
      });
      expect(profile.limits).toMatchObject({
        contextWindow: document.limits?.context_window ?? 0,
        maxInputTokens: document.limits?.max_input_tokens ?? 0,
        maxInputCharacters: document.limits?.max_input_characters ?? 0,
        maxInputBytes: document.limits?.max_input_bytes ?? 0,
        maxOutputTokens:
          document.limits?.max_output_tokens ??
          document.limits?.max_tokens ??
          0,
      });
    }
  });

  it('keeps model-specific endpoint and transport contracts explicit', async () => {
    const expected: Record<
      string,
      { endpoints: string[]; transports: string[] }
    > = {
      'gpt-5.6-sol': {
        endpoints: ['chat_completions', 'responses', 'batches'],
        transports: ['http', 'sse'],
      },
      'gpt-5.5-pro': {
        endpoints: ['responses', 'batches'],
        transports: ['http'],
      },
      'gpt-image-2': {
        endpoints: ['image_generations', 'image_edits', 'batches'],
        transports: ['http'],
      },
      'gpt-transcribe': {
        endpoints: ['audio_transcriptions'],
        transports: ['http', 'sse'],
      },
      'tts-1': {
        endpoints: ['audio_speech'],
        transports: ['http'],
      },
      'text-embedding-3-large': {
        endpoints: ['embeddings', 'batches'],
        transports: ['http'],
      },
      'omni-moderation-latest': {
        endpoints: ['moderations'],
        transports: ['http'],
      },
    };
    for (const [model, contract] of Object.entries(expected)) {
      const document = parse(
        await readFile(join(providerDir, `${model}.yaml`), 'utf8'),
      ) as OpenAIProfileFixture;
      const profile = modelRuntimeProfileSchema.parse(document.runtimeProfile);
      expect(profile.endpoints).toEqual(
        expect.arrayContaining(contract.endpoints),
      );
      expect(profile.transports).toEqual(contract.transports);
    }
  });

  it('rejects a missing or downgraded runtime profile', async () => {
    const root = await copyProviderFixture();
    const modelPath = join(root, 'openai', 'tts-1.yaml');
    const yaml = await readFile(modelPath, 'utf8');
    try {
      await writeFile(
        modelPath,
        yaml.replace(/\nruntimeProfile:[\s\S]*$/, '\n'),
      );
      await expect(validateOpenAIProviderDirectory(root)).rejects.toThrow(
        'invalid_runtime_profile',
      );

      await writeFile(
        modelPath,
        yaml.replace(/(runtimeProfile:\s*\r?\n\s*version:) 3/, '$1 2'),
      );
      await expect(validateOpenAIProviderDirectory(root)).rejects.toThrow(
        'runtime_profile_version',
      );
    } finally {
      await rm(root, { recursive: true, force: true });
    }
  });

  it('rejects filename/model mismatches and malformed YAML', async () => {
    const root = await copyProviderFixture();
    const modelPath = join(root, 'openai', 'tts-1.yaml');
    const yaml = await readFile(modelPath, 'utf8');
    try {
      await writeFile(
        modelPath,
        yaml.replace('model: tts-1', 'model: tts-1-hd'),
      );
      await expect(validateOpenAIProviderDirectory(root)).rejects.toThrow(
        'filename_mismatch',
      );

      await writeFile(modelPath, 'model: [\n');
      await expect(validateOpenAIProviderDirectory(root)).rejects.toThrow(
        'malformed_yaml',
      );
    } finally {
      await rm(root, { recursive: true, force: true });
    }
  });
});

async function copyProviderFixture(): Promise<string> {
  const root = await mkdtemp(join(tmpdir(), 'openai-import-'));
  await cp(
    join(__dirname, '..', '..', 'providers', 'openai'),
    join(root, 'openai'),
    { recursive: true },
  );
  return root;
}

type OpenAIProfileFixture = {
  status?: string;
  modalities?: { input?: string[]; output?: string[] };
  tools?: string[];
  serviceTiers?: string[];
  rateLimits?: Array<Record<string, unknown>>;
  limits?: Record<string, number>;
  runtimeProfile?: unknown;
};
