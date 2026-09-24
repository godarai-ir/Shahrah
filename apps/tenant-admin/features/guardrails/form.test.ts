import { describe, expect, it } from "vitest";

import { buildPayload, emptyForm } from "./form";
import { guardrailTypeSchema, type GuardrailFormInput } from "./types";

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
