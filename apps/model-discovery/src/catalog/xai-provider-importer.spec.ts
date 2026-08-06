import { cp, mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { parse } from 'yaml';
import {
  XAI_RUNTIME_PROFILE_VERSIONS,
  XAI_SUPPORTED_MODEL_IDS,
  validateXAIProviderDirectory,
  xaiProviderImporter,
} from '../../scripts/model-import/providers/xai';
import type { ProviderPricingRate } from '../../scripts/model-import/types';
import { modelRuntimeProfileSchema } from './runtime-profile.schema';

describe('xAI provider importer', () => {
  const providersRoot = join(__dirname, '..', '..', 'providers');

  it('validates exactly four curated models and their non-colliding aliases', async () => {
    const result = await validateXAIProviderDirectory(providersRoot);
    expect(result.modelIds).toEqual([...XAI_SUPPORTED_MODEL_IDS].sort());
    expect(result.modelFiles).toHaveLength(4);
    expect(result.aliases['grok-imagine-image-pro']).toBe(
      'grok-imagine-image-quality',
    );
  });

  it('keeps model YAML aliases and reviewed runtime versions synchronized', async () => {
    for (const model of XAI_SUPPORTED_MODEL_IDS) {
      const document = parse(
        await readFile(join(providersRoot, 'xai', `${model}.yaml`), 'utf8'),
      ) as { aliases?: string[]; runtimeProfile?: unknown };
      const profile = modelRuntimeProfileSchema.parse(document.runtimeProfile);
      expect(profile.version).toBe(XAI_RUNTIME_PROFILE_VERSIONS[model]);
      expect(profile.aliases).toEqual(document.aliases);
      expect(profile.usageAccounting).toMatchObject({
        exactProviderCost: true,
        requiredResponseFields: ['cost_in_usd_ticks'],
      });
      expect(profile.batch.eligible).toBe(true);
      expect(profile.batch.discounted).toBe(model === 'grok-4.3');
    }
  });

  it('rejects a downgraded xAI runtime profile before import', async () => {
    const root = await mkdtemp(join(tmpdir(), 'xai-profile-version-'));
    await cp(join(providersRoot, 'xai'), join(root, 'xai'), {
      recursive: true,
    });

    const modelPath = join(root, 'xai', 'grok-4.3.yaml');
    try {
      const yaml = await readFile(modelPath, 'utf8');
      await writeFile(
        modelPath,
        yaml.replace(/(runtimeProfile:\s*\r?\n\s*version:) 4/, '$1 3'),
      );

      await expect(validateXAIProviderDirectory(root)).rejects.toThrow(
        'runtime_profile_version',
      );
    } finally {
      await rm(root, { recursive: true, force: true });
    }
  });

  it('owns xAI long-context, Batch, resolution, and tool rate mappings', () => {
    expect(xaiProviderImporter.batchPricingCondition).toEqual({
      api: 'batches',
    });
    expect(xaiProviderImporter.rateMappings).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          key: 'xai_long_context_input_cost_per_token',
          conditionJson: { min_prompt_tokens: 200000 },
        }),
        expect.objectContaining({
          key: 'output_cost_per_image_2k',
          dimension: 'output_images',
          conditionJson: { resolution: '2k' },
        }),
      ]),
    );

    const pricingRates: ProviderPricingRate[] = [];
    xaiProviderImporter.appendToolPricingRates({
      defaultToolPricing: {
        web_search: { per_thousand_requests: 5 },
        x_search: { per_thousand_requests: 5 },
        attachment_search: { per_thousand_requests: 10 },
      },
      documentedTools: new Set(['web_search', 'x_search']),
      pricingRates,
      sources: ['https://docs.x.ai/developers/pricing'],
    });
    expect(pricingRates).toEqual([
      expect.objectContaining({
        dimension: 'web_search_requests',
        unitPriceUsd: '0.005',
      }),
      expect.objectContaining({
        dimension: 'x_search_requests',
        unitPriceUsd: '0.005',
      }),
    ]);
  });

  it('rejects empty documents and stale provider sources', async () => {
    const root = await mkdtemp(join(tmpdir(), 'xai-import-'));
    await cp(join(providersRoot, 'xai'), join(root, 'xai'), {
      recursive: true,
    });

    const modelPath = join(root, 'xai', 'grok-4.5.yaml');
    await writeFile(modelPath, '');
    await expect(validateXAIProviderDirectory(root)).rejects.toThrow(
      'xAI model allowlist validation failed',
    );

    await cp(join(providersRoot, 'xai', 'grok-4.5.yaml'), modelPath, {
      force: true,
    });
    await writeFile(
      join(root, 'xai', 'default.yaml'),
      'documentation: [https://console.groq.com/docs]\n',
    );
    await expect(validateXAIProviderDirectory(root)).rejects.toThrow(
      'official docs.x.ai sources',
    );
    await rm(root, { recursive: true, force: true });
  });
});
