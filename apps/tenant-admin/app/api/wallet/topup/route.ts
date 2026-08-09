import { NextRequest, NextResponse } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";
import { withRequestAuth } from "@/lib/http/server/route-handler";

export const dynamic = "force-dynamic";

type TopUpPayload = {
  amount_irr?: unknown;
  amount_usd?: unknown;
  provider?: unknown;
  return_url?: unknown;
  metadata?: unknown;
};

export function POST(request: NextRequest) {
  return withRequestAuth(request, async (auth) => {
    const body = (await request.json().catch(() => null)) as TopUpPayload | null;
    if (!body || typeof body !== "object") {
      return NextResponse.json(
        { error: { message: "بدنه درخواست نامعتبر است." } },
        { status: 400 },
      );
    }

    const amountIRR = parseAmount(body.amount_irr);
    const amountUSD = parseAmount(body.amount_usd);
    if ((amountIRR === null || amountIRR <= 0) && (amountUSD === null || amountUSD <= 0)) {
      return NextResponse.json(
        { error: { message: "مقدار amount_irr باید یک عدد مثبت باشد." } },
        { status: 400 },
      );
    }

    const provider =
      typeof body.provider === "string" && body.provider.trim() !== ""
        ? body.provider.trim().toLowerCase()
        : "mock";

    const returnURL =
      typeof body.return_url === "string" && body.return_url.trim() !== ""
        ? body.return_url.trim()
        : undefined;

    const metadata =
      body.metadata && typeof body.metadata === "object" && !Array.isArray(body.metadata)
        ? body.metadata
        : {};

    return proxyToControlPlane({
      request,
      auth,
      method: "POST",
      path: "/api/v1/wallet/topup/intents",
      body: {
        ...(amountIRR !== null && amountIRR > 0
          ? { amount_irr: Math.round(amountIRR) }
          : { amount_usd: amountUSD }),
        provider,
        return_url: returnURL,
        metadata,
      },
    });
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
