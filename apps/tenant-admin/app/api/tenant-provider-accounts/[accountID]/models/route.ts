import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{
    accountID: string;
  }>;
};

type UpdateModelsPayload = {
  models?: unknown;
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

  const body = (await request.json().catch(() => null)) as UpdateModelsPayload | null;
  if (!body || typeof body !== "object" || !Array.isArray(body.models)) {
    return NextResponse.json(
      { error: { message: "آرایه models الزامی است." } },
      { status: 400 },
    );
  }

  const items = body.models.filter(
    (item) => !!item && typeof item === "object" && !Array.isArray(item),
  );
  if (items.length === 0) {
    return NextResponse.json(
      { error: { message: "حداقل یک مدل باید انتخاب شود." } },
      { status: 400 },
    );
  }

  return proxyToControlPlane({
    request,
    method: "PUT",
    path: `/api/v1/tenant-provider-accounts/${encodeURIComponent(normalizedAccountID)}/models`,
    body: {
      models: items,
    },
  });
}
