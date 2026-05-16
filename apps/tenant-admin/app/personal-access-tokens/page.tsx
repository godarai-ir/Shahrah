import AdminShell from "@/app/components/admin-shell";
import PersonalAccessTokenManagement from "@/app/components/personal-access-token-management";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type CreatedKey = {
  id: string;
  key_alias: string;
  key_name?: string;
  key_type?: string;
  models?: string[];
  created_at?: string;
  expires_at?: string;
  last_used_at?: string;
  revoked_at?: string;
};

type PATListResponse = {
  data?: CreatedKey[];
  total?: number;
};

type ModelListResponse = {
  data?: Array<{ id: string }>;
};

export const dynamic = "force-dynamic";

export default async function PersonalAccessTokensPage() {
  const auth = await requireAuth("/login?next=/personal-access-tokens");
  const tenantSlug = auth.tenantSlug;
  await requirePermissionOr403(tenantSlug, auth.user.id, ["settings:list"]);

  const [patsResult, modelsResult] = await Promise.all([
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/personal-access-tokens?limit=25&offset=0",
      userID: auth.user.id,
    }),
    fetchControlPlaneJSON({
      tenantSlug,
      method: "GET",
      path: "/api/v1/models/allowed?key_type=pat",
      userID: auth.user.id,
    }),
  ]);

  const patPayload = (patsResult.payload ?? {}) as PATListResponse;
  const modelPayload = (modelsResult.payload ?? {}) as ModelListResponse;

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <PersonalAccessTokenManagement
        initialTokens={patsResult.ok ? patPayload.data ?? [] : []}
        initialTotal={patsResult.ok ? patPayload.total ?? 0 : 0}
        initialModels={(modelsResult.ok ? modelPayload.data ?? [] : []).map((m) => m.id)}
        initialError={patsResult.ok ? null : patsResult.errorMessage ?? "Request failed."}
      />
    </AdminShell>
  );
}
