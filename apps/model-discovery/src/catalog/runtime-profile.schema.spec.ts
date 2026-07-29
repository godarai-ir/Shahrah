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
});
