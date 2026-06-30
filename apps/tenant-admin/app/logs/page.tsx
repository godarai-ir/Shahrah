import Link from "next/link";
import { getTranslations } from "next-intl/server";

import AdminShell from "@/app/components/admin-shell";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type RequestLog = {
  log_id: string;
  request_id: string;
  trace_id?: string;
  provider?: string;
  model?: string;
  route?: string;
  status?: string;
  status_code: number;
  started_at: string;
  duration_ms: number;
  total_tokens?: number;
  cost_usd?: number;
  cache_status?: string;
  fallback_status?: string;
};

type ResponsePayload = {
  data?: RequestLog[];
};

export const dynamic = "force-dynamic";

export default async function LogsPage() {
  const auth = await requireAuth("/login?next=/logs");
  const tenantSlug = auth.tenantSlug;
  const t = await getTranslations("logs");
  await requirePermissionOr403(tenantSlug, auth.user.id, ["logs:read_metadata", "logs:read_content"]);

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: "/api/v2/logs?limit=25&offset=0",
    userID: auth.user.id,
  });

  const payload = (result.payload ?? {}) as ResponsePayload;
  const logs = result.ok ? payload.data ?? [] : [];
  const currentUser = { id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role };

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <section className="space-y-4 rounded-2xl border border-border bg-card p-5">
        <div className="space-y-1">
          <h2 className="text-xl font-semibold">{t("title")}</h2>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>

        {!result.ok ? (
          <p className="rounded-xl border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">
            {result.errorMessage ?? t("loadError")}
          </p>
        ) : logs.length === 0 ? (
          <p className="rounded-xl border border-dashed border-border p-6 text-sm text-muted-foreground">{t("empty")}</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full text-sm">
              <thead>
                <tr className="border-b border-border text-muted-foreground">
                  <th className="px-3 py-2 text-right">{t("summary.status")}</th>
                  <th className="px-3 py-2 text-right">{t("summary.provider")}</th>
                  <th className="px-3 py-2 text-right">{t("summary.model")}</th>
                  <th className="px-3 py-2 text-right">{t("summary.route")}</th>
                  <th className="px-3 py-2 text-right">{t("summary.latency")}</th>
                  <th className="px-3 py-2 text-right">{t("summary.tokens")}</th>
                  <th className="px-3 py-2 text-right">{t("summary.cost")}</th>
                </tr>
              </thead>
              <tbody>
                {logs.map((item) => (
                  <tr key={item.log_id} className="border-b border-border/60">
                    <td className="px-3 py-3">
                      <Link className="font-medium text-primary hover:underline" href={`/logs/${item.log_id}`}>
                        {item.status_code}
                      </Link>
                    </td>
                    <td className="px-3 py-3">{item.provider ?? "-"}</td>
                    <td className="px-3 py-3">{item.model ?? "-"}</td>
                    <td className="px-3 py-3">{item.route ?? "-"}</td>
                    <td className="px-3 py-3">{item.duration_ms} ms</td>
                    <td className="px-3 py-3">{item.total_tokens ?? 0}</td>
                    <td className="px-3 py-3">${(item.cost_usd ?? 0).toFixed(4)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </AdminShell>
  );
}
