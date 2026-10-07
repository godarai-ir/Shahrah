/**
 * Server-side client for the control-plane OIDC login bridge.
 *
 * A relying party such as LibreChat sends the browser to the control-plane authorize
 * endpoint, which records the request and redirects here with `oidc_request`. Once the
 * member has a tenant-admin session, this server accepts the request for that session and
 * the control plane answers with the relying party's redirect carrying the code.
 */

import { z } from "zod";

import { fetchControlPlaneJSON } from "@/lib/control-plane";

export const OIDC_REQUEST_PARAM = "oidc_request";
export const OIDC_CONTINUE_PATH = "/auth/oidc/continue";

const OIDC_REQUEST_ID_PATTERN = /^[A-Za-z0-9_-]{32,128}$/;

const loginRequestSchema = z.object({
  id: z.string(),
  client_name: z.string(),
  prompt: z.string().optional(),
});

const decisionSchema = z.object({ redirect_to: z.string().url() });

const bridgeErrorSchema = z.object({
  error: z.object({ code: z.string() }),
});

export type OIDCLoginRequest = {
  id: string;
  clientName: string;
  prompt: "none" | "login" | "";
};

export type OIDCBridgeErrorCode = "expired" | "session_invalid" | "unavailable";

export type OIDCBridgeResult<T> =
  | { ok: true; value: T }
  | { ok: false; code: OIDCBridgeErrorCode };

export function normalizeOIDCRequestID(value: string | null | undefined): string | null {
  const normalized = value?.trim() ?? "";
  return OIDC_REQUEST_ID_PATTERN.test(normalized) ? normalized : null;
}

export function oidcContinuePath(requestID: string): string {
  return `${OIDC_CONTINUE_PATH}?${new URLSearchParams({ [OIDC_REQUEST_PARAM]: requestID })}`;
}

function bridgeFailure(status: number, payload: unknown): OIDCBridgeErrorCode {
  const parsed = bridgeErrorSchema.safeParse(payload);
  const code = parsed.success ? parsed.data.error.code : "";

  if (code === "oidc_session_invalid" || status === 401) return "session_invalid";
  if (code === "oidc_request_not_found" || code === "oidc_request_expired" || status === 404 || status === 410) {
    return "expired";
  }
  return "unavailable";
}

function bridgePath(requestID: string, action?: "accept" | "reject"): string {
  const base = `/api/v1/auth/oidc/requests/${encodeURIComponent(requestID)}`;
  return action ? `${base}/${action}` : base;
}

export async function getOIDCLoginRequest(
  tenantSlug: string,
  requestID: string,
): Promise<OIDCBridgeResult<OIDCLoginRequest>> {
  const result = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: `${bridgePath(requestID)}?${new URLSearchParams({ tenant_slug: tenantSlug })}`,
  });
  if (!result.ok) return { ok: false, code: bridgeFailure(result.status, result.payload) };

  const parsed = loginRequestSchema.safeParse(result.payload);
  if (!parsed.success) return { ok: false, code: "unavailable" };

  const prompt = parsed.data.prompt === "none" || parsed.data.prompt === "login" ? parsed.data.prompt : "";
  return { ok: true, value: { id: parsed.data.id, clientName: parsed.data.client_name, prompt } };
}

async function decide(
  tenantSlug: string,
  requestID: string,
  action: "accept" | "reject",
  sessionID?: string,
): Promise<OIDCBridgeResult<string>> {
  const result = await fetchControlPlaneJSON({
    tenantSlug,
    method: "POST",
    path: bridgePath(requestID, action),
    body: { tenant_slug: tenantSlug, session_id: sessionID ?? "" },
  });
  if (!result.ok) return { ok: false, code: bridgeFailure(result.status, result.payload) };

  const parsed = decisionSchema.safeParse(result.payload);
  if (!parsed.success) return { ok: false, code: "unavailable" };

  return { ok: true, value: parsed.data.redirect_to };
}

/** Accepts the request for the member who owns sessionID and returns the client redirect. */
export function acceptOIDCLogin(tenantSlug: string, requestID: string, sessionID: string) {
  return decide(tenantSlug, requestID, "accept", sessionID);
}

/** Ends the request without signing in and returns the client's error redirect. */
export function rejectOIDCLogin(tenantSlug: string, requestID: string) {
  return decide(tenantSlug, requestID, "reject");
}
