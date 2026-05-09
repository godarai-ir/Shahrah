import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type RouteContext = {
  params: Promise<{
    accountID: string;
  }>;
};

type CredentialsPayload = {
  auth_mode?: unknown;
  allow_virtual_keys_only?: unknown;
  provider_config?: unknown;
  credentials?: unknown;
};

function isPlainObject(input: unknown): input is Record<string, unknown> {
  return !!input && typeof input === "object" && !Array.isArray(input);
}

export async function PUT(request: NextRequest, context: RouteContext) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const { accountID } = await context.params;
  const normalizedAccountID = accountID.trim();
  if (!normalizedAccountID) {
    return NextResponse.json(
      { error: { message: "accountID معتبر نیست." } },
      { status: 400 },
    );
  }

  const body = (await request.json().catch(() => null)) as CredentialsPayload | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "بدنه درخواست نامعتبر است." } },
      { status: 400 },
    );
  }

  const authMode = typeof body.auth_mode === "string" ? body.auth_mode.trim().toLowerCase() : undefined;
  const allowVirtualKeysOnly = typeof body.allow_virtual_keys_only === "boolean" ? body.allow_virtual_keys_only : undefined;
  const providerConfig = isPlainObject(body.provider_config) ? body.provider_config : undefined;
  const credentials = isPlainObject(body.credentials) ? body.credentials : undefined;

  return proxyToControlPlane({
    request,
    method: "PUT",
    path: `/api/v1/tenant-provider-accounts/${encodeURIComponent(normalizedAccountID)}/credentials`,
    body: {
      auth_mode: authMode,
      allow_virtual_keys_only: allowVirtualKeysOnly,
      provider_config: providerConfig,
      credentials,
    },
  });
}

