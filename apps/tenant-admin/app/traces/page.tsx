import AdminShell from "@/app/components/admin-shell";
import TraceManagement from "@/app/components/trace-management";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type RequestTrace = {
  trace_id: string;
  tenant_id: string;
  request_id?: string;
  model?: string;
  provider?: string;
  route: string;
  status_code: number;
  started_at: string;
  duration_ms: number;
  error_code?: string;
  sampled: boolean;
  attempts?: number;
};

type TraceListResponse = {
  data?: RequestTrace[];
  total?: number;
};

export const dynamic = "force-dynamic";

export default async function TracesPage() {
  const auth = await requireAuth("/login?next=/traces");
  const tenantSlug = auth.tenantSlug;

  const tracesResult = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v2/traces?limit=25&offset=0",
    userID: auth.user.id,
  });

  const payload = (tracesResult.payload ?? {}) as TraceListResponse;
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <TraceManagement
        initialData={tracesResult.ok ? payload.data ?? [] : []}
        initialTotal={tracesResult.ok ? payload.total ?? 0 : 0}
        initialError={tracesResult.ok ? null : tracesResult.errorMessage ?? "Request failed."}
      />
    </AdminShell>
  );
}
