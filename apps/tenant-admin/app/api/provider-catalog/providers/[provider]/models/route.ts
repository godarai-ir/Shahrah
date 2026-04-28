import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{
    provider: string;
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

  const { provider } = await context.params;
  const normalizedProvider = provider.trim().toLowerCase();
  if (!normalizedProvider) {
    return NextResponse.json(
      { error: { message: "provider معتبر نیست." } },
      { status: 400 },
    );
  }

  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/api/v1/provider-catalog/providers/${encodeURIComponent(normalizedProvider)}/models`,
  });
}
