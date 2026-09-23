import { cp, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { parse } from 'yaml';
import {
  OPENAI_PRIORITY_LONG_CONTEXT_RATE_MAPPINGS,
  OPENAI_SUPPORTED_MODEL_IDS,
  validateOpenAIProviderDirectory,
} from '../../scripts/model-import/providers/openai';
import { importProviderYaml } from '../../scripts/import-models-from-yaml';
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
      'gpt-6-sol': {
        endpoints: ['chat_completions', 'responses', 'batches'],
        transports: ['http', 'sse'],
      },
      'gpt-6-luna': {
        endpoints: ['chat_completions', 'responses', 'batches'],
        transports: ['http', 'sse'],
      },
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

  it('maps Fast mode long-context prices to priority speed dimensions', () => {
    const condition = { service_tier: 'priority', min_prompt_tokens: 272001 };
    expect(
      OPENAI_PRIORITY_LONG_CONTEXT_RATE_MAPPINGS.map(
        ({ dimension, unit, conditionJson }) => ({
          dimension,
          unit,
          conditionJson,
        }),
      ),
    ).toEqual([
      {
        dimension: 'speed_long_context_input_tokens',
        unit: 'token',
        conditionJson: condition,
      },
      {
        dimension: 'speed_long_context_output_tokens',
        unit: 'token',
        conditionJson: condition,
      },
      {
        dimension: 'speed_long_context_cache_read_tokens',
        unit: 'token',
        conditionJson: condition,
      },
      {
        dimension: 'speed_long_context_cache_write_tokens',
        unit: 'token',
        conditionJson: condition,
      },
    ]);
  });

  // Official per-1M-token prices from https://developers.openai.com/api/docs/pricing
  // in the order input, cached input, cache writes, output.
  it.each([
    [
      'gpt-6-sol',
      {
        standard: [2, 0.2, 2.5, 10],
        standardLong: [4, 0.4, 5, 15],
        batchFlex: [1, 0.1, 1.25, 5],
        batchFlexLong: [2, 0.2, 2.5, 7.5],
        fast: [4, 0.4, 5, 20],
        fastLong: [8, 0.8, 10, 30],
      },
    ],
    [
      'gpt-6-luna',
      {
        standard: [0.1, 0.01, 0.125, 0.5],
        standardLong: [0.2, 0.02, 0.25, 0.75],
        batchFlex: [0.05, 0.005, 0.0625, 0.25],
        batchFlexLong: [0.1, 0.01, 0.125, 0.375],
        fast: [0.2, 0.02, 0.25, 1],
        fastLong: [0.4, 0.04, 0.5, 1.5],
      },
    ],
  ])(
    'pins the published %s token prices for every tier',
    async (model, prices) => {
      const document = parse(
        await readFile(join(providerDir, `${model}.yaml`), 'utf8'),
      ) as OpenAIProfileFixture;
      const cost = document.costs?.[0] ?? {};
      const priority = cost.priority_pricing as Record<string, number>;
      const tiered = cost.tiered_pricing as Record<
        string,
        Array<{ cost_per_token: number; from: number }>
      >;
      const perMillion = (value: unknown) =>
        Math.round(Number(value) * 1e6 * 1e6) / 1e6;
      const fields = [
        'input_cost_per_token',
        'cache_read_input_token_cost',
        'cache_creation_input_token_cost',
        'output_cost_per_token',
      ];

      expect(fields.map((field) => perMillion(cost[field]))).toEqual(
        prices.standard,
      );
      expect(
        ['input', 'cache_read', 'cache_write', 'output'].map((key) => {
          expect(tiered[key][0].from).toBe(272000);
          return perMillion(tiered[key][0].cost_per_token);
        }),
      ).toEqual(prices.standardLong);
      for (const tier of ['batches', 'flex']) {
        expect(
          fields.map((field) => perMillion(cost[`${field}_${tier}`])),
        ).toEqual(prices.batchFlex);
        expect(
          fields.map((field) =>
            perMillion(cost[`${field}_${tier}_above_272k_tokens`]),
          ),
        ).toEqual(prices.batchFlexLong);
      }
      expect(fields.map((field) => perMillion(priority[field]))).toEqual(
        prices.fast,
      );
      expect(
        OPENAI_PRIORITY_LONG_CONTEXT_RATE_MAPPINGS.map((mapping) =>
          perMillion(cost[mapping.key]),
        ),
      ).toEqual([
        prices.fastLong[0],
        prices.fastLong[3],
        prices.fastLong[1],
        prices.fastLong[2],
      ]);
    },
  );

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

  it('makes the generic production importer enforce tier pricing coverage', async () => {
    const root = await copyProviderFixture();
    const modelPath = join(root, 'openai', 'gpt-4.1.yaml');
    const yaml = await readFile(modelPath, 'utf8');
    try {
      await writeFile(
        modelPath,
        yaml.replace(
          /\n      priority_pricing:\r?\n(?:          .*\r?\n)+/,
          '\n',
        ),
      );

      await expect(
        importProviderYaml({
          providersRoot: root,
          providerSlugs: ['openai'],
          dryRun: true,
          log: () => undefined,
        }),
      ).rejects.toThrow('missing_tier_pricing');
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
  costs?: Array<Record<string, unknown>>;
  runtimeProfile?: unknown;
};
