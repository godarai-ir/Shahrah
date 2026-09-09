import type { Dispatch, ReactNode, SetStateAction } from "react";
import { useTranslations } from "next-intl";

import { DefinitionList, DefinitionListItem } from "@/components/app-ui/DefinitionList";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

import { SummaryKPI } from "./overview";
import type { AccessPrincipal, ProviderAuthMode, ProviderCatalogModel } from "./types";
import { formatDisplayPrice } from "./utils";
import type { CreateWizardSectionID } from "./wizard-layout";
import { CatalogCreateSection } from "./wizard-shell";

export function CreateReviewPanel({
  ux,
  catalogT,
  setActiveCreateSection,
  onFinish,
  createReady,
  canCreateProviderAccount,
  isActivating,
  isTesting,
  isLoading,
  hasCredential,
  usePlatformManagedCredentials,
  apiKey,
  selectedModels,
  selectedModelCount,
  unresolvedPricingCount,
  managers,
  users,
  providerName,
  accountName,
  authMode,
}: {
  ux: ReturnType<typeof useTranslations>;
  catalogT: ReturnType<typeof useTranslations>;
  setActiveCreateSection: Dispatch<SetStateAction<CreateWizardSectionID>>;
  onFinish: () => Promise<void>;
  createReady: boolean;
  canCreateProviderAccount: boolean;
  isActivating: boolean;
  isTesting: boolean;
  isLoading: boolean;
  hasCredential: boolean;
  usePlatformManagedCredentials: boolean;
  apiKey: string;
  selectedModels: ProviderCatalogModel[];
  selectedModelCount: number;
  unresolvedPricingCount: number;
  managers: AccessPrincipal[];
  users: AccessPrincipal[];
  providerName: string;
  accountName: string;
  authMode: ProviderAuthMode;
}) {
  return (
    <CatalogCreateSection
      id="review-confirm"
      title={ux("integrations.wizard.reviewConfirm")}
      description={ux("integrations.wizard.reviewConfirmDescription")}
      fillHeight={false}
      footer={(
        <div className="flex w-full flex-wrap items-center justify-between gap-3">
          <Button type="button" variant="outline" onClick={() => setActiveCreateSection("access")}>
            {ux("actions.back")}
          </Button>
          <div className="flex flex-wrap items-center justify-end gap-3">
            {!createReady ? <span className="text-sm text-muted-foreground">{catalogT("create.completeRequired")}</span> : null}
            <Button
              type="button"
              onClick={() => void onFinish()}
              disabled={!canCreateProviderAccount || !createReady || isActivating || isTesting || isLoading}
              aria-busy={isActivating}
            >
              {isActivating ? ux("actions.activating") : ux("actions.activateAccount")}
            </Button>
          </div>
        </div>
      )}
    >
      <div className="flex flex-col gap-4">
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <SummaryKPI label={ux("kpi.credentialStatus")} value={hasCredential || usePlatformManagedCredentials || apiKey.trim() ? ux("readiness.ready") : ux("readiness.needsCredentials")} />
          <SummaryKPI label={ux("kpi.enabledModels")} value={selectedModelCount.toString()} />
          <SummaryKPI label={ux("kpi.unresolvedPricing")} value={unresolvedPricingCount.toString()} />
          <SummaryKPI label={ux("kpi.managersUsers")} value={`${managers.length} / ${users.length}`} />
        </div>
        <ReviewSection title={ux("review.connectionTitle")} editLabel={ux("actions.edit")} onEdit={() => setActiveCreateSection("details")}>
          <DefinitionList columns={2}>
            <DefinitionListItem label={ux("review.provider")} value={providerName} valueDir="ltr" surface />
            <DefinitionListItem label={ux("review.accountName")} value={accountName} valueDir="ltr" mono surface />
            <DefinitionListItem label={ux("review.authMode")} value={authMode.replaceAll("_", " ")} valueDir="ltr" surface />
            <DefinitionListItem label={ux("review.credentialSource")} value={usePlatformManagedCredentials ? ux("review.godarKey") : ux("review.providerKey")} surface />
          </DefinitionList>
        </ReviewSection>
        <ReviewSection title={ux("review.modelsTitle")} editLabel={ux("actions.edit")} onEdit={() => setActiveCreateSection("models")}>
          <div className="grid gap-2 sm:grid-cols-2">
            {selectedModels.map((model) => (
              <div key={model.model} className="min-w-0 rounded-lg border border-border bg-muted/20 px-3 py-2">
                <p className="truncate text-sm font-medium" dir="ltr" title={model.display_name || model.model}>{model.display_name || model.model}</p>
                <p className="mt-1 text-xs text-muted-foreground">
                  {ux("models.card.context")}: <bdi>{model.context_window?.toLocaleString() ?? ux("models.na")}</bdi>
                  {" · "}{ux("models.labels.input")}: <bdi dir="ltr">{formatReviewPrice(model.public_pricing?.input_usd_per_1m, ux)}</bdi>
                  {" · "}{ux("models.labels.output")}: <bdi dir="ltr">{formatReviewPrice(model.public_pricing?.output_usd_per_1m, ux)}</bdi>
                </p>
              </div>
            ))}
          </div>
        </ReviewSection>
        <ReviewSection title={ux("review.accessTitle")} editLabel={ux("actions.edit")} onEdit={() => setActiveCreateSection("access")}>
          <DefinitionList columns={2}>
            <DefinitionListItem label={ux("review.managers")} value={managers.map(principalLabel).join(ux("listSeparator"))} wrap surface />
            <DefinitionListItem label={ux("review.users")} value={users.length > 0 ? users.map(principalLabel).join(ux("listSeparator")) : ux("review.noUsers")} wrap surface />
          </DefinitionList>
        </ReviewSection>
        <Alert>
          <AlertTitle>{ux("review.activationCheckTitle")}</AlertTitle>
          <AlertDescription>{ux("microcopy.pricingPolicy")}</AlertDescription>
        </Alert>
      </div>
    </CatalogCreateSection>
  );
}

function ReviewSection({ title, editLabel, onEdit, children }: { title: string; editLabel: string; onEdit: () => void; children: ReactNode }) {
  return (
    <section className="flex flex-col gap-3 rounded-lg border border-border bg-background p-4">
      <div className="flex items-center justify-between gap-3">
        <h3 className="font-semibold">{title}</h3>
        <Button type="button" variant="link" size="sm" onClick={onEdit}>{editLabel}</Button>
      </div>
      {children}
    </section>
  );
}

function principalLabel(item: AccessPrincipal) {
  return item.display_name || item.principal_id || item.type;
}

function formatReviewPrice(value: number | undefined, ux: ReturnType<typeof useTranslations>) {
  return formatDisplayPrice(value, {
    currency: "USD",
    scale: "PER_MILLION_TOKENS",
    usdToIrr: 0,
    naLabel: ux("models.na"),
  });
}
