import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { extractTenantSlugFromHost } from "@/lib/tenant";
import { NextRequest, NextResponse } from "next/server";
import { headers } from "next/headers";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ orgUnitID: string }> }
) {
  const { orgUnitID } = await params;
  
  const incomingHeaders = await headers();
  const incomingHost =
    incomingHeaders.get("x-forwarded-host") ??
    incomingHeaders.get("host") ??
    "";

  const tenantSlug =
    extractTenantSlugFromHost(incomingHost) ||
    process.env.SHAHRAH_DEFAULT_TENANT ||
    "default";

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: `/api/v1/org-units/${orgUnitID}`,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to fetch org unit" },
      { status: result.statusCode ?? 500 }
    );
  }

  return NextResponse.json(result.payload);
}

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ orgUnitID: string }> }
) {
  const { orgUnitID } = await params;
  
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
    method: "PATCH",
    path: `/api/v1/org-units/${orgUnitID}`,
    body,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to update org unit" },
      { status: result.statusCode ?? 500 }
    );
  }

  return NextResponse.json(result.payload);
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ orgUnitID: string }> }
) {
  const { orgUnitID } = await params;
  
  const incomingHeaders = await headers();
  const incomingHost =
    incomingHeaders.get("x-forwarded-host") ??
    incomingHeaders.get("host") ??
    "";

  const tenantSlug =
    extractTenantSlugFromHost(incomingHost) ||
    process.env.SHAHRAH_DEFAULT_TENANT ||
    "default";

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    method: "DELETE",
    path: `/api/v1/org-units/${orgUnitID}`,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to delete org unit" },
      { status: result.statusCode ?? 500 }
    );
  }

  return new NextResponse(null, { status: 204 });
}