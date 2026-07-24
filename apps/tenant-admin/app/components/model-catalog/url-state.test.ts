import { describe, expect, it } from "vitest";

import { parseModelCatalogURLState, patchModelCatalogURL } from "./url-state";

describe("model catalog URL state", () => {
  it("maps legacy and invalid tabs to providers", () => {
    expect(parseModelCatalogURLState(new URLSearchParams("tab=all")).tab).toBe("providers");
    expect(parseModelCatalogURLState(new URLSearchParams("tab=connected")).tab).toBe("providers");
  });

  it("parses allowlisted state and normalizes CSV values", () => {
    const state = parseModelCatalogURLState(
      new URLSearchParams(
        "tab=models&providers=openai,anthropic,openai&accounts=a1,a2&sort=context&order=desc&section=models",
      ),
    );
    expect(state).toMatchObject({
      tab: "models",
      providers: ["openai", "anthropic"],
      accounts: ["a1", "a2"],
      sort: "context",
      order: "desc",
      section: "models",
    });
  });

  it("patches state without discarding unrelated catalog filters", () => {
    const next = patchModelCatalogURL(new URLSearchParams("tab=models&q=gpt&providers=openai"), {
      model_provider: "openai",
      model: "gpt-4.1",
    });
    expect(next).toContain("q=gpt");
    expect(next).toContain("providers=openai");
    expect(next).toContain("model=gpt-4.1");
  });
});
