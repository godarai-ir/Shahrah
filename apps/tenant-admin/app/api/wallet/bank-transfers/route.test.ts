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

describe("wallet/bank-transfers POST route", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("passes the Jalali bank transfer contract through to control-plane", async () => {
    const request = new NextRequest("http://localhost/api/wallet/bank-transfers", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        amount_irr: 250000,
        tracking_number: "ABC-123",
        bank_transaction_date_jalali: "1405-04-17",
      }),
    });

    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "tenant-one",
    });
    proxyToControlPlaneMock.mockResolvedValueOnce(new NextResponse(null, { status: 201 }));

    const response = await POST(request);

    expect(proxyToControlPlaneMock).toHaveBeenCalledWith({
      request,
      method: "POST",
      path: "/api/v1/wallet/bank-transfers",
      body: { amount_irr: 250000, tracking_number: "ABC-123", bank_transaction_date_jalali: "1405-04-17" },
    });
    expect(response.status).toBe(201);
  });

  it("rejects caller-supplied receipt URLs on JSON requests", async () => {
    const request = new NextRequest("http://localhost/api/wallet/bank-transfers", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        amount_irr: 250000,
        tracking_number: "ABC-123",
        bank_transaction_date_jalali: "1405-04-17",
        receipt_file_url: "/api/wallet/bank-transfer-receipts/t_other_tenant_file.png",
      }),
    });

    readAuthFromRequestMock.mockResolvedValueOnce({
      ok: true,
      tenantSlug: "tenant-one",
    });

    const response = await POST(request);

    expect(response.status).toBe(400);
    expect(proxyToControlPlaneMock).not.toHaveBeenCalled();
  });
});
