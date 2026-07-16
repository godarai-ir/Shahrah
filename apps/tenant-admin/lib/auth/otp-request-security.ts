import { createHash, randomBytes } from "crypto";

import type { CreateOTPResult, NormalizedIdentifier } from "./types";

const OTP_TTL_SECONDS = 5 * 60;
const IDENTIFIER_RATE_LIMIT = 5;
const IDENTIFIER_RATE_LIMIT_WINDOW_MS = 15 * 60 * 1000;
const IP_RATE_LIMIT = 30;
const IP_RATE_LIMIT_WINDOW_MS = 60 * 1000;
const MAX_RATE_LIMIT_BUCKETS = 10_000;

type RateLimitBucket = {
  count: number;
  resetAt: number;
};

type RateLimitPolicy = {
  key: string;
  limit: number;
  windowMs: number;
};

type RateLimitDecision =
  | { ok: true }
  | { ok: false; retryAfterSeconds: number };

const buckets = new Map<string, RateLimitBucket>();

export function checkOTPRequestRateLimit(input: {
  tenantSlug: string;
  identifier: NormalizedIdentifier;
  clientAddress: string;
}): RateLimitDecision {
  const policies: RateLimitPolicy[] = [
    {
      key: rateLimitKey(
        "identifier",
        input.tenantSlug,
        input.identifier.channel,
        input.identifier.normalized,
      ),
      limit: IDENTIFIER_RATE_LIMIT,
      windowMs: IDENTIFIER_RATE_LIMIT_WINDOW_MS,
    },
    {
      key: rateLimitKey("ip", input.tenantSlug, input.clientAddress),
      limit: IP_RATE_LIMIT,
      windowMs: IP_RATE_LIMIT_WINDOW_MS,
    },
  ];

  let retryAfterSeconds = 0;
  for (const policy of policies) {
    const decision = consumeRateLimit(policy);
    if (!decision.ok) {
      retryAfterSeconds = Math.max(
        retryAfterSeconds,
        decision.retryAfterSeconds,
      );
    }
  }

  if (retryAfterSeconds > 0) {
    return { ok: false, retryAfterSeconds };
  }

  return { ok: true };
}

export function buildGenericOTPRequestResult(
  identifier: NormalizedIdentifier,
): CreateOTPResult {
  return {
    request_id: randomBytes(16).toString("hex"),
    channel: identifier.channel,
    destination_masked: maskOTPIdentifier(identifier),
    expires_in_seconds: OTP_TTL_SECONDS,
  };
}

export function clientAddressFromHeaders(headers: Headers): string {
  const forwardedFor = headers.get("x-forwarded-for")?.split(",")[0]?.trim();
  if (forwardedFor) {
    return forwardedFor;
  }

  return headers.get("x-real-ip")?.trim() || "unknown";
}

export function isAccountEnumerationAuthError(error: unknown): boolean {
  if (!(error instanceof Error)) {
    return false;
  }

  const message = error.message.toLowerCase();
  const status = extractAuthAPIStatus(error);
  if (status !== 403 && status !== 404) {
    return false;
  }

  return (
    message.includes("tenant member not found") ||
    message.includes("not found") ||
    message.includes("do not have access") ||
    message.includes("access to this tenant")
  );
}

export function extractAuthAPIStatus(error: unknown): number | null {
  if (!(error instanceof Error)) {
    return null;
  }

  const statusMatch = error.message.match(/Auth API error:\s+(\d{3})/);
  return statusMatch ? Number(statusMatch[1]) : null;
}

function consumeRateLimit(policy: RateLimitPolicy): RateLimitDecision {
  const now = Date.now();
  pruneExpiredBuckets(now);
  const existing = buckets.get(policy.key);

  if (!existing || existing.resetAt <= now) {
    buckets.set(policy.key, {
      count: 1,
      resetAt: now + policy.windowMs,
    });
    return { ok: true };
  }

  if (existing.count >= policy.limit) {
    return {
      ok: false,
      retryAfterSeconds: Math.max(
        1,
        Math.ceil((existing.resetAt - now) / 1000),
      ),
    };
  }

  existing.count += 1;
  buckets.set(policy.key, existing);
  return { ok: true };
}

function pruneExpiredBuckets(now: number) {
  if (buckets.size <= MAX_RATE_LIMIT_BUCKETS) {
    return;
  }

  for (const [key, bucket] of buckets.entries()) {
    if (bucket.resetAt <= now) {
      buckets.delete(key);
    }
  }

  for (const key of buckets.keys()) {
    if (buckets.size <= MAX_RATE_LIMIT_BUCKETS) {
      break;
    }
    buckets.delete(key);
  }
}

function rateLimitKey(...parts: string[]): string {
  return createHash("sha256").update(parts.join("\0")).digest("hex");
}

function maskOTPIdentifier(identifier: NormalizedIdentifier): string {
  if (identifier.channel === "email") {
    const [local = "", domain = ""] = identifier.normalized.split("@");
    if (!local) {
      return `***@${domain}`;
    }
    if (local.length <= 2) {
      return `${local[0] ?? "*"}***@${domain}`;
    }
    return `${local.slice(0, 2)}***@${domain}`;
  }

  const compact = identifier.normalized.replace(/\s+/g, "");
  if (compact.length <= 4) {
    return "***";
  }
  return `${compact.slice(0, 3)}****${compact.slice(-3)}`;
}
