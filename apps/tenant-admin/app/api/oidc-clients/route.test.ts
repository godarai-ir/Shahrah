import { NextRequest, NextResponse } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { GET, POST } from "./route";
import { PATCH } from "./[id]/route";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());
const proxyToControlPlaneMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({ readAuthFromRequest: readAuthFromRequestMock }));
vi.mock("@/lib/control-plane", () => ({ proxyToControlPlane: proxyToControlPlaneMock }));

function jsonRequest(method: string, url: string, body: unknown) {
  return new NextRequest(url, { method, body: JSON.stringify(body), headers: { "content-type": "application/json" } });
}

describe("oidc-clients route handlers", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    readAuthFromRequestMock.mockResolvedValue({ ok: true, user: { id: "user-1" } });
    proxyToControlPlaneMock.mockResolvedValue(NextResponse.json({ ok: true }));
  });

  it("lists clients as the signed-in member", async () => {
    await GET(new NextRequest("http://acme.localhost/api/oidc-clients"));

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith(
      expect.objectContaining({ method: "GET", path: "/api/v1/oidc-clients", userID: "user-1" }),
    );
  });

  it("forwards only the normalized registration fields", async () => {
    await POST(
      jsonRequest("POST", "http://acme.localhost/api/oidc-clients", {
        name: " LibreChat ",
        redirect_uris: ["https://chat.example.com/cb"],
        active: false,
        client_secret: "attacker-chosen",
      }),
    );

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith(
      expect.objectContaining({
        method: "POST",
        userID: "user-1",
        body: { name: "LibreChat", redirect_uris: ["https://chat.example.com/cb"], post_logout_redirect_uris: [] },
      }),
    );
  });

  it("rejects an invalid registration without calling the control plane", async () => {
    const response = await POST(
      jsonRequest("POST", "http://acme.localhost/api/oidc-clients", { name: "App", redirect_uris: ["http://evil.example/cb"] }),
    );

    expect(response.status).toBe(400);
    expect(await response.json()).toMatchObject({ error: { code: "oidc_client_invalid" } });
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });

  it("encodes the client identifier in update paths", async () => {
    await PATCH(jsonRequest("PATCH", "http://acme.localhost/api/oidc-clients/a%2Fb", { active: true }), {
      params: Promise.resolve({ id: "a/b" }),
    });

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith(
      expect.objectContaining({ method: "PATCH", path: "/api/v1/oidc-clients/a%2Fb", body: { active: true } }),
    );
  });

  it("returns the auth failure for signed-out requests", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({ ok: false, status: 401, message: "Unauthorized" });

    const response = await GET(new NextRequest("http://acme.localhost/api/oidc-clients"));

    expect(response.status).toBe(401);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });
});
