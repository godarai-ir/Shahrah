import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";

export const dynamic = "force-dynamic";

const REMOVAL_MESSAGE =
  "Legacy /api/keys endpoint is removed. Use /api/personal-access-tokens for PAT and /api/v1/keys/vat for VAT.";

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  return NextResponse.json({ error: { message: REMOVAL_MESSAGE } }, { status: 410 });
}

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: { message: auth.message } }, { status: auth.status });
  }

  return NextResponse.json({ error: { message: REMOVAL_MESSAGE } }, { status: 410 });
}
