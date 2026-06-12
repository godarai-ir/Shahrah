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

describe("provider-catalog/providers/[provider]/models GET route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("proxies provider model reads to control-plane", async () => {
    const req = new NextRequest("http://localhost/api/provider-catalog/providers/openai/models");
    readAuthFromRequestMock.mockResolvedValueOnce({ ok: true, user: { id: "user-1" } });
    proxyToControlPlaneMock.mockResolvedValueOnce(
      NextResponse.json({ data: [{ model: "gpt-4.1-mini" }] }, { status: 200 }),
    );

    const res = await GET(req, { params: Promise.resolve({ provider: "OpenAI" }) });

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request: req,
      method: "GET",
      path: "/api/v1/provider-catalog/providers/openai/models",
    });
    expect(res.status).toBe(200);
  });
});
