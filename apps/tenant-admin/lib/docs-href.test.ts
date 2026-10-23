import { describe, expect, it } from "vitest";

import { buildApexHref, buildDocsBaseHref } from "./docs-href";

describe("buildDocsBaseHref", () => {
  it("strips the tenant label from the forwarded host", () => {
    const headers = new Headers({ "x-forwarded-proto": "https, http", "x-forwarded-host": "acme.godarai.ir, proxy", host: "internal:3000" });

    expect(buildDocsBaseHref(headers, "acme")).toBe("https://godarai.ir/docs");
  });

  it("uses http for localhost and keeps hosts without the tenant label", () => {
    expect(buildDocsBaseHref(new Headers({ host: "acme.localhost:3020" }), "acme")).toBe("http://localhost:3020/docs");
    expect(buildDocsBaseHref(new Headers({ host: "docs.example.com" }), "acme")).toBe("https://docs.example.com/docs");
  });

  it("falls back to a relative path without a host", () => {
    expect(buildDocsBaseHref(new Headers(), "acme")).toBe("/docs");
  });
});

describe("buildApexHref", () => {
  it("links the public pricing changelog on the apex host", () => {
    const headers = new Headers({ "x-forwarded-host": "acme.godarai.ir" });

    expect(buildApexHref(headers, "acme", "/pricing/changelog")).toBe("https://godarai.ir/pricing/changelog");
    expect(buildApexHref(new Headers(), "acme", "/pricing/changelog")).toBe("/pricing/changelog");
  });
});
