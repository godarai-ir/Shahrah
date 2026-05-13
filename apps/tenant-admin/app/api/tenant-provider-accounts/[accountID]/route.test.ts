import { NextRequest, NextResponse } from "next/server";
import { describe, expect, it, vi, beforeEach } from "vitest";

import { DELETE } from "./route";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());
const proxyToControlPlaneMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({
  readAuthFromRequest: readAuthFromRequestMock,
}));

vi.mock("@/lib/control-plane", () => ({
  proxyToControlPlane: proxyToControlPlaneMock,
}));

describe("tenant-provider-accounts/[accountID] DELETE route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("proxies delete request when auth is valid", async () => {
    const req = new NextRequest("http://localhost/api/tenant-provider-accounts/abc", { method: "DELETE" });
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      user: { id: "user-1" },
    });
    proxyToControlPlaneMock.mockResolvedValueOnce(new NextResponse(null, { status: 204 }));

    const res = await DELETE(req, { params: Promise.resolve({ accountID: "acc-123" }) });

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request: req,
      method: "DELETE",
      path: "/api/v1/tenant-provider-accounts/acc-123",
      userID: "user-1",
    });
    expect(res.status).toBe(204);
  });

  it("returns 400 for empty account id", async () => {
    const req = new NextRequest("http://localhost/api/tenant-provider-accounts", { method: "DELETE" });
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      user: { id: "user-1" },
    });

    const res = await DELETE(req, { params: Promise.resolve({ accountID: "   " }) });
    const body = await res.json();

    expect(res.status).toBe(400);
    expect(body?.error?.message).toBeTruthy();
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });
});

