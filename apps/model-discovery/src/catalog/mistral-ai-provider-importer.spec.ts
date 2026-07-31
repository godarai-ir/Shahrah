import { join } from 'node:path';
import {
  MISTRAL_AI_SUPPORTED_MODEL_IDS,
  mistralAIProviderImporter,
  validateMistralAIProviderDirectory,
} from '../../scripts/model-import/providers/mistral-ai';

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
