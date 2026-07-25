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

describe("models POST route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    readAuthFromRequestMock.mockResolvedValue({
      ok: true,
      user: { id: "user-1" },
    });
    proxyToControlPlaneMock.mockResolvedValue(
      NextResponse.json({ id: "deployment-loadtest" }, { status: 201 }),
    );
  });

  it("allows the internal mock provider without upstream credentials", async () => {
    const request = new NextRequest("http://localhost/api/models", {
      method: "POST",
      body: JSON.stringify({
        model_name: "loadtest/mock",
        provider: "mock",
        model: "mock/chat",
        rpm: 100000,
        tpm: 10000000,
        weight: 1,
        timeout: 60,
      }),
    });

    const response = await POST(request);

    expect(response.status).toBe(201);
    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request,
      method: "POST",
      path: "/api/v1/models",
      body: {
        model_name: "loadtest/mock",
        provider: "mock",
        model: "mock/chat",
        api_base: "",
        api_key: "",
        api_version: "",
        region_name: "",
        rpm: 100000,
        tpm: 10000000,
        weight: 1,
        timeout: 60,
      },
    });
  });

  it("still requires credentials for external providers", async () => {
    const request = new NextRequest("http://localhost/api/models", {
      method: "POST",
      body: JSON.stringify({
        model_name: "openai/gpt-4o-mini",
        provider: "openai",
        model: "gpt-4o-mini",
      }),
    });

    const response = await POST(request);

    expect(response.status).toBe(400);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });
});
