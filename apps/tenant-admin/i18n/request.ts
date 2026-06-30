import { getRequestConfig } from "next-intl/server";

const SUPPORTED_LOCALES = new Set(["fa", "en"]);
const MESSAGE_NAMESPACES = [
  "theme",
  "shell",
  "common",
  "dashboard",
  "topupPage",
  "topupResult",
  "login",
  "settings",
  "providerAccounts",
  "virtualModels",
  "virtualUsers",
  "accessManagement",
  "orgUnits",
  "teams",
  "personalAccessTokens",
  "traces",
  "logs",
  "metrics",
  "invite",
  "rateLimitRules",
  "budgetLimitRules",
  "guardrails",
  "prompts",
] as const;

async function loadMessages(locale: "fa" | "en") {
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

export default getRequestConfig(async ({ requestLocale }) => {
  const localeCandidate = await requestLocale;
  const locale = (localeCandidate && SUPPORTED_LOCALES.has(localeCandidate) ? localeCandidate : "fa") as "fa" | "en";

  return {
    locale,
    messages: await loadMessages(locale),
    timeZone: "Asia/Tehran",
  };
});
