import type { BetterAuthOptions } from "better-auth";
import { betterAuth } from "better-auth";

import { resolveClientIPAddressOptions } from "./client-ip-config";
import { socialAuthPool } from "./database";

export const SOCIAL_AUTH_PROVIDERS = ["google", "github", "microsoft"] as const;
export type SocialAuthProvider = (typeof SOCIAL_AUTH_PROVIDERS)[number];

const authBaseURL = normalizeOrigin(
  process.env.GODARAI_AUTH_BASE_URL ?? "http://auth.localhost:3000",
);
const clientIPAddressOptions = resolveClientIPAddressOptions();

const socialProviders: NonNullable<BetterAuthOptions["socialProviders"]> = {};

if (process.env.GOOGLE_CLIENT_ID && process.env.GOOGLE_CLIENT_SECRET) {
  socialProviders.google = {
    clientId: process.env.GOOGLE_CLIENT_ID,
    clientSecret: process.env.GOOGLE_CLIENT_SECRET,
    prompt: "select_account",
  };
}

if (process.env.GITHUB_CLIENT_ID && process.env.GITHUB_CLIENT_SECRET) {
  socialProviders.github = {
    clientId: process.env.GITHUB_CLIENT_ID,
    clientSecret: process.env.GITHUB_CLIENT_SECRET,
    scope: ["read:user", "user:email"],
  };
}

if (process.env.MICROSOFT_CLIENT_ID && process.env.MICROSOFT_CLIENT_SECRET) {
  socialProviders.microsoft = {
    clientId: process.env.MICROSOFT_CLIENT_ID,
    clientSecret: process.env.MICROSOFT_CLIENT_SECRET,
    tenantId: process.env.MICROSOFT_TENANT_ID ?? "common",
    prompt: "select_account",
  };
}

export const socialAuth = betterAuth({
  appName: "GodarAI",
  baseURL: authBaseURL,
  basePath: "/api/auth",
  database: socialAuthPool,
  secret: process.env.BETTER_AUTH_SECRET,
  trustedOrigins: [authBaseURL],
  socialProviders,
  user: {
    modelName: "social_auth_users",
  },
  session: {
    modelName: "social_auth_sessions",
    expiresIn: 60 * 60 * 12,
    updateAge: 60 * 60,
    freshAge: 60 * 10,
  },
  account: {
    modelName: "social_auth_accounts",
    encryptOAuthTokens: true,
    storeStateStrategy: "database",
    storeAccountCookie: false,
    updateAccountOnSignIn: false,
    accountLinking: {
      enabled: true,
      disableImplicitLinking: true,
      allowDifferentEmails: false,
      allowUnlinkingAll: false,
      updateUserInfoOnLink: false,
      trustedProviders: [],
    },
  },
  verification: {
    modelName: "social_auth_verifications",
  },
  rateLimit: {
    enabled: process.env.NODE_ENV === "production",
    storage: "database",
    modelName: "social_auth_rate_limits",
    window: 60,
    max: 30,
    customRules: {
      "/sign-in/social": { window: 60, max: 10 },
      "/callback/*": { window: 60, max: 20 },
    },
  },
  advanced: {
    ipAddress: clientIPAddressOptions,
    useSecureCookies: authBaseURL.startsWith("https://"),
    disableOriginCheck: false,
    defaultCookieAttributes: {
      httpOnly: true,
      secure: authBaseURL.startsWith("https://"),
      sameSite: "lax",
      path: "/",
    },
    crossSubDomainCookies: {
      enabled: false,
    },
  },
});

export function socialAuthBaseURL() {
  return authBaseURL;
}

export function isSocialAuthProvider(value: string): value is SocialAuthProvider {
  return SOCIAL_AUTH_PROVIDERS.includes(value as SocialAuthProvider);
}

export function isSocialAuthProviderEnabled(provider: SocialAuthProvider) {
  return Object.prototype.hasOwnProperty.call(socialProviders, provider);
}

function normalizeOrigin(value: string) {
  const url = new URL(value);
  return url.origin;
}
