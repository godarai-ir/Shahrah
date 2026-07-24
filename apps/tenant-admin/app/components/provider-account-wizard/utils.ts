import type { ProviderCatalogModel, SupportedCurlGuideMode } from "./types";
import { SUPPORTED_CURL_GUIDE_MODES } from "./types";
export async function requestJSON<T = unknown>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");

  const response = await fetch(path, { ...options, headers, cache: "no-store" });
  const text = await response.text();
  const parsed = tryParseResponsePayload(text);
  if (!response.ok) {
    throw new Error(
      (parsed.ok ? extractMessageFromPayload(parsed.data) : null) ??
        `Request failed (${response.status})`,
    );
  }
  if (!parsed.ok) {
    throw new Error(`Unexpected response format (${response.status})`);
  }
  return parsed.data as T;
}

export function parseResponsePayload(text: string): unknown {
  return tryParseResponsePayload(text).data;
}

function tryParseResponsePayload(
  text: string,
): { ok: true; data: unknown } | { ok: false; data: Record<string, never> } {
  if (!text.trim()) return { ok: true, data: {} };
  try {
    return { ok: true, data: JSON.parse(text) as unknown };
  } catch {
    return { ok: false, data: {} };
  }
}

export function unwrapData<T>(payload: unknown): T | null {
  if (!payload || typeof payload !== "object") return null;
  const direct = payload as T;
  const candidate = payload as { data?: T };
  if (candidate.data !== undefined) return candidate.data;
  return direct;
}

export function extractMessageFromPayload(payload: unknown): string | null {
  if (typeof payload === "string") return payload;
  if (!payload || typeof payload !== "object") return null;
  const candidate = payload as { error?: { message?: string }; message?: string };
  return candidate.error?.message ?? candidate.message ?? null;
}

export function normalizeError(error: unknown, fallback: string): string {
  if (error instanceof Error) return error.message;
  return fallback;
}

export function toNumberOrUndefined(value: string): number | undefined {
  const raw = value.trim();
  if (!raw) return undefined;
  const parsed = Number(raw);
  if (!Number.isFinite(parsed)) return undefined;
  return parsed;
}

export function normalizeCapabilityList(values: string[] | undefined): string[] {
  if (!Array.isArray(values)) return [];
  const seen = new Set<string>();
  const next: string[] = [];
  for (const item of values) {
    const normalized = item.trim().toLowerCase();
    if (!normalized || seen.has(normalized)) continue;
    seen.add(normalized);
    next.push(normalized);
  }
  return next;
}

export function getModelModes(model: Pick<ProviderCatalogModel, "mode" | "modes">): string[] {
  return normalizeCapabilityList([...(model.mode ?? []), ...(model.modes ?? [])]);
}

export function resolveCurlGuideMode(
  model: Pick<
    ProviderCatalogModel,
    "provider" | "model" | "display_name" | "mode" | "modes" | "types" | "capabilities" | "features" | "input_modalities" | "output_modalities"
  >,
): SupportedCurlGuideMode {
  const modes = [
    ...getModelModes(model),
    ...normalizeCapabilityList(model.types),
    ...normalizeCapabilityList(model.capabilities),
    ...normalizeCapabilityList(model.features),
  ];
  const inputModalities = normalizeCapabilityList(model.input_modalities);
  const outputModalities = normalizeCapabilityList(model.output_modalities);
  if (modes.some(isModerationCapability)) return "moderation";
  if (modes.some(isTextToSpeechCapability)) return "text_to_speech";
  if (modes.some(isAudioTranscriptionCapability)) return "audio_transcription";
  if (inputModalities.includes("text") && outputModalities.includes("audio")) return "text_to_speech";
  if (inputModalities.includes("audio") && outputModalities.includes("text")) return "audio_transcription";
  if ([...modes, ...outputModalities].some(isImageGenerationCapability)) {
    return "image_generation";
  }
  for (const mode of modes) {
    if (isSupportedCurlGuideMode(mode)) return mode;
  }
  return "chat";
}

export function isSupportedCurlGuideMode(mode: string): mode is SupportedCurlGuideMode {
  return (SUPPORTED_CURL_GUIDE_MODES as string[]).includes(mode);
}

export function isModerationCapability(value: string): boolean {
  const normalized = value.trim().toLowerCase();
  return normalized === "moderation" || normalized.includes("moderation");
}

export function isImageGenerationCapability(value: string): boolean {
  const normalized = value.trim().toLowerCase().replaceAll("-", "_");
  return normalized === "image" || normalized === "image_generation" || normalized === "image_generation_api";
}

export function isTextToSpeechCapability(value: string): boolean {
  const normalized = normalizeCurlGuideCapability(value);
  return normalized === "text_to_speech" || normalized === "tts" || normalized === "audio_speech" || normalized === "speech_synthesis";
}

export function isAudioTranscriptionCapability(value: string): boolean {
  const normalized = normalizeCurlGuideCapability(value);
  return (
    normalized === "audio_transcription" ||
    normalized === "audio_transcriptions" ||
    normalized === "speech_to_text" ||
    normalized === "stt" ||
    normalized === "transcription" ||
    normalized === "transcribe"
  );
}

export function getCurlGuideEndpoint(mode: SupportedCurlGuideMode): string {
  switch (mode) {
    case "completion":
      return "/v1/completions";
    case "embedding":
      return "/v1/embeddings";
    case "responses":
      return "/v1/responses";
    case "moderation":
      return "/v1/moderations";
    case "image_generation":
      return "/v1/images/generations";
    case "audio_transcription":
      return "/v1/audio/transcriptions";
    case "text_to_speech":
      return "/v1/audio/speech";
    case "chat":
    default:
      return "/v1/chat/completions";
  }
}

export function buildCurlGuidePayload(mode: SupportedCurlGuideMode, modelFqn: string): Record<string, unknown> {
  switch (mode) {
    case "completion":
      return {
        model: modelFqn,
        prompt: "سلام، یک پاسخ کوتاه بده",
        max_tokens: 64,
      };
    case "embedding":
      return {
        model: modelFqn,
        input: "سلام، این متن را برای جستجو برداری کن",
      };
    case "responses":
      return {
        model: modelFqn,
        input: "سلام، یک پاسخ کوتاه بده",
      };
    case "moderation":
      return {
        model: modelFqn,
        input: "I want to check this text for policy violations.",
      };
    case "image_generation":
      return {
        model: modelFqn,
        prompt: "A cinematic photo of Tehran at sunrise",
        n: 1,
        size: "1024x1024",
        quality: "standard",
        response_format: "url",
      };
    case "audio_transcription":
      return {
        model: modelFqn,
        language: "fa",
        response_format: "json",
      };
    case "text_to_speech":
      return {
        model: modelFqn,
        input: "سلام، به گدارAI خوش آمدید.",
        voice: "alloy",
        response_format: "mp3",
      };
    case "chat":
    default:
      return {
        model: modelFqn,
        messages: [
          {
            role: "user",
            content: "سلام، یک پاسخ کوتاه بده",
          },
        ],
      };
  }
}

export function buildBashCurlGuideCommand(input: {
  baseUrl: string;
  endpoint: string;
  payload: Record<string, unknown>;
}): string {
  if (input.endpoint === "/v1/audio/transcriptions") {
    const model = escapeShellSingleQuoted(String(input.payload.model ?? ""));
    const language = escapeShellSingleQuoted(String(input.payload.language ?? "fa"));
    const responseFormat = escapeShellSingleQuoted(String(input.payload.response_format ?? "json"));
    return [
      `curl --location '${input.baseUrl}${input.endpoint}' \\`,
      `  --header 'Authorization: Bearer <PAT_TOKEN>' \\`,
      `  --form 'model=${model}' \\`,
      `  --form 'language=${language}' \\`,
      `  --form 'response_format=${responseFormat}' \\`,
      `  --form 'file=@sample.wav;type=audio/wav'`,
    ].join("\n");
  }

  const payload = JSON.stringify(input.payload, null, 2).replaceAll("'", "'\\''");
  if (input.endpoint === "/v1/audio/speech") {
    return [
      `curl --location '${input.baseUrl}${input.endpoint}' \\`,
      `  --header 'Authorization: Bearer <PAT_TOKEN>' \\`,
      `  --header 'Content-Type: application/json' \\`,
      `  --data '${payload}' \\`,
      `  --output speech.mp3`,
    ].join("\n");
  }

  return [
    `curl --location '${input.baseUrl}${input.endpoint}' \\`,
    `  --header 'Authorization: Bearer <PAT_TOKEN>' \\`,
    `  --header 'Content-Type: application/json' \\`,
    `  --data '${payload}'`,
  ].join("\n");
}

export function buildPowerShellCurlGuideCommand(input: {
  baseUrl: string;
  endpoint: string;
  mode: SupportedCurlGuideMode;
  modelFqn: string;
}): string {
  const modelFqn = input.modelFqn.replaceAll('"', '`"');
  const textPrompt = "سلام، یک پاسخ کوتاه بده";
  const embeddingInput = "سلام، این متن را برای جستجو برداری کن";
  const moderationInput = "I want to check this text for policy violations.";
  const imagePrompt = "A cinematic photo of Tehran at sunrise";
  const speechInput = "سلام، به گدارAI خوش آمدید.";
  const bodyLines = (() => {
    switch (input.mode) {
      case "completion":
        return [
          "$body = @{",
          `  model = "${modelFqn}"`,
          `  prompt = "${textPrompt}"`,
          "  max_tokens = 64",
          "} | ConvertTo-Json -Depth 5",
        ];
      case "embedding":
        return [
          "$body = @{",
          `  model = "${modelFqn}"`,
          `  input = "${embeddingInput}"`,
          "} | ConvertTo-Json -Depth 5",
        ];
      case "responses":
        return [
          "$body = @{",
          `  model = "${modelFqn}"`,
          `  input = "${textPrompt}"`,
          "} | ConvertTo-Json -Depth 5",
        ];
      case "moderation":
        return [
          "$body = @{",
          `  model = "${modelFqn}"`,
          `  input = "${moderationInput}"`,
          "} | ConvertTo-Json -Depth 5",
        ];
      case "image_generation":
        return [
          "$body = @{",
          `  model = "${modelFqn}"`,
          `  prompt = "${imagePrompt}"`,
          "  n = 1",
          '  size = "1024x1024"',
          '  quality = "standard"',
          '  response_format = "url"',
          "} | ConvertTo-Json -Depth 5",
        ];
      case "audio_transcription":
        return [
          "$form = @{",
          `  model = "${modelFqn}"`,
          '  language = "fa"',
          '  response_format = "json"',
          '  file = Get-Item ".\\sample.wav"',
          "}",
        ];
      case "text_to_speech":
        return [
          "$body = @{",
          `  model = "${modelFqn}"`,
          `  input = "${speechInput}"`,
          '  voice = "alloy"',
          '  response_format = "mp3"',
          "} | ConvertTo-Json -Depth 5",
        ];
      case "chat":
      default:
        return [
          "$body = @{",
          `  model = "${modelFqn}"`,
          "  messages = @(",
          "    @{",
          '      role = "user"',
          `      content = "${textPrompt}"`,
          "    }",
          "  )",
          "} | ConvertTo-Json -Depth 5",
        ];
    }
  })();

  const requestLines = [
    "$headers = @{",
    '  Authorization = "Bearer <PAT_TOKEN>"',
    ...(input.mode === "audio_transcription" ? [] : ['  "Content-Type" = "application/json"']),
    "}",
    "",
    ...bodyLines,
    "",
    input.mode === "text_to_speech" ? "Invoke-WebRequest -Method Post `" : "Invoke-RestMethod -Method Post `",
    `  -Uri "${input.baseUrl}${input.endpoint}" \``,
    "  -Headers $headers `",
  ];

  if (input.mode === "audio_transcription") {
    return [...requestLines, "  -Form $form"].join("\n");
  }

  if (input.mode === "text_to_speech") {
    return [...requestLines, "  -Body $body `", '  -OutFile ".\\speech.mp3"'].join("\n");
  }

  return [...requestLines, "  -Body $body"].join("\n");
}

export function buildCurlGuideSampleResponse(mode: SupportedCurlGuideMode, modelName: string): string | undefined {
  if (mode === "audio_transcription") {
    return JSON.stringify(
      {
        text: "سلام، این یک نمونه تبدیل گفتار به متن است.",
      },
      null,
      2,
    );
  }

  if (mode !== "moderation") return undefined;

  return JSON.stringify(
    {
      id: "modr_example",
      model: modelName,
      results: [
        {
          flagged: false,
          categories: {
            sexual: false,
            hate: false,
            harassment: false,
            "self-harm": false,
            violence: false,
          },
          category_scores: {
            sexual: 0.0001,
            hate: 0.0002,
            harassment: 0.0001,
            "self-harm": 0,
            violence: 0.0003,
          },
        },
      ],
    },
    null,
    2,
  );
}

function normalizeCurlGuideCapability(value: string): string {
  return value.trim().toLowerCase().replaceAll("-", "_").replaceAll(" ", "_");
}

function escapeShellSingleQuoted(value: string): string {
  return value.replaceAll("'", "'\\''");
}

export function getModelCapabilities(model: Pick<ProviderCatalogModel, "capabilities" | "features">): string[] {
  const capabilities = normalizeCapabilityList(model.capabilities);
  if (capabilities.length > 0) return capabilities;
  return normalizeCapabilityList(model.features);
}

export function getModelSelectionTypes(model: Pick<ProviderCatalogModel, "types">): string[] {
  return normalizeCapabilityList(model.types);
}

export const FRIENDLY_ACCOUNT_ADJECTIVES = [
  "amber",
  "brisk",
  "calm",
  "clear",
  "cobalt",
  "daring",
  "eager",
  "ember",
  "gentle",
  "golden",
  "lively",
  "lucky",
  "minty",
  "nimble",
  "pearl",
  "rapid",
  "silver",
  "sunny",
  "tidy",
  "velvet",
];

export const FRIENDLY_ACCOUNT_NOUNS = [
  "anchor",
  "breeze",
  "comet",
  "falcon",
  "harbor",
  "lighthouse",
  "lollipop",
  "lotus",
  "meadow",
  "nebula",
  "oasis",
  "orbit",
  "pioneer",
  "ribbon",
  "sparrow",
  "summit",
  "tidal",
  "voyager",
  "willow",
  "zephyr",
];

export function getProviderLogoSrc(provider: string, logoURL?: string): string {
  const proxiedLogoURL = toTenantProviderAssetPath(logoURL);
  if (proxiedLogoURL) return proxiedLogoURL;
  return getProviderFallbackLogoPath(provider);
}

export function toTenantProviderAssetPath(logoURL?: string): string {
  const raw = logoURL?.trim() ?? "";
  if (!raw) return "";
  const candidate = (() => {
    if (raw.startsWith("/")) return raw;
    try {
      return new URL(raw).pathname;
    } catch {
      return raw;
    }
  })();
  const match = candidate.match(/^\/api\/v1\/providers\/assets\/([^/]+)\/([^/]+)$/i);
  if (!match) return raw;
  const [, bucket, fileName] = match;
  return `/api/provider-catalog/assets/${encodeURIComponent(bucket)}/${encodeURIComponent(fileName)}`;
}

export function getModelDiscoveryCatalogBaseURL(): string {
  if (typeof window === "undefined") return "http://localhost:3003";

  return getModelDiscoveryCatalogBaseURLForOrigin(window.location.origin);
}

export function buildModelDiscoveryProviderHref(provider: string, baseURL?: string): string {
  const base = baseURL
    ? normalizeModelDiscoveryCatalogBaseURL(baseURL)
    : getModelDiscoveryCatalogBaseURL();
  return `${base}/providers/${encodeURIComponent(normalizeProviderSlug(provider))}`;
}

export function getModelDiscoveryCatalogBaseURLForOrigin(origin: string): string {
  const publicOrigin = new URL(getPublicSiteOrigin(origin));
  if (isLocalHostname(publicOrigin.hostname)) {
    publicOrigin.port = "3003";
    return `${publicOrigin.origin}/catalog`;
  }
  return `${publicOrigin.origin}/catalog`;
}

export function normalizeModelDiscoveryCatalogBaseURL(value: string): string {
  const trimmed = value.trim().replace(/\/+$/, "");
  if (!trimmed) return trimmed;
  try {
    const url = new URL(trimmed);
    const hostname = url.hostname.toLowerCase();
    if (hostname.startsWith("catalog.") && !isLocalHostname(hostname)) {
      url.hostname = hostname.slice("catalog.".length);
      if (url.pathname === "/" || url.pathname === "") {
        url.pathname = "/catalog";
      }
      url.search = "";
      url.hash = "";
      return url.toString().replace(/\/+$/, "");
    }
  } catch {
    return trimmed;
  }
  return trimmed;
}

export function getTenantApiBaseURL(): string {
  if (typeof window === "undefined") return "http://tenant1.localhost:4000";

  return getTenantApiBaseURLForOrigin(window.location.origin);
}

export function getTenantApiBaseURLForOrigin(origin: string): string {
  const url = new URL(origin);
  if (isLocalHostname(url.hostname)) {
    url.port = "4000";
    return url.origin;
  }
  if (url.hostname.toLowerCase().startsWith("www.")) {
    url.hostname = url.hostname.slice(4);
  }
  url.protocol = "https:";
  url.port = "";
  return url.origin;
}

export function getPublicSiteOrigin(origin: string): string {
  const url = new URL(origin);
  const hostname = url.hostname.toLowerCase();
  if (isLocalHostname(hostname)) {
    url.hostname = "localhost";
    url.port = "";
    return url.origin;
  }
  const parts = hostname.split(".");
  if (parts[0] === "www") {
    url.hostname = parts.slice(1).join(".");
  } else if (parts.length > 2) {
    url.hostname = parts.slice(1).join(".");
  }
  url.protocol = "https:";
  url.port = "";
  return url.origin;
}

function isLocalHostname(hostname: string): boolean {
  const normalized = hostname.toLowerCase();
  return normalized === "localhost" || normalized.endsWith(".localhost") || normalized === "127.0.0.1" || normalized === "[::1]";
}

export function getProviderFallbackLogoPath(provider: string): string {
  const key = normalizeProviderSlug(provider);
  if (key.includes("openai")) return "/providers/openai.svg";
  if (key.includes("gemini") || key.includes("google")) return "/providers/google.svg";
  if (key.includes("vertex")) return "/providers/vertex.svg";
  if (key.includes("azure")) return "/providers/azure.svg";
  if (key.includes("aws") || key.includes("bedrock") || key.includes("sagemaker")) return "/providers/aws.svg";
  if (key.includes("anthropic")) return "/providers/anthropic.svg";
  if (key.includes("cohere")) return "/providers/cohere.svg";
  if (key.includes("groq")) return "/providers/groq.svg";
  if (key.includes("mistral")) return "/providers/mistral.svg";
  return "/globe.svg";
}

export function shouldGenerateProviderAccountName(input: {
  currentName: string;
  currentProvider: string;
  isAutoGenerated: boolean;
}): boolean {
  const currentName = input.currentName.trim().toLowerCase();
  const currentProvider = normalizeProviderSlug(input.currentProvider);
  if (!currentName) return true;
  if (input.isAutoGenerated) return true;
  if (!currentProvider) return false;
  return currentName === currentProvider || currentName.startsWith(`${currentProvider}-`);
}

export function generateProviderAccountName(provider: string, existingNames: string[]): string {
  const normalizedProvider = normalizeProviderSlug(provider) || "provider";
  const usedNames = new Set(existingNames.map((item) => item.trim().toLowerCase()).filter(Boolean));

  for (let attempt = 0; attempt < 24; attempt += 1) {
    const adjective = pickRandom(FRIENDLY_ACCOUNT_ADJECTIVES);
    const noun = pickRandom(FRIENDLY_ACCOUNT_NOUNS);
    const candidate = `${normalizedProvider}-${adjective}-${noun}`;
    if (!usedNames.has(candidate.toLowerCase())) return candidate;
  }

  return `${normalizedProvider}-${pickRandom(FRIENDLY_ACCOUNT_ADJECTIVES)}-${Math.floor(100 + Math.random() * 900)}`;
}

export function pickRandom(values: readonly string[]): string {
  return values[Math.floor(Math.random() * values.length)] ?? "account";
}

export function normalizeProviderSlug(value: string): string {
  const normalized = value.trim().toLowerCase();
  switch (normalized) {
    case "perplexity":
      return "perplexity-ai";
    case "zai":
      return "xai";
    default:
      return normalized;
  }
}

export function stringsEqualFold(a: string, b: string): boolean {
  return normalizeProviderSlug(a) === normalizeProviderSlug(b);
}

export function formatProviderDisplayName(provider: string, displayName?: string): string {
  const normalized = normalizeProviderSlug(provider);
  const knownNames: Record<string, string> = {
    openai: "OpenAI",
    "azure-openai": "Azure OpenAI",
    "azure-open-ai": "Azure OpenAI",
    azure_open_ai: "Azure OpenAI",
    "azure-ai-foundry": "Azure AI Foundry",
    azure_ai_foundry: "Azure AI Foundry",
    azure: "Azure OpenAI",
    ai21: "AI21",
    xai: "xAI",
    "aws-bedrock": "AWS Bedrock",
    bedrock: "AWS Bedrock",
    "google-vertex-ai": "Google Vertex AI",
    vertexai: "Google Vertex AI",
    "mistral-ai": "Mistral AI",
    mistral_ai: "Mistral AI",
    deepinfra: "DeepInfra",
    elevenlabs: "ElevenLabs",
    openrouter: "OpenRouter",
    "perplexity-ai": "Perplexity AI",
    sambanova: "SambaNova",
    "together-ai": "Together AI",
  };
  return knownNames[normalized] ?? displayName?.trim() ?? provider;
}

export function formatDisplayPrice(
  usdPerMillion: number | undefined,
  options: {
    currency: "USD" | "TOMAN";
    scale: "PER_MILLION_TOKENS" | "PER_TOKEN";
    usdToIrr: number;
    naLabel: string;
  },
): string {
  if (typeof usdPerMillion !== "number" || !Number.isFinite(usdPerMillion)) return options.naLabel;

  let value = usdPerMillion;
  if (options.scale === "PER_TOKEN") {
    value = value / 1_000_000;
  }

  if (options.currency === "TOMAN") {
    const tomanValue = (value * options.usdToIrr) / 10;
    const formatted = new Intl.NumberFormat("en-US", {
      maximumFractionDigits: options.scale === "PER_TOKEN" ? 2 : 0,
    }).format(tomanValue);
    const scaleSuffix = options.scale === "PER_TOKEN" ? " / token" : "";
    return `${formatted} toman${scaleSuffix}`;
  }

  const formattedUsd = new Intl.NumberFormat("en-US", {
    maximumFractionDigits: options.scale === "PER_TOKEN" ? 8 : 6,
  }).format(value);
  const scaleSuffix = options.scale === "PER_TOKEN" ? " / token" : "";
  return `$${formattedUsd}${scaleSuffix}`;
}


