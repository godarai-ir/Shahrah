import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { getCurrentMePermissions, getNormalizedPermissionSet } from "@/lib/auth/me-permissions";
import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { NextRequest, NextResponse } from "next/server";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ teamID: string }> },
) {
  const { teamID } = await params;
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) return NextResponse.json({ error: auth.message }, { status: auth.status });
  const tenantSlug = auth.tenantSlug;

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    userID: auth.user.id,
    method: "GET",
    path: `/api/v1/teams/${teamID}`,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to fetch team" },
      { status: result.statusCode ?? 500 },
    );
  }

  return NextResponse.json(result.payload);
}

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ teamID: string }> },
) {
  const { teamID } = await params;
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) return NextResponse.json({ error: auth.message }, { status: auth.status });
  const tenantSlug = auth.tenantSlug;
  const body = await request.json();

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    userID: auth.user.id,
    method: "PATCH",
    path: `/api/v1/teams/${teamID}`,
    body,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to update team" },
      { status: result.statusCode ?? 500 },
    );
  }

  return NextResponse.json(result.payload);
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ teamID: string }> },
) {
  const { teamID } = await params;
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) return NextResponse.json({ error: auth.message }, { status: auth.status });
  const tenantSlug = auth.tenantSlug;
  try {
    const permissionsPayload = await getCurrentMePermissions(tenantSlug, auth.user.id);
    const permissionSet = getNormalizedPermissionSet(permissionsPayload);
    if (!permissionSet.has("team:delete")) {
      return NextResponse.json({ error: "Insufficient permissions (team:delete required)." }, { status: 403 });
    }
  } catch {
    return NextResponse.json({ error: "Failed to verify permissions." }, { status: 403 });
  }

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    userID: auth.user.id,
    method: "DELETE",
    path: `/api/v1/teams/${teamID}`,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to delete team" },
      { status: result.statusCode ?? 500 },
    );
  }

  return new NextResponse(null, { status: 204 });
}


