import { describe, expect, it } from "vitest";

import {
  buildProviderCredentialPayload,
  defaultProviderCredentialValues,
  missingRequiredProviderFields,
  requiresProviderCredentialValidation,
} from "./credential-schema";
import type { ProviderConnectionSchema } from "./types";

const openAISchema: ProviderConnectionSchema = {
  provider_slug: "openai",
  auth_mode: "api_key",
  version: 1,
  runtime_ready: true,
  credential_validation_supported: true,
  fields: [
    { key: "api_key", label: "API key", type: "password", secret: true, required: true, scope: "credential", runtime_mapping: "authorization_bearer" },
    { key: "base_url", label: "Base URL", type: "url", secret: false, scope: "connection", validation: { default: "https://api.openai.com/v1" } },
    { key: "organization", label: "Organization", type: "text", secret: false, scope: "connection" },
  ],
};

describe("provider credential schema helpers", () => {
  it("hydrates documented defaults and identifies missing required values", () => {
    const values = defaultProviderCredentialValues(openAISchema);
    expect(values).toEqual({ base_url: "https://api.openai.com/v1" });
    expect(missingRequiredProviderFields(openAISchema, values).map((field) => field.key)).toEqual(["api_key"]);
  });

  it("separates connection fields from write-only credentials", () => {
    const payload = buildProviderCredentialPayload(openAISchema, {
      api_key: " sk-user ",
      base_url: " https://example.test/v1 ",
      organization: " org-1 ",
    });

    expect(payload.providerConfig).toEqual({
      base_url: "https://example.test/v1",
      organization: "org-1",
    });
    expect(payload.credentials).toEqual({
      api_key: "sk-user",
      raw: { api_key: "sk-user" },
    });
  });

  it("maps provider bearer tokens to the runtime-compatible API key slot", () => {
    const databricksSchema: ProviderConnectionSchema = {
      provider_slug: "databricks",
      auth_mode: "personal_access_token",
      version: 1,
      runtime_ready: true,
      credential_validation_supported: true,
      fields: [
        { key: "access_token", label: "Access token", type: "password", secret: true, required: true, scope: "credential", runtime_mapping: "authorization_bearer" },
      ],
    };

    expect(buildProviderCredentialPayload(databricksSchema, { access_token: "token" }).credentials).toEqual({
      api_key: "token",
      raw: { access_token: "token" },
    });
  });

  it("requires preflight only when the shared backend registry supports it", () => {
    expect(requiresProviderCredentialValidation(openAISchema, false)).toBe(true);
    expect(requiresProviderCredentialValidation(openAISchema, true)).toBe(false);
    expect(requiresProviderCredentialValidation({
      ...openAISchema,
      provider_slug: "deepgram",
      runtime_ready: false,
      credential_validation_supported: false,
    }, false)).toBe(false);
  });
});
