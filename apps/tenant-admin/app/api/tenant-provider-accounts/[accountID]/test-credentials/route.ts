import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{
    accountID: string;
  }>;
};

type TestCredentialsPayload = {
  model_hint?: unknown;
};

export async function POST(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const { accountID } = await context.params;
  const normalizedAccountID = accountID.trim();
  if (!normalizedAccountID) {
    return NextResponse.json(
      { error: { message: "accountID معتبر نیست." } },
      { status: 400 },
    );
  }

  const body = (await request.json().catch(() => ({}))) as TestCredentialsPayload;
  const modelHint =
    typeof body.model_hint === "string" ? body.model_hint.trim() : undefined;

  return proxyToControlPlane({
    request,
    method: "POST",
    path: `/api/v1/tenant-provider-accounts/${encodeURIComponent(normalizedAccountID)}/test-credentials`,
    body: modelHint ? { model_hint: modelHint } : {},
  });
}
