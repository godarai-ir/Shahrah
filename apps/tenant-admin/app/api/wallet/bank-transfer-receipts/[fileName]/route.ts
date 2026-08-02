import { readFile } from "node:fs/promises";
import { extname, relative, resolve } from "node:path";
import { NextResponse, type NextRequest } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlaneBodyStream } from "@/lib/control-plane";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";
const LEGACY_RECEIPT_UPLOAD_DIR = resolve(process.cwd(), "uploads", "bank-transfer-receipts");

type RouteContext = {
  params: Promise<{
    fileName: string;
  }>;
};

export async function GET(request: NextRequest, context: RouteContext) {
  const { fileName } = await context.params;
  const upstream = await proxyToControlPlaneBodyStream({
    request,
    method: "GET",
    path: `/api/v1/wallet/bank-transfer-receipts/${encodeURIComponent(fileName)}`,
  });
  if (upstream.status !== 404 || !fileName.startsWith("t_")) {
    return upstream;
  }

  return readLegacyReceipt(request, fileName);
}

async function readLegacyReceipt(request: NextRequest, fileName: string) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok || !fileName.startsWith(legacyTenantPrefix(auth.tenantSlug))) {
    return NextResponse.json({ error: { message: "file not found." } }, { status: 404 });
  }

  const filePath = resolve(LEGACY_RECEIPT_UPLOAD_DIR, fileName);
  const relativePath = relative(LEGACY_RECEIPT_UPLOAD_DIR, filePath);
  if (relativePath.startsWith("..") || resolve(LEGACY_RECEIPT_UPLOAD_DIR, relativePath) !== filePath) {
    return NextResponse.json({ error: { message: "file not found." } }, { status: 404 });
  }

  try {
    return new NextResponse(await readFile(filePath), {
      status: 200,
      headers: {
        "Content-Type": legacyReceiptContentType(fileName),
        "Cache-Control": "private, max-age=3600",
        "X-Content-Type-Options": "nosniff",
      },
    });
  } catch {
    return NextResponse.json({ error: { message: "file not found." } }, { status: 404 });
  }
}

function legacyTenantPrefix(tenantSlug: string): string {
  const tenantKey = Buffer.from(tenantSlug.trim().toLowerCase() || "tenant").toString("base64url");
  return `t_${tenantKey}_`;
}

function legacyReceiptContentType(fileName: string): string {
  switch (extname(fileName).toLowerCase()) {
    case ".jpg":
    case ".jpeg":
      return "image/jpeg";
    case ".png":
      return "image/png";
    case ".webp":
      return "image/webp";
    case ".gif":
      return "image/gif";
    default:
      return "application/octet-stream";
  }
}
