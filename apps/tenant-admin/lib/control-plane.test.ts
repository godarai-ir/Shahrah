import { beforeEach, describe, expect, it, vi } from "vitest";
import { NextRequest } from "next/server";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({
  readAuthFromRequest: readAuthFromRequestMock,
}));

import {
  createSSEHeartbeatStream,
  proxyPublicInvitationAcceptance,
  proxyToControlPlane,
  proxyToControlPlaneBinary,
  proxyToControlPlaneBodyStream,
  proxyToControlPlaneStream,
  readUpstreamErrorCode,
} from "@/lib/control-plane";

function tenantRequest(path: string) {
  return new NextRequest(`http://acme.localhost${path}`, {
    headers: {
      host: "acme.localhost",
    },
  });
}

describe("authenticated control-plane proxy", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    vi.stubEnv("SHAHRAH_MASTER_KEY", "test-master-key");
    readAuthFromRequestMock.mockReset();
  });

  it("rejects an unauthenticated JSON proxy request before contacting upstream", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: false,
      status: 401,
      message: "authentication required",
    });
    const fetchMock = vi.spyOn(globalThis, "fetch");

    const response = await proxyToControlPlane({
      request: tenantRequest("/api/logs"),
      method: "GET",
      path: "/api/v2/logs",
    });

    expect(response.status).toBe(401);
    await expect(response.json()).resolves.toEqual({
      error: { message: "authentication required" },
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects an unauthenticated binary proxy request before contacting upstream", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: false,
      status: 403,
      message: "invalid session",
    });
    const fetchMock = vi.spyOn(globalThis, "fetch");

    const response = await proxyToControlPlaneBinary({
      request: tenantRequest("/api/logs/exports/export-1/download"),
      method: "GET",
      path: "/api/v2/logs/exports/export-1/download",
    });

    expect(response.status).toBe(403);
    await expect(response.json()).resolves.toEqual({
      error: { message: "invalid session" },
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("derives upstream user identity from the authenticated session", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "acme",
      user: { id: "user-1" },
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      Response.json({ data: [] }),
    );

    const response = await proxyToControlPlane({
      request: tenantRequest("/api/logs"),
      method: "GET",
      path: "/api/v2/logs",
    });

    expect(response.status).toBe(200);
    const [, init] = fetchMock.mock.calls[0];
    expect(new Headers(init?.headers).get("x-godarai-user-id")).toBe("user-1");
  });

  it("buffers multipart uploads and forwards an exact content length", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "acme",
      user: { id: "user-1" },
    });
    const form = new FormData();
    form.set("payload", JSON.stringify({ subject: "A support ticket" }));
    form.append("files", new File(["diagnostic-data"], "diagnostic.txt", { type: "text/plain" }));
    const request = new NextRequest("http://acme.localhost/api/support/tickets", {
      method: "POST",
      headers: { host: "acme.localhost" },
      body: form,
    });
    const expectedBody = await request.clone().arrayBuffer();
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      Response.json({ ok: true }, { status: 201 }),
    );

    const response = await proxyToControlPlaneBodyStream({
      request,
      method: "POST",
      path: "/api/v1/support/tickets",
    });

    expect(response.status).toBe(201);
    const [, init] = fetchMock.mock.calls[0];
    const upstreamHeaders = new Headers(init?.headers);
    expect(init?.body).toBeInstanceOf(ArrayBuffer);
    expect(init).not.toHaveProperty("duplex");
    expect(upstreamHeaders.get("content-type")).toBe(request.headers.get("content-type"));
    expect(upstreamHeaders.get("content-length")).toBe(String(expectedBody.byteLength));
    expect(new Uint8Array(init?.body as ArrayBuffer)).toEqual(new Uint8Array(expectedBody));
  });

  it("parses a JSON response even when the upstream content type is missing", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "acme",
      user: { id: "user-1" },
    });
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response('{"data":[{"b64_json":"YWJj"}]}'),
    );

    const response = await proxyToControlPlane({
      request: tenantRequest("/api/playground/run"),
      method: "POST",
      path: "/v1/images/generations",
      body: { model: "gpt-image-1.5", prompt: "A bridge" },
    });

    await expect(response.json()).resolves.toEqual({
      data: [{ b64_json: "YWJj" }],
    });
  });

  it("rejects a caller-supplied user identity that differs from the session", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "acme",
      user: { id: "user-1" },
    });
    const fetchMock = vi.spyOn(globalThis, "fetch");

    const response = await proxyToControlPlane({
      request: tenantRequest("/api/logs"),
      method: "GET",
      path: "/api/v2/logs",
      userID: "user-2",
    });

    expect(response.status).toBe(403);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("rejects a tenant that differs from the authenticated session", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "other-tenant",
      user: { id: "user-1" },
    });
    const fetchMock = vi.spyOn(globalThis, "fetch");

    const response = await proxyToControlPlane({
      request: tenantRequest("/api/logs"),
      method: "GET",
      path: "/api/v2/logs",
    });

    expect(response.status).toBe(403);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("does not allow optional upstream headers to replace trusted identity headers", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "acme",
      user: { id: "user-1" },
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      Response.json({ data: [] }),
    );

    await proxyToControlPlane({
      request: tenantRequest("/api/logs"),
      method: "GET",
      path: "/api/v2/logs",
      upstreamHeaders: {
        Authorization: "Bearer attacker-key",
        "x-godarai-tenant": "other-tenant",
        "x-godarai-user-id": "user-2",
        "x-shahrah-user-id": "legacy-user-2",
      },
    });

    const [, init] = fetchMock.mock.calls[0];
    const headers = new Headers(init?.headers);
    expect(headers.get("authorization")).toBe("Bearer test-master-key");
    expect(headers.get("x-godarai-tenant")).toBe("acme");
    expect(headers.get("x-godarai-user-id")).toBe("user-1");
    expect(headers.has("x-shahrah-user-id")).toBe(false);
  });

  it("relays SSE bytes without buffering and exposes only allowlisted operational headers", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "acme",
      user: { id: "user-1" },
    });
    const streamBody = "data: {\"choices\":[]}\n\ndata: [DONE]\n\n";
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(streamBody, {
        headers: {
          "Content-Type": "text/event-stream",
          "x-godarai-cost-irr": "1200",
          "x-internal-secret": "hidden",
          "set-cookie": "session=hidden",
        },
      }),
    );

    const response = await proxyToControlPlaneStream({
      request: tenantRequest("/api/playground/run"),
      method: "POST",
      path: "/v1/chat/completions",
      body: { model: "gpt-test", stream: true },
      exposeResponseHeaders: ["x-godarai-cost-irr", "x-internal-secret", "set-cookie"],
    });

    expect(response.status).toBe(200);
    await expect(response.text()).resolves.toBe(`: gateway-ready\n\n${streamBody}`);
    expect(response.headers.get("cache-control")).toBe("no-cache, no-transform");
    expect(response.headers.get("x-accel-buffering")).toBe("no");
    expect(response.headers.get("x-godarai-cost-irr")).toBe("1200");
    expect(response.headers.has("x-internal-secret")).toBe(false);
    expect(response.headers.has("set-cookie")).toBe(false);
    const [, init] = fetchMock.mock.calls[0];
    expect(init?.body).toBe(JSON.stringify({ model: "gpt-test", stream: true }));
  });

  it("emits periodic SSE heartbeats while an upstream stream is idle", async () => {
    vi.useFakeTimers();
    const encoder = new TextEncoder();
    const decoder = new TextDecoder();
    let sourceController: ReadableStreamDefaultController<Uint8Array> | undefined;
    const source = new ReadableStream<Uint8Array>({
      start(controller) {
        sourceController = controller;
      },
    });
    const reader = createSSEHeartbeatStream(source, { heartbeatIntervalMs: 1_000 }).getReader();

    try {
      const ready = await reader.read();
      expect(decoder.decode(ready.value)).toBe(": gateway-ready\n\n");

      await vi.advanceTimersByTimeAsync(1_000);
      const heartbeat = await reader.read();
      expect(decoder.decode(heartbeat.value)).toBe(": keep-alive\n\n");

      sourceController?.enqueue(encoder.encode("data: {\"choices\":[]}\n\n"));
      const data = await reader.read();
      expect(decoder.decode(data.value)).toBe("data: {\"choices\":[]}\n\n");
      sourceController?.close();
      await expect(reader.read()).resolves.toMatchObject({ done: true });
      expect(vi.getTimerCount()).toBe(0);
    } finally {
      vi.useRealTimers();
    }
  });

  it("never inserts a heartbeat inside a partial SSE event", async () => {
    vi.useFakeTimers();
    const encoder = new TextEncoder();
    const decoder = new TextDecoder();
    let sourceController: ReadableStreamDefaultController<Uint8Array> | undefined;
    const source = new ReadableStream<Uint8Array>({
      start(controller) {
        sourceController = controller;
      },
    });
    const reader = createSSEHeartbeatStream(source, { heartbeatIntervalMs: 1_000 }).getReader();

    try {
      await reader.read();
      sourceController?.enqueue(encoder.encode("data: {\"choices\":"));
      const partial = await reader.read();
      expect(decoder.decode(partial.value)).toBe("data: {\"choices\":");

      await vi.advanceTimersByTimeAsync(1_000);
      sourceController?.enqueue(encoder.encode("[]}\n\n"));
      const remainder = await reader.read();
      expect(decoder.decode(remainder.value)).toBe("[]}\n\n");

      await vi.advanceTimersByTimeAsync(1_000);
      const heartbeat = await reader.read();
      expect(decoder.decode(heartbeat.value)).toBe(": keep-alive\n\n");
      await reader.cancel();
    } finally {
      vi.useRealTimers();
    }
  });

  it("cancels the upstream SSE reader and heartbeat timer when the client disconnects", async () => {
    vi.useFakeTimers();
    const cancel = vi.fn();
    const source = new ReadableStream<Uint8Array>({ cancel });
    const reader = createSSEHeartbeatStream(source, { heartbeatIntervalMs: 1_000 }).getReader();

    try {
      await reader.read();
      await reader.cancel("client disconnected");

      expect(cancel).toHaveBeenCalledWith("client disconnected");
      expect(vi.getTimerCount()).toBe(0);
    } finally {
      vi.useRealTimers();
    }
  });

  it("cancels the upstream SSE reader when the request signal is aborted", async () => {
    vi.useFakeTimers();
    const cancel = vi.fn();
    const abortController = new AbortController();
    const source = new ReadableStream<Uint8Array>({ cancel });
    const reader = createSSEHeartbeatStream(source, {
      heartbeatIntervalMs: 1_000,
      signal: abortController.signal,
    }).getReader();

    try {
      await reader.read();
      const pendingRead = reader.read();
      abortController.abort(new DOMException("request aborted", "AbortError"));

      await expect(pendingRead).rejects.toMatchObject({ name: "AbortError" });
      expect(cancel).toHaveBeenCalledOnce();
      expect(vi.getTimerCount()).toBe(0);
    } finally {
      vi.useRealTimers();
    }
  });

  it("keeps concurrent SSE heartbeat timers isolated", async () => {
    vi.useFakeTimers();
    const firstReader = createSSEHeartbeatStream(new ReadableStream<Uint8Array>(), { heartbeatIntervalMs: 1_000 }).getReader();
    const secondReader = createSSEHeartbeatStream(new ReadableStream<Uint8Array>(), { heartbeatIntervalMs: 1_000 }).getReader();
    const decoder = new TextDecoder();

    try {
      await firstReader.read();
      await secondReader.read();
      expect(vi.getTimerCount()).toBe(2);

      await vi.advanceTimersByTimeAsync(1_000);
      expect(decoder.decode((await firstReader.read()).value)).toBe(": keep-alive\n\n");
      expect(decoder.decode((await secondReader.read()).value)).toBe(": keep-alive\n\n");

      await firstReader.cancel();
      expect(vi.getTimerCount()).toBe(1);
      await vi.advanceTimersByTimeAsync(1_000);
      expect(decoder.decode((await secondReader.read()).value)).toBe(": keep-alive\n\n");
      await secondReader.cancel();
      expect(vi.getTimerCount()).toBe(0);
    } finally {
      vi.useRealTimers();
    }
  });

  it("propagates an upstream SSE failure and clears its heartbeat timer", async () => {
    vi.useFakeTimers();
    let sourceController: ReadableStreamDefaultController<Uint8Array> | undefined;
    const source = new ReadableStream<Uint8Array>({
      start(controller) {
        sourceController = controller;
      },
    });
    const reader = createSSEHeartbeatStream(source, { heartbeatIntervalMs: 1_000 }).getReader();

    try {
      await reader.read();
      const pendingRead = reader.read();
      sourceController?.error(new Error("upstream stream failed"));

      await expect(pendingRead).rejects.toThrow("upstream stream failed");
      expect(vi.getTimerCount()).toBe(0);
    } finally {
      vi.useRealTimers();
    }
  });

  it("does not prepend SSE comments to a non-SSE upstream error", async () => {
    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "acme",
      user: { id: "user-1" },
    });
    vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      Response.json({ error: { message: "provider rejected request" } }, { status: 429 }),
    );

    const response = await proxyToControlPlaneStream({
      request: tenantRequest("/api/playground/run"),
      method: "POST",
      path: "/v1/chat/completions",
      body: { model: "gpt-test", stream: true },
    });

    expect(response.status).toBe(429);
    await expect(response.json()).resolves.toEqual({ error: { message: "provider rejected request" } });
  });

  it("preserves public invitation acceptance through its fixed allowlisted upstream path", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValueOnce(
      Response.json({ invitation: { id: "invite-1" } }),
    );

    const response = await proxyPublicInvitationAcceptance(
      tenantRequest("/api/users/invitations/accept"),
      "invite-token",
    );

    expect(response.status).toBe(200);
    expect(readAuthFromRequestMock).not.toHaveBeenCalled();
    expect(fetchMock).toHaveBeenCalledOnce();
    const [url, init] = fetchMock.mock.calls[0];
    expect(String(url)).toBe("http://127.0.0.1:4000/api/v1/users/invitations/accept");
    expect(init?.method).toBe("POST");
    expect(init?.body).toBe(JSON.stringify({ token: "invite-token" }));
    expect(new Headers(init?.headers).has("x-godarai-user-id")).toBe(false);
  });
});

describe("readUpstreamErrorCode", () => {
  it("forwards stable string codes and ignores numeric status codes", () => {
    expect(readUpstreamErrorCode({ error: { code: "plan_user_limit_reached", message: "full" } })).toBe("plan_user_limit_reached");
    expect(readUpstreamErrorCode({ error: { code: 403, message: "forbidden" } })).toBeUndefined();
    expect(readUpstreamErrorCode({ message: "plain" })).toBeUndefined();
    expect(readUpstreamErrorCode("text")).toBeUndefined();
  });
});
