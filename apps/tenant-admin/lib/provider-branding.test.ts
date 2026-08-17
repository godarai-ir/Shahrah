import { describe, expect, it } from "vitest";

import { normalizeProviderSlug } from "./provider-branding";

describe("provider branding", () => {
  it("keeps Z.ai isolated from xAI", () => {
    expect(normalizeProviderSlug("zai")).toBe("zai");
    expect(normalizeProviderSlug("xai")).toBe("xai");
  });
});
