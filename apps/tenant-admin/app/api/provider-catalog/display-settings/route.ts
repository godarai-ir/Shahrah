import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { fetchModelDiscoveryJSON } from "@/lib/model-discovery";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  try {
    const settings = await fetchModelDiscoveryJSON<{
      currencyUnit: "USD" | "IRR" | "TOMAN";
      priceScale: "PER_MILLION_TOKENS" | "PER_TOKEN";
      exchangeRate: {
        usdToIrr: number;
        fetchedAt: string;
        source: string;
      };
    }>("/settings/display");

    return NextResponse.json({ data: settings }, { status: 200, headers: { "Cache-Control": "public, max-age=300, s-maxage=1800" } });
  } catch (error) {
    return NextResponse.json(
      { error: { message: error instanceof Error ? error.message : "Failed to load display settings." } },
      { status: 502 },
    );
  }
}

