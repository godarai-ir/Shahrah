import { NextRequest, NextResponse } from "next/server";

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

  return proxyToControlPlaneBinary({
    request,
    method: "GET",
    path: `/api/v1/provider-catalog/assets/${encodeURIComponent(normalizedBucket)}/${encodeURIComponent(normalizedFileName)}`,
  });
}
