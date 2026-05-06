import AdminShell from "@/app/components/admin-shell";
import ModelMetricsDashboard from "@/app/components/model-metrics-dashboard";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

export default async function MetricsPage() {
  const auth = await requireAuth("/login?next=/metrics");
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell tenantSlug={auth.tenantSlug} currentUser={currentUser}>
      <ModelMetricsDashboard />
    </AdminShell>
  );
}
