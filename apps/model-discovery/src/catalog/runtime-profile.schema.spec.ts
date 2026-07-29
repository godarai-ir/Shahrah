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
});
