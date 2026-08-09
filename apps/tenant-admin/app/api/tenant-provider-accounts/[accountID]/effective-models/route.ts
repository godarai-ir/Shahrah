import { NextRequest, NextResponse } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{
    accountID: string;
  }>;
};

export async function GET(request: NextRequest, context: RouteContext) {

  const { accountID } = await context.params;
  const normalizedAccountID = accountID.trim();
  if (!normalizedAccountID) {
    return NextResponse.json(
      { error: { message: "accountID معتبر نیست." } },
      { status: 400 },
    );
  }

  const query = request.nextUrl.searchParams;
  const params = new URLSearchParams();
  const keyType = query.get("key_type")?.trim() ?? "";
  const userID = query.get("user_id")?.trim() ?? "";
  const teamID = query.get("team_id")?.trim() ?? "";
  if (keyType !== "pat" && keyType !== "vat") {
    return NextResponse.json(
      { error: { message: "key_type باید pat یا vat باشد." } },
      { status: 400 },
    );
  }
  if (Boolean(userID) === Boolean(teamID)) {
    return NextResponse.json(
      { error: { message: "ارسال دقیقاً یکی از user_id یا team_id الزامی است." } },
      { status: 400 },
    );
  }
  if (keyType) params.set("key_type", keyType);
  if (userID) params.set("user_id", userID);
  if (teamID) params.set("team_id", teamID);

  const path = `/api/v1/tenant-provider-accounts/${encodeURIComponent(normalizedAccountID)}/effective-models${params.toString() ? `?${params.toString()}` : ""}`;
  return proxyToControlPlane({
    request,
    method: "GET",
    path,
  });
}

