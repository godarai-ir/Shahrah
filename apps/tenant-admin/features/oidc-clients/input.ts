/**
 * Validation for OIDC client registrations, shared by the form and the route handlers.
 * It mirrors the control-plane rules so most mistakes are reported before a round trip;
 * the control plane stays authoritative.
 */

export const MAX_CLIENT_NAME_LENGTH = 120;
export const MAX_REDIRECT_URIS = 10;

/** Product labels a client can report its gateway usage under (see the credit pricing PRD). */
export const OIDC_CLIENT_PRODUCTS = ["oidc_app", "librechat"] as const;
export type OIDCClientProduct = (typeof OIDC_CLIENT_PRODUCTS)[number];
export const DEFAULT_OIDC_CLIENT_PRODUCT: OIDCClientProduct = "oidc_app";
const MAX_REDIRECT_URI_LENGTH = 2048;

export type OIDCClientInputErrorCode =
  | "invalid_payload"
  | "name_required"
  | "name_too_long"
  | "redirect_required"
  | "redirect_invalid"
  | "redirect_too_many"
  | "product_invalid";

export type OIDCClientInput = {
  name?: string;
  redirect_uris?: string[];
  post_logout_redirect_uris?: string[];
  credit_product?: OIDCClientProduct;
  active?: boolean;
};

export type OIDCClientInputResult =
  | { ok: true; value: OIDCClientInput }
  | { ok: false; code: OIDCClientInputErrorCode; message: string; field?: keyof OIDCClientInput; value?: string };

const MESSAGES: Record<OIDCClientInputErrorCode, string> = {
  invalid_payload: "Invalid request payload.",
  name_required: "The name field is required.",
  name_too_long: `The name must be at most ${MAX_CLIENT_NAME_LENGTH} characters.`,
  redirect_required: "At least one redirect URI is required.",
  redirect_invalid: "Redirect URIs must be absolute https URLs without credentials or fragments.",
  redirect_too_many: `At most ${MAX_REDIRECT_URIS} URIs are allowed.`,
  product_invalid: `The product must be one of: ${OIDC_CLIENT_PRODUCTS.join(", ")}.`,
};

function isOIDCClientProduct(value: unknown): value is OIDCClientProduct {
  return typeof value === "string" && (OIDC_CLIENT_PRODUCTS as readonly string[]).includes(value);
}

type OIDCClientInputFailure = Extract<OIDCClientInputResult, { ok: false }>;

function failure(code: OIDCClientInputErrorCode, field?: keyof OIDCClientInput, value?: string): OIDCClientInputFailure {
  return { ok: false, code, message: MESSAGES[code], field, value };
}

/** Splits a textarea value into unique, non-empty lines. */
export function splitURIList(raw: string): string[] {
  const seen = new Set<string>();
  for (const line of raw.split(/\r?\n/)) {
    const value = line.trim();
    if (value) seen.add(value);
  }
  return [...seen];
}

function isDevelopmentHost(hostname: string): boolean {
  const host = hostname.toLowerCase().replace(/^\[|\]$/g, "");
  return host === "localhost" || host.endsWith(".localhost") || host === "::1" || /^127\.\d+\.\d+\.\d+$/.test(host);
}

export function isValidRedirectURI(value: string): boolean {
  if (value.length > MAX_REDIRECT_URI_LENGTH || value.includes("#")) return false;

  let parsed: URL;
  try {
    parsed = new URL(value);
  } catch {
    return false;
  }

  if (!parsed.hostname || parsed.username || parsed.password) return false;
  if (parsed.protocol === "https:") return true;
  return parsed.protocol === "http:" && isDevelopmentHost(parsed.hostname);
}

function readURIList(
  raw: unknown,
  field: "redirect_uris" | "post_logout_redirect_uris",
  required: boolean,
): { ok: true; values: string[] } | OIDCClientInputFailure {
  if (!Array.isArray(raw) || raw.some((item) => typeof item !== "string")) {
    return failure("invalid_payload", field);
  }

  const values = [...new Set((raw as string[]).map((item) => item.trim()).filter(Boolean))];
  if (required && values.length === 0) return failure("redirect_required", field);
  if (values.length > MAX_REDIRECT_URIS) return failure("redirect_too_many", field);

  const invalid = values.find((value) => !isValidRedirectURI(value));
  if (invalid) return failure("redirect_invalid", field, invalid);

  return { ok: true, values };
}

/**
 * Normalizes a create (`partial: false`) or update (`partial: true`) payload. Updates only
 * carry the fields that are present.
 */
export function parseOIDCClientInput(payload: unknown, options: { partial: boolean }): OIDCClientInputResult {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) return failure("invalid_payload");

  const body = payload as Record<string, unknown>;
  const value: OIDCClientInput = {};

  if (!options.partial || "name" in body) {
    if (typeof body.name !== "string" || !body.name.trim()) return failure("name_required", "name");
    const name = body.name.trim();
    if ([...name].length > MAX_CLIENT_NAME_LENGTH) return failure("name_too_long", "name");
    value.name = name;
  }

  if (!options.partial || "redirect_uris" in body) {
    const redirects = readURIList(body.redirect_uris, "redirect_uris", true);
    if (!redirects.ok) return redirects;
    value.redirect_uris = redirects.values;
  }

  if ("post_logout_redirect_uris" in body) {
    const logouts = readURIList(body.post_logout_redirect_uris, "post_logout_redirect_uris", false);
    if (!logouts.ok) return logouts;
    value.post_logout_redirect_uris = logouts.values;
  } else if (!options.partial) {
    value.post_logout_redirect_uris = [];
  }

  if ("credit_product" in body) {
    if (!isOIDCClientProduct(body.credit_product)) return failure("product_invalid", "credit_product");
    value.credit_product = body.credit_product;
  }

  if (options.partial && "active" in body) {
    if (typeof body.active !== "boolean") return failure("invalid_payload", "active");
    value.active = body.active;
  }

  if (options.partial && Object.keys(value).length === 0) return failure("invalid_payload");

  return { ok: true, value };
}
