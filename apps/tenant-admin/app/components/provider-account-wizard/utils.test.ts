import { describe, expect, it } from "vitest";

import {
  buildBashCurlGuideCommand,
  buildCurlGuidePayload,
  buildCurlGuideSampleResponse,
  getCurlGuideEndpoint,
  getPublicSiteOrigin,
  getTenantApiBaseURLForOrigin,
  resolveCurlGuideMode,
} from "./utils";

describe("provider account curl guide helpers", () => {
  it("uses the moderation endpoint when a model mode is moderation", () => {
    const mode = resolveCurlGuideMode({
      provider: "openai",
      model: "omni-moderation-latest",
      display_name: "Omni Moderation",
      mode: ["moderation"],
      types: [],
    });

    expect(mode).toBe("moderation");
    expect(getCurlGuideEndpoint(mode)).toBe("/v1/moderations");
  });

  it("detects moderation from types and capability values containing moderation", () => {
    expect(resolveCurlGuideMode({
      provider: "openai",
      model: "omni-moderation-latest",
      display_name: "Omni Moderation",
      types: ["safety-moderation"],
    })).toBe("moderation");

    expect(resolveCurlGuideMode({
      provider: "openai",
      model: "omni-moderation-latest",
      display_name: "Omni Moderation",
      mode: ["chat"],
      types: ["chat"],
      capabilities: ["moderation"],
    })).toBe("moderation");
  });

  it("does not use moderation for non-moderation models", () => {
    const mode = resolveCurlGuideMode({
      provider: "openai",
      model: "gpt-4o-mini",
      display_name: "GPT-4o mini",
      mode: ["chat"],
      types: ["chat"],
    });

    expect(mode).toBe("chat");
    expect(getCurlGuideEndpoint(mode)).toBe("/v1/chat/completions");
  });

  it("builds moderation request and response examples with the selected model", () => {
    const modelFqn = "openai:default:omni-moderation-latest";
    const payload = buildCurlGuidePayload("moderation", modelFqn);
    const command = buildBashCurlGuideCommand({
      baseUrl: "http://tenant1.localhost:4000",
      endpoint: getCurlGuideEndpoint("moderation"),
      payload,
    });
    const response = buildCurlGuideSampleResponse("moderation", "omni-moderation-latest");

    expect(payload).toEqual({
      model: modelFqn,
      input: "I want to check this text for policy violations.",
    });
    expect(command).toContain("/v1/moderations");
    expect(command).toContain("Authorization: Bearer <PAT_TOKEN>");
    expect(command).toContain("Content-Type: application/json");
    expect(command).toContain(`"model": "${modelFqn}"`);
    expect(response).toContain('"id": "modr_example"');
    expect(response).toContain('"model": "omni-moderation-latest"');
    expect(response).toContain('"category_scores"');
  });

  it("keeps production tenant API URLs on the tenant subdomain without ports", () => {
    expect(getTenantApiBaseURLForOrigin("http://tenant1.godarai.ir:4000")).toBe("https://tenant1.godarai.ir");
    expect(getTenantApiBaseURLForOrigin("https://tenant2.godarai.ir")).toBe("https://tenant2.godarai.ir");
  });

  it("normalizes www to the root domain instead of treating it as a tenant", () => {
    expect(getTenantApiBaseURLForOrigin("http://www.godarai.ir:4000")).toBe("https://godarai.ir");
    expect(getPublicSiteOrigin("http://www.godarai.ir:4000")).toBe("https://godarai.ir");
  });

  it("builds public catalog origins on the root site instead of tenant subdomains", () => {
    expect(getPublicSiteOrigin("https://delpak.godarai.ir")).toBe("https://godarai.ir");
    expect(getPublicSiteOrigin("http://tenant1.localhost:3000")).toBe("http://localhost");
  });
});
