import AdminShell from "@/app/components/admin-shell";
import { getTranslations } from "next-intl/server";
import ModelMetricsDashboard from "@/app/components/model-metrics-dashboard";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

export const dynamic = "force-dynamic";

type MetricsPageProps = {
  searchParams?: Promise<{
    from?: string;
    to?: string;
    model?: string;
    provider?: string;
    request_id?: string;
  }>;
};

export default async function MetricsPage({ searchParams }: MetricsPageProps) {
  const auth = await requireAuth("/login?next=/metrics");
  const t = await getTranslations("metrics");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["settings:list"]);
  const params = (await searchParams) ?? {};
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
      <ModelMetricsDashboard
        initialFrom={params.from}
        initialTo={params.to}
        initialModel={params.model}
        initialProvider={params.provider}
        initialRequestID={params.request_id}
      />
    </AdminShell>
  );
}
