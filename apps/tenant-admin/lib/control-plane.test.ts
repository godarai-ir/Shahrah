import { beforeEach, describe, expect, it, vi } from "vitest";
import { NextRequest } from "next/server";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({
  readAuthFromRequest: readAuthFromRequestMock,
}));

import {
  proxyPublicInvitationAcceptance,
  proxyToControlPlane,
  proxyToControlPlaneBinary,
} from "@/lib/control-plane";

function tenantRequest(path: string) {
  return new NextRequest(`http://acme.localhost${path}`, {
    headers: {
      host: "acme.localhost",
    },
  });
}

describe("authenticated control-plane proxy", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.stubEnv("SHAHRAH_MASTER_KEY", "test-master-key");
    readAuthFromRequestMock.mockReset();
  });

  it("rejects an unauthenticated JSON proxy request before contacting upstream", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: false,
      status: 401,
      message: "authentication required",
    });
    const fetchMock = vi.spyOn(globalThis, "fetch");

    const response = await proxyToControlPlane({
      request: tenantRequest("/api/logs"),
      method: "GET",
      path: "/api/v2/logs",
    });

    expect(response.status).toBe(401);
    await expect(response.json()).resolves.toEqual({
      error: { message: "authentication required" },
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects an unauthenticated binary proxy request before contacting upstream", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: false,
      status: 403,
      message: "invalid session",
    });
    const fetchMock = vi.spyOn(globalThis, "fetch");

    const response = await proxyToControlPlaneBinary({
      request: tenantRequest("/api/logs/exports/export-1/download"),
      method: "GET",
      path: "/api/v2/logs/exports/export-1/download",
    });

    expect(response.status).toBe(403);
    await expect(response.json()).resolves.toEqual({
      error: { message: "invalid session" },
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("derives upstream user identity from the authenticated session", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "acme",
      user: { id: "user-1" },
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      Response.json({ data: [] }),
    );

    const response = await proxyToControlPlane({
      request: tenantRequest("/api/logs"),
      method: "GET",
      path: "/api/v2/logs",
    });

    expect(response.status).toBe(200);
    const [, init] = fetchMock.mock.calls[0];
    expect(new Headers(init?.headers).get("x-godarai-user-id")).toBe("user-1");
  });

  it("rejects a caller-supplied user identity that differs from the session", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "acme",
      user: { id: "user-1" },
    });
    const fetchMock = vi.spyOn(globalThis, "fetch");

    const response = await proxyToControlPlane({
      request: tenantRequest("/api/logs"),
      method: "GET",
      path: "/api/v2/logs",
      userID: "user-2",
    });

    expect(response.status).toBe(403);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects a tenant that differs from the authenticated session", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "other-tenant",
      user: { id: "user-1" },
    });
    const fetchMock = vi.spyOn(globalThis, "fetch");

    const response = await proxyToControlPlane({
      request: tenantRequest("/api/logs"),
      method: "GET",
      path: "/api/v2/logs",
    });

    expect(response.status).toBe(403);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("does not allow optional upstream headers to replace trusted identity headers", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "acme",
      user: { id: "user-1" },
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      Response.json({ data: [] }),
    );

    await proxyToControlPlane({
      request: tenantRequest("/api/logs"),
      method: "GET",
      path: "/api/v2/logs",
      upstreamHeaders: {
        Authorization: "Bearer attacker-key",
        "x-godarai-tenant": "other-tenant",
        "x-godarai-user-id": "user-2",
        "x-shahrah-user-id": "legacy-user-2",
      },
    });

    const [, init] = fetchMock.mock.calls[0];
    const headers = new Headers(init?.headers);
    expect(headers.get("authorization")).toBe("Bearer test-master-key");
    expect(headers.get("x-godarai-tenant")).toBe("acme");
    expect(headers.get("x-godarai-user-id")).toBe("user-1");
    expect(headers.has("x-shahrah-user-id")).toBe(false);
  });

  it("preserves public invitation acceptance through its fixed allowlisted upstream path", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      Response.json({ invitation: { id: "invite-1" } }),
    );

    const response = await proxyPublicInvitationAcceptance(
      tenantRequest("/api/users/invitations/accept"),
      "invite-token",
    );

    expect(response.status).toBe(200);
    expect(readAuthFromRequestMock).not.toHaveBeenCalled();
    expect(fetchMock).toHaveBeenCalledOnce();
    const [url, init] = fetchMock.mock.calls[0];
    expect(String(url)).toBe("http://127.0.0.1:4000/api/v1/users/invitations/accept");
    expect(init?.method).toBe("POST");
    expect(init?.body).toBe(JSON.stringify({ token: "invite-token" }));
    expect(new Headers(init?.headers).has("x-godarai-user-id")).toBe(false);
  });
});
