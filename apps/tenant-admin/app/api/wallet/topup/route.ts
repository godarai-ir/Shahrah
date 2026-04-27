import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type TopUpPayload = {
  amount_irr?: unknown;
  metadata?: unknown;
};

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const body = (await request.json().catch(() => null)) as TopUpPayload | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "بدنه درخواست نامعتبر است." } },
      { status: 400 },
    );
  }

  const amount = parseAmount(body.amount_irr);
  if (amount === null || amount <= 0) {
    return NextResponse.json(
      { error: { message: "amount_irr باید عددی مثبت باشد." } },
      { status: 400 },
    );
  }

  const metadata =
    body.metadata && typeof body.metadata === "object" && !Array.isArray(body.metadata)
      ? body.metadata
      : {};

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/wallet/topup",
    body: {
      amount_irr: amount,
      metadata,
    },
  });
}

function parseAmount(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string" && value.trim() !== "") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return null;
}
