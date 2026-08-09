"use client";

import { useState } from "react";
import { ShieldCheck, X } from "lucide-react";
import { useTranslations } from "next-intl";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { OperationNotice, type OperationNoticeValue } from "@/components/app-ui/OperationNotice";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel } from "@/components/ui/field";
import { Spinner } from "@/components/ui/spinner";
import {
  PrincipalEntitySelect,
  type PrincipalEntitySelection,
  type PrincipalEntitySelectLabels,
  type PrincipalSearchKind,
} from "@/app/components/principal-search-input";
import type {
  AccessPrincipalCandidate,
  EffectiveModelsResponse,
  EffectiveModelView,
  TenantProviderAccount,
} from "@/app/components/provider-account-wizard/types";
import { EffectiveModelsTable } from "@/app/components/provider-account-wizard/effective-models-table";
import { normalizeError, requestJSON, unwrapData } from "@/app/components/provider-account-wizard/utils";

export function EffectiveAccessTester(props: {
  account: TenantProviderAccount;
  onSearchPrincipal: (kind: PrincipalSearchKind, query: string) => Promise<AccessPrincipalCandidate[]>;
  principalSelectorLabels: PrincipalEntitySelectLabels;
}) {
  const ux = useTranslations("providerAccounts.ux");
  const [keyType, setKeyType] = useState<"pat" | "vat">("pat");
  const [principal, setPrincipal] = useState<PrincipalEntitySelection | null>(null);
  const [models, setModels] = useState<EffectiveModelView[]>([]);
  const [notice, setNotice] = useState<OperationNoticeValue | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [hasRun, setHasRun] = useState(false);

  const hasPrincipal = Boolean((principal?.principal_id ?? "").trim());

  async function loadEffectiveModels() {
    if (!hasPrincipal || !principal) {
      setModels([]);
      return;
    }

    setIsLoading(true);
    setHasRun(false);
    setNotice(null);
    try {
      const query = new URLSearchParams();
      query.set("key_type", keyType);
      const principalID = (principal.principal_id ?? "").trim();
      if (principal.type === "user") query.set("user_id", principalID);
      if (principal.type === "team") query.set("team_id", principalID);

      const response = await requestJSON<EffectiveModelsResponse>(
        `/api/tenant-provider-accounts/${encodeURIComponent(props.account.id)}/effective-models?${query.toString()}`,
      );
      const payload = unwrapData<EffectiveModelView[]>(response) ?? [];
      setModels(
        payload.map((item) => ({
          ...item,
          types: item.types ?? item.effective_capabilities ?? [],
        })),
      );
      setHasRun(true);
    } catch (error) {
      setModels([]);
      setNotice({ kind: "error", message: normalizeError(error, ux("accessTester.error")) });
    } finally {
      setIsLoading(false);
    }
  }

  function clearPrincipal() {
    setPrincipal(null);
    setModels([]);
    setHasRun(false);
    setNotice(null);
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-3">
        <AppSelectField
          id="effectiveAccessKeyType"
          label={ux("access.keyTypeLabel")}
          value={keyType}
          onValueChange={(value) => setKeyType(value as "pat" | "vat")}
          options={[
            { value: "pat", label: ux("access.pat") },
            { value: "vat", label: ux("access.vat") },
          ]}
        />
        <Field>
          <FieldLabel htmlFor="effectiveAccessPrincipal">{`${ux("access.userIdLabel")} / ${ux("access.teamIdLabel")}`}</FieldLabel>
          <div className="flex min-w-0 gap-2">
            <PrincipalEntitySelect
              id="effectiveAccessPrincipal"
              selected={principal}
              onSelectedChange={(selection) => {
                setPrincipal(selection);
                setModels([]);
                setHasRun(false);
                setNotice(null);
              }}
              onSearch={props.onSearchPrincipal}
              labels={props.principalSelectorLabels}
              kinds={["user", "team"]}
              className="min-w-0 flex-1"
              ariaLabel={ux("access.previewPrincipalSearch")}
            />
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={clearPrincipal}
              disabled={!hasPrincipal || isLoading}
              title={ux("actions.clear")}
              aria-label={ux("actions.clear")}
            >
              <X />
            </Button>
          </div>
        </Field>
      </div>

      <Button type="button" onClick={() => void loadEffectiveModels()} disabled={isLoading || !hasPrincipal} aria-busy={isLoading}>
        {isLoading ? <Spinner data-icon="inline-start" /> : <ShieldCheck data-icon="inline-start" />}
        {isLoading ? ux("actions.loading") : ux("accessTester.run")}
      </Button>

      <OperationNotice
        notice={notice}
        labels={{
          success: ux("accessTester.successTitle"),
          error: ux("accessTester.errorTitle"),
          warning: ux("accessTester.warningTitle"),
          info: ux("accessTester.infoTitle"),
        }}
      />

      {hasPrincipal && hasRun ? (
        <EffectiveModelsTable models={models} />
      ) : (
        <p className="rounded-lg border border-dashed border-border px-4 py-6 text-center text-sm text-muted-foreground">
          {hasPrincipal ? ux("accessTester.readyHint") : ux("access.previewEmptyHint")}
        </p>
      )}
    </div>
  );
}
