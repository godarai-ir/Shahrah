import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type CreateKeyPayload = {
  key_alias?: unknown;
  models?: unknown;
  role?: unknown;
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
    path: "/api/v1/keys",
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

  const body = (await request.json().catch(() => null)) as CreateKeyPayload | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "بدنه درخواست نامعتبر است." } },
      { status: 400 },
    );
  }

  const keyAlias =
    typeof body.key_alias === "string" ? body.key_alias.trim() : "";
  const models = Array.isArray(body.models)
    ? body.models
        .filter((item): item is string => typeof item === "string")
        .map((item) => item.trim())
        .filter(Boolean)
    : [];
  const role = typeof body.role === "string" ? body.role.trim() : "virtual_key";

  if (!keyAlias) {
    return NextResponse.json(
      { error: { message: "ارسال key_alias الزامی است." } },
      { status: 400 },
    );
  }
  if (models.length === 0) {
    return NextResponse.json(
      { error: { message: "حداقل یک مدل وارد کنید." } },
      { status: 400 },
    );
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/keys",
    body: {
      key_alias: keyAlias,
      models,
      role: role || "virtual_key",
    },
  });
}
