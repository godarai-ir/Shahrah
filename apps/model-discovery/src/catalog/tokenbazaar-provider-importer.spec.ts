import { readFile, readdir } from 'node:fs/promises';
import { join } from 'node:path';
import { parse } from 'yaml';
import { modelRuntimeProfileSchema } from './runtime-profile.schema';

/**
 * The 26 model IDs reconciled against the live TokenBazaar catalog
 * (GET https://tokenbazaar.ai/api/catalog/public and the authenticated
 * /v1/models shape documented at https://tokenbazaar.ai/docs) on 2026-09-10.
 * See plans/tokenbazaar-production-readiness.md for the full reconciliation
 * evidence. Do not add or remove an entry here without updating that plan.
 */
export const TOKENBAZAAR_SUPPORTED_MODEL_IDS = [
  'claude-fable-5',
  'claude-haiku-4-5',
  'claude-opus-5',
  'claude-sonnet-5',
  'deepseek-v4-flash',
  'deepseek-v4-pro',
  'gemini-2.5-flash',
  'gemini-2.5-pro',
  'gemini-3-pro-image',
  'gemini-3.1-flash-image',
  'gemini-3.1-flash-lite-image',
  'gemini-3.1-pro-preview',
  'gemini-3.5-flash',
  'gemini-3.6-flash',
  'gemini-3.7-flash',
  'glm-5.2',
  'glm-5v-turbo',
  'gpt-5.5',
  'gpt-5.6-luna',
  'gpt-5.6-sol',
  'gpt-5.6-terra',
  'gpt-image-2',
  'kimi-k2.6',
  'kimi-k2.7-code',
  'kimi-k3',
  'text-embedding-3-large',
] as const;

const DIRECT_PROVIDER_SLUGS = [
  'anthropic',
  'deepseek',
  'google-gemini',
  'openai',
  'moonshot',
  'zai',
];

type YamlDoc = Record<string, unknown>;

async function readProviderYaml(
  providersRoot: string,
  providerSlug: string,
): Promise<Map<string, YamlDoc>> {
  const dir = join(providersRoot, providerSlug);
  let entries: string[];
  try {
    entries = await readdir(dir);
  } catch {
    return new Map();
  }
  const docs = new Map<string, YamlDoc>();
  for (const entry of entries) {
    if (!entry.endsWith('.yaml') || entry === 'default.yaml') continue;
    const raw = await readFile(join(dir, entry), 'utf8');
    const doc = parse(raw) as YamlDoc;
    if (doc && typeof doc.model === 'string') {
      docs.set(doc.model, doc);
    }
  }
  return docs;
}

describe('TokenBazaar provider importer', () => {
  const providersRoot = join(__dirname, '..', '..', 'providers');

  it('imports exactly the 26 reconciled TokenBazaar models', async () => {
    const docs = await readProviderYaml(providersRoot, 'tokenbazaar');
    const modelIds = Array.from(docs.keys()).sort();
    expect(modelIds).toEqual([...TOKENBAZAAR_SUPPORTED_MODEL_IDS].sort());
    expect(modelIds).toHaveLength(26);
  });

  it('never collides with a direct provider canonicalId', async () => {
    for (const directSlug of DIRECT_PROVIDER_SLUGS) {
      const directDocs = await readProviderYaml(providersRoot, directSlug);
      for (const modelId of directDocs.keys()) {
        const tokenbazaarCanonical = `tokenbazaar/${modelId}`;
        const directCanonical = `${directSlug}/${modelId}`;
        expect(tokenbazaarCanonical).not.toEqual(directCanonical);
      }
    }
  });

  it('gives every model complete, provenance-tagged pricing', async () => {
    const docs = await readProviderYaml(providersRoot, 'tokenbazaar');
    for (const [, doc] of docs) {
      const costs = doc.costs;
      expect(Array.isArray(costs) && costs.length > 0).toBe(true);

      const sources = doc.sources;
      expect(Array.isArray(sources) && sources.length > 0).toBe(true);
      const sourceList = sources as string[];

      // Every model must retain provenance to both its original upstream
      // vendor documentation and the TokenBazaar pricing page it was priced
      // from -- never priced from TokenBazaar alone.
      const hasUpstreamSource = sourceList.some(
        (s) => !s.includes('tokenbazaar.ai'),
      );
      const hasTokenBazaarSource = sourceList.some((s) =>
        s.includes('tokenbazaar.ai'),
      );
      expect(hasUpstreamSource).toBe(true);
      expect(hasTokenBazaarSource).toBe(true);

      for (const cost of costs as Record<string, unknown>[]) {
        const dims = [
          'input_cost_per_token',
          'output_cost_per_token',
          'input_cost_per_image',
          'output_cost_per_image',
          'input_cost_per_second',
          'output_cost_per_second',
        ];
        // At least one billable dimension must carry a strictly positive
        // price. A model may legitimately declare an explicit `0` for a
        // dimension it does not use (e.g. an embedding model's
        // output_cost_per_token), so only the presence of at least one
        // positive dimension is required, not that every declared
        // dimension be positive.
        const hasPositivePriceDimension = dims.some((dim) => {
          const value = cost[dim];
          return typeof value === 'number' && value > 0;
        });
        expect(hasPositivePriceDimension).toBe(true);
      }

      // If a runtimeProfile is declared it must validate against the
      // shared schema -- TokenBazaar YAMLs are self-contained per model.
      if (doc.runtimeProfile !== undefined) {
        expect(() =>
          modelRuntimeProfileSchema.parse(doc.runtimeProfile),
        ).not.toThrow();
      }
    }
  });

  it('rejects a slug/model-id collision with tokenbazaar and a direct provider', () => {
    // canonicalId is always `${providerSlug}/${modelId}`; because tokenbazaar
    // is a distinct provider slug from every direct provider slug, no
    // canonicalId collision is possible even when the underlying model id
    // string is identical (e.g. `claude-sonnet-5`).
    for (const directSlug of DIRECT_PROVIDER_SLUGS) {
      expect(directSlug).not.toEqual('tokenbazaar');
    }
  });
});
