import { NextRequest, NextResponse } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { POST } from "./route";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());
const proxyToControlPlaneMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({
  readAuthFromRequest: readAuthFromRequestMock,
}));

vi.mock("@/lib/control-plane", () => ({
  proxyToControlPlane: proxyToControlPlaneMock,
}));

describe("tenant-provider-accounts/[accountID]/test-credentials POST route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    readAuthFromRequestMock.mockResolvedValue({
      ok: true,
      user: { id: "user-1" },
    });
  });

  it("proxies a credential test with a trimmed model hint", async () => {
    const req = new NextRequest("http://localhost/api/tenant-provider-accounts/acc-123/test-credentials", {
      method: "POST",
      body: JSON.stringify({ model_hint: "  gpt-4o-mini  " }),
    });
    proxyToControlPlaneMock.mockResolvedValueOnce(NextResponse.json({ success: true }));

    const res = await POST(req, { params: Promise.resolve({ accountID: "acc-123" }) });

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request: req,
      method: "POST",
      path: "/api/v1/tenant-provider-accounts/acc-123/test-credentials",
      body: { model_hint: "gpt-4o-mini" },
    });
    expect(res.status).toBe(200);
  });

  it("omits an empty model hint from the proxied body", async () => {
    const req = new NextRequest("http://localhost/api/tenant-provider-accounts/acc-123/test-credentials", {
      method: "POST",
      body: JSON.stringify({ model_hint: "   " }),
    });
    proxyToControlPlaneMock.mockResolvedValueOnce(NextResponse.json({ success: true }));

    await POST(req, { params: Promise.resolve({ accountID: "acc-123" }) });

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request: req,
      method: "POST",
      path: "/api/v1/tenant-provider-accounts/acc-123/test-credentials",
      body: {},
    });
  });
});
