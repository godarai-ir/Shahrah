import { NextRequest, NextResponse } from "next/server";

import {
  authorizeAPIRequest,
  PLAYGROUND_VIDEO_MUTATE_PERMISSIONS,
  PLAYGROUND_VIDEO_READ_PERMISSIONS,
} from "@/lib/auth/api-permissions";
import { fetchControlPlaneJSON, proxyToControlPlane } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

const VIDEO_FIELDS = new Set([
  "model", "prompt", "seconds", "resolution", "size", "aspect_ratio", "n",
  "service_tier", "seed", "input_reference",
]);

export async function GET(request: NextRequest) {
  const authorization = await authorizeAPIRequest(request, PLAYGROUND_VIDEO_READ_PERMISSIONS);
  if (!authorization.ok) return authorization.response;

  return proxyToControlPlane({
    request,
    method: "GET",
    path: `/v1/videos${request.nextUrl.search}`,
  });
}

export async function POST(request: NextRequest) {
  const authorization = await authorizeAPIRequest(request, PLAYGROUND_VIDEO_MUTATE_PERMISSIONS);
  if (!authorization.ok) return authorization.response;
  const { auth } = authorization;

  const form = await request.formData().catch(() => null);
  if (!form) return NextResponse.json({ error: { message: "Invalid video request." } }, { status: 400 });
  const unknownFields = [...new Set([...form.keys()].filter((key) => !VIDEO_FIELDS.has(key)))];
  if (unknownFields.length > 0) {
    return NextResponse.json({ error: { message: `Unsupported video fields: ${unknownFields.join(", ")}` } }, { status: 400 });
  }

  const validationError = validateVideoForm(form);
  if (validationError) return NextResponse.json({ error: { message: validationError } }, { status: 400 });
  const model = String(form.get("model") ?? "");
  const capability = await validateVideoModelCapability(request, auth.tenantSlug, auth.user.id, model);
  if (!capability.ok) {
    return NextResponse.json({ error: { message: capability.message } }, { status: capability.status });
  }

  return proxyToControlPlane({
    request,
    method: "POST",
    path: "/v1/videos",
    body: form,
    userID: auth.user.id,
    upstreamHeaders: request.headers.get("idempotency-key")
      ? { "Idempotency-Key": request.headers.get("idempotency-key")! }
      : undefined,
  });
}

function validateVideoForm(form: FormData): string | null {
  const model = form.get("model");
  const prompt = form.get("prompt");
  if (typeof model !== "string" || !model.trim() || model.length > 512) return "A model is required.";
  if (typeof prompt !== "string" || !prompt.trim() || prompt.length > 100_000) return "A bounded video prompt is required.";

  const integer = (name: string, min: number, max: number, optional = false) => {
    const raw = form.get(name);
    if ((raw === null || raw === "") && optional) return true;
    if (typeof raw !== "string" || !/^[0-9]+$/.test(raw)) return false;
    const value = Number(raw);
    return Number.isSafeInteger(value) && value >= min && value <= max;
  };
  if (!integer("seconds", 1, 120)) return "seconds must be an integer between 1 and 120.";
  if (!integer("n", 1, 4)) return "n must be an integer between 1 and 4.";
  if (!integer("seed", 0, Number.MAX_SAFE_INTEGER, true)) return "seed must be a non-negative integer.";

  const resolution = form.get("resolution");
  if (typeof resolution !== "string" || !["540p", "720p", "1080p"].includes(resolution)) return "resolution is invalid.";
  const aspectRatio = form.get("aspect_ratio");
  if (typeof aspectRatio !== "string" || !["16:9", "9:16", "1:1"].includes(aspectRatio)) return "aspect_ratio is invalid.";
  const serviceTier = form.get("service_tier");
  if (typeof serviceTier !== "string" || !["standard", "priority"].includes(serviceTier)) return "service_tier is invalid.";

  const input = form.get("input_reference");
  if (input instanceof File) {
    if (input.size > 20 * 1024 * 1024) return "The input image exceeds 20 MiB.";
    if (!["image/png", "image/jpeg", "image/webp"].includes(input.type)) return "The input reference must be a PNG, JPEG, or WebP image.";
  }
  return null;
}

async function validateVideoModelCapability(request: NextRequest, tenantSlug: string, userID: string, modelID: string) {
  const upstream = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v1/models",
    userID,
    requestHeaders: request.headers,
  });
  if (!upstream.ok) return { ok: false as const, status: 502, message: "The model capability list could not be verified." };
  const payload = isRecord(upstream.payload) ? upstream.payload : {};
  const models = Array.isArray(payload.data) ? payload.data : [];
  const model = models.find((entry) => isRecord(entry) && entry.id === modelID);
  if (!isRecord(model)) return { ok: false as const, status: 400, message: "This model is not available in the current workspace." };
  const modes = Array.isArray(model.supported_modes) ? model.supported_modes : Array.isArray(model.types) ? model.types : [];
  if (!modes.some((value) => typeof value === "string" && value.trim().toLowerCase() === "video_generation")) {
    return { ok: false as const, status: 400, message: "This model is not enabled for video generation in the current workspace." };
  }
  return { ok: true as const };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
