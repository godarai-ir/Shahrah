import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { extractTenantSlugFromHost } from "@/lib/tenant";
import { NextRequest, NextResponse } from "next/server";
import { headers } from "next/headers";

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ orgUnitID: string; userID: string }> }
) {
  const { orgUnitID, userID } = await params;
  
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
    path: `/api/v1/org-units/${orgUnitID}/members/${userID}`,
    body,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to update org unit member" },
      { status: result.status }
    );
  }

  return NextResponse.json(result.payload);
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ orgUnitID: string; userID: string }> }
) {
  const { orgUnitID, userID } = await params;
  
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
    path: `/api/v1/org-units/${orgUnitID}/members/${userID}`,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to remove org unit member" },
      { status: result.status }
    );
  }

  return new NextResponse(null, { status: 204 });
}