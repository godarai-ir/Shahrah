import Link from "next/link";
import { ChevronRight, SearchX } from "lucide-react";

import { EmptyState } from "@/components/app-ui/EmptyState";
import { DebouncedTextFilter } from "@/components/app-ui/FilterField";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";

import type { ProviderCatalogProvider } from "./types";
import { formatProviderDisplayName } from "./utils";
import { ProviderLogo } from "@/components/app-ui/ProviderLogo";
import { CatalogCreateSection } from "./wizard-shell";

export function ProviderPicker({
  providers,
  isLoading,
  search,
  onSearchChange,
  onSelect,
  labels,
}: {
  providers: ProviderCatalogProvider[];
  isLoading: boolean;
  search: string;
  onSearchChange: (value: string) => void;
  onSelect: (provider: ProviderCatalogProvider) => void;
  labels: {
    cancel: string;
    search: string;
    searchPlaceholder: string;
    emptyTitle: string;
    emptyDescription: string;
  };
}) {
  return (
    <CatalogCreateSection
      title=""
      footer={(
        <Button type="button" variant="outline" asChild>
          <Link href="/provider-accounts">{labels.cancel}</Link>
        </Button>
      )}
    >
      <div className="flex flex-col gap-4 sm:gap-5">
        <DebouncedTextFilter
          label={labels.search}
          value={search}
          onValueChange={onSearchChange}
          placeholder={labels.searchPlaceholder}
        />
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          {isLoading ? Array.from({ length: 6 }).map((_, index) => (
            <div key={`provider-picker-skeleton-${index}`} className="flex min-h-20 items-center gap-3 rounded-lg border border-card-border bg-background p-3">
              <Skeleton className="size-10 shrink-0 rounded-lg" />
              <Skeleton className="h-4 w-32 max-w-full" />
            </div>
          )) : providers.length === 0 ? (
            <EmptyState
              icon={SearchX}
              title={labels.emptyTitle}
              description={labels.emptyDescription}
              variant="inline"
              density="compact"
              className="sm:col-span-2 xl:col-span-3"
            />
          ) : providers.map((item) => (
            <button
              key={item.provider}
              type="button"
              className="group flex min-h-20 min-w-0 items-center gap-3 rounded-lg border border-card-border bg-background p-3 text-start shadow-xs transition hover:border-primary/40 hover:bg-muted/50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring active:translate-y-px"
              onClick={() => onSelect(item)}
            >
              <ProviderLogo
                provider={item.provider}
                displayName={formatProviderDisplayName(item.provider, item.display_name)}
                logoURL={item.logo_url}
                size={28}
                decorative
              />
              <p className="min-w-0 flex-1 truncate text-sm font-semibold" dir="ltr">
                {formatProviderDisplayName(item.provider, item.display_name)}
              </p>
              <ChevronRight aria-hidden="true" className="size-4 shrink-0 text-muted-foreground transition-colors group-hover:text-primary rtl:rotate-180" />
            </button>
          ))}
        </div>
      </div>
    </CatalogCreateSection>
  );
}
