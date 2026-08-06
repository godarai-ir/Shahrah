import { cp, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { parse } from 'yaml';
import {
  MISTRAL_AI_RUNTIME_PROFILE_VERSIONS,
  MISTRAL_AI_SUPPORTED_MODEL_IDS,
  mistralAIProviderImporter,
  validateMistralAIProviderDirectory,
} from '../../scripts/model-import/providers/mistral-ai';
import { modelRuntimeProfileSchema } from './runtime-profile.schema';

describe('mistralAIProviderImporter', () => {
  it('validates the curated ten-model Mistral AI directory', async () => {
    const result = await validateMistralAIProviderDirectory(
      join(__dirname, '..', '..', 'providers'),
    );
    expect(result.modelIds).toEqual([...MISTRAL_AI_SUPPORTED_MODEL_IDS].sort());
    expect(result.modelFiles).toHaveLength(
      MISTRAL_AI_SUPPORTED_MODEL_IDS.length,
    );
  });

  it('requires the reviewed immutable profile version for every model', async () => {
    const providerDir = join(__dirname, '..', '..', 'providers', 'mistral-ai');
    for (const modelId of MISTRAL_AI_SUPPORTED_MODEL_IDS) {
      const document = parse(
        await readFile(join(providerDir, `${modelId}.yaml`), 'utf8'),
      ) as { runtimeProfile?: unknown };
      const profile = modelRuntimeProfileSchema.parse(document.runtimeProfile);
      expect(profile.version).toBe(
        MISTRAL_AI_RUNTIME_PROFILE_VERSIONS[modelId],
      );
    }
  });

  it('rejects a downgraded batch profile before import', async () => {
    const providersRoot = await mkdtemp(
      join(tmpdir(), 'mistral-profile-version-'),
    );
    const sourceDir = join(__dirname, '..', '..', 'providers', 'mistral-ai');
    const providerDir = join(providersRoot, 'mistral-ai');
    const modelPath = join(providerDir, 'codestral-2508.yaml');
    try {
      await cp(sourceDir, providerDir, { recursive: true });
      const yaml = await readFile(modelPath, 'utf8');
      await writeFile(
        modelPath,
        yaml.replace(/(runtimeProfile:\s*\r?\n\s*version:) 3/, '$1 2'),
      );

      await expect(
        validateMistralAIProviderDirectory(providersRoot),
      ).rejects.toThrow('runtime_profile_version');
    } finally {
      await rm(providersRoot, { recursive: true, force: true });
    }
  });

  it('maps non-token billing dimensions without float conversions', () => {
    expect(mistralAIProviderImporter.rateMappings).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          key: 'input_cost_per_second',
          dimension: 'audio_input_seconds',
          unit: 'second',
        }),
        expect.objectContaining({
          key: 'input_cost_per_page',
          dimension: 'ocr_input_pages',
          unit: 'page',
        }),
        expect.objectContaining({
          key: 'input_cost_per_annotated_page',
          dimension: 'ocr_annotated_pages',
          conditionJson: { document_ai: true },
        }),
      ]),
    );
  });

  it('maps the official premium web-search tool price', () => {
    const pricingRates: Parameters<
      NonNullable<typeof mistralAIProviderImporter.appendToolPricingRates>
    >[0]['pricingRates'] = [];
    mistralAIProviderImporter.appendToolPricingRates?.({
      defaultToolPricing: { web_search_premium: { per_call: 0.05 } },
      pricingRates,
      sources: ['https://mistral.ai/pricing/api/'],
    });
    expect(pricingRates).toContainEqual(
      expect.objectContaining({
        dimension: 'web_search_premium_requests',
        unit: 'request',
        unitPriceUsd: '0.05',
        sourceUrl: 'https://mistral.ai/pricing/api/',
      }),
    );
  });
});
