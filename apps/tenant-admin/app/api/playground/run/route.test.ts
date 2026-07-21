import { NextRequest, NextResponse } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { POST } from "./route";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());
const getCurrentMePermissionsMock = vi.hoisted(() => vi.fn());
const getNormalizedPermissionSetMock = vi.hoisted(() => vi.fn());
const proxyToControlPlaneMock = vi.hoisted(() => vi.fn());
const proxyToControlPlaneStreamMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({ readAuthFromRequest: readAuthFromRequestMock }));
vi.mock("@/lib/auth/me-permissions", () => ({
  getCurrentMePermissions: getCurrentMePermissionsMock,
  getNormalizedPermissionSet: getNormalizedPermissionSetMock,
}));
vi.mock("@/lib/control-plane", () => ({
  proxyToControlPlane: proxyToControlPlaneMock,
  proxyToControlPlaneStream: proxyToControlPlaneStreamMock,
}));

describe("playground run route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    readAuthFromRequestMock.mockResolvedValue({ ok: true, tenantSlug: "acme", user: { id: "user-1" } });
    getCurrentMePermissionsMock.mockResolvedValue({ permissions: ["policy:manage"] });
    getNormalizedPermissionSetMock.mockReturnValue(new Set(["policy:manage"]));
    proxyToControlPlaneMock.mockResolvedValue(NextResponse.json({ choices: [] }));
    proxyToControlPlaneStreamMock.mockResolvedValue(new NextResponse("data: [DONE]\n\n", { headers: { "Content-Type": "text/event-stream" } }));
  });

  it("blocks billable execution for a read-only principal", async () => {
    getNormalizedPermissionSetMock.mockReturnValue(new Set(["policy:list"]));
    const response = await POST(requestFor({ request: { model: "gpt-test", messages: [{ role: "user", content: "hi" }] } }));
    expect(response.status).toBe(403);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });

  it("rejects trusted or unsupported request fields", async () => {
    const response = await POST(requestFor({ request: { model: "gpt-test", messages: [{ role: "user", content: "hi" }], authorization: "secret" } }));
    expect(response.status).toBe(400);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });

  it("rejects malformed envelope IDs instead of throwing while sanitizing metadata", async () => {
    const response = await POST(requestFor({
      client_run_id: 42,
      request: { model: "gpt-test", messages: [{ role: "user", content: "hi" }] },
    }));

    expect(response.status).toBe(400);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });

  it("rejects attempts to override trusted playground metadata", async () => {
    const response = await POST(requestFor({
      request: {
        model: "gpt-test",
        messages: [{ role: "user", content: "hi" }],
        metadata: { source: "attacker" },
      },
    }));

    expect(response.status).toBe(400);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });

  it("rejects unknown nested message and cache fields", async () => {
    const invalidMessage = await POST(requestFor({
      request: { model: "gpt-test", messages: [{ role: "user", content: "hi", name: "hidden" }] },
    }));
    const invalidCache = await POST(requestFor({
      request: { model: "gpt-test", messages: [{ role: "user", content: "hi" }] },
      godarai: { cache_config: { type: "simple", ttl: 60, authorization: "secret" } },
    }));

    expect(invalidMessage.status).toBe(400);
    expect(invalidCache.status).toBe(400);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });

  it("routes streaming requests through the non-buffering proxy with an operational header allowlist", async () => {
    const request = requestFor({
      client_run_id: "run-1",
      comparison_id: "cmp-1",
      column_id: "column-1",
      request: { model: "gpt-test", messages: [{ role: "user", content: "hi" }], stream: true, metadata: { experiment: "a" } },
    });
    await POST(request);

    expect(proxyToControlPlaneStreamMock).toHaveBeenCalledOnce();
    const options = proxyToControlPlaneStreamMock.mock.calls[0][0];
    expect(options.path).toBe("/v1/chat/completions");
    expect(options.body.metadata).toEqual({ source: "tenant_admin_playground", experiment: "a", client_run_id: "run-1", comparison_id: "cmp-1", column_id: "column-1" });
    expect(options.exposeResponseHeaders).toContain("x-godarai-cost-irr");
  });
});

function requestFor(body: unknown) {
  return new NextRequest("http://acme.localhost/api/playground/run", {
    method: "POST",
    headers: { "Content-Type": "application/json", host: "acme.localhost" },
    body: JSON.stringify(body),
  });
}
