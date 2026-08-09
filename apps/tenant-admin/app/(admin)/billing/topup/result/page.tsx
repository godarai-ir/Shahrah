import Link from "next/link";
import { getTranslations } from "next-intl/server";

import AdminShell from "@/features/admin-shell/AdminShell";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

type ResultPageProps = {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export const dynamic = "force-dynamic";

export default async function TopUpResultPage(props: ResultPageProps) {
  const auth = await requireAuth("/login");
  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["settings:list"]);
  const t = await getTranslations("topupResult");
  const tenantSlug = auth.tenantSlug;

  const currentUser = {
    id: auth.user.id,
    email: auth.user.email,
    mobile: auth.user.mobile,
    role: auth.user.role,
  };

  const searchParams = await props.searchParams;
  const status = first(searchParams.status) ?? "unknown";
  const reason = first(searchParams.reason);
  const refID = first(searchParams.ref_id);
  const authority = first(searchParams.authority);

  const success = status === "success";
  const statusLabel = labelForStatus(status, t);

  return (
    <AdminShell messageNamespaces={["topupResult"]}
      tenantSlug={tenantSlug}
      currentUser={currentUser}
      pageHeader={{ title: t("title"), description: success ? t("successDescription") : t("failureDescription") }}
    >
      <div className="mx-auto max-w-2xl rounded-xl border bg-card p-6">
        <dl className="flex flex-col gap-2 text-sm">
          <div className="flex items-center justify-between gap-3 border-b pb-2">
            <dt className="text-muted-foreground">{t("fields.status")}</dt>
            <dd className="font-medium">{statusLabel}</dd>
          </div>
          {refID ? (
            <div className="flex items-center justify-between gap-3 border-b pb-2">
              <dt className="text-muted-foreground">{t("fields.refId")}</dt>
              <dd className="font-mono">{refID}</dd>
            </div>
          ) : null}
          {authority ? (
            <div className="flex items-center justify-between gap-3 border-b pb-2">
              <dt className="text-muted-foreground">{t("fields.authority")}</dt>
              <dd className="font-mono">{authority}</dd>
            </div>
          ) : null}
          {reason ? (
            <div className="flex items-center justify-between gap-3 border-b pb-2">
              <dt className="text-muted-foreground">{t("fields.reason")}</dt>
              <dd>{reason}</dd>
            </div>
          ) : null}
        </dl>

        <div className="mt-6">
          <Link href="/billing/topup" className="inline-flex h-10 items-center rounded-md border px-4 text-sm font-medium">
            {t("backToTopup")}
          </Link>
        </div>
      </div>
    </AdminShell>
  );
}

function first(value: string | string[] | undefined): string | undefined {
  if (Array.isArray(value)) {
    return value[0];
  }
  return value;
}

function labelForStatus(status: string, t: Awaited<ReturnType<typeof getTranslations<"topupResult">>>) {
  switch (status) {
    case "success":
      return t("status.success");
    case "failed":
      return t("status.failed");
    case "cancelled":
    case "canceled":
      return t("status.cancelled");
    case "pending":
      return t("status.pending");
    case "unknown":
      return t("status.unknown");
    default:
      return status;
  }
}
