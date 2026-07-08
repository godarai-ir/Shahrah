import { randomUUID } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { extname, join, relative, resolve } from "node:path";
import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";

export const dynamic = "force-dynamic";

const MAX_RECEIPT_IMAGE_SIZE = 2 * 1024 * 1024;
const RECEIPT_UPLOAD_DIR = resolve(process.cwd(), "uploads", "bank-transfer-receipts");
const ALLOWED_RECEIPT_IMAGE_TYPES = new Map<string, string>([
  ["image/jpeg", ".jpg"],
  ["image/png", ".png"],
  ["image/webp", ".webp"],
  ["image/gif", ".gif"],
]);

type RouteContext = {
  params: Promise<{
    fileName: string;
  }>;
};

export async function GET(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  const { fileName } = await context.params;
  const normalizedFileName = fileName.trim();
  if (!normalizedFileName) {
    return NextResponse.json({ error: { message: "fileName is required." } }, { status: 400 });
  }

  const filePath = resolve(RECEIPT_UPLOAD_DIR, normalizedFileName);
  const relativePath = relative(RECEIPT_UPLOAD_DIR, filePath);
  if (relativePath.startsWith("..") || resolve(RECEIPT_UPLOAD_DIR, relativePath) !== filePath) {
    return NextResponse.json({ error: { message: "invalid file name." } }, { status: 400 });
  }

  try {
    const file = await readFile(filePath);
    return new NextResponse(file, {
      status: 200,
      headers: {
        "Content-Type": getReceiptContentType(normalizedFileName),
        "Cache-Control": "private, max-age=3600",
      },
    });
  } catch {
    return NextResponse.json({ error: { message: "file not found." } }, { status: 404 });
  }
}

export async function saveReceiptImage(file: File, tenantSlug: string) {
  if (!ALLOWED_RECEIPT_IMAGE_TYPES.has(file.type)) {
    throw new Error("فقط فایل‌های JPG، PNG، WEBP و GIF مجاز هستند.");
  }
  if (file.size > MAX_RECEIPT_IMAGE_SIZE) {
    throw new Error("حجم تصویر رسید باید حداکثر ۲ مگابایت باشد.");
  }

  const preferredExtension = ALLOWED_RECEIPT_IMAGE_TYPES.get(file.type) ?? ".bin";
  const originalExtension = extname(file.name || "").toLowerCase();
  const extension = originalExtension && originalExtension.length <= 5 ? originalExtension : preferredExtension;
  const safeTenant = tenantSlug.trim().replace(/[^a-z0-9-_]/gi, "-") || "tenant";
  const fileName = `${safeTenant}-${Date.now()}-${randomUUID()}${extension}`;
  const filePath = join(RECEIPT_UPLOAD_DIR, fileName);

  await mkdir(RECEIPT_UPLOAD_DIR, { recursive: true });
  await writeFile(filePath, Buffer.from(await file.arrayBuffer()));

  return {
    fileName,
    url: `/api/wallet/bank-transfer-receipts/${encodeURIComponent(fileName)}`,
  };
}

function getReceiptContentType(fileName: string): string {
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
