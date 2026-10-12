import { NextRequest, NextResponse } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { GET, POST } from "./route";
import { POST as REVOKE } from "./[tokenID]/revoke/route";
import { POST as ROTATE } from "./[tokenID]/rotate/route";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());
const proxyToControlPlaneMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({
  readAuthFromRequest: readAuthFromRequestMock,
}));

vi.mock("@/lib/control-plane", () => ({
  proxyToControlPlane: proxyToControlPlaneMock,
}));

describe("virtual user token routes", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    readAuthFromRequestMock.mockResolvedValue({
      ok: true,
      user: { id: "user-1" },
    });
    proxyToControlPlaneMock.mockResolvedValue(NextResponse.json({ ok: true }));
  });

  it("proxies VAT list with pagination", async () => {
    const req = new NextRequest("http://localhost/api/virtual-users/vu-1/tokens?limit=10&offset=20&query=prod");

    await GET(req, { params: Promise.resolve({ virtualUserID: "vu-1" }) });

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request: req,
      method: "GET",
      path: "/api/v1/virtual-users/vu-1/tokens?query=prod&limit=10&offset=20",
      userID: "user-1",
    });
  });

  it("proxies VAT creation", async () => {
    const req = new NextRequest("http://localhost/api/virtual-users/vu-1/tokens", {
      method: "POST",
      body: JSON.stringify({ key_alias: "prod", allowed_models: ["openai:default:gpt-4o-mini"] }),
    });

    await POST(req, { params: Promise.resolve({ virtualUserID: "vu-1" }) });

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request: req,
      method: "POST",
      path: "/api/v1/virtual-users/vu-1/tokens",
      userID: "user-1",
      body: { key_alias: "prod", allowed_models: ["openai:default:gpt-4o-mini"] },
    });
  });

  it("proxies VAT revoke", async () => {
    const req = new NextRequest("http://localhost/api/virtual-users/vu-1/tokens/key-1/revoke", {
      method: "POST",
      body: JSON.stringify({ reason: "test" }),
    });

    await REVOKE(req, { params: Promise.resolve({ virtualUserID: "vu-1", tokenID: "key-1" }) });

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request: req,
      method: "POST",
      path: "/api/v1/virtual-users/vu-1/tokens/key-1/revoke",
      userID: "user-1",
      body: { reason: "test" },
    });
  });

  it("proxies VAT rotation", async () => {
    const req = new NextRequest("http://localhost/api/virtual-users/vu-1/tokens/key-1/rotate", {
      method: "POST",
      body: JSON.stringify({ grace_period: "10m" }),
    });

    await ROTATE(req, { params: Promise.resolve({ virtualUserID: "vu-1", tokenID: "key-1" }) });

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request: req,
      method: "POST",
      path: "/api/v1/virtual-users/vu-1/tokens/key-1/rotate",
      userID: "user-1",
      body: { grace_period: "10m" },
    });
  });
});
