import { describe, expect, it } from "vitest";

import {
  applyColumnJSON,
  buildPlaygroundManifest,
  clonePlaygroundColumn,
  createPlaygroundColumn,
  createPlaygroundTool,
  getColumnVariables,
  hydratePlaygroundColumn,
  normalizeImageGenerationResponse,
  serializeColumnJSON,
  validateAndBuildRunPayload,
} from "./playground-domain";

describe("playground domain", () => {
  it("round-trips supported visual fields through JSON", () => {
    const column = createPlaygroundColumn("Option 1", "System {{tone}}");
    column.messages[1].content = "Hello {{name}}";
    column.tools = [{ ...createPlaygroundTool(), name: "lookup", description: "Look up a record" }];
    column.jsonDraft = serializeColumnJSON(column);

    const parsed = applyColumnJSON(column);

    expect(parsed.ok).toBe(true);
    if (!parsed.ok) return;
    expect(parsed.column.messages.map((message) => message.role)).toEqual(["system", "user"]);
    expect(parsed.column.tools[0].name).toBe("lookup");
    expect(parsed.column.config.temperature).toBe("0.2");
  });

  it("preserves invalid JSON and reports its error", () => {
    const column = createPlaygroundColumn("Option 1", "System");
    column.jsonDraft = '{"messages": [';
    expect(applyColumnJSON(column)).toMatchObject({ ok: false });
    expect(column.jsonDraft).toBe('{"messages": [');
  });

  it("rejects unsupported raw fields before they can be silently dropped", () => {
    const column = createPlaygroundColumn("Option 1", "System");
    column.jsonDraft = JSON.stringify({ messages: [], authorization: "secret" });
    expect(applyColumnJSON(column)).toEqual({
      ok: false,
      errorKey: "validation.jsonUnsupportedFields",
      errorValues: { fields: "authorization" },
    });
  });

  it("round-trips structured response format and disables it when JSON removes the field", () => {
    const column = createPlaygroundColumn("Option 1", "System");
    column.config.structuredEnabled = true;
    column.config.structuredFormat = "json_schema";
    column.config.structuredName = "answer";
    column.jsonDraft = serializeColumnJSON(column);

    expect(JSON.parse(column.jsonDraft).parameters.response_format).toMatchObject({
      type: "json_schema",
      json_schema: { name: "answer" },
    });
    const applied = applyColumnJSON(column);
    expect(applied.ok).toBe(true);
    if (!applied.ok) return;
    expect(applied.column.config.structuredEnabled).toBe(true);

    const raw = JSON.parse(applied.column.jsonDraft);
    delete raw.parameters.response_format;
    applied.column.jsonDraft = JSON.stringify(raw);
    const removed = applyColumnJSON(applied.column);
    expect(removed.ok).toBe(true);
    if (!removed.ok) return;
    expect(removed.column.config.structuredEnabled).toBe(false);
  });

  it("rejects unsupported nested message fields", () => {
    const column = createPlaygroundColumn("Option 1", "System");
    column.jsonDraft = JSON.stringify({ messages: [{ role: "user", content: "hi", name: "attacker" }] });
    expect(applyColumnJSON(column)).toEqual({
      ok: false,
      errorKey: "validation.jsonMessageUnsupportedFields",
      errorValues: { index: 1, path: "messages[0]", fields: "name" },
    });
  });

  it("shares detected variables with the runtime payload and renders values", () => {
    const column = createPlaygroundColumn("Option 1", "Use {{tone}}");
    column.selectedModel = "openai:gpt-test";
    column.messages[1].content = "Hello {{name}}";
    column.variableValues = { tone: "brief", name: "Ada" };

    expect(getColumnVariables(column)).toEqual(["name", "tone"]);
    const built = validateAndBuildRunPayload(column, "comparison_1");
    expect(built.ok).toBe(true);
    if (!built.ok) return;
    expect(built.payload.request.messages).toEqual([
      { role: "system", content: "Use brief" },
      { role: "user", content: "Hello Ada" },
    ]);
  });

  it("keeps immutable clone identifiers and stores tools in the manifest parameters", () => {
    const column = createPlaygroundColumn("Option 1", "System");
    column.selectedModel = "virtual-support";
    column.messages[1].content = "Hello";
    column.tools = [{ ...createPlaygroundTool(), name: "lookup" }];
    const clone = clonePlaygroundColumn(column, "Option 2");
    const manifest = buildPlaygroundManifest(clone, "support_prompt", "Support");

    expect(clone.id).not.toBe(column.id);
    expect(clone.messages[0].id).not.toBe(column.messages[0].id);
    expect(manifest.parameters?.tools).toHaveLength(1);
  });

  it("keeps prompt identity fields needed to save another immutable version", () => {
    const hydrated = hydratePlaygroundColumn({
      id: "version-2",
      prompt_id: "prompt-1",
      version: 2,
      fqn: "chat_prompt:acme/support_prompt:2",
      name: "support_prompt",
      description: "Support prompt",
      messages: [{ role: "user", content: "Hello" }],
    }, "Option 1");

    expect(hydrated.sourcePromptName).toBe("support_prompt");
    expect(hydrated.sourcePromptDescription).toBe("Support prompt");
  });

  it("blocks invalid tool schemas before both run and save can proceed", () => {
    const column = createPlaygroundColumn("Option 1", "System");
    column.selectedModel = "gpt-test";
    column.messages[1].content = "Hello";
    column.tools = [{ ...createPlaygroundTool(), name: "lookup", schema: "not-json" }];

    expect(validateAndBuildRunPayload(column, "comparison-1")).toEqual({
      ok: false,
      errorKey: "validation.toolSchemaInvalid",
      errorValues: { name: "lookup" },
    });
  });

  it("builds a dedicated non-streaming image generation request", () => {
    const column = createPlaygroundColumn("Option 1", "System");
    column.selectedModel = "openai:images:gpt-image-1";
    column.imagePrompt = "A {{tone}} portrait";
    column.variableValues = { tone: "cinematic" };
    column.imageConfig = {
      ...column.imageConfig,
      responseFormat: "b64_json",
      size: "1024x1024",
      outputFormat: "webp",
      outputCompression: "80",
    };

    const built = validateAndBuildRunPayload(column, "comparison-1", "image_generation");

    expect(built.ok).toBe(true);
    if (!built.ok) return;
    expect(built.payload).toMatchObject({
      mode: "image_generation",
      request: {
        model: "openai:images:gpt-image-1",
        prompt: "A cinematic portrait",
        n: 1,
        response_format: "b64_json",
        size: "1024x1024",
        output_format: "webp",
        output_compression: 80,
      },
    });
    expect(built.payload).not.toHaveProperty("godarai");
  });

  it("normalizes safe image results and redacts base64 from the raw response", () => {
    const normalized = normalizeImageGenerationResponse({
      data: [
        { b64_json: "YWJj", revised_prompt: "A revised prompt" },
        { url: "javascript:alert(1)" },
        { url: "https://cdn.example.test/image.png" },
      ],
    }, "png");

    expect(normalized.images).toEqual([
      { index: 0, b64Json: "YWJj", revisedPrompt: "A revised prompt", mimeType: "image/png" },
      { index: 2, url: "https://cdn.example.test/image.png", mimeType: "image/png" },
    ]);
    expect(JSON.stringify(normalized.rawResponse)).not.toContain("YWJj");
    expect(JSON.stringify(normalized.rawResponse)).toContain("base64 omitted");
  });
});
