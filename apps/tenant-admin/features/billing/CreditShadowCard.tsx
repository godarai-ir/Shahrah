import { Coins } from "lucide-react";
import type { ReactNode } from "react";

import { AppCard } from "@/components/app-ui/AppCard";
import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { ModelNameValue } from "@/components/app-ui/ModelNameValue";
import { StatusBadge } from "@/components/app-ui/StatusBadge";
import type { AppLocale } from "@/lib/locale";

import type { CreditUsageItem, CreditUsageSummary } from "./api";
import { formatCreditAmount } from "./domain";

export type CreditShadowCardLabels = {
  title: string;
  description: string;
  badge: string;
  unit: string;
  total: string;
  cacheSaved: string;
  byModel: string;
  byTeam: string;
  noTeam: string;
  emptyTitle: string;
  emptyDescription: string;
  requestCount: (count: number) => string;
};

type CreditShadowCardProps = {
  summary: CreditUsageSummary;
  labels: CreditShadowCardLabels;
  locale: AppLocale;
  limit?: number;
};

/**
 * Shows shadow-mode credit usage next to the IRR billing figures. Credits are
 * computed for comparison only and are never deducted while shadow mode is on.
 */
export function CreditShadowCard({ summary, labels, locale, limit = 5 }: CreditShadowCardProps) {
  const formatCredits = (value: string) => `${formatCreditAmount(value, locale)} ${labels.unit}`;
  const isEmpty = summary.request_count === 0;

  return (
    <AppCard
      title={labels.title}
      description={labels.description}
      actions={<StatusBadge status="info">{labels.badge}</StatusBadge>}
    >
      <div className="flex flex-col gap-6">
        <DefinitionList columns={2}>
          <DefinitionListItem icon={Coins} label={labels.total} value={formatCredits(summary.total_credits)} emphasis />
          <DefinitionListItem label={labels.cacheSaved} value={formatCredits(summary.cache_saved_credits)} />
        </DefinitionList>
        {isEmpty ? (
          <EmptyState variant="inline" density="compact" title={labels.emptyTitle} description={labels.emptyDescription} />
        ) : (
          <div className="grid gap-6 lg:grid-cols-2">
            <CreditUsageList
              title={labels.byModel}
              items={summary.by_model.slice(0, limit)}
              renderKey={(item) => <ModelNameValue value={item.key} />}
              formatCredits={formatCredits}
              requestCount={labels.requestCount}
            />
            <CreditUsageList
              title={labels.byTeam}
              items={summary.by_team.slice(0, limit)}
              renderKey={(item) => (
                <span className="truncate font-medium text-foreground">{item.label || (item.key ? item.key : labels.noTeam)}</span>
              )}
              formatCredits={formatCredits}
              requestCount={labels.requestCount}
            />
          </div>
        )}
      </div>
    </AppCard>
  );
}

type CreditUsageListProps = {
  title: string;
  items: CreditUsageItem[];
  renderKey: (item: CreditUsageItem) => ReactNode;
  formatCredits: (value: string) => string;
  requestCount: (count: number) => string;
};

function CreditUsageList({ title, items, renderKey, formatCredits, requestCount }: CreditUsageListProps) {
  return (
    <section className="flex min-w-0 flex-col gap-2">
      <h3 className="text-sm font-medium text-foreground">{title}</h3>
      <ul className="divide-y divide-border">
        {items.map((item) => (
          <li key={item.key || "none"} className="flex items-center justify-between gap-4 py-2">
            <div className="flex min-w-0 flex-col gap-1">
              {renderKey(item)}
              <span className="text-xs text-muted-foreground">{requestCount(item.request_count)}</span>
            </div>
            <StatusBadge status="neutral" className="shrink-0 tabular-nums">{formatCredits(item.credits)}</StatusBadge>
          </li>
        ))}
      </ul>
    </section>
  );
}
