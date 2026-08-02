type ProvisionedTenantPayload = {
  tenant?: { slug?: unknown };
  login_url?: unknown;
  [key: string]: unknown;
};

const tenantSlugPattern = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

export function normalizeOnboardingLoginURL(
  payload: unknown,
  tenantOriginTemplate: string,
) {
  if (!payload || typeof payload !== "object") {
    return payload;
  }

  const candidate = payload as ProvisionedTenantPayload;
  const tenantSlug = normalizeTenantSlug(candidate.tenant?.slug);
  if (
    !tenantSlug ||
    !tenantOriginTemplate.includes("{tenant}") ||
    typeof candidate.login_url !== "string"
  ) {
    return payload;
  }

  try {
    const upstreamLoginURL = new URL(candidate.login_url);
    const tenantOrigin = new URL(
      tenantOriginTemplate.replace("{tenant}", tenantSlug),
    ).origin;
    const loginURL = new URL(
      `${upstreamLoginURL.pathname}${upstreamLoginURL.search}`,
      tenantOrigin,
    );

    return {
      ...candidate,
      login_url: loginURL.toString(),
    };
  } catch {
    return payload;
  }
}

function normalizeTenantSlug(value: unknown) {
  if (typeof value !== "string") {
    return "";
  }
  const normalized = value.trim().toLowerCase();
  return tenantSlugPattern.test(normalized) ? normalized : "";
}
