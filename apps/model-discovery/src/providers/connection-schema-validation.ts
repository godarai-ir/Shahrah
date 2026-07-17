import { UnprocessableEntityException } from '@nestjs/common';

const allowedFieldProperties = new Set([
  'key',
  'label',
  'helptext',
  'type',
  'secret',
  'required',
  'requiredwhen',
  'validation',
  'scope',
  'runtimemapping',
  'documentationurl',
  'redactionpolicy',
]);

const secretShapedNames = new Set([
  'apikey',
  'clientsecret',
  'secretaccesskey',
  'sessiontoken',
  'serviceaccountjson',
  'credential',
  'credentials',
  'token',
  'password',
]);

const secretReferencePattern =
  /(?:tfy-(?:k8s-)?secret:\/\/|(?:vault|secret):\/\/|::\/)/i;

const normalized = (value: string) =>
  value.toLowerCase().replace(/[^a-z0-9]/g, '');

function reject(path: string, reason: string): never {
  throw new UnprocessableEntityException({
    error: {
      code: 'invalid_provider_connection_schema',
      message:
        'Schema definitions must not contain connection values or secret references.',
      field: path,
      reason,
    },
  });
}

function scanMetadata(value: unknown, path: string): void {
  if (typeof value === 'string') {
    if (secretReferencePattern.test(value))
      reject(path, 'secret_reference_forbidden');
    return;
  }
  if (Array.isArray(value)) {
    value.forEach((item, index) => scanMetadata(item, `${path}[${index}]`));
    return;
  }
  if (!value || typeof value !== 'object') return;
  for (const [key, nested] of Object.entries(value)) {
    if (secretShapedNames.has(normalized(key)))
      reject(`${path}.${key}`, 'secret_value_key_forbidden');
    scanMetadata(nested, `${path}.${key}`);
  }
}

export function validateConnectionSchemaFields(
  fields: Record<string, unknown>[],
): void {
  fields.forEach((field, index) => {
    if (!field || Array.isArray(field) || typeof field !== 'object') {
      reject(`fields[${index}]`, 'field_definition_must_be_an_object');
    }
    for (const [property, value] of Object.entries(field)) {
      if (!allowedFieldProperties.has(normalized(property))) {
        const reason = secretShapedNames.has(normalized(property))
          ? 'secret_value_key_forbidden'
          : 'unknown_field_property';
        reject(`fields[${index}].${property}`, reason);
      }
      // key is a field identifier (for example "api_key"), not a submitted
      // credential value. All other nested metadata remains value-free.
      if (normalized(property) !== 'key') {
        scanMetadata(value, `fields[${index}].${property}`);
      }
    }
    if (typeof field.key !== 'string' || field.key.trim() === '') {
      reject(`fields[${index}].key`, 'field_key_required');
    }
    if (field.scope !== 'credential' && field.scope !== 'connection') {
      reject(
        `fields[${index}].scope`,
        'scope_must_be_credential_or_connection',
      );
    }
    if (field.scope === 'credential' && field.secret !== true) {
      reject(`fields[${index}].secret`, 'credential_fields_must_be_secret');
    }
  });
}
