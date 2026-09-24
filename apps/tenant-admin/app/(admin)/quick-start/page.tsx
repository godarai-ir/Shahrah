import AdminShell from "@/features/admin-shell/AdminShell";
import QuickStartManagement from "@/features/quick-start/QuickStartManagement";
import {
  type QuickStartResources,
} from "@/features/quick-start/journey";
import {
  countQuickStartResources,
  parseAllowedModelIDs,
  parseQuickStartResourceLabel,
  parseQuickStartEntitlements,
  parseQuickStartSubscription,
} from "@/features/quick-start/api";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";
import { headers } from "next/headers";
import { getTranslations } from "next-intl/server";
import { buildDocsBaseHref } from "@/lib/docs-href";

export const dynamic = "force-dynamic";

export default async function QuickStartPage() {
  const auth = await requireAuth("/login?next=/quick-start");
  const t = await getTranslations("quickStart");
  const tenantSlug = auth.tenantSlug;
  const requestHeaders = await headers();
  const permissionsPayload = await getCurrentMePermissions(tenantSlug, auth.user.id).catch(() => ({ permissions: [] as string[] }));
  const permissions = (permissionsPayload.permissions ?? []).map((item) => item.trim().toLowerCase());
  const docsBaseHref = buildDocsBaseHref(requestHeaders, tenantSlug);

  const [subscriptionResult, entitlementsResult, providerAccountsResult, patsResult, virtualUsersResult, teamsResult, orgUnitsResult, rateRulesResult, budgetRulesResult, guardrailsResult, allowedPatModelsResult, allowedVatModelsResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/subscription",
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/entitlements",
      userID: auth.user.id,
    }),
    canAnyPermission(permissions, ["provider_account:create", "provider_account:manage", "provider_account:read"])
      ? fetchControlPlaneJSON({
          tenantSlug,
          method: "GET",
          path: "/api/v1/tenant-provider-accounts",
          userID: auth.user.id,
        })
      : Promise.resolve(emptyResponse()),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/personal-access-tokens?limit=1&offset=0",
      userID: auth.user.id,
    }),
    canAnyPermission(permissions, ["virtual_account:create", "virtual_account:manage", "virtual_account:read"])
      ? fetchControlPlaneJSON({
          tenantSlug,
          method: "GET",
          path: "/api/v1/virtual-users?limit=1&offset=0",
          userID: auth.user.id,
        })
      : Promise.resolve(emptyResponse()),
    canAnyPermission(permissions, ["team:create", "team:manage", "team:read"])
      ? fetchControlPlaneJSON({
          tenantSlug,
          method: "GET",
          path: "/api/v1/teams?limit=1&offset=0",
          userID: auth.user.id,
        })
      : Promise.resolve(emptyResponse()),
    canAnyPermission(permissions, ["org_unit:create", "org_unit:manage", "org_unit:read", "org_unit:list"])
      ? fetchControlPlaneJSON({
          tenantSlug,
          method: "GET",
          path: "/api/v1/org-units?limit=1&offset=0",
          userID: auth.user.id,
        })
      : Promise.resolve(emptyResponse()),
    canAnyPermission(permissions, ["policy:list", "policy:manage"])
      ? fetchControlPlaneJSON({
          tenantSlug,
          method: "GET",
          path: "/api/v1/gateway-policies/rate-limit-rules?limit=1&offset=0",
          userID: auth.user.id,
        })
      : Promise.resolve(emptyResponse()),
    canAnyPermission(permissions, ["policy:list", "policy:manage"])
      ? fetchControlPlaneJSON({
          tenantSlug,
          method: "GET",
          path: "/api/v1/gateway-policies/budget-limit-rules?limit=1&offset=0",
          userID: auth.user.id,
        })
      : Promise.resolve(emptyResponse()),
    canAnyPermission(permissions, ["policy:list", "policy:manage"])
      ? fetchControlPlaneJSON({
          tenantSlug,
          method: "GET",
          path: "/api/v1/guardrails?limit=1&offset=0",
          userID: auth.user.id,
        })
      : Promise.resolve(emptyResponse()),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/models/allowed?key_type=pat",
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/models/allowed?key_type=vat",
      userID: auth.user.id,
    }),
  ]);

  const initialSubscription = subscriptionResult.ok ? parseQuickStartSubscription(subscriptionResult.payload) : null;
  const initialEntitlements = entitlementsResult.ok ? parseQuickStartEntitlements(entitlementsResult.payload) : null;

  const initialResources: QuickStartResources = {
    providerAccountCount: countQuickStartResources(providerAccountsResult.payload),
    personalAccessTokenCount: countQuickStartResources(patsResult.payload),
    virtualUserCount: countQuickStartResources(virtualUsersResult.payload),
    teamCount: countQuickStartResources(teamsResult.payload),
    orgUnitCount: countQuickStartResources(orgUnitsResult.payload),
    rateLimitRuleCount: countQuickStartResources(rateRulesResult.payload),
    budgetLimitRuleCount: countQuickStartResources(budgetRulesResult.payload),
    guardrailCount: countQuickStartResources(guardrailsResult.payload),
    allowedPatModels: parseAllowedModelIDs(allowedPatModelsResult.payload),
    allowedVatModels: parseAllowedModelIDs(allowedVatModelsResult.payload),
    providerAccountLabel: parseQuickStartResourceLabel(providerAccountsResult.payload),
    personalAccessTokenLabel: parseQuickStartResourceLabel(patsResult.payload),
    virtualUserLabel: parseQuickStartResourceLabel(virtualUsersResult.payload),
    teamLabel: parseQuickStartResourceLabel(teamsResult.payload),
    orgUnitLabel: parseQuickStartResourceLabel(orgUnitsResult.payload),
    rateLimitRuleLabel: parseQuickStartResourceLabel(rateRulesResult.payload),
    budgetLimitRuleLabel: parseQuickStartResourceLabel(budgetRulesResult.payload),
    guardrailLabel: parseQuickStartResourceLabel(guardrailsResult.payload),
  };

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["quickStart"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("subtitle"), helpHref: docsBaseHref }}
    >
      <QuickStartManagement
        initialSubscription={initialSubscription}
        initialEntitlements={initialEntitlements}
        initialPermissions={permissions}
        initialResources={initialResources}
        docsBaseHref={docsBaseHref}
      />
    </AdminShell>
  );
}

function canAnyPermission(permissions: string[], required: string[]): boolean {
  return required.some((permission) => permissions.includes(permission.trim().toLowerCase()));
}

function emptyResponse() {
  return {
    ok: true,
    status: 200,
    payload: {},
  };
}
