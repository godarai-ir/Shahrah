import AdminShell from "@/features/admin-shell/AdminShell";
import { getTranslations } from "next-intl/server";
import ModelMetricsDashboard from "@/features/metrics/ModelMetricsDashboard";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

/** Filters live in the URL and are read by the dashboard itself; see lib/observe-filters. */
export default async function MetricsPage() {
  const auth = await requireAuth("/login?next=/metrics");
  const t = await getTranslations("metrics");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["settings:list"]);
  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  return (
    <AdminShell messageNamespaces={["metrics"]}
      tenantSlug={auth.tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: t("description"), helpHref: "/docs/observability-model-metrics" }}
    >
      <ModelMetricsDashboard />
    </AdminShell>
  );
}
