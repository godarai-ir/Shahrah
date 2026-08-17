export function normalizeProviderSlug(value: string): string {
  const normalized = value.trim().toLowerCase();
  switch (normalized) {
    case "perplexity":
      return "perplexity-ai";
    default:
      return normalized;
  }
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

export function getProviderLogoSrc(provider: string, logoURL?: string): string {
  return toTenantProviderAssetPath(logoURL) || getProviderFallbackLogoPath(provider);
}
