import { modelRuntimeProfileSchema } from './runtime-profile.schema';

const validProfile = {
  version: 1,
  endpoints: ['messages', 'count_tokens'],
  thinking: {
    defaultMode: 'adaptive',
    allowedModes: ['adaptive'],
    manualDeprecated: false,
    effortValues: ['low', 'high'],
  },
  assistantPrefill: false,
  sampling: {
    parameters: [],
    disallowedWithThinking: [],
  },
  tools: {
    custom: true,
    server: true,
    parallel: true,
    structuredOutput: true,
  },
  modalities: {
    input: ['text', 'image', 'pdf'],
    output: ['text'],
  },
  betaHeaders: [],
  limits: {
    contextWindow: 1_000_000,
    maxOutputTokens: 128_000,
  },
};

describe('modelRuntimeProfileSchema', () => {
  it('accepts a versioned Anthropic runtime profile and supplies size limits', () => {
    const parsed = modelRuntimeProfileSchema.parse(validProfile);

    expect(parsed.limits.messagesBytes).toBe(32 * 1024 * 1024);
    expect(parsed.limits.batchBytes).toBe(256 * 1024 * 1024);
    expect(parsed.limits.fileBytes).toBe(500 * 1024 * 1024);
  });

  it('rejects a thinking default that is not allowed', () => {
    const result = modelRuntimeProfileSchema.safeParse({
      ...validProfile,
      thinking: {
        ...validProfile.thinking,
        defaultMode: 'manual',
      },
    });

    expect(result.success).toBe(false);
  });

  it('rejects unknown runtime fields', () => {
    const result = modelRuntimeProfileSchema.safeParse({
      ...validProfile,
      unsupportedBehavior: true,
    });

    expect(result.success).toBe(false);
  });

  it('validates endpoint-scoped beta limits and effective speed', () => {
    const parsed = modelRuntimeProfileSchema.parse({
      ...validProfile,
      endpoints: [...validProfile.endpoints, 'message_batches'],
      requestControls: {
        serviceTierValues: ['auto'],
        inferenceGeoValues: ['global'],
        speedValues: ['standard', 'fast'],
        effectiveSpeed: { standard: 'standard', fast: 'standard' },
      },
      betaHeaders: ['output-300k-2026-03-24'],
      betaCapabilities: [
        {
          header: 'output-300k-2026-03-24',
          endpoints: ['message_batches'],
          maxOutputTokens: 300_000,
        },
      ],
    });

    expect(parsed.requestControls.effectiveSpeed.fast).toBe('standard');
    expect(parsed.betaCapabilities[0]?.maxOutputTokens).toBe(300_000);
  });

  it('rejects a beta capability whose header is not allowlisted', () => {
    const result = modelRuntimeProfileSchema.safeParse({
      ...validProfile,
      betaCapabilities: [
        {
          header: 'output-300k-2026-03-24',
          endpoints: ['messages'],
          maxOutputTokens: 300_000,
        },
      ],
    });

    expect(result.success).toBe(false);
  });

  it('accepts Gemini native and OpenAI-compatible endpoints and modalities', () => {
    const parsed = modelRuntimeProfileSchema.parse({
      ...validProfile,
      endpoints: [
        'models',
        'chat_completions',
        'embeddings',
        'image_generations',
        'videos',
        'interactions',
        'live',
        'batches',
        'generate_content',
        'generate_answer',
        'embed_content',
        'imagen_generate',
      ],
      thinking: {
        defaultMode: 'manual',
        allowedModes: ['manual'],
        manualDeprecated: false,
        effortValues: ['none', 'minimal', 'low', 'medium', 'high'],
      },
      modalities: {
        input: ['text', 'image', 'pdf', 'audio', 'video'],
        output: ['text', 'image', 'audio', 'video', 'embedding'],
      },
      requestControls: {
        serviceTierValues: ['standard', 'flex', 'priority'],
        inferenceGeoValues: ['global'],
        speedValues: [],
        effectiveSpeed: {},
      },
      limits: {
        contextWindow: 0,
        maxOutputTokens: 0,
      },
    });

    expect(parsed.endpoints).toContain('interactions');
    expect(parsed.endpoints).toContain('generate_content');
    expect(parsed.endpoints).toContain('embed_content');
    expect(parsed.modalities.output).toContain('embedding');
    expect(parsed.thinking.effortValues).toContain('minimal');
  });

  it('accepts OpenAI Responses tools and the Fast service tier', () => {
    const parsed = modelRuntimeProfileSchema.parse({
      ...validProfile,
      endpoints: ['models', 'responses'],
      tools: {
        ...validProfile.tools,
        supported: [
          'function_calling',
          'web_search',
          'file_search',
          'computer_use',
          'mcp',
        ],
      },
      requestControls: {
        serviceTierValues: ['standard', 'priority', 'fast'],
        inferenceGeoValues: ['global'],
        speedValues: ['standard', 'fast'],
        effectiveSpeed: { standard: 'standard', fast: 'fast' },
      },
    });

    expect(parsed.tools.supported).toContain('computer_use');
    expect(parsed.requestControls.serviceTierValues).toContain('fast');
  });
});
