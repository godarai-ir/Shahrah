import { getRequestConfig } from "next-intl/server";

export default getRequestConfig(async () => {
  const locale = "fa";

  return {
    locale,
    messages: (await import("../messages/fa.json")).default,
    timeZone: "Asia/Tehran",
  };
});
