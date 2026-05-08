import { getRequestConfig } from "next-intl/server";

export default getRequestConfig(async () => {
  const locale = "fa";
  const baseMessages =
    locale === "en"
      ? (await import("../messages/en.json")).default
      : (await import("../messages/fa.json")).default;
  const providerAccountWizardMessages =
    locale === "en"
      ? (await import("../messages/provider-account-wizard/en.json")).default
      : (await import("../messages/provider-account-wizard/fa.json")).default;

  return {
    locale,
    messages: {
      ...baseMessages,
      providerAccounts: providerAccountWizardMessages,
    },
    timeZone: "Asia/Tehran",
  };
});
