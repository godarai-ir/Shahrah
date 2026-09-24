import { describe, expect, it } from "vitest";

import { buildPayload, compilePattern, emptyForm, formFromGuardrail, slugFromName, slugFromType } from "./form";
import { guardrailTypeSchema, type GuardrailFormInput, type JwtVerifier, type TenantGuardrail } from "./types";

/** Every config field is populated so the payload shows exactly which fields each type sends. */
const filledForm: GuardrailFormInput = {
  ...emptyForm,
  slug: " pii-guard ",
  name: " PII guard ",
  description: " Masks personal data ",
  operation: "mutate",
  enforcement: "audit",
  enabled: false,
  terms: "ignore previous instructions\n\n drop table ",
  patterns: "(?i)password\\s*[:=]\\s*\\S+\n\\b\\d{10}\\b",
  pattern: " secret-\\d+ ",
  replacement: "<hidden>",
  redactionText: " [redacted] ",
  min: "2",
  max: "40",
  schema: '{ "type": "object", "required": ["answer"] }',
  keys: "answer\nconfidence",
  keyMode: "any",
  languages: "sql\npython",
  codeMode: "must_not_contain",
  tokenRegex: " Bearer (.+) ",
  issuer: " https://issuer.example.com ",
  audience: " gateway ",
  sharedSecret: " s3cret ",
  inlineJwks: '{ "keys": [] }',
  jwksUri: " https://issuer.example.com/jwks.json ",
  introspectionUrl: " https://issuer.example.com/introspect ",
  introspectionAuthHeader: " Bearer abc ",
  maxBytes: "2097152",
  allowedHosts: "cdn.example.com\nimages.example.com",
  timeoutMs: "5000",
  allowPrivateNetworks: true,
};

describe("buildPayload", () => {
  it.each(guardrailTypeSchema.options)("keeps the %s payload stable", (type) => {
    expect(buildPayload({ ...filledForm, type }, "create")).toMatchSnapshot("create");
    expect(buildPayload({ ...filledForm, type }, "edit")).toMatchSnapshot("edit");
  });
});

describe("buildPayload for JWT validation", () => {
  const verifierKeys = {
    jwks: ["jwks_uri"],
    inline: ["inline_jwks"],
    shared: ["shared_secret"],
    introspection: ["introspection_url", "introspection_auth_header"],
  } as const;

  it.each(Object.entries(verifierKeys))("sends only the %s verifier fields", (verifier, keys) => {
    const payload = buildPayload({ ...filledForm, type: "jwt_validation", jwtVerifier: verifier as JwtVerifier }, "create");

    expect(Object.keys(payload.config ?? {}).sort()).toEqual(["audience", "issuer", "token_regex", ...keys].sort());
  });
});

describe("formFromGuardrail", () => {
  function jwtGuardrail(config: Record<string, unknown>): TenantGuardrail {
    return {
      id: "g1",
      slug: "jwt",
      selector: "guardrails/jwt",
      name: "JWT",
      type: "jwt_validation",
      operation: "validate",
      enforcement: "enforce",
      enabled: true,
      config,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    };
  }

  it("infers the JWT verifier from the first filled method", () => {
    expect(formFromGuardrail(jwtGuardrail({ shared_secret: "s3cret" })).jwtVerifier).toBe("shared");
    expect(formFromGuardrail(jwtGuardrail({ inline_jwks: { keys: [] } })).jwtVerifier).toBe("inline");
    expect(formFromGuardrail(jwtGuardrail({ introspection_url: "https://auth/introspect" })).jwtVerifier).toBe("introspection");
    expect(formFromGuardrail(jwtGuardrail({ jwks_uri: "https://auth/jwks", shared_secret: "s3cret" })).jwtVerifier).toBe("jwks");
    expect(formFromGuardrail(jwtGuardrail({})).jwtVerifier).toBe("jwks");
  });

  it("round-trips a saved guardrail through buildPayload", () => {
    const item = jwtGuardrail({ introspection_url: "https://auth/introspect", introspection_auth_header: "Bearer x", issuer: "iss" });

    expect(buildPayload(formFromGuardrail(item), "edit").config).toEqual(item.config);
  });
});

describe("compilePattern", () => {
  it("turns a leading (?i) into the i flag", () => {
    expect(compilePattern("(?i)password")?.flags).toBe("gi");
    expect(compilePattern("(?i)password")?.test("PASSWORD")).toBe(true);
  });

  it("accepts RE2 named groups", () => {
    expect(compilePattern("(?P<key>sk-\w+)")).not.toBeNull();
  });

  it("returns null for an invalid pattern instead of throwing", () => {
    expect(compilePattern("([a-z")).toBeNull();
  });
});

describe("slug helpers", () => {
  it("derives a slug from a Latin name only", () => {
    expect(slugFromName("  Block API Keys! ")).toBe("block-api-keys");
    expect(slugFromName("جلوگیری از نشت")).toBeNull();
    expect(slugFromType("secrets_detection")).toBe("secrets-detection");
  });
});
