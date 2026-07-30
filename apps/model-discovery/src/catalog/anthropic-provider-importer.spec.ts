import { join } from 'node:path';
import {
  ANTHROPIC_SUPPORTED_MODEL_IDS,
  anthropicProviderImporter,
  validateAnthropicProviderDirectory,
} from '../../scripts/model-import/providers/anthropic';
import type { ProviderPricingRate } from '../../scripts/model-import/types';

describe('anthropicProviderImporter', () => {
  it('validates the curated Anthropic provider directory', async () => {
    const result = await validateAnthropicProviderDirectory(
      join(__dirname, '..', '..', 'providers'),
    );

    expect(result.modelIds).toEqual([...ANTHROPIC_SUPPORTED_MODEL_IDS].sort());
    expect(result.modelFiles).toHaveLength(
      ANTHROPIC_SUPPORTED_MODEL_IDS.length,
    );
  });

  it('owns Anthropic Message Batches pricing behavior', () => {
    const pricingRates: ProviderPricingRate[] = [];

    anthropicProviderImporter.appendDerivedCostPricingRates({
      cost: {
        input_cost_per_token_batches: 0.0000015,
        output_cost_per_token_batches: 0.0000075,
        cache_creation_input_token_cost: 0.00000375,
        cache_creation_input_token_cost_per_hour: 0.000006,
        cache_read_input_token_cost: 0.0000003,
      },
      pricingRates,
      region: 'global',
      sourceUrl: 'https://platform.claude.com/docs/en/about-claude/pricing',
    });

    expect(pricingRates).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          dimension: 'batch_cache_write_5m_tokens',
          unitPriceUsd: '0.000001875',
          conditionJson: { api: 'message_batches' },
        }),
        expect.objectContaining({
          dimension: 'batch_cache_write_1h_tokens',
          unitPriceUsd: '0.000003',
          conditionJson: { api: 'message_batches' },
        }),
        expect.objectContaining({
          dimension: 'batch_cache_read_tokens',
          unitPriceUsd: '1.5e-7',
          conditionJson: { api: 'message_batches' },
        }),
      ]),
    );
  });

  it('owns Anthropic fast-mode and code-execution pricing behavior', () => {
    const pricingRates: ProviderPricingRate[] = [];

    expect(anthropicProviderImporter.rateMappings).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          key: 'fast_mode_input_cost_per_token',
          dimension: 'speed_input_tokens',
          conditionJson: { speed: 'fast' },
        }),
        expect.objectContaining({
          key: 'data_residency_fast_mode_output_cost_per_token',
          dimension: 'data_residency_speed_output_tokens',
          conditionJson: { speed: 'fast', inference_geo: 'configured' },
        }),
      ]),
    );

    anthropicProviderImporter.appendToolPricingRates({
      defaultToolPricing: {
        code_execution: {
          per_container_hour: 0.05,
          free_organization_hours_per_month: 1550,
          minimum_minutes: 5,
          reconciliation: 'anthropic_cost_api',
        },
      },
      pricingRates,
      sources: [
        'https://platform.claude.com/docs/en/agents-and-tools/tool-use/code-execution-tool',
      ],
    });

    expect(pricingRates).toEqual([
      expect.objectContaining({
        dimension: 'code_execution_container_hours',
        unit: 'container_hour',
        unitPriceUsd: '0.05',
        conditionJson: expect.objectContaining({
          reconciliation: 'anthropic_cost_api',
          request_attribution: 'unavailable',
        }),
      }),
    ]);
  });
});
