import { NextRequest, NextResponse } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { GET } from "./route";

const proxyToControlPlaneMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/control-plane", () => ({
  proxyToControlPlane: proxyToControlPlaneMock,
}));

describe("tenant-provider-account-models GET route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    proxyToControlPlaneMock.mockResolvedValue(NextResponse.json({ data: [], total: 0 }));
  });

  it("normalizes and proxies allowlisted catalog filters", async () => {
    const request = new NextRequest(
      "http://localhost/api/tenant-provider-account-models?q=%20gpt%20&providers=openai,%20anthropic,openai&account_ids=a1,a2&sort=context&order=DESC&limit=50&offset=10&secret=ignored",
    );

    await GET(request);

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request,
      method: "GET",
      path: "/api/v1/tenant-provider-account-models?q=gpt&providers=openai%2Canthropic&account_ids=a1%2Ca2&sort=context&order=desc&limit=50&offset=10",
    });
  });

  it("drops invalid sort, order, and pagination values", async () => {
    const request = new NextRequest(
      "http://localhost/api/tenant-provider-account-models?sort=secret&order=random&limit=1000&offset=-1",
    );

    await GET(request);

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request,
      method: "GET",
      path: "/api/v1/tenant-provider-account-models",
    });
  });
});
