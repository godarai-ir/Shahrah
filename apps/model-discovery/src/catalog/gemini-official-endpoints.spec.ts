import { readFile, readdir } from 'node:fs/promises';
import { join } from 'node:path';
import { parse } from 'yaml';
import { modelRuntimeProfileSchema } from './runtime-profile.schema';

/**
 * Official Gemini API surface for the text models, per
 * https://ai.google.dev/api, https://ai.google.dev/gemini-api/docs/openai and
 * https://ai.google.dev/gemini-api/docs/interactions-overview. Gemini exposes
 * native generateContent, its OpenAI-compatible chat completions, and its own
 * Interactions API. It has no OpenAI Responses or Anthropic Messages endpoint.
 */
const GEMINI_NON_OFFICIAL_ENDPOINTS = ['responses', 'messages'];

/**
 * TokenBazaar resells Gemini through chat/completions and generateContent
 * only; its catalog lists no interactions, files, caching, batch, or
 * countTokens surface for these models. It also lists responses and
 * messages, which are TokenBazaar translation layers, not Gemini APIs.
 */
const TOKENBAZAAR_GEMINI_ENDPOINTS = [
  'models',
  'chat_completions',
  'generate_content',
  'stream_generate_content',
];

type YamlDoc = Record<string, unknown>;

async function readGeminiTextModels(providerSlug: string): Promise<YamlDoc[]> {
  const dir = join(__dirname, '..', '..', 'providers', providerSlug);
  const docs: YamlDoc[] = [];
  for (const entry of await readdir(dir)) {
    if (!entry.startsWith('gemini-') || !entry.endsWith('.yaml')) continue;
    const doc = parse(await readFile(join(dir, entry), 'utf8')) as YamlDoc;
    if (doc.mode === 'chat') docs.push(doc);
  }
  return docs;
}

describe.each(['google-gemini', 'tokenbazaar'])(
  '%s Gemini text models follow the official Gemini API',
  (providerSlug) => {
    it('advertise chat and generate_content, never responses', async () => {
      const docs = await readGeminiTextModels(providerSlug);
      expect(docs.length).toBeGreaterThan(0);
      for (const doc of docs) {
        expect(doc.supportedModes).toEqual(['chat', 'generate_content']);
      }
    });

    it('declare no endpoint Gemini does not officially offer', async () => {
      for (const doc of await readGeminiTextModels(providerSlug)) {
        const profile = modelRuntimeProfileSchema.parse(doc.runtimeProfile);
        for (const endpoint of GEMINI_NON_OFFICIAL_ENDPOINTS) {
          expect(profile.endpoints).not.toContain(endpoint);
        }
        expect(profile.endpoints).toEqual(
          expect.arrayContaining(['chat_completions', 'generate_content']),
        );
      }
    });
  },
);

describe('TokenBazaar Gemini text models', () => {
  it('declare only the endpoints TokenBazaar serves for Gemini', async () => {
    for (const doc of await readGeminiTextModels('tokenbazaar')) {
      const profile = modelRuntimeProfileSchema.parse(doc.runtimeProfile);
      expect(profile.endpoints).toEqual(TOKENBAZAAR_GEMINI_ENDPOINTS);
      expect(doc.endpoints).toEqual(TOKENBAZAAR_GEMINI_ENDPOINTS);
    }
  });
});
