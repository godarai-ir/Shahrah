export const SOCIAL_AUTH_PROVIDERS = ["google", "github", "microsoft"] as const;

export type SocialAuthProvider = (typeof SOCIAL_AUTH_PROVIDERS)[number];

export function isSocialAuthProvider(value: string): value is SocialAuthProvider {
  return SOCIAL_AUTH_PROVIDERS.includes(value as SocialAuthProvider);
}

export function resolveCentralSocialAuthBaseURL(
  environment: NodeJS.ProcessEnv = process.env,
): URL | null {
  const configured = environment.GODARAI_AUTH_BASE_URL?.trim();
  const candidate = configured || (environment.NODE_ENV === "production" ? "" : "http://localhost:3010");
  if (!candidate) {
    return null;
  }

  try {
    const url = new URL(candidate);
    if (!["http:", "https:"].includes(url.protocol) || url.username || url.password) {
      return null;
    }
    return new URL(url.origin);
  } catch {
    return null;
  }
}

export function normalizeSocialReturnPath(value: string | null | undefined): string {
  const normalized = value?.trim() ?? "";
  if (!normalized.startsWith("/") || normalized.startsWith("//")) {
    return "/";
  }
  return normalized.slice(0, 1024);
}
