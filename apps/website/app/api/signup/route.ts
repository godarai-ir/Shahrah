import { NextResponse } from "next/server";

import {
  insertSignupSubmission,
  type SignupSubmission,
} from "@/app/lib/signup-db";

const TEAM_SIZE_MAX = 64;
const USAGE_TYPE_MAX = 128;
const ACTIVITY_FIELD_MAX = 128;
const MAX_FEATURES = 10;
const RATE_LIMIT_WINDOW_MS = 10 * 60 * 1000;
const RATE_LIMIT_MAX_REQUESTS = 20;

const allowedFeatures = new Set([
  "Observability",
  "AI Gateway",
  "Guardrails",
  "Prompt Engineering",
  "Agents",
  "Security and Compliance",
  "MCP Gateway",
]);

const requestLog = new Map<string, number[]>();

function getClientIp(req: Request): string {
  const forwardedFor = req.headers.get("x-forwarded-for");
  if (forwardedFor) {
    return forwardedFor.split(",")[0]?.trim() || "unknown";
  }
  return req.headers.get("x-real-ip") || "unknown";
}

function isRateLimited(ip: string): boolean {
  const now = Date.now();
  const start = now - RATE_LIMIT_WINDOW_MS;
  const entries = requestLog.get(ip) ?? [];
  const recent = entries.filter((ts) => ts >= start);

  if (recent.length >= RATE_LIMIT_MAX_REQUESTS) {
    requestLog.set(ip, recent);
    return true;
  }

  recent.push(now);
  requestLog.set(ip, recent);
  return false;
}

function isValidMobile(mobile: string): boolean {
  return /^09\d{9}$/.test(mobile);
}

function isValidSubmission(input: Partial<SignupSubmission>): input is SignupSubmission {
  return (
    typeof input.mobile === "string" &&
    isValidMobile(input.mobile.trim()) &&
    typeof input.teamSize === "string" &&
    input.teamSize.trim().length > 0 &&
    input.teamSize.trim().length <= TEAM_SIZE_MAX &&
    typeof input.usageType === "string" &&
    input.usageType.trim().length > 0 &&
    input.usageType.trim().length <= USAGE_TYPE_MAX &&
    typeof input.activityField === "string" &&
    input.activityField.trim().length > 0 &&
    input.activityField.trim().length <= ACTIVITY_FIELD_MAX &&
    Array.isArray(input.features) &&
    input.features.length > 0 &&
    input.features.length <= MAX_FEATURES &&
    input.features.every(
      (feature) =>
        typeof feature === "string" && allowedFeatures.has(feature),
    )
  );
}

export async function POST(req: Request) {
  try {
    const clientIp = getClientIp(req);
    if (isRateLimited(clientIp)) {
      return NextResponse.json(
        { ok: false, error: "Too many requests" },
        {
          status: 429,
          headers: { "retry-after": String(Math.floor(RATE_LIMIT_WINDOW_MS / 1000)) },
        },
      );
    }

    const body = (await req.json()) as Partial<SignupSubmission>;
    if (!isValidSubmission(body)) {
      return NextResponse.json(
        { ok: false, error: "Invalid payload" },
        { status: 400 },
      );
    }

    const id = await insertSignupSubmission({
      mobile: body.mobile.trim(),
      teamSize: body.teamSize.trim(),
      usageType: body.usageType.trim(),
      activityField: body.activityField.trim(),
      features: body.features.map((v) => String(v)),
    });

    return NextResponse.json({ ok: true, id });
  } catch (error) {
    return NextResponse.json(
      {
        ok: false,
        error: "Failed to save submission",
      },
      { status: 500 },
    );
  }
}
