import { describe, expect, it } from "vitest";

import {
  buildPlaygroundHref,
  modelModesSupportPlaygroundMode,
  normalizePlaygroundMode,
} from "./playground-mode";

describe("playground mode helpers", () => {
  it("normalizes supported endpoint aliases", () => {
    expect(normalizePlaygroundMode("chat_completions")).toBe("chat");
    expect(normalizePlaygroundMode("image")).toBe("image_generation");
    expect(normalizePlaygroundMode("text-to-image")).toBe("image_generation");
  });

  it("does not map unsupported endpoint modes to chat", () => {
    expect(normalizePlaygroundMode("audio_transcription")).toBeNull();
    expect(normalizePlaygroundMode("text_to_speech")).toBeNull();
    expect(normalizePlaygroundMode("embedding")).toBeNull();
  });

  it("builds links only for modes implemented by the playground", () => {
    expect(buildPlaygroundHref("google-gemini:account:gemini-3.5-flash-lite", "chat")).toBe(
      "/playground?model=google-gemini%3Aaccount%3Agemini-3.5-flash-lite&mode=chat",
    );
    expect(buildPlaygroundHref("openai:account:gpt-image-2", "image")).toBe(
      "/playground?model=openai%3Aaccount%3Agpt-image-2&mode=image_generation",
    );
    expect(buildPlaygroundHref("openai:account:gpt-transcribe", "audio_transcription")).toBeNull();
  });

  it("matches model mode aliases without treating unrelated modes as chat", () => {
    expect(modelModesSupportPlaygroundMode(["image"], "image_generation")).toBe(true);
    expect(modelModesSupportPlaygroundMode(["audio_transcription"], "chat")).toBe(false);
    expect(modelModesSupportPlaygroundMode([], "chat")).toBe(true);
  });
});
