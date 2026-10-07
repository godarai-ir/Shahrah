import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { GET } from "./route";

const fetchControlPlaneJSONMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/control-plane", () => ({
  fetchControlPlaneJSON: fetchControlPlaneJSONMock,
}));

const REQUEST_ID = "A".repeat(43);

function continueRequest(options: { requestID?: string; session?: string } = {}) {
  const url = new URL("http://acme.localhost:3000/auth/oidc/continue");
  url.searchParams.set("oidc_request", options.requestID ?? REQUEST_ID);
  const headers = new Headers({ host: "acme.localhost:3000", "x-godarai-tenant-resolved": "acme" });
  if (options.session) {
    headers.set("cookie", `tenant_admin_session=${options.session}`);
  }
  return new NextRequest(url, { headers });
}

function location(response: Response) {
  return new URL(response.headers.get("location") ?? "", "http://acme.localhost:3000");
}

describe("OIDC continue route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("accepts the request for the current session and redirects to the client", async () => {
    fetchControlPlaneJSONMock.mockResolvedValueOnce({
      ok: true,
      status: 200,
      payload: { redirect_to: "https://chat.acme.example/oauth/openid/callback?code=gdc_x&state=s" },
      responseHeaders: {},
    });

    const response = await GET(continueRequest({ session: "session-1" }));

    expect(response.status).toBe(303);
    expect(response.headers.get("location")).toBe("https://chat.acme.example/oauth/openid/callback?code=gdc_x&state=s");
    expect(response.headers.get("cache-control")).toBe("no-store");
    expect(fetchControlPlaneJSONMock).toHaveBeenCalledWith(expect.objectContaining({
      method: "POST",
      path: `/api/v1/auth/oidc/requests/${REQUEST_ID}/accept`,
      body: { tenant_slug: "acme", session_id: "session-1" },
    }));
  });

  it("sends a visitor without a session to the login page with the pending request", async () => {
    const response = await GET(continueRequest());

    expect(response.status).toBe(303);
    expect(location(response).pathname).toBe("/login");
    expect(location(response).searchParams.get("oidc_request")).toBe(REQUEST_ID);
    expect(fetchControlPlaneJSONMock).not.toHaveBeenCalled();
  });

  it("asks for a fresh sign-in when the control plane rejects the session", async () => {
    fetchControlPlaneJSONMock.mockResolvedValueOnce({
      ok: false,
      status: 401,
      payload: { error: { code: "oidc_session_invalid" } },
      responseHeaders: {},
    });

    const response = await GET(continueRequest({ session: "stale" }));

    expect(location(response).searchParams.get("oidc_reauth")).toBe("1");
    expect(location(response).searchParams.get("oidc_request")).toBe(REQUEST_ID);
  });

  it("reports expired requests and outages with stable codes", async () => {
    fetchControlPlaneJSONMock.mockResolvedValueOnce({
      ok: false,
      status: 410,
      payload: { error: { code: "oidc_request_expired" } },
      responseHeaders: {},
    });
    const expired = await GET(continueRequest({ session: "session-1" }));
    expect(location(expired).searchParams.get("oidc_error")).toBe("expired");

    fetchControlPlaneJSONMock.mockResolvedValueOnce({ ok: false, status: 500, payload: null, responseHeaders: {} });
    const outage = await GET(continueRequest({ session: "session-1" }));
    expect(location(outage).searchParams.get("oidc_error")).toBe("unavailable");
  });

  it("rejects malformed request identifiers without calling the control plane", async () => {
    const response = await GET(continueRequest({ requestID: "../../etc", session: "session-1" }));

    expect(location(response).searchParams.get("oidc_error")).toBe("expired");
    expect(fetchControlPlaneJSONMock).not.toHaveBeenCalled();
  });

  it("never follows a control-plane redirect that is not an absolute URL", async () => {
    fetchControlPlaneJSONMock.mockResolvedValueOnce({
      ok: true,
      status: 200,
      payload: { redirect_to: "javascript-free relative/path" },
      responseHeaders: {},
    });

    const response = await GET(continueRequest({ session: "session-1" }));

    expect(location(response).searchParams.get("oidc_error")).toBe("unavailable");
  });
});
