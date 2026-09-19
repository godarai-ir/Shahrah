import { resolveSectionLandingTab } from "@/features/admin-shell/section-landing";
import { requirePermissionOr403 } from "@/lib/auth/permissions";
import { requireAuth } from "@/lib/auth/session";

import LogsPage from "../logs/page";
import MetricsPage from "../metrics/page";
import TracesPage from "../traces/page";

export const dynamic = "force-dynamic";

type ObservePageProps = {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
};

/** Section landing for logs, traces and metrics. Renders the first permitted tab in place; the legacy routes stay as deep links. */
export default async function ObservePage({ searchParams }: ObservePageProps) {
  const auth = await requireAuth("/login?next=/observe");
  const tab = await resolveSectionLandingTab("/observe", auth.tenantSlug, auth.user.id);

  if (tab === "logs") return <LogsPage searchParams={searchParams} />;
  if (tab === "traces") return <TracesPage searchParams={searchParams} />;
  if (tab === "metrics") return <MetricsPage />;

  await requirePermissionOr403(auth.tenantSlug, auth.user.id, ["logs:read_metadata", "logs:read_content", "settings:list"]);
  return null;
}
