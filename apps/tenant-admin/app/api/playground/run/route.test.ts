import { NextRequest, NextResponse } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { POST } from "./route";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());
const getCurrentMePermissionsMock = vi.hoisted(() => vi.fn());
const getNormalizedPermissionSetMock = vi.hoisted(() => vi.fn());
const proxyToControlPlaneMock = vi.hoisted(() => vi.fn());
const proxyToControlPlaneStreamMock = vi.hoisted(() => vi.fn());
const fetchControlPlaneJSONMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({ readAuthFromRequest: readAuthFromRequestMock }));
vi.mock("@/lib/auth/me-permissions", () => ({
  getCurrentMePermissions: getCurrentMePermissionsMock,
  getNormalizedPermissionSet: getNormalizedPermissionSetMock,
}));
vi.mock("@/lib/control-plane", () => ({
  fetchControlPlaneJSON: fetchControlPlaneJSONMock,
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
    fetchControlPlaneJSONMock.mockResolvedValue({
      ok: true,
      status: 200,
      payload: { data: [
        { id: "gpt-image-1", supported_modes: ["image"] },
        { id: "dall-e-3", supported_modes: ["image_generation"] },
      ] },
      responseHeaders: {},
    });
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
      request: { model: "gpt-test", messages: [{ role: "user", content: "hi" }], stream: true, stream_options: { include_usage: true }, metadata: { experiment: "a" } },
    });
    await POST(request);

    expect(proxyToControlPlaneStreamMock).toHaveBeenCalledOnce();
    const options = proxyToControlPlaneStreamMock.mock.calls[0][0];
    expect(options.path).toBe("/v1/chat/completions");
    expect(options.body.metadata).toEqual({ source: "tenant_admin_playground", experiment: "a", client_run_id: "run-1", comparison_id: "cmp-1", column_id: "column-1" });
    expect(options.body.stream_options).toEqual({ include_usage: true });
    expect(options.upstreamHeaders.get("x-godarai-include-stream-telemetry")).toBe("true");
    expect(options.exposeResponseHeaders).toContain("x-godarai-cost-irr");
  });

  it("routes image generation through the dedicated non-streaming endpoint", async () => {
    await POST(requestFor({
      mode: "image_generation",
      client_run_id: "run-image-1",
      request: {
        model: "gpt-image-1",
        prompt: "A quiet mountain lake",
        n: 1,
        output_format: "png",
      },
    }));

    expect(proxyToControlPlaneStreamMock).not.toHaveBeenCalled();
    expect(proxyToControlPlaneMock).toHaveBeenCalledOnce();
    const options = proxyToControlPlaneMock.mock.calls[0][0];
    expect(options.path).toBe("/v1/images/generations");
    expect(options.body).toEqual({
      model: "gpt-image-1",
      prompt: "A quiet mountain lake",
      n: 1,
      output_format: "png",
    });
  });

  it("forwards an explicitly selected response format for compatible image models", async () => {
    await POST(requestFor({
      mode: "image_generation",
      request: {
        model: "dall-e-3",
        prompt: "A quiet mountain lake",
        n: 1,
        response_format: "b64_json",
      },
    }));

    expect(proxyToControlPlaneMock).toHaveBeenCalledOnce();
    expect(proxyToControlPlaneMock.mock.calls[0][0].body).toMatchObject({
      model: "dall-e-3",
      response_format: "b64_json",
    });
  });

  it("rejects chat-only options in image generation mode", async () => {
    const response = await POST(requestFor({
      mode: "image_generation",
      request: { model: "gpt-image-1", prompt: "A lake", messages: [] },
    }));

    expect(response.status).toBe(400);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });

  it("rejects a tampered image request when the selected model lacks the capability", async () => {
    fetchControlPlaneJSONMock.mockResolvedValue({
      ok: true,
      status: 200,
      payload: { data: [{ id: "chat-only", supported_modes: ["chat"] }] },
      responseHeaders: {},
    });

    const response = await POST(requestFor({
      mode: "image_generation",
      request: { model: "chat-only", prompt: "A lake", n: 1 },
    }));

    expect(response.status).toBe(400);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });
});

function requestFor(body: unknown) {
  return new NextRequest("http://acme.localhost/api/playground/run", {
    method: "POST",
    headers: { "Content-Type": "application/json", host: "acme.localhost" },
    body: JSON.stringify(body),
  });
}
