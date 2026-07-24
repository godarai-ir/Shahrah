import { describe, expect, it } from "vitest";

import { formatCatalogModelDisplayName } from "./formatters";

describe("model catalog formatters", () => {
  it("preserves official display names and normalizes known acronyms", () => {
    expect(formatCatalogModelDisplayName("gpt-4o", "GPT-4o mini")).toBe("GPT-4o mini");
    expect(formatCatalogModelDisplayName("tts-1")).toBe("TTS-1");
    expect(formatCatalogModelDisplayName("pdf-api")).toBe("PDF-API");
  });

  it("does not title-case raw model identifiers", () => {
    expect(formatCatalogModelDisplayName("claude-3-5-sonnet")).toBe("claude-3-5-sonnet");
  });
});
