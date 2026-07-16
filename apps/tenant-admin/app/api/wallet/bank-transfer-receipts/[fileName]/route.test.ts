import { mkdir, rm, writeFile } from "node:fs/promises";
import { resolve } from "node:path";
import { NextRequest } from "next/server";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { GET, saveReceiptImage } from "./route";

const readAuthFromRequestMock = vi.hoisted(() => vi.fn());

vi.mock("@/lib/auth/request-auth", () => ({
  readAuthFromRequest: readAuthFromRequestMock,
}));

const uploadDir = resolve(process.cwd(), "uploads", "bank-transfer-receipts");

describe("bank transfer receipt files", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("does not serve a receipt owned by another tenant", async () => {
    const fileName = `t_${tenantKey("tenant-two")}_11111111-1111-4111-8111-111111111111.png`;
    const filePath = resolve(uploadDir, fileName);
    await mkdir(uploadDir, { recursive: true });
    await writeFile(filePath, Buffer.from([0x89, 0x50, 0x4e, 0x47]));

    try {
      readAuthFromRequestMock.mockResolvedValueOnce({
        ok: true,
        tenantSlug: "tenant-one",
      });

      const response = await GET(new NextRequest(`http://localhost/api/wallet/bank-transfer-receipts/${fileName}`), {
        params: Promise.resolve({ fileName }),
      });

      expect(response.status).toBe(404);
    } finally {
      await rm(filePath, { force: true });
    }
  });

  it("serves a valid receipt only to the tenant that uploaded it", async () => {
    const file = new File([minimalPNG()], "receipt.png", { type: "image/png" });
    const uploaded = await saveReceiptImage(file, "tenant-one");
    const filePath = resolve(uploadDir, uploaded.fileName);

    try {
      readAuthFromRequestMock.mockResolvedValueOnce({
        ok: true,
        tenantSlug: "tenant-one",
      });

      const response = await GET(new NextRequest(`http://localhost${uploaded.url}`), {
        params: Promise.resolve({ fileName: uploaded.fileName }),
      });

      expect(response.status).toBe(200);
      expect(response.headers.get("Content-Type")).toBe("image/png");
      expect(uploaded.fileName).toMatch(new RegExp(`^t_${tenantKey("tenant-one")}_`));
    } finally {
      await rm(filePath, { force: true });
    }
  });

  it("rejects content that does not match the declared image type", async () => {
    const file = new File([Buffer.from("not an image")], "receipt.png", { type: "image/png" });

    await expect(saveReceiptImage(file, "tenant-one")).rejects.toThrow("معتبر");
  });
});

function tenantKey(tenantSlug: string): string {
  return Buffer.from(tenantSlug.trim().toLowerCase()).toString("base64url");
}

function minimalPNG(): Buffer {
  return Buffer.from([
    0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
    0x49, 0x48, 0x44, 0x52,
  ]);
}
