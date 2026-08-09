import type { ProviderConnectionSchema } from "./types";

export type ProviderCredentialValues = Record<string, string>;

export function defaultProviderCredentialValues(schema: ProviderConnectionSchema | null): ProviderCredentialValues {
  return Object.fromEntries(
    (schema?.fields ?? []).flatMap((field) => {
      const value = field.validation?.default;
      if (typeof value === "string" && value.trim()) return [[field.key, value]];
      if (Array.isArray(value)) {
        const normalized = value.filter((item): item is string => typeof item === "string" && !!item.trim());
        return normalized.length ? [[field.key, normalized.join(",")]] : [];
      }
      return [];
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
    else if (Array.isArray(value)) {
      values[field.key] = value.filter((item): item is string => typeof item === "string").join(",");
    }
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

export function buildProviderCredentialPayload(
  schema: ProviderConnectionSchema | null,
  values: ProviderCredentialValues,
) {
  const providerConfig: Record<string, unknown> = {};
  const rawCredentials: Record<string, string> = {};

  for (const field of schema?.fields ?? []) {
    const value = values[field.key]?.trim();
    if (!value) continue;
    if (field.scope === "credential") rawCredentials[field.key] = value;
    else if (field.type === "multi_select") {
      providerConfig[field.key] = value.split(",").map((item) => item.trim()).filter(Boolean);
    } else providerConfig[field.key] = value;
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
