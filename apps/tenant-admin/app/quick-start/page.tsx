import AdminShell from "@/app/components/admin-shell";
import QuickStartManagement from "@/app/components/quick-start-management";
import {
  type QuickStartEntitlements,
  type QuickStartResources,
  type QuickStartSubscription,
} from "@/app/components/quick-start-journey";
import { getCurrentMePermissions } from "@/lib/auth/me-permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

export const dynamic = "force-dynamic";

type DataList<T> = {
  data?: T[];
  total?: number;
};

type AllowedModel = {
  id: string;
};

export default async function QuickStartPage() {
  const auth = await requireAuth("/login?next=/quick-start");
  const tenantSlug = auth.tenantSlug;
  const permissionsPayload = await getCurrentMePermissions(tenantSlug, auth.user.id).catch(() => ({ permissions: [] as string[] }));
  const permissions = (permissionsPayload.permissions ?? []).map((item) => item.trim().toLowerCase());

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

  const initialSubscription = subscriptionResult.ok ? (subscriptionResult.payload as QuickStartSubscription) : null;
  const initialEntitlements = entitlementsResult.ok ? (entitlementsResult.payload as QuickStartEntitlements) : null;

  const initialResources: QuickStartResources = {
    providerAccountCount: countList(providerAccountsResult.payload),
    personalAccessTokenCount: countList(patsResult.payload),
    virtualUserCount: countList(virtualUsersResult.payload),
    teamCount: countList(teamsResult.payload),
    orgUnitCount: countList(orgUnitsResult.payload),
    rateLimitRuleCount: countList(rateRulesResult.payload),
    budgetLimitRuleCount: countList(budgetRulesResult.payload),
    guardrailCount: countList(guardrailsResult.payload),
    allowedPatModels: allowedModelIDs(allowedPatModelsResult.payload),
    allowedVatModels: allowedModelIDs(allowedVatModelsResult.payload),
  };

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <QuickStartManagement
        initialSubscription={initialSubscription}
        initialEntitlements={initialEntitlements}
        initialPermissions={permissions}
        initialResources={initialResources}
      />
    </AdminShell>
  );
}

function canAnyPermission(permissions: string[], required: string[]): boolean {
  return required.some((permission) => permissions.includes(permission.trim().toLowerCase()));
}

function countList(payload: unknown): number {
  if (!payload || typeof payload !== "object") return 0;
  const candidate = payload as DataList<unknown>;
  if (typeof candidate.total === "number" && Number.isFinite(candidate.total)) {
    return candidate.total;
  }
  if (Array.isArray(candidate.data)) {
    return candidate.data.length;
  }
  return 0;
}

function allowedModelIDs(payload: unknown): string[] {
  if (!payload || typeof payload !== "object") return [];
  const candidate = payload as DataList<AllowedModel>;
  if (!Array.isArray(candidate.data)) return [];
  return candidate.data.map((item) => item.id?.trim()).filter(Boolean);
}

function emptyResponse() {
  return {
    ok: true,
    status: 200,
    payload: {},
  };
}
