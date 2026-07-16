import { NextRequest } from "next/server";
import { describe, expect, it } from "vitest";

import { proxy } from "./proxy";

describe("tenant-admin proxy tenant resolution", () => {
  it.each([
    ["x-godarai-tenant-resolved", "victim"],
    ["x-godarai-tenant", "victim"],
    ["x-forwarded-host", "victim.shahrah.ir"],
  ])("ignores client-controlled %s", (headerName, headerValue) => {
    const request = new NextRequest("https://acme.shahrah.ir/api/keys", {
      headers: {
        host: "acme.shahrah.ir",
        [headerName]: headerValue,
      },
    });

    const response = proxy(request);

    expect(
      response.headers.get("x-middleware-request-x-godarai-tenant-resolved"),
    ).toBe("acme");
    expect(
      response.headers.get("x-middleware-request-x-shahrah-tenant-source"),
    ).toBe("host");
    expect(
      response.headers.get("x-middleware-request-x-shahrah-tenant-host"),
    ).toBe("acme");
    expect(
      response.headers.get("x-middleware-request-x-godarai-tenant"),
    ).toBeNull();
    expect(
      response.headers.get("x-middleware-request-x-forwarded-host"),
    ).toBe("acme.shahrah.ir");
  });

  it("preserves legitimate hostname-based tenant resolution", () => {
    const request = new NextRequest("http://internal-server:3001/api/keys", {
      headers: { host: "mytenant.localhost:3001" },
    });

    const response = proxy(request);

    expect(
      response.headers.get("x-middleware-request-x-godarai-tenant-resolved"),
    ).toBe("mytenant");
    expect(response.status).toBe(200);
  });
});
