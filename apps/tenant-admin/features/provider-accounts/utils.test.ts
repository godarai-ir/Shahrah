import { describe, expect, it } from "vitest";

import {
  buildBashCurlGuideCommand,
  buildCurlGuidePayload,
  buildPowerShellCurlGuideCommand,
  buildCurlGuideSampleResponse,
  buildModelDiscoveryProviderHref,
  formatProviderDisplayName,
  getCurlGuideEndpoint,
  getModelDiscoveryCatalogBaseURLForOrigin,
  getPublicSiteOrigin,
  getTenantApiBaseURLForOrigin,
  normalizeModelDiscoveryCatalogBaseURL,
  resolveCurlGuideMode,
} from "./utils";

describe("provider account curl guide helpers", () => {
  it("uses official provider casing without title-casing unknown identifiers", () => {
    expect(formatProviderDisplayName("openai")).toBe("OpenAI");
    expect(formatProviderDisplayName("azure-openai")).toBe("Azure OpenAI");
    expect(formatProviderDisplayName("azure_open_ai")).toBe("Azure OpenAI");
    expect(formatProviderDisplayName("azure-ai-foundry")).toBe("Azure AI Foundry");
    expect(formatProviderDisplayName("mistral-ai")).toBe("Mistral AI");
    expect(formatProviderDisplayName("zai")).toBe("Z.ai");
    expect(formatProviderDisplayName("tokenbazaar")).toBe("TokenBazaar");
    expect(formatProviderDisplayName("custom_provider")).toBe("custom_provider");
  });

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

  it("detects moderation from types when no endpoint mode is declared", () => {
    expect(resolveCurlGuideMode({
      provider: "openai",
      model: "omni-moderation-latest",
      display_name: "Omni Moderation",
      types: ["safety-moderation"],
    })).toBe("moderation");
  });

  it("keeps the declared endpoint mode ahead of secondary capabilities", () => {
    expect(resolveCurlGuideMode({
      provider: "openai",
      model: "multimodal-chat",
      display_name: "Multimodal Chat",
      mode: ["chat"],
      types: ["chat"],
      capabilities: ["moderation", "stt", "tts", "image_generation"],
    })).toBe("chat");
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

  it("detects image aliases from endpoint metadata without provider-specific branching", () => {
    expect(resolveCurlGuideMode({
      provider: "future-compatible-provider",
      model: "future-image-model",
      display_name: "Future Image Model",
      mode: ["image"],
      types: [],
    })).toBe("image_generation");

    expect(resolveCurlGuideMode({
      provider: "another-provider",
      model: "poster-maker",
      display_name: "Poster Maker",
      output_modalities: ["image"],
      types: [],
    })).toBe("chat");
  });

  it("does not infer dedicated audio endpoints from multimodal chat content", () => {
    expect(resolveCurlGuideMode({
      provider: "google-gemini",
      model: "gemini-3.5-flash-lite",
      display_name: "Gemini 3.5 Flash Lite",
      mode: ["chat", "responses"],
      types: ["chat", "responses"],
      input_modalities: ["text", "image", "audio", "video", "pdf"],
      output_modalities: ["text"],
    })).toBe("chat");

    expect(resolveCurlGuideMode({
      provider: "future-compatible-provider",
      model: "multimodal-audio-chat",
      display_name: "Multimodal Audio Chat",
      mode: ["chat"],
      types: [],
      input_modalities: ["text"],
      output_modalities: ["audio"],
    })).toBe("chat");
  });

  it("builds multipart speech-to-text samples for audio transcription models", () => {
    const modelFqn = "openai:default:gpt-4o-transcribe";
    const mode = resolveCurlGuideMode({
      provider: "openai",
      model: "gpt-4o-transcribe",
      display_name: "GPT-4o Transcribe",
      mode: ["audio_transcription"],
      types: ["audio_transcription"],
      input_modalities: ["audio"],
      output_modalities: ["text"],
    });
    const payload = buildCurlGuidePayload(mode, modelFqn);
    const command = buildBashCurlGuideCommand({
      baseUrl: "http://tenant1.localhost:4000",
      endpoint: getCurlGuideEndpoint(mode),
      payload,
    });
    const response = buildCurlGuideSampleResponse(mode, "gpt-4o-transcribe");

    expect(mode).toBe("audio_transcription");
    expect(getCurlGuideEndpoint(mode)).toBe("/v1/audio/transcriptions");
    expect(payload).toEqual({
      model: modelFqn,
      language: "fa",
      response_format: "json",
    });
    expect(command).toContain("/v1/audio/transcriptions");
    expect(command).toContain("Authorization: Bearer <PAT_TOKEN>");
    expect(command).not.toContain("Content-Type: application/json");
    expect(command).toContain(`--form 'model=${modelFqn}'`);
    expect(command).toContain("--form 'language=fa'");
    expect(command).toContain("--form 'response_format=json'");
    expect(command).toContain("--form 'file=@sample.wav;type=audio/wav'");
    expect(response).toContain('"text"');
  });

  it("detects STT and transcription aliases for the transcription sample", () => {
    expect(resolveCurlGuideMode({
      provider: "deepgram",
      model: "nova-3-general",
      display_name: "Nova 3",
      types: ["stt"],
    })).toBe("audio_transcription");

    expect(resolveCurlGuideMode({
      provider: "cohere",
      model: "cohere-transcribe-03-2026",
      display_name: "Cohere Transcribe",
      mode: ["transcription"],
      types: [],
    })).toBe("audio_transcription");
  });

  it("builds JSON text-to-speech samples for TTS models", () => {
    const modelFqn = "openai:default:gpt-4o-mini-tts";
    const mode = resolveCurlGuideMode({
      provider: "openai",
      model: "gpt-4o-mini-tts",
      display_name: "GPT-4o mini TTS",
      mode: ["text_to_speech"],
      types: ["tts"],
      input_modalities: ["text"],
      output_modalities: ["audio"],
    });
    const payload = buildCurlGuidePayload(mode, modelFqn);
    const bashCommand = buildBashCurlGuideCommand({
      baseUrl: "http://tenant1.localhost:4000",
      endpoint: getCurlGuideEndpoint(mode),
      payload,
    });
    const powershellCommand = buildPowerShellCurlGuideCommand({
      baseUrl: "http://tenant1.localhost:4000",
      endpoint: getCurlGuideEndpoint(mode),
      mode,
      modelFqn,
    });

    expect(mode).toBe("text_to_speech");
    expect(getCurlGuideEndpoint(mode)).toBe("/v1/audio/speech");
    expect(payload).toEqual({
      model: modelFqn,
      input: "سلام، به گدارAI خوش آمدید.",
      voice: "alloy",
      response_format: "mp3",
    });
    expect(bashCommand).toContain("/v1/audio/speech");
    expect(bashCommand).toContain("Content-Type: application/json");
    expect(bashCommand).toContain('"voice": "alloy"');
    expect(bashCommand).toContain("--output speech.mp3");
    expect(powershellCommand).toContain("/v1/audio/speech");
    expect(powershellCommand).toContain("Invoke-WebRequest");
    expect(powershellCommand).toContain('voice = "alloy"');
    expect(powershellCommand).toContain('response_format = "mp3"');
    expect(powershellCommand).toContain('-OutFile ".\\speech.mp3"');
  });

  it("builds PowerShell transcription samples as multipart form requests", () => {
    const command = buildPowerShellCurlGuideCommand({
      baseUrl: "http://tenant1.localhost:4000",
      endpoint: getCurlGuideEndpoint("audio_transcription"),
      mode: "audio_transcription",
      modelFqn: "openai:default:gpt-4o-transcribe",
    });

    expect(command).toContain("/v1/audio/transcriptions");
    expect(command).toContain('Authorization = "Bearer <PAT_TOKEN>"');
    expect(command).not.toContain('"Content-Type" = "application/json"');
    expect(command).toContain("$form = @{");
    expect(command).toContain('model = "openai:default:gpt-4o-transcribe"');
    expect(command).toContain('language = "fa"');
    expect(command).toContain('file = Get-Item ".\\sample.wav"');
    expect(command).toContain("-Form $form");
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
    expect(getModelDiscoveryCatalogBaseURLForOrigin("https://delpak.godarai.ir")).toBe("https://godarai.ir/catalog");
    expect(getModelDiscoveryCatalogBaseURLForOrigin("http://tenant1.localhost:3000")).toBe("http://localhost:3003/catalog");
  });

  it("normalizes the legacy catalog subdomain to the root catalog path", () => {
    expect(normalizeModelDiscoveryCatalogBaseURL("https://catalog.godarai.ir/")).toBe("https://godarai.ir/catalog");
    expect(normalizeModelDiscoveryCatalogBaseURL("https://godarai.ir/catalog/")).toBe("https://godarai.ir/catalog");
  });

  it("builds provider profile links on the root catalog path", () => {
    expect(buildModelDiscoveryProviderHref("anthropic", "https://catalog.godarai.ir")).toBe(
      "https://godarai.ir/catalog/providers/anthropic",
    );
  });
});
