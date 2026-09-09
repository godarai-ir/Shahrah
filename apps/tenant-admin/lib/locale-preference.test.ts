import { describe, expect, it } from "vitest";

import { LOCALE_COOKIE_MAX_AGE_SECONDS, LOCALE_COOKIE_NAME } from "./locale";
import { serializeLocaleCookie } from "./locale-preference";

describe("serializeLocaleCookie", () => {
  it("writes a year-long, root-scoped cookie", () => {
    const cookie = serializeLocaleCookie("en", false);

    expect(cookie).toContain(`${LOCALE_COOKIE_NAME}=en`);
    expect(cookie).toContain("Path=/");
    expect(cookie).toContain(`Max-Age=${LOCALE_COOKIE_MAX_AGE_SECONDS}`);
    expect(cookie).toContain("SameSite=Lax");
  });

  it("marks the cookie Secure only over HTTPS", () => {
    expect(serializeLocaleCookie("fa", true)).toContain("Secure");
    expect(serializeLocaleCookie("fa", false)).not.toContain("Secure");
  });
});
