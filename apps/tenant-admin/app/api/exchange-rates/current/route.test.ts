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

describe("exchange-rates/current GET route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("proxies an authenticated request to the current rate service", async () => {
    const request = new NextRequest("http://localhost/api/exchange-rates/current");
    readAuthFromRequestMock.mockResolvedValueOnce({ ok: true, user: { id: "user-1" } });
    proxyToControlPlaneMock.mockResolvedValueOnce(
      NextResponse.json({ base: "USD", quote: "IRR", rate: 2_000_000 }, { status: 200 }),
    );

    const response = await GET(request);

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request,
      method: "GET",
      path: "/api/v1/exchange-rates/current",
    });
    expect(response.status).toBe(200);
  });
});
