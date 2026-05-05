import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { NextRequest, NextResponse } from "next/server";

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ teamID: string; userID: string }> },
) {
  const { teamID, userID } = await params;
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) return NextResponse.json({ error: auth.message }, { status: auth.status });
  const tenantSlug = auth.tenantSlug;
  const body = await request.json();

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    method: "PATCH",
    path: `/api/v1/teams/${teamID}/members/${userID}`,
    body,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to update team member" },
      { status: result.statusCode ?? 500 },
    );
  }

  return NextResponse.json(result.payload);
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ teamID: string; userID: string }> },
) {
  const { teamID, userID } = await params;
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) return NextResponse.json({ error: auth.message }, { status: auth.status });
  const tenantSlug = auth.tenantSlug;

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    method: "DELETE",
    path: `/api/v1/teams/${teamID}/members/${userID}`,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to remove team member" },
      { status: result.statusCode ?? 500 },
    );
  }

  return new NextResponse(null, { status: 204 });
}

