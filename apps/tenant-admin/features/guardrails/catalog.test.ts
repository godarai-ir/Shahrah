import { describe, expect, it } from "vitest";

import { defaultOperationForType, GUARDRAIL_CATALOG, isMutateOnlyType, isValidateOnlyType, operationForType } from "./catalog";
import { guardrailTypeSchema } from "./types";

describe("GUARDRAIL_CATALOG", () => {
  it("covers every guardrail type exactly once", () => {
    const catalogTypes = GUARDRAIL_CATALOG.map((entry) => entry.type);

    expect(new Set(catalogTypes).size).toBe(catalogTypes.length);
    expect([...catalogTypes].sort()).toEqual([...guardrailTypeSchema.options].sort());
  });

  it("gives every count type a unit and no other type one", () => {
    for (const entry of GUARDRAIL_CATALOG) {
      expect(Boolean(entry.unit)).toBe(entry.type.endsWith("_count"));
    }
  });

  it.each(GUARDRAIL_CATALOG)("derives the operation of $type from its ops", ({ type, ops }) => {
    expect(isValidateOnlyType(type)).toBe(ops === "validate");
    expect(isMutateOnlyType(type)).toBe(ops === "mutate");

    if (ops === "both") {
      expect(operationForType(type, "validate")).toBe("validate");
      expect(operationForType(type, "mutate")).toBe("mutate");
    } else {
      expect(operationForType(type, "validate")).toBe(ops);
      expect(operationForType(type, "mutate")).toBe(ops);
      expect(defaultOperationForType(type)).toBe(ops);
    }
  });

  it("matches the operation lock the form enforced before the catalog existed", () => {
    const validateOnly = ["keyword_blocklist", "sentence_count", "word_count", "character_count", "json_schema", "json_keys", "contains_code", "not_null", "jwt_validation", "semantic_policy", "document_confidentiality"];
    const mutateOnly = ["regex_replace", "inline_image_urls"];

    for (const { type } of GUARDRAIL_CATALOG) {
      expect(isValidateOnlyType(type)).toBe(validateOnly.includes(type));
      expect(isMutateOnlyType(type)).toBe(mutateOnly.includes(type));
    }
  });

  it("preselects mutate for flexible types except regex", () => {
    expect(defaultOperationForType("secrets_detection")).toBe("mutate");
    expect(defaultOperationForType("basic_pii")).toBe("mutate");
    expect(defaultOperationForType("regex")).toBe("validate");
  });
});
