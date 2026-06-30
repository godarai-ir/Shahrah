import Link from "next/link";
import { notFound } from "next/navigation";
import { getTranslations } from "next-intl/server";

import AdminShell from "@/app/components/admin-shell";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type Params = {
  params: Promise<{ logID: string }>;
};

export const dynamic = "force-dynamic";

export default async function LogDetailPage({ params }: Params) {
  const { logID } = await params;
  const auth = await requireAuth(`/login?next=/logs/${logID}`);
  const tenantSlug = auth.tenantSlug;
  const t = await getTranslations("logs");
  await requirePermissionOr403(tenantSlug, auth.user.id, ["logs:read_metadata", "logs:read_content"]);

  const result = await fetchControlPlaneJSON({
    tenantSlug,
    method: "GET",
    path: `/api/v2/logs/${encodeURIComponent(logID)}`,
    userID: auth.user.id,
  });

  if (result.status === 404) {
    notFound();
  }

  const payload = (result.payload ?? {}) as Record<string, unknown>;
  const currentUser = { id: auth.user.id, email: auth.user.email, mobile: auth.user.mobile, role: auth.user.role };
  const log = (payload.log ?? {}) as Record<string, unknown>;
  const visibility = (payload.content_visibility ?? {}) as Record<string, unknown>;
  const requestPayload = payload.request_payload as Record<string, unknown> | undefined;
  const responsePayload = payload.response_payload as Record<string, unknown> | undefined;

  return (
    <AdminShell tenantSlug={tenantSlug} currentUser={currentUser}>
      <section className="space-y-4 rounded-2xl border border-border bg-card p-5">
        <div className="space-y-1">
          <h2 className="text-xl font-semibold">{t("detailTitle")}</h2>
          <p className="text-sm text-muted-foreground">{String(log.log_id ?? logID)}</p>
        </div>

        {!result.ok ? (
          <p className="rounded-xl border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">
            {result.errorMessage ?? t("detailError")}
          </p>
        ) : (
          <>
            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              <Info label={t("summary.status")} value={`${String(log.status_code ?? "-")} / ${String(log.status ?? "-")}`} />
              <Info label={t("summary.provider")} value={String(log.provider ?? "-")} />
              <Info label={t("summary.model")} value={String(log.model ?? "-")} />
              <Info label={t("summary.route")} value={String(log.route ?? "-")} />
              <Info label={t("summary.latency")} value={`${String(log.duration_ms ?? 0)} ms`} />
              <Info label={t("summary.tokens")} value={String(log.total_tokens ?? 0)} />
              <Info label={t("summary.cost")} value={`$${Number(log.cost_usd ?? 0).toFixed(4)}`} />
              <Info
                label={t("summary.trace")}
                value={
                  log.trace_id ? (
                    <Link className="text-primary hover:underline" href={`/traces/${String(log.trace_id)}`}>
                      {String(log.trace_id)}
                    </Link>
                  ) : "-"
                }
              />
            </div>

            <div className="grid gap-4 lg:grid-cols-2">
              <PayloadCard title={t("content.request")} payload={requestPayload} hidden={!visibility.can_view_content} hiddenText={t("content.hidden")} />
              <PayloadCard title={t("content.response")} payload={responsePayload} hidden={!visibility.can_view_content} hiddenText={t("content.hidden")} />
            </div>
          </>
        )}
      </section>
    </AdminShell>
  );
}

function Info({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-border/70 p-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 text-sm font-medium">{value}</div>
    </div>
  );
}

function PayloadCard({
  title,
  payload,
  hidden,
  hiddenText,
}: {
  title: string;
  payload?: Record<string, unknown>;
  hidden: boolean;
  hiddenText: string;
}) {
  return (
    <div className="rounded-xl border border-border/70 p-4">
      <h3 className="mb-3 text-sm font-semibold">{title}</h3>
      {hidden ? (
        <p className="text-sm text-muted-foreground">{hiddenText}</p>
      ) : (
        <pre className="overflow-x-auto rounded-lg bg-muted/50 p-3 text-xs leading-6">
          {JSON.stringify(payload ?? {}, null, 2)}
        </pre>
      )}
    </div>
  );
}
