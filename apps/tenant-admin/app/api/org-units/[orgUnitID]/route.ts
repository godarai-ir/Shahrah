import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { NextRequest, NextResponse } from "next/server";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ orgUnitID: string }> }
) {
  const { orgUnitID } = await params;
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) return NextResponse.json({ error: auth.message }, { status: auth.status });
  const tenantSlug = auth.tenantSlug;

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    userID: auth.user.id,
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
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) return NextResponse.json({ error: auth.message }, { status: auth.status });
  const tenantSlug = auth.tenantSlug;

  const body = stripOrgUnitCostCenter(await request.json());

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    userID: auth.user.id,
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

function stripOrgUnitCostCenter(body: unknown): unknown {
  if (!body || typeof body !== "object" || Array.isArray(body)) {
    return body;
  }

  const { metadata, ...rest } = body as Record<string, unknown>;
  const cleanBody: Record<string, unknown> = { ...rest };
  delete cleanBody.cost_center;
  delete cleanBody.costCenter;

  if (metadata && typeof metadata === "object" && !Array.isArray(metadata)) {
    const cleanMetadata = { ...(metadata as Record<string, unknown>) };
    delete cleanMetadata.cost_center;
    delete cleanMetadata.costCenter;
    cleanBody.metadata = cleanMetadata;
  } else if (metadata !== undefined) {
    cleanBody.metadata = metadata;
  }

  return cleanBody;
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ orgUnitID: string }> }
) {
  const { orgUnitID } = await params;
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) return NextResponse.json({ error: auth.message }, { status: auth.status });
  const tenantSlug = auth.tenantSlug;

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    userID: auth.user.id,
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

