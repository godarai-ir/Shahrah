import { getRequestConfig } from "next-intl/server";

export default getRequestConfig(async () => {
  const locale = "fa";
  const baseMessages = (await import("../messages/fa.json")).default;
  const providerAccountWizardMessages = (await import("../messages/provider-account-wizard/fa.json")).default;

  return {
    locale,
    messages: {
      ...baseMessages,
      providerAccounts: providerAccountWizardMessages,
    },
    timeZone: "Asia/Tehran",
  };
});
