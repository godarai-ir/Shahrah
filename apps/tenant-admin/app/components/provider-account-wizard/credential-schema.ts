import type { ProviderConnectionSchema } from "./types";

export type ProviderCredentialValues = Record<string, string>;

export type ProviderCredentialValidationResult = {
  success: boolean;
  unsupported?: boolean;
  category?: string;
};

export type ProviderCredentialValidationMessageKey =
  | "messages.credentialValid"
  | "messages.credentialQuotaExhausted"
  | "messages.credentialInvalid"
  | "messages.credentialForbidden"
  | "messages.credentialConfigurationInvalid"
  | "messages.credentialMissing"
  | "messages.credentialNetworkError"
  | "messages.credentialTimeout"
  | "messages.credentialProviderError"
  | "messages.credentialUnsupported"
  | "messages.credentialManaged"
  | "messages.credentialNotRequired";

export function providerCredentialValidationMessageKey(
  result: ProviderCredentialValidationResult,
): ProviderCredentialValidationMessageKey {
  if (result.unsupported || result.category === "unsupported") return "messages.credentialUnsupported";

  switch (result.category) {
    case "ok":
      return "messages.credentialValid";
    case "quota_exhausted":
      return "messages.credentialQuotaExhausted";
    case "invalid_credential":
      return "messages.credentialInvalid";
    case "forbidden":
      return "messages.credentialForbidden";
    case "configuration":
      return "messages.credentialConfigurationInvalid";
    case "missing_credential":
      return "messages.credentialMissing";
    case "network":
      return "messages.credentialNetworkError";
    case "timeout":
      return "messages.credentialTimeout";
    case "provider":
      return "messages.credentialProviderError";
    case "managed":
      return "messages.credentialManaged";
    case "no_credential":
      return "messages.credentialNotRequired";
    default:
      return result.success ? "messages.credentialValid" : "messages.credentialInvalid";
  }
}

export function providerCredentialValidationNoticeKind(
  result: ProviderCredentialValidationResult,
): "success" | "warning" | "error" {
  if (result.unsupported || result.category === "unsupported") return "warning";
  return result.success ? "success" : "error";
}

export function defaultProviderCredentialValues(schema: ProviderConnectionSchema | null): ProviderCredentialValues {
  return Object.fromEntries(
    (schema?.fields ?? []).flatMap((field) => {
      const value = field.validation?.default;
      return typeof value === "string" && value.trim() ? [[field.key, value]] : [];
    }),
  );
}

export function providerCredentialValuesForSchema(
  schema: ProviderConnectionSchema | null,
  providerConfig?: Record<string, unknown>,
): ProviderCredentialValues {
  const values = defaultProviderCredentialValues(schema);

  for (const field of schema?.fields ?? []) {
    if (field.scope !== "connection") continue;
    const value = providerConfig?.[field.key];
    if (typeof value === "string") values[field.key] = value;
    else if (typeof value === "number" || typeof value === "boolean") values[field.key] = String(value);
  }

  return values;
}

export function missingRequiredProviderFields(
  schema: ProviderConnectionSchema | null,
  values: ProviderCredentialValues,
) {
  return (schema?.fields ?? []).filter((field) => field.required && !values[field.key]?.trim());
}

export function requiresProviderCredentialValidation(
  schema: ProviderConnectionSchema | null,
  usePlatformManagedCredentials: boolean,
) {
  return !usePlatformManagedCredentials && schema?.credential_validation_supported === true;
}

export function buildProviderCredentialPayload(
  schema: ProviderConnectionSchema | null,
  values: ProviderCredentialValues,
) {
  const providerConfig: Record<string, string> = {};
  const rawCredentials: Record<string, string> = {};

  for (const field of schema?.fields ?? []) {
    const value = values[field.key]?.trim();
    if (!value) continue;
    if (field.scope === "credential") rawCredentials[field.key] = value;
    else providerConfig[field.key] = value;
  }

  const credentials: Record<string, unknown> = {};
  if (Object.keys(rawCredentials).length > 0) credentials.raw = rawCredentials;

  const runtimeAPIKeyField = (schema?.fields ?? []).find((field) => {
    if (field.scope !== "credential") return false;
    const mapping = field.runtime_mapping?.trim().toLowerCase();
    return field.key === "api_key"
      || field.key === "access_token"
      || mapping === "authorization_bearer"
      || mapping === "authorization_token"
      || mapping === "api_key_header"
      || mapping === "x_api_key"
      || mapping === "x_goog_api_key";
  });
  const runtimeAPIKey = runtimeAPIKeyField ? rawCredentials[runtimeAPIKeyField.key] : undefined;
  if (runtimeAPIKey) credentials.api_key = runtimeAPIKey;

  return { providerConfig, credentials, rawCredentials };
}
