import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { extractTenantSlugFromHost } from "@/lib/tenant";
import { NextRequest, NextResponse } from "next/server";
import { headers } from "next/headers";

export async function GET(request: NextRequest) {
  const incomingHeaders = await headers();
  const incomingHost =
    incomingHeaders.get("x-forwarded-host") ??
    incomingHeaders.get("host") ??
    "";

  const tenantSlug =
    extractTenantSlugFromHost(incomingHost) ||
    process.env.SHAHRAH_DEFAULT_TENANT ||
    "default";

  const { searchParams } = new URL(request.url);
  const query = searchParams.get("query") ?? "";
  const limit = searchParams.get("limit") ?? "25";
  const offset = searchParams.get("offset") ?? "0";

  const result = await fetchControlPlaneJSON({
    tenantSlug,
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
  const incomingHeaders = await headers();
  const incomingHost =
    incomingHeaders.get("x-forwarded-host") ??
    incomingHeaders.get("host") ??
    "";

  const tenantSlug =
    extractTenantSlugFromHost(incomingHost) ||
    process.env.SHAHRAH_DEFAULT_TENANT ||
    "default";

  const body = await request.json();

  const result = await fetchControlPlaneJSON({
    tenantSlug,
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