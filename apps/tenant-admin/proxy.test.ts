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

describe("tenant-admin proxy CSRF protection", () => {
  it("rejects a state-changing request from a sibling subdomain", () => {
    const request = new NextRequest("https://acme.shahrah.ir/api/keys", {
      method: "POST",
      headers: {
        host: "acme.shahrah.ir",
        origin: "https://attacker.shahrah.ir",
        "sec-fetch-site": "same-site",
        "content-type": "application/json",
      },
      body: JSON.stringify({ name: "stolen-session-key" }),
    });

    const response = proxy(request);

    expect(response.status).toBe(403);
  });

  it("rejects a simple text/plain JSON request even when the origin matches", () => {
    const request = new NextRequest("https://acme.shahrah.ir/api/keys", {
      method: "POST",
      headers: {
        host: "acme.shahrah.ir",
        origin: "https://acme.shahrah.ir",
        "sec-fetch-site": "same-origin",
        "content-type": "text/plain",
      },
      body: JSON.stringify({ name: "csrf" }),
    });

    const response = proxy(request);

    expect(response.status).toBe(415);
  });

  it("allows same-origin JSON requests", () => {
    const request = new NextRequest("https://acme.shahrah.ir/api/keys", {
      method: "POST",
      headers: {
        host: "acme.shahrah.ir",
        origin: "https://acme.shahrah.ir",
        "sec-fetch-site": "same-origin",
        "content-type": "application/json",
      },
      body: JSON.stringify({ name: "legitimate-key" }),
    });

    const response = proxy(request);

    expect(response.status).toBe(200);
    expect(
      response.headers.get("x-middleware-request-x-godarai-tenant-resolved"),
    ).toBe("acme");
  });

  it("rejects a state-changing request without origin or Fetch Metadata", () => {
    const request = new NextRequest("https://acme.shahrah.ir/api/keys/123", {
      method: "DELETE",
      headers: { host: "acme.shahrah.ir" },
    });

    const response = proxy(request);

    expect(response.status).toBe(403);
  });

  it("allows same-origin multipart receipt uploads on the dedicated route", () => {
    const formData = new FormData();
    formData.set("amount_irr", "100000");
    const request = new NextRequest(
      "https://acme.shahrah.ir/api/wallet/bank-transfers",
      {
        method: "POST",
        headers: {
          host: "acme.shahrah.ir",
          origin: "https://acme.shahrah.ir",
          "sec-fetch-site": "same-origin",
        },
        body: formData,
      },
    );

    const response = proxy(request);

    expect(response.status).toBe(200);
  });
});
