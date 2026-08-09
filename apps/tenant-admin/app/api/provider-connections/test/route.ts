import { NextRequest, NextResponse } from "next/server";

import { proxyToControlPlane } from "@/lib/control-plane";
import { withRequestAuth } from "@/lib/http/server/route-handler";

export const dynamic = "force-dynamic";

type ProviderConnectionTestPayload = {
  provider?: unknown;
  auth_mode?: unknown;
  provider_config?: unknown;
  credentials?: unknown;
  use_platform_managed_credentials?: unknown;
  model_hint?: unknown;
};

export function POST(request: NextRequest) {
  return withRequestAuth(request, async (auth) => {
    const body = (await request.json().catch(() => null)) as ProviderConnectionTestPayload | null;
    if (!body || typeof body !== "object") {
      return NextResponse.json(
        { error: { message: "بدنه درخواست نامعتبر است." } },
        { status: 400 },
      );
    }

    const provider = asTrimmedString(body.provider).toLowerCase();
    const authMode = asTrimmedString(body.auth_mode).toLowerCase();
    if (!provider || !authMode) {
      return NextResponse.json(
        { error: { message: "ارسال provider و auth_mode الزامی است." } },
        { status: 400 },
      );
    }

    return proxyToControlPlane({
      request,
      auth,
      method: "POST",
      path: "/api/v1/provider-connections/test",
      body: {
        provider,
        auth_mode: authMode,
        provider_config: isPlainObject(body.provider_config) ? body.provider_config : {},
        credentials: isPlainObject(body.credentials) ? body.credentials : undefined,
        use_platform_managed_credentials: body.use_platform_managed_credentials === true,
        model_hint: asTrimmedString(body.model_hint),
      },
    });
  });
}

function asTrimmedString(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === "object" && !Array.isArray(value);
}
