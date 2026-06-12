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

describe("provider-catalog/display-settings GET route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("proxies the request to control-plane when auth is valid", async () => {
    const req = new NextRequest("http://localhost/api/provider-catalog/display-settings");
    readAuthFromRequestMock.mockResolvedValueOnce({ ok: true, user: { id: "user-1" } });
    proxyToControlPlaneMock.mockResolvedValueOnce(
      NextResponse.json({ data: { currencyUnit: "USD" } }, { status: 200 }),
    );

    const res = await GET(req);

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request: req,
      method: "GET",
      path: "/api/v1/provider-catalog/display-settings",
    });
    expect(res.status).toBe(200);
  });
});
