import { mkdir, rm, writeFile } from "node:fs/promises";
import { resolve } from "node:path";
import { NextRequest, NextResponse } from "next/server";
import { describe, expect, it, vi } from "vitest";

import { GET } from "./route";

const proxyToControlPlaneBodyStreamMock = vi.hoisted(() => vi.fn());
const readAuthFromRequestMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/control-plane", () => ({
  proxyToControlPlaneBodyStream: proxyToControlPlaneBodyStreamMock,
}));
vi.mock("@/lib/auth/request-auth", () => ({
  readAuthFromRequest: readAuthFromRequestMock,
}));

describe("bank transfer receipt files", () => {
  it("streams the tenant-scoped receipt from control-plane object storage", async () => {
    const request = new NextRequest("http://localhost/api/wallet/bank-transfer-receipts/11111111-1111-4111-8111-111111111111.png");
    proxyToControlPlaneBodyStreamMock.mockResolvedValueOnce(new NextResponse("image", {
      status: 200,
      headers: { "content-type": "image/png" },
    }));

    const response = await GET(request, {
      params: Promise.resolve({ fileName: "11111111-1111-4111-8111-111111111111.png" }),
    });

    expect(proxyToControlPlaneBodyStreamMock).toHaveBeenCalledWith({
      request,
      method: "GET",
      path: "/api/v1/wallet/bank-transfer-receipts/11111111-1111-4111-8111-111111111111.png",
    });
    expect(response.status).toBe(200);
    expect(response.headers.get("content-type")).toBe("image/png");
  });

  it("keeps legacy local receipts readable for their owning tenant", async () => {
    const fileName = `t_${tenantKey("tenant-one")}_11111111-1111-4111-8111-111111111111.png`;
    const uploadDir = resolve(process.cwd(), "uploads", "bank-transfer-receipts");
    const filePath = resolve(uploadDir, fileName);
    await mkdir(uploadDir, { recursive: true });
    await writeFile(filePath, Buffer.from([0x89, 0x50, 0x4e, 0x47]));
    proxyToControlPlaneBodyStreamMock.mockResolvedValueOnce(new NextResponse(null, { status: 404 }));
    readAuthFromRequestMock.mockResolvedValueOnce({ ok: true, tenantSlug: "tenant-one" });

    try {
      const request = new NextRequest(`http://localhost/api/wallet/bank-transfer-receipts/${fileName}`);
      const response = await GET(request, { params: Promise.resolve({ fileName }) });

      expect(response.status).toBe(200);
      expect(response.headers.get("content-type")).toBe("image/png");
    } finally {
      await rm(filePath, { force: true });
    }
  });
});

function tenantKey(tenantSlug: string): string {
  return Buffer.from(tenantSlug.trim().toLowerCase()).toString("base64url");
}
