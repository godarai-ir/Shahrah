import { beforeEach, describe, expect, it, vi } from "vitest";
import { NextRequest, NextResponse } from "next/server";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({
  readAuthFromRequest: readAuthFromRequestMock,
}));

import { withRequestAuth } from "./route-handler";

describe("withRequestAuth", () => {
  beforeEach(() => {
    readAuthFromRequestMock.mockReset();
  });

  it("returns the canonical auth error without calling the handler", async () => {
    readAuthFromRequestMock.mockResolvedValue({ ok: false, status: 401, message: "Authentication required" });
    const handler = vi.fn();

    const response = await withRequestAuth(new NextRequest("http://acme.localhost/api/items"), handler);

    expect(response.status).toBe(401);
    await expect(response.json()).resolves.toEqual({ error: { message: "Authentication required" } });
    expect(handler).not.toHaveBeenCalled();
  });

  it("passes the trusted auth context to the route implementation", async () => {
    const auth = { ok: true, tenantSlug: "acme", user: { id: "user-1" } };
    readAuthFromRequestMock.mockResolvedValue(auth);
    const handler = vi.fn().mockResolvedValue(NextResponse.json({ ok: true }));

    const response = await withRequestAuth(new NextRequest("http://acme.localhost/api/items"), handler);

    expect(response.status).toBe(200);
    expect(handler).toHaveBeenCalledWith(auth);
  });
});
