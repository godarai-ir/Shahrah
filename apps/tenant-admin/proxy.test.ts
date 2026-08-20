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

  it("allows the same-origin bodyless logout request", () => {
    const request = new NextRequest("https://acme.shahrah.ir/api/auth/logout", {
      method: "POST",
      headers: {
        host: "acme.shahrah.ir",
        origin: "https://acme.shahrah.ir",
        "sec-fetch-site": "same-origin",
      },
    });

    const response = proxy(request);

    expect(response.status).toBe(200);
    expect(
      response.headers.get("x-middleware-request-x-godarai-tenant-resolved"),
    ).toBe("acme");
  });

  it("allows a same-origin bodyless delete request", () => {
    const request = new NextRequest(
      "https://acme.shahrah.ir/api/tenant-provider-accounts/account-1",
      {
        method: "DELETE",
        headers: {
          host: "acme.shahrah.ir",
          origin: "https://acme.shahrah.ir",
          "sec-fetch-site": "same-origin",
        },
      },
    );

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

  it.each([
    "/api/support/tickets",
    "/api/support/tickets/ticket-123/messages",
  ])("allows same-origin multipart support requests on %s", (pathname) => {
    const formData = new FormData();
    formData.set("payload", JSON.stringify({ body: "Please help" }));
    const request = new NextRequest(`https://acme.shahrah.ir${pathname}`, {
      method: "POST",
      headers: {
        host: "acme.shahrah.ir",
        origin: "https://acme.shahrah.ir",
        "sec-fetch-site": "same-origin",
      },
      body: formData,
    });

    const response = proxy(request);

    expect(response.status).toBe(200);
  });

  it("allows same-origin multipart document uploads for managed RAG", () => {
    const formData = new FormData();
    formData.set(
      "file",
      new File(["docx-content"], "راهنما.docx", {
        type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
      }),
    );
    const request = new NextRequest(
      "https://acme.shahrah.ir/api/managed-rag/knowledge-bases/kb-123/documents",
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

  it("rejects multipart requests on managed RAG routes that do not accept files", () => {
    const formData = new FormData();
    formData.set("file", new File(["content"], "file.docx"));
    const request = new NextRequest(
      "https://acme.shahrah.ir/api/managed-rag/knowledge-bases/kb-123",
      {
        method: "PATCH",
        headers: {
          host: "acme.shahrah.ir",
          origin: "https://acme.shahrah.ir",
          "sec-fetch-site": "same-origin",
        },
        body: formData,
      },
    );

    const response = proxy(request);

    expect(response.status).toBe(415);
  });

  it("rejects multipart requests on support routes that do not accept uploads", () => {
    const formData = new FormData();
    formData.set("payload", "{}");
    const request = new NextRequest(
      "https://acme.shahrah.ir/api/support/tickets/ticket-123/rating",
      {
        method: "PUT",
        headers: {
          host: "acme.shahrah.ir",
          origin: "https://acme.shahrah.ir",
          "sec-fetch-site": "same-origin",
        },
        body: formData,
      },
    );

    const response = proxy(request);

    expect(response.status).toBe(415);
  });
});
