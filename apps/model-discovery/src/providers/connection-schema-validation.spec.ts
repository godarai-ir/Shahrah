import { UnprocessableEntityException } from '@nestjs/common';
import { validateConnectionSchemaFields } from './connection-schema-validation';

describe('validateConnectionSchemaFields', () => {
  it('accepts value-free secret and connection field definitions', () => {
    expect(() =>
      validateConnectionSchemaFields([
        {
          key: 'api_key',
          label: { en: 'API key', fa: 'کلید API' },
          type: 'password',
          secret: true,
          required: true,
          scope: 'credential',
          runtime_mapping: 'api_key',
          redaction_policy: 'write_only',
        },
        {
          key: 'base_url',
          label: { en: 'Base URL' },
          type: 'url',
          secret: false,
          required: false,
          scope: 'connection',
          runtime_mapping: 'base_url',
        },
      ]),
    ).not.toThrow();
  });

  it.each(['api_key', 'apiKey', 'client_secret', 'secretAccessKey', 'sessionToken', 'password'])(
    'rejects a value-shaped %s property without echoing its value',
    (property) => {
      const submittedSecret = 'must-not-appear-in-errors';
      let thrown: unknown;
      try {
        validateConnectionSchemaFields([
          {
            key: 'api_key',
            type: 'password',
            secret: true,
            scope: 'credential',
            [property]: submittedSecret,
          },
        ]);
      } catch (error) {
        thrown = error;
      }
      expect(thrown).toBeInstanceOf(UnprocessableEntityException);
      expect(JSON.stringify((thrown as UnprocessableEntityException).getResponse())).not.toContain(
        submittedSecret,
      );
    },
  );

  it('rejects nested secret references', () => {
    expect(() =>
      validateConnectionSchemaFields([
        {
          key: 'api_key',
          type: 'password',
          secret: true,
          scope: 'credential',
          validation: { pattern: 'tfy-secret://production/openai' },
        },
      ]),
    ).toThrow(UnprocessableEntityException);
  });
});
