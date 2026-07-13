import Link from "next/link";
import { notFound } from "next/navigation";
import { getLocale, getTranslations } from "next-intl/server";

import AdminShell from "@/app/components/admin-shell";
import { CompactValue, DefinitionList, DefinitionListItem } from "@/components/app-ui";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";
import { formatCompactCurrency, formatCompactNumber } from "@/lib/compact-number";
import { fetchControlPlaneJSON } from "@/lib/control-plane";

type Params = {
  params: Promise<{ logID: string }>;
};

export const dynamic = "force-dynamic";

export default async function LogDetailPage({ params }: Params) {
  const { logID } = await params;
  const locale = await getLocale();
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
    <AdminShell
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("detailTitle"), description: String(log.log_id ?? logID) }}
    >
      <section className="space-y-4 rounded-2xl border border-border bg-card p-5">
        {!result.ok ? (
          <p className="rounded-xl border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive">
            {result.errorMessage ?? t("detailError")}
          </p>
        ) : (
          <>
            <DefinitionList columns={4}>
              <DefinitionListItem surface label={t("summary.status")} value={`${String(log.status_code ?? "-")} / ${String(log.status ?? "-")}`} />
              <DefinitionListItem surface mono valueDir="ltr" label={t("summary.provider")} value={String(log.provider ?? "-")} />
              <DefinitionListItem surface mono valueDir="ltr" label={t("summary.model")} value={String(log.model ?? "-")} />
              <DefinitionListItem surface mono valueDir="ltr" label={t("summary.route")} value={String(log.route ?? "-")} />
              <DefinitionListItem surface valueDir="ltr" label={t("summary.latency")} value={`${String(log.duration_ms ?? 0)} ms`} />
              <DefinitionListItem surface label={t("summary.tokens")} value={<CompactValue display={formatCompactNumber(Number(log.total_tokens ?? 0), { locale })} />} />
              <DefinitionListItem surface label={t("summary.cost")} value={<CompactValue display={formatCompactCurrency(Number(log.cost_usd ?? 0), { locale, currency: "USD", maximumFractionDigits: 4 })} />} />
              <DefinitionListItem surface mono valueDir="ltr"
                label={t("summary.trace")}
                value={
                  log.trace_id ? (
                    <Link className="text-primary hover:underline" href={`/traces/${String(log.trace_id)}`}>
                      {String(log.trace_id)}
                    </Link>
                  ) : "-"
                }
              />
            </DefinitionList>

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
