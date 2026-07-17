import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { TENANT_ADMIN_SESSION_COOKIE } from "@/lib/auth/session";
import { POST } from "./route";

const revokeSessionMock = vi.hoisted(() => vi.fn());
const getAuthServiceMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/factory", () => ({
  getAuthService: getAuthServiceMock,
}));

describe("auth/logout route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    getAuthServiceMock.mockReturnValue({ revokeSession: revokeSessionMock });
    revokeSessionMock.mockResolvedValue(undefined);
  });

  it("revokes the backend session and expires the browser cookie", async () => {
    const response = await POST(buildRequest());

    expect(response.status).toBe(200);
    expect(revokeSessionMock).toHaveBeenCalledWith("session-1");
    expect(response.headers.get("set-cookie")).toContain(
      `${TENANT_ADMIN_SESSION_COOKIE}=`,
    );
    expect(response.headers.get("set-cookie")).toContain("Max-Age=0");
  });

  it("still expires the browser cookie when backend revocation fails", async () => {
    revokeSessionMock.mockRejectedValueOnce(new Error("control-plane unavailable"));
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => undefined);

    try {
      const response = await POST(buildRequest());

      expect(response.status).toBe(502);
      expect(response.headers.get("set-cookie")).toContain(
        `${TENANT_ADMIN_SESSION_COOKIE}=`,
      );
      expect(response.headers.get("set-cookie")).toContain("Max-Age=0");
    } finally {
      consoleError.mockRestore();
    }
  });
});

function buildRequest() {
  return new NextRequest("https://tenant1.example.com/api/auth/logout", {
    method: "POST",
    headers: {
      cookie: `${TENANT_ADMIN_SESSION_COOKIE}=session-1`,
      host: "tenant1.example.com",
      origin: "https://tenant1.example.com",
      "sec-fetch-site": "same-origin",
    },
  });
}
