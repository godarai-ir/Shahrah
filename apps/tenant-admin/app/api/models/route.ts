import { NextRequest, NextResponse } from "next/server";

import { readAuthFromRequest } from "@/lib/auth/request-auth";
import { proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type CreateModelPayload = {
  model_name?: unknown;
  provider?: unknown;
  model?: unknown;
  api_base?: unknown;
  api_key?: unknown;
  api_version?: unknown;
  region_name?: unknown;
  rpm?: unknown;
  tpm?: unknown;
  weight?: unknown;
  timeout?: unknown;
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
    path: "/api/v1/models",
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

  const body = (await request.json().catch(() => null)) as CreateModelPayload | null;
  if (!body || typeof body !== "object") {
    return NextResponse.json(
      { error: { message: "بدنه درخواست نامعتبر است." } },
      { status: 400 },
    );
  }

  const modelName = asTrimmedString(body.model_name);
  const provider = asTrimmedString(body.provider);
  const model = asTrimmedString(body.model);
  const apiBase = asTrimmedString(body.api_base);
  const apiKey = asTrimmedString(body.api_key);
  const isInternalMock = provider.toLowerCase() === "mock";

  if (
    !modelName ||
    !provider ||
    !model ||
    (!isInternalMock && (!apiBase || !apiKey))
  ) {
    return NextResponse.json(
      {
        error: {
          message:
            "ارسال model_name، provider، model، api_base و api_key الزامی است.",
        },
      },
      { status: 400 },
    );
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/api/v1/models",
    body: {
      model_name: modelName,
      provider,
      model,
      api_base: isInternalMock ? "" : apiBase,
      api_key: isInternalMock ? "" : apiKey,
      api_version: asTrimmedString(body.api_version),
      region_name: asTrimmedString(body.region_name),
      rpm: asNumber(body.rpm),
      tpm: asNumber(body.tpm),
      weight: asNumber(body.weight),
      timeout: asNumber(body.timeout),
    },
  });
}

function asTrimmedString(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}

function asNumber(value: unknown): number | undefined {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string" && value.trim() !== "") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return undefined;
}
