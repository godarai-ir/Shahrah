import { readFileSync } from 'fs';
import { resolve } from 'path';

const expectedProviders = [
  'ai21',
  'anthropic',
  'aws-bedrock',
  'azure-ai-foundry',
  'azure-open-ai',
  'cerebras',
  'cohere',
  'databricks',
  'deepgram',
  'deepinfra',
  'elevenlabs',
  'google-gemini',
  'google-vertex',
  'groq',
  'mistral-ai',
  'openai',
  'openrouter',
  'perplexity-ai',
  'sambanova',
  'together-ai',
  'xai',
];

const expectedRuntimeReadyProviders = [
  'anthropic',
  'azure-open-ai',
  'cerebras',
  'cohere',
  'databricks',
  'deepinfra',
  'google-gemini',
  'groq',
  'mistral-ai',
  'openai',
  'openrouter',
  'together-ai',
  'xai',
];

describe('provider connection schema migration', () => {
  it('contains parseable, explicit schemas for every active database provider', () => {
    const sql = readFileSync(
      resolve(
        process.cwd(),
        'prisma/migrations/202607170001_platform_provider_connection_schemas/migration.sql',
      ),
      'utf8',
    );
    const rows = [
      ...sql.matchAll(
        /\('([^']+)',\s*'([^']+)',\s*(true|false),\s*'(\[[\s\S]*?\])'::jsonb\)/g,
      ),
    ];
    const providers = [...new Set(rows.map((row) => row[1]))].sort();
    const runtimeReadyProviders = [
      ...new Set(rows.filter((row) => row[3] === 'true').map((row) => row[1])),
    ].sort();

    expect(rows).toHaveLength(26);
    expect(providers).toEqual(expectedProviders);
    expect(runtimeReadyProviders).toEqual(expectedRuntimeReadyProviders);
    for (const [, provider, authMode, runtimeReady, encodedFields] of rows) {
      const fields = JSON.parse(encodedFields) as Array<{
        key: string;
        scope: string;
        secret: boolean;
        runtime_mapping?: string;
      }>;
      expect(provider).toBeTruthy();
      expect(authMode).toBeTruthy();
      expect(['true', 'false']).toContain(runtimeReady);
      expect(fields.length).toBeGreaterThan(0);
      expect(fields.every((field) => field.key && field.scope)).toBe(true);
      expect(
        fields
          .filter((field) => field.scope === 'credential')
          .every((field) => field.secret === true && field.runtime_mapping),
      ).toBe(true);
    }
  });
});
