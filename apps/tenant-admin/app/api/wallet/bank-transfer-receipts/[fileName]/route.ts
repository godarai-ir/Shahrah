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
  if (!isReceiptOwnedByTenant(normalizedFileName, auth.tenantSlug)) {
    return NextResponse.json({ error: { message: "file not found." } }, { status: 404 });
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

  const content = Buffer.from(await file.arrayBuffer());
  const detectedType = detectReceiptImageType(content);
  if (!detectedType || detectedType !== file.type) {
    throw new Error("فایل رسید معتبر نیست یا با نوع تصویر انتخاب‌شده همخوانی ندارد.");
  }

  const preferredExtension = ALLOWED_RECEIPT_IMAGE_TYPES.get(file.type) ?? ".bin";
  const fileName = `${tenantReceiptPrefix(tenantSlug)}${randomUUID()}${preferredExtension}`;
  const filePath = join(RECEIPT_UPLOAD_DIR, fileName);

  await mkdir(RECEIPT_UPLOAD_DIR, { recursive: true });
  await writeFile(filePath, content);

  return {
    fileName,
    url: `/api/wallet/bank-transfer-receipts/${encodeURIComponent(fileName)}`,
  };
}

function isReceiptOwnedByTenant(fileName: string, tenantSlug: string): boolean {
  return fileName.startsWith(tenantReceiptPrefix(tenantSlug));
}

function tenantReceiptPrefix(tenantSlug: string): string {
  const normalizedTenant = tenantSlug.trim().toLowerCase();
  const tenantKey = Buffer.from(normalizedTenant || "tenant").toString("base64url");
  return `t_${tenantKey}_`;
}

function detectReceiptImageType(content: Buffer): string | null {
  if (content.length >= 3 && content[0] === 0xff && content[1] === 0xd8 && content[2] === 0xff) {
    return "image/jpeg";
  }

  if (
    content.length >= 8 &&
    content[0] === 0x89 &&
    content[1] === 0x50 &&
    content[2] === 0x4e &&
    content[3] === 0x47 &&
    content[4] === 0x0d &&
    content[5] === 0x0a &&
    content[6] === 0x1a &&
    content[7] === 0x0a
  ) {
    return "image/png";
  }

  if (
    content.length >= 12 &&
    content.subarray(0, 4).toString("ascii") === "RIFF" &&
    content.subarray(8, 12).toString("ascii") === "WEBP"
  ) {
    return "image/webp";
  }

  if (
    content.length >= 6 &&
    (content.subarray(0, 6).toString("ascii") === "GIF87a" ||
      content.subarray(0, 6).toString("ascii") === "GIF89a")
  ) {
    return "image/gif";
  }

  return null;
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
