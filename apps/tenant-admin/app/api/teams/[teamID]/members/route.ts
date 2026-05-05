import { fetchControlPlaneJSON } from "@/lib/control-plane";
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
    method: "GET",
    path: `/api/v1/teams/${teamID}/members`,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to fetch team members" },
      { status: result.statusCode ?? 500 },
    );
  }

  return NextResponse.json(result.payload);
}

export async function POST(
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
    method: "POST",
    path: `/api/v1/teams/${teamID}/members`,
    body,
  });

  if (!result.ok) {
    return NextResponse.json(
      { error: result.errorMessage ?? "Failed to add team member" },
      { status: result.statusCode ?? 500 },
    );
  }

  return NextResponse.json(result.payload, { status: 201 });
}

