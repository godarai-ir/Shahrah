import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { GET } from "./route";

const fetchMock = vi.fn();

function callback(provider: string, query = "authority=mock_51d0e963643f&status=ok") {
  return GET(new Request(`http://tenant1.localhost:3020/api/v1/seats/payment/callback/${provider}?${query}`), {
    params: Promise.resolve({ provider }),
  });
}

describe("seat payment callback relay", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
    vi.stubEnv("SHAHRAH_CONTROL_PLANE_URL", "http://control-plane.test:4000");
  });

  afterEach(() => {
    fetchMock.mockReset();
    vi.unstubAllGlobals();
    vi.unstubAllEnvs();
  });

  it("forwards the gateway's return to the control plane and follows its redirect in the browser", async () => {
    const returnURL = "http://tenant1.localhost:3020/billing/seats?payment_status=success&authority=mock_51d0e963643f";
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 302, headers: { Location: returnURL } }));

    const response = await callback("mock");

    expect(fetchMock).toHaveBeenCalledWith(
      new URL("http://control-plane.test:4000/api/v1/seats/payment/callback/mock?authority=mock_51d0e963643f&status=ok"),
      expect.objectContaining({ method: "GET", redirect: "manual" }),
    );
    expect(response.status).toBe(302);
    expect(response.headers.get("location")).toBe(returnURL);
  });

  it("resolves a relative return location against the request", async () => {
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 302, headers: { Location: "/billing/seats?payment_status=failed" } }));

    const response = await callback("zarinpal");

    expect(response.headers.get("location")).toBe("http://tenant1.localhost:3020/billing/seats?payment_status=failed");
  });

  it("passes a control-plane error on instead of redirecting", async () => {
    fetchMock.mockResolvedValueOnce(Response.json({ status: "failed", reason: "order_not_found" }, { status: 400 }));

    const response = await callback("mock", "authority=unknown");

    expect(response.status).toBe(400);
    expect(await response.json()).toEqual({ status: "failed", reason: "order_not_found" });
  });

  it("answers 502 when the control plane is unreachable", async () => {
    fetchMock.mockRejectedValueOnce(new Error("connect ECONNREFUSED"));

    expect((await callback("mock")).status).toBe(502);
  });

  it("never relays the mock gateway in production", async () => {
    vi.stubEnv("NODE_ENV", "production");

    const response = await callback("mock");

    expect(response.status).toBe(404);
    expect(fetchMock).not.toHaveBeenCalled();

    fetchMock.mockResolvedValueOnce(new Response(null, { status: 302, headers: { Location: "/billing/seats" } }));
    expect((await callback("zarinpal")).status).toBe(302);
  });
});
