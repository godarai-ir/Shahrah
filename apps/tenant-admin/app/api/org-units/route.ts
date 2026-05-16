import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { NextRequest, NextResponse } from "next/server";

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.message }, { status: auth.status });
  }
  const tenantSlug = auth.tenantSlug;

  const { searchParams } = new URL(request.url);
  const query = searchParams.get("query") ?? "";
  const limit = searchParams.get("limit") ?? "25";
  const offset = searchParams.get("offset") ?? "0";

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    userID: auth.user.id,
    method: "GET",
    path: `/api/v1/org-units?query=${encodeURIComponent(query)}&limit=${limit}&offset=${offset}`,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to fetch org units" },
      { status: result.statusCode ?? 500 }
    );
  }

  return NextResponse.json(result.payload);
}

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json({ error: auth.message }, { status: auth.status });
  }
  const tenantSlug = auth.tenantSlug;

  const body = await request.json();

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    userID: auth.user.id,
    method: "POST",
    path: "/api/v1/org-units",
    body,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to create org unit" },
      { status: result.statusCode ?? 500 }
    );
  }

  return NextResponse.json(result.payload, { status: 201 });
}

