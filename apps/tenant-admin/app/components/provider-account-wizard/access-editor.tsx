"use client";

import { memo } from "react";
import { ShieldCheck, X } from "lucide-react";
import { useTranslations } from "next-intl";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { Button } from "@/components/ui/button";
import {
  PrincipalEntitySelect,
  type PrincipalEntitySelection,
  type PrincipalEntitySelectLabels,
  type PrincipalSearchKind,
} from "@/app/components/principal-search-input";
import { PrincipalPicker } from "@/app/components/principal-picker";
import type { AccessPrincipal, AccessPrincipalCandidate, EffectiveModelView } from "./types";
import { EffectiveModelsTable } from "./effective-models-table";
import { Field } from "./field";
export const AccessEditor = memo(function AccessEditor(props: {
  t: ReturnType<typeof useTranslations>;
  managers: AccessPrincipal[];
  setManagers: (value: AccessPrincipal[]) => void;
  users: AccessPrincipal[];
  setUsers: (value: AccessPrincipal[]) => void;
  onSearchPrincipal: (kind: PrincipalSearchKind, query: string) => Promise<AccessPrincipalCandidate[]>;
  principalSelectorLabels: PrincipalEntitySelectLabels;
  onSave: () => void;
  isSaving: boolean;
  canManage: boolean;
  effectiveKeyType: "pat" | "vat";
  setEffectiveKeyType: (value: "pat" | "vat") => void;
  effectivePrincipal: PrincipalEntitySelection | null;
  setEffectivePrincipal: (value: PrincipalEntitySelection | null) => void;
  effectiveModels: EffectiveModelView[];
  isLoadingEffective: boolean;
  onLoadEffective: () => void;
}) {
  const ux = useTranslations("providerAccounts.ux");
  const hasEffectivePrincipal = Boolean((props.effectivePrincipal?.principal_id ?? "").trim());

  return (
    <section className="space-y-5 rounded-lg border border-border/70 bg-background p-4 sm:p-5">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <h3 className="text-base font-semibold">{ux("tabs.access")}</h3>
          <p className="mt-1 text-sm leading-6 text-muted-foreground">{ux("access.description")}</p>
        </div>
        <Button type="button" onClick={props.onSave} disabled={!props.canManage || props.isSaving} className="shrink-0">
          <ShieldCheck className="h-4 w-4" />
          {props.isSaving ? props.t("saving") : props.t("saveAccessControl")}
        </Button>
      </div>

      <div className="grid items-stretch gap-4">
        <PrincipalPicker
          id="managers-principal-picker"
          title={props.t("step3.managersTitle")}
          onSearchPrincipal={props.onSearchPrincipal}
          principalSelectorLabels={props.principalSelectorLabels}
          selected={props.managers}
          setSelected={props.setManagers}
          disabled={!props.canManage}
          allowEveryone={false}
          addLabel={props.t("principal.add")}
          removeLabel={props.t("principal.remove")}
        />

        <PrincipalPicker
          id="users-principal-picker"
          title={props.t("step3.usersTitle")}
          onSearchPrincipal={props.onSearchPrincipal}
          principalSelectorLabels={props.principalSelectorLabels}
          selected={props.users}
          setSelected={props.setUsers}
          disabled={!props.canManage}
          allowEveryone
          addLabel={props.t("principal.add")}
          removeLabel={props.t("principal.remove")}
        />
      </div>

      <div className="space-y-4 rounded-lg border border-border/70 bg-muted/10 p-4">
        <h4 className="text-base font-semibold">{ux("access.effectivePreviewTitle")}</h4>
        <div className="grid gap-3 lg:grid-cols-[150px_minmax(0,1fr)_auto]">
          <AppSelectField
            id="previewKeyType"
            label={ux("access.keyTypeLabel")}
            value={props.effectiveKeyType}
            onValueChange={(value) => props.setEffectiveKeyType(value as "pat" | "vat")}
            options={[
              { value: "pat", label: ux("access.pat") },
              { value: "vat", label: ux("access.vat") },
            ]}
          />
          <Field label={`${ux("access.userIdLabel")} / ${ux("access.teamIdLabel")}`} id="previewPrincipal">
            <div className="grid grid-cols-[minmax(0,1fr)_40px] gap-2">
              <PrincipalEntitySelect
                id="previewPrincipal"
                selected={props.effectivePrincipal}
                onSelectedChange={props.setEffectivePrincipal}
                onSearch={props.onSearchPrincipal}
                labels={props.principalSelectorLabels}
                kinds={["user", "team"]}
                className="min-w-0"
              />
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => props.setEffectivePrincipal(null)}
                disabled={!hasEffectivePrincipal}
                title={ux("actions.clear")}
                aria-label={ux("actions.clear")}
                className="mt-0"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </Field>
          <div className="flex items-end">
            <Button
              type="button"
              variant="outline"
              onClick={props.onLoadEffective}
              disabled={props.isLoadingEffective || !hasEffectivePrincipal}
              className="h-10 w-full xl:w-auto"
            >
              {props.isLoadingEffective ? ux("actions.loading") : ux("access.previewAction")}
            </Button>
          </div>
        </div>
        <EffectiveModelsTable models={props.effectiveModels} />
      </div>
    </section>
  );
});


