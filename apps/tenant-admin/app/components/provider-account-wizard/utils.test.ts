import { describe, expect, it } from "vitest";

import {
  buildBashCurlGuideCommand,
  buildCurlGuidePayload,
  buildPowerShellCurlGuideCommand,
  buildCurlGuideSampleResponse,
  getCurlGuideEndpoint,
  getModelDiscoveryCatalogBaseURLForOrigin,
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

  it("keeps chat models on the chat completions sample", () => {
    const modelFqn = "openai:default:gpt-4o-mini";
    const mode = resolveCurlGuideMode({
      provider: "openai",
      model: "gpt-4o-mini",
      display_name: "GPT-4o mini",
      mode: ["chat"],
      types: ["chat"],
      capabilities: ["tools"],
    });
    const payload = buildCurlGuidePayload(mode, modelFqn);
    const command = buildBashCurlGuideCommand({
      baseUrl: "http://tenant1.localhost:4000",
      endpoint: getCurlGuideEndpoint(mode),
      payload,
    });

    expect(mode).toBe("chat");
    expect(command).toContain("/v1/chat/completions");
    expect(command).toContain(`"model": "${modelFqn}"`);
    expect(command).toContain('"messages"');
  });

  it("uses the image generations endpoint for image generation mode", () => {
    const modelFqn = "xai:prod:grok-imagine-image";
    const mode = resolveCurlGuideMode({
      provider: "xai",
      model: "grok-imagine-image",
      display_name: "Grok Imagine",
      mode: ["image_generation"],
      types: ["image_generation"],
      capabilities: [],
    });
    const payload = buildCurlGuidePayload(mode, modelFqn);
    const command = buildBashCurlGuideCommand({
      baseUrl: "http://tenant1.localhost:4000",
      endpoint: getCurlGuideEndpoint(mode),
      payload,
    });

    expect(mode).toBe("image_generation");
    expect(getCurlGuideEndpoint(mode)).toBe("/v1/images/generations");
    expect(payload).toEqual({
      model: modelFqn,
      prompt: "A cinematic photo of Tehran at sunrise",
      n: 1,
      size: "1024x1024",
      quality: "standard",
      response_format: "url",
    });
    expect(command).toContain("/v1/images/generations");
    expect(command).toContain("Authorization: Bearer <PAT_TOKEN>");
    expect(command).toContain("Content-Type: application/json");
    expect(command).toContain(`"model": "${modelFqn}"`);
    expect(command).toContain('"prompt": "A cinematic photo of Tehran at sunrise"');
    expect(command).toContain('"n": 1');
    expect(command).toContain('"size": "1024x1024"');
    expect(command).toContain('"response_format": "url"');
  });

  it("detects image mode generically without provider-specific branching", () => {
    expect(resolveCurlGuideMode({
      provider: "future-compatible-provider",
      model: "future-image-model",
      display_name: "Future Image Model",
      types: ["image"],
    })).toBe("image_generation");

    expect(resolveCurlGuideMode({
      provider: "another-provider",
      model: "poster-maker",
      display_name: "Poster Maker",
      output_modalities: ["image"],
      types: [],
    })).toBe("image_generation");
  });

  it("builds PowerShell image generation samples with the selected model", () => {
    const command = buildPowerShellCurlGuideCommand({
      baseUrl: "http://tenant1.localhost:4000",
      endpoint: getCurlGuideEndpoint("image_generation"),
      mode: "image_generation",
      modelFqn: "openai:default:gpt-image-1",
    });

    expect(command).toContain("/v1/images/generations");
    expect(command).toContain('Authorization = "Bearer <PAT_TOKEN>"');
    expect(command).toContain('model = "openai:default:gpt-image-1"');
    expect(command).toContain('prompt = "A cinematic photo of Tehran at sunrise"');
    expect(command).toContain("n = 1");
    expect(command).toContain('size = "1024x1024"');
    expect(command).toContain('response_format = "url"');
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
    expect(getModelDiscoveryCatalogBaseURLForOrigin("https://delpak.godarai.ir")).toBe("https://catalog.godarai.ir");
    expect(getModelDiscoveryCatalogBaseURLForOrigin("http://tenant1.localhost:3000")).toBe("http://localhost:3003");
  });
});
