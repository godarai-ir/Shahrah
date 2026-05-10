import { getRequestConfig } from "next-intl/server";

const SUPPORTED_LOCALES = new Set(["fa", "en"]);

export default getRequestConfig(async ({ requestLocale }) => {
  const localeCandidate = await requestLocale;
  const locale = localeCandidate && SUPPORTED_LOCALES.has(localeCandidate) ? localeCandidate : "fa";

  const baseMessages = (await import(`../messages/${locale}.json`)).default;
  const providerAccountWizardMessages = (await import(`../messages/provider-account-wizard/${locale}.json`)).default;
  const budgetLimitRulesMessages = (await import(`../messages/budget-limit-rules/${locale}.json`)).default;

  return {
    locale,
    messages: {
      ...baseMessages,
      providerAccounts: providerAccountWizardMessages,
      budgetLimitRules: budgetLimitRulesMessages,
    },
    timeZone: "Asia/Tehran",
  };
});
