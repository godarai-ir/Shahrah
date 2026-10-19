import { cookies, headers } from "next/headers";
import { getRequestConfig } from "next-intl/server";

import {
  DEFAULT_LOCALE,
  LOCALE_COOKIE_NAME,
  isSupportedLocale,
  type AppLocale,
} from "@/lib/locale";

const MESSAGE_NAMESPACES = [
  "theme",
  "language",
  "shell",
  "common",
  "dashboard",
  "quickStart",
  "billing",
  "billingPlan",
  "login",
  "settings",
  "providerAccounts",
  "modelCatalog",
  "virtualModels",
  "virtualUsers",
  "accessManagement",
  "teams",
  "teamsPage",
  "personalAccessTokens",
  "oidcClients",
  "traces",
  "logs",
  "metrics",
  "invite",
  "gatewayPolicies",
  "rateLimitRules",
  "budgetLimitRules",
  "guardrails",
  "prompts",
  "playground",
  "support",
] as const;

async function loadMessages(locale: AppLocale) {
  const baseMessages = (await import(`../messages/${locale}.json`)).default as Record<string, unknown>;
  const messages: Record<string, unknown> = {};

  for (const namespace of MESSAGE_NAMESPACES) {
    try {
      const moduleMessages = (await import(`../messages/modules/${namespace}/${locale}.json`)).default;
      messages[namespace] = moduleMessages;
    } catch {
      // Temporary additive fallback while modularization rolls out.
      if (namespace in baseMessages) {
        messages[namespace] = baseMessages[namespace];
      }
    }
  }

  return messages;
}

/**
 * Best supported locale advertised by the browser. Used only when the visitor
 * has never picked a language, so the explicit cookie always wins.
 */
export function matchAcceptLanguage(headerValue: string | null | undefined): AppLocale | null {
  if (!headerValue) return null;

  const ranked = headerValue
    .split(",")
    .map((part) => {
      const [tag, ...params] = part.trim().split(";");
      const quality = params
        .map((param) => param.trim())
        .find((param) => param.startsWith("q="))
        ?.slice(2);
      const parsedQuality = quality === undefined ? 1 : Number.parseFloat(quality);
      return { tag: tag.trim(), quality: Number.isFinite(parsedQuality) ? parsedQuality : 0 };
    })
    .filter((entry) => entry.tag.length > 0 && entry.quality > 0)
    .sort((a, b) => b.quality - a.quality);

  for (const entry of ranked) {
    if (entry.tag === "*") return DEFAULT_LOCALE;
    const language = entry.tag.toLowerCase().split(/[-_]/, 1)[0];
    if (isSupportedLocale(language)) return language;
  }

  return null;
}

export async function resolveRequestLocale(requestLocale: string | undefined): Promise<AppLocale> {
  if (isSupportedLocale(requestLocale)) return requestLocale;

  const cookieStore = await cookies();
  const cookieLocale = cookieStore.get(LOCALE_COOKIE_NAME)?.value;
  if (isSupportedLocale(cookieLocale)) return cookieLocale;

  const requestHeaders = await headers();
  return matchAcceptLanguage(requestHeaders.get("accept-language")) ?? DEFAULT_LOCALE;
}

export default getRequestConfig(async ({ requestLocale }) => {
  const locale = await resolveRequestLocale(await requestLocale);

  return {
    locale,
    messages: await loadMessages(locale),
    timeZone: "Asia/Tehran",
  };
});
