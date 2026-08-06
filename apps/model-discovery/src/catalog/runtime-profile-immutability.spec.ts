import { assertRuntimeProfileVersionIsImmutable } from './runtime-profile-immutability';

describe('assertRuntimeProfileVersionIsImmutable', () => {
  const storedProfile = {
    version: 2,
    profileJson: {
      version: 2,
      endpoints: ['chat_completions', 'batches'],
      batch: { eligible: false, supportedEndpoints: [] },
    },
  };

  it('accepts an unchanged immutable version', () => {
    expect(() =>
      assertRuntimeProfileVersionIsImmutable(
        'mistral-ai/example',
        storedProfile,
        {
          batch: { supportedEndpoints: [], eligible: false },
          endpoints: ['chat_completions', 'batches'],
          version: 2,
        },
      ),
    ).not.toThrow();
  });

  it('rejects changed content with the same version', () => {
    expect(() =>
      assertRuntimeProfileVersionIsImmutable(
        'mistral-ai/example',
        storedProfile,
        {
          version: 2,
          endpoints: ['chat_completions', 'batches'],
          batch: {
            eligible: true,
            supportedEndpoints: ['chat_completions'],
          },
        },
      ),
    ).toThrow(
      'mistral-ai/example: runtime profile version 2 is immutable; increment runtimeProfile.version',
    );
  });

  it('accepts changed content under a new version', () => {
    expect(() =>
      assertRuntimeProfileVersionIsImmutable(
        'mistral-ai/example',
        storedProfile,
        {
          version: 3,
          endpoints: ['chat_completions', 'batches'],
          batch: {
            eligible: true,
            supportedEndpoints: ['chat_completions'],
          },
        },
      ),
    ).not.toThrow();
  });
});
