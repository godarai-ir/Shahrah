import type { ProviderCatalogModel, SupportedCurlGuideMode } from "./types";
import { SUPPORTED_CURL_GUIDE_MODES } from "./types";
import {
  normalizeProviderSlug,
} from "@/lib/provider-branding";

export {
  getProviderFallbackLogoPath,
  getProviderLogoSrc,
  normalizeProviderSlug,
  toTenantProviderAssetPath,
} from "@/lib/provider-branding";
export { buildBashCurlGuideCommand } from "@/lib/api-guides";
export { getTenantApiBaseURL, getTenantApiBaseURLForOrigin } from "@/lib/tenant-api-origin";
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
  const capabilityGroups = [
    getModelModes(model),
    normalizeCapabilityList(model.types),
    normalizeCapabilityList(model.capabilities),
    normalizeCapabilityList(model.features),
  ];

  // Endpoint modes are authoritative. Modalities only describe supported
  // content and cannot distinguish multimodal chat from dedicated audio or
  // image endpoints.
  for (const capabilities of capabilityGroups) {
    const mode = resolveCurlGuideModeFromCapabilities(capabilities);
    if (mode) return mode;
  }
  return "chat";
}

function resolveCurlGuideModeFromCapabilities(capabilities: string[]): SupportedCurlGuideMode | undefined {
  for (const capability of capabilities) {
    const normalized = normalizeCurlGuideCapability(capability);
    if (isModerationCapability(normalized)) return "moderation";
    if (isTextToSpeechCapability(normalized)) return "text_to_speech";
    if (isAudioTranscriptionCapability(normalized)) return "audio_transcription";
    if (isImageGenerationCapability(normalized)) return "image_generation";
    if (isSupportedCurlGuideMode(normalized)) return normalized;
  }
  return undefined;
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


