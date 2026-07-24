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

describe("provider-connections/test POST route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    readAuthFromRequestMock.mockResolvedValue({ ok: true, user: { id: "user-1" } });
    proxyToControlPlaneMock.mockResolvedValue(NextResponse.json({ success: true }));
  });

  it("proxies a normalized write-only preflight payload", async () => {
    const request = new NextRequest("http://localhost/api/provider-connections/test", {
      method: "POST",
      body: JSON.stringify({
        provider: " OpenAI ",
        auth_mode: " API_KEY ",
        provider_config: { organization: "org-1" },
        credentials: { api_key: "super-secret" },
        model_hint: " gpt-4.1 ",
      }),
    });

    await POST(request);

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request,
      method: "POST",
      path: "/api/v1/provider-connections/test",
      body: {
        provider: "openai",
        auth_mode: "api_key",
        provider_config: { organization: "org-1" },
        credentials: { api_key: "super-secret" },
        use_platform_managed_credentials: false,
        model_hint: "gpt-4.1",
      },
    });
  });

  it("rejects a malformed request before proxying credentials", async () => {
    const request = new NextRequest("http://localhost/api/provider-connections/test", {
      method: "POST",
      body: JSON.stringify({ provider: "openai", credentials: { api_key: "super-secret" } }),
    });

    const response = await POST(request);
    const payload = await response.json();

    expect(response.status).toBe(400);
    expect(JSON.stringify(payload)).not.toContain("super-secret");
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });
});
