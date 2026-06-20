import { NextRequest, NextResponse } from "next/server";
import { readFile } from "node:fs/promises";
import { extname, relative, resolve } from "node:path";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlaneBinary } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{
    bucket: string;
    fileName: string;
  }>;
};

export async function GET(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const { bucket, fileName } = await context.params;
  const normalizedBucket = bucket.trim();
  const normalizedFileName = fileName.trim();
  if (!normalizedBucket || !normalizedFileName) {
    return NextResponse.json(
      { error: { message: "bucket and fileName are required." } },
      { status: 400 },
    );
  }

  const proxied = await proxyToControlPlaneBinary({
    request,
    method: "GET",
    path: `/api/v1/provider-catalog/assets/${encodeURIComponent(normalizedBucket)}/${encodeURIComponent(normalizedFileName)}`,
  });
  if (proxied.status !== 404) {
    return proxied;
  }

  const localAsset = await readLocalModelDiscoveryAsset(normalizedBucket, normalizedFileName);
  if (!localAsset) {
    return proxied;
  }
  return new NextResponse(localAsset.body, {
    status: 200,
    headers: {
      "Content-Type": localAsset.contentType,
      "Cache-Control": "public, max-age=3600",
    },
  });
}

async function readLocalModelDiscoveryAsset(bucket: string, fileName: string): Promise<{
  body: ArrayBuffer;
  contentType: string;
} | null> {
  const roots = [
    resolve(process.cwd(), "..", "model-discovery", "uploads", bucket),
    resolve(process.cwd(), "apps", "model-discovery", "uploads", bucket),
  ];

  for (const root of roots) {
    const filePath = resolve(root, fileName);
    const rel = relative(root, filePath);
    if (rel.startsWith("..") || resolve(root, rel) !== filePath) {
      continue;
    }
    try {
      const file = await readFile(filePath);
      return {
        body: file.buffer.slice(file.byteOffset, file.byteOffset + file.byteLength) as ArrayBuffer,
        contentType: getAssetContentType(fileName),
      };
    } catch {
      // Try the next workspace layout.
    }
  }
  return null;
}

function getAssetContentType(fileName: string): string {
  switch (extname(fileName).toLowerCase()) {
    case ".jpg":
    case ".jpeg":
      return "image/jpeg";
    case ".png":
      return "image/png";
    case ".svg":
      return "image/svg+xml";
    case ".webp":
      return "image/webp";
    default:
      return "application/octet-stream";
  }
}
