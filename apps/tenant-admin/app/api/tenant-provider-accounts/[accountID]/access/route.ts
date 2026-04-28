import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{
    accountID: string;
  }>;
};

type UpdateAccessPayload = {
  access?: unknown;
};

export async function PUT(request: NextRequest, context: RouteContext) {
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

  const body = (await request.json().catch(() => null)) as UpdateAccessPayload | null;
  if (!body || typeof body !== "object" || !isPlainObject(body.access)) {
    return NextResponse.json(
      { error: { message: "فیلد access الزامی است." } },
      { status: 400 },
    );
  }

  return proxyToControlPlane({
    request,
    method: "PUT",
    path: `/api/v1/tenant-provider-accounts/${encodeURIComponent(normalizedAccountID)}/access`,
    body: {
      access: body.access,
    },
  });
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === "object" && !Array.isArray(value);
}
