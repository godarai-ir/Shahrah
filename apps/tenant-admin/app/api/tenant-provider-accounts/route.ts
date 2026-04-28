import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type CreateAccountPayload = {
  provider?: unknown;
  name?: unknown;
  auth_mode?: unknown;
  provider_config?: unknown;
  credentials?: unknown;
  allow_virtual_keys_only?: unknown;
};

export async function GET(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  return proxyToControlPlane({
    request,
    method: "GET",
    path: "/api/v1/tenant-provider-accounts",
  });
}

export async function POST(request: NextRequest) {
  const auth = await readAuthFromRequest(request);
  if (!auth.ok) {
    return NextResponse.json(
      { error: { message: auth.message } },
      { status: auth.status },
    );
  }

  const body = (await request.json().catch(() => null)) as CreateAccountPayload | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "بدنه درخواست نامعتبر است." } },
      { status: 400 },
    );
  }

  const provider = asTrimmedString(body.provider).toLowerCase();
  const name = asTrimmedString(body.name);
  const authMode = asTrimmedString(body.auth_mode);
  if (!provider || !name || !authMode) {
    return NextResponse.json(
      { error: { message: "ارسال provider، name و auth_mode الزامی است." } },
      { status: 400 },
    );
  }

  const providerConfig = isPlainObject(body.provider_config) ? body.provider_config : {};
  const credentials = isPlainObject(body.credentials) ? body.credentials : undefined;
  const allowVirtualKeysOnly =
    typeof body.allow_virtual_keys_only === "boolean"
      ? body.allow_virtual_keys_only
      : false;

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/tenant-provider-accounts",
    body: {
      provider,
      name,
      auth_mode: authMode,
      provider_config: providerConfig,
      credentials,
      allow_virtual_keys_only: allowVirtualKeysOnly,
    },
  });
}

function asTrimmedString(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === "object" && !Array.isArray(value);
}
