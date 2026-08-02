import test from "node:test";
import assert from "node:assert/strict";

import { ApiAuthRepository } from "./api-auth-repository.ts";

const BASE_URL = "http://control-plane.local";

test("maps session payload from control-plane snake_case to app camelCase", async () => {
  const restoreFetch = mockFetch(async (input, init) => {
    const url = String(input);
    if (url.endsWith("/api/v1/auth/session") && init?.method === "POST") {
      return jsonResponse({
        id: "sess-1",
        tenant_slug: "tenant1",
        user_id: "user-1",
        created_at_iso: "2026-05-09T10:00:00Z",
        expires_at_iso: "2026-05-16T10:00:00Z",
      });
    }

    if (url.endsWith("/api/v1/auth/session/sess-1") && init?.method === "GET") {
      return jsonResponse({
        id: "sess-1",
        tenant_slug: "tenant1",
        user_id: "user-1",
        created_at_iso: "2026-05-09T10:00:00Z",
        expires_at_iso: "2026-05-16T10:00:00Z",
      });
    }

    return new Response("not found", { status: 404 });
  });

  try {
    const repo = new ApiAuthRepository(BASE_URL, "master-key");
    const created = await repo.createSession({
      tenantSlug: "tenant1",
      userID: "user-1",
    });
    const loaded = await repo.getSession("sess-1");

    assert.equal(created.userID, "user-1");
    assert.equal(created.tenantSlug, "tenant1");
    assert.equal(created.expiresAtISO, "2026-05-16T10:00:00Z");
    assert.equal(loaded?.userID, "user-1");
    assert.equal(loaded?.tenantSlug, "tenant1");
    assert.equal(loaded?.expiresAtISO, "2026-05-16T10:00:00Z");
  } finally {
    restoreFetch();
  }
});

test("maps OTP verification record from control-plane snake_case keys", async () => {
  const restoreFetch = mockFetch(async (input, init) => {
    const url = String(input);
    if (url.endsWith("/api/v1/auth/otp/verify") && init?.method === "POST") {
      return jsonResponse({
        ok: true,
        record: {
          id: "req-1",
          tenant_slug: "tenant1",
          purpose: "login",
          channel: "email",
          identifier: "admin@tenant1.com",
          user_id: "user-1",
          pending_value: null,
          otp_code: "123456",
          attempts: 0,
          expires_at_iso: "2026-05-09T10:05:00Z",
          created_at_iso: "2026-05-09T10:00:00Z",
        },
      });
    }
    return new Response("not found", { status: 404 });
  });

  try {
    const repo = new ApiAuthRepository(BASE_URL, "master-key");
    const result = await repo.verifyOTPRequest({
      tenantSlug: "tenant1",
      requestID: "req-1",
      otpCode: "123456",
      purpose: "login",
    });

    assert.equal(result.ok, true);
    if (!result.ok) {
      throw new Error("verification unexpectedly failed");
    }
    assert.equal(result.record.tenantSlug, "tenant1");
    assert.equal(result.record.userID, "user-1");
    assert.equal(result.record.otpCode, "123456");
    assert.equal(result.record.expiresAtISO, "2026-05-09T10:05:00Z");
  } finally {
    restoreFetch();
  }
});

test("maps linked social providers onto the tenant user", async () => {
  const restoreFetch = mockFetch(async (input, init) => {
    const url = String(input);
    if (url.includes("/api/v1/auth/user/user-1?") && init?.method === "GET") {
      return jsonResponse({
        id: "user-1",
        tenant_slug: "tenant1",
        email: "user@example.com",
        role: "admin",
        active: true,
        social_providers: ["google"],
        created_at_iso: "2026-05-09T10:00:00Z",
        updated_at_iso: "2026-05-09T10:00:00Z",
      });
    }

    return new Response("not found", { status: 404 });
  });

  try {
    const repo = new ApiAuthRepository(BASE_URL, "master-key");
    const user = await repo.getUser("tenant1", "user-1");

    assert.deepEqual(user?.socialProviders, ["google"]);
  } finally {
    restoreFetch();
  }
});

function mockFetch(
  handler: (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>,
) {
  const original = globalThis.fetch;
  globalThis.fetch = handler as typeof fetch;
  return () => {
    globalThis.fetch = original;
  };
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
