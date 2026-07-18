import { NextRequest, NextResponse } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { GET } from "./route";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());
const proxyToControlPlaneMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({
  readAuthFromRequest: readAuthFromRequestMock,
}));

vi.mock("@/lib/control-plane", () => ({
  proxyToControlPlane: proxyToControlPlaneMock,
}));

describe("tenant-provider-accounts/[accountID]/effective-models GET route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    readAuthFromRequestMock.mockResolvedValue({
      ok: true,
      user: { id: "user-1" },
    });
  });

  it("proxies an effective model lookup for one selected principal", async () => {
    const req = new NextRequest(
      "http://localhost/api/tenant-provider-accounts/acc-123/effective-models?key_type=pat&user_id=user-123",
    );
    proxyToControlPlaneMock.mockResolvedValueOnce(NextResponse.json({ data: [] }));

    const res = await GET(req, { params: Promise.resolve({ accountID: "acc-123" }) });

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request: req,
      method: "GET",
      path: "/api/v1/tenant-provider-accounts/acc-123/effective-models?key_type=pat&user_id=user-123",
    });
    expect(res.status).toBe(200);
  });

  it("returns 400 for an unsupported key type", async () => {
    const req = new NextRequest(
      "http://localhost/api/tenant-provider-accounts/acc-123/effective-models?key_type=session&user_id=user-123",
    );

    const res = await GET(req, { params: Promise.resolve({ accountID: "acc-123" }) });

    expect(res.status).toBe(400);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });

  it("returns 400 when the principal scope is ambiguous", async () => {
    const req = new NextRequest(
      "http://localhost/api/tenant-provider-accounts/acc-123/effective-models?key_type=vat&user_id=user-123&team_id=team-123",
    );

    const res = await GET(req, { params: Promise.resolve({ accountID: "acc-123" }) });

    expect(res.status).toBe(400);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });
});
