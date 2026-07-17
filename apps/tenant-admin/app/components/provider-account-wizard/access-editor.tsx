"use client";

import { memo } from "react";
import { ShieldCheck, X } from "lucide-react";
import { useTranslations } from "next-intl";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel } from "@/components/ui/field";
import {
  PrincipalEntitySelect,
  type PrincipalEntitySelection,
  type PrincipalEntitySelectLabels,
  type PrincipalSearchKind,
} from "@/app/components/principal-search-input";
import { PrincipalPicker } from "@/app/components/principal-picker";
import type { AccessPrincipal, AccessPrincipalCandidate, EffectiveModelView } from "./types";
import { EffectiveModelsTable } from "./effective-models-table";
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
  showSaveAction?: boolean;
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
        {props.showSaveAction ?? true ? (
          <Button type="button" onClick={props.onSave} disabled={!props.canManage || props.isSaving} className="shrink-0">
            <ShieldCheck data-icon="inline-start" />
            {props.isSaving ? props.t("saving") : props.t("saveAccessControl")}
          </Button>
        ) : null}
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
          addLabel={ux("access.addManager")}
          removeLabel={props.t("principal.remove")}
          description={ux("access.managerRequiredHint")}
          searchLabel={ux("access.searchManager")}
          getRemoveLabel={(item) => ux("access.removeManager", { identity: item.display_name || item.principal_id || item.type })}
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
          addLabel={ux("access.addUser")}
          removeLabel={props.t("principal.remove")}
          description={ux("access.usersOptionalHint")}
          searchLabel={ux("access.searchUser")}
          getRemoveLabel={(item) => ux("access.removeUser", { identity: item.display_name || item.principal_id || item.type })}
          everyoneConfirmation={{
            title: ux("access.everyoneConfirmTitle"),
            description: ux("access.everyoneConfirmDescription"),
            confirmLabel: ux("access.everyoneConfirmAction"),
            cancelLabel: ux("actions.cancelDelete"),
          }}
        />
      </div>

      <div className="space-y-4 rounded-lg border border-border/70 bg-muted/10 p-4">
        <h4 className="text-base font-semibold">{ux("access.effectivePreviewTitle")}</h4>
        <div className="grid gap-3 lg:grid-cols-4">
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
          <Field className="lg:col-span-2">
            <FieldLabel htmlFor="previewPrincipal">{`${ux("access.userIdLabel")} / ${ux("access.teamIdLabel")}`}</FieldLabel>
            <div className="flex min-w-0 gap-2">
              <PrincipalEntitySelect
                id="previewPrincipal"
                selected={props.effectivePrincipal}
                onSelectedChange={props.setEffectivePrincipal}
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
        {hasEffectivePrincipal ? (
          <EffectiveModelsTable models={props.effectiveModels} />
        ) : (
          <p className="rounded-lg border border-dashed border-border px-4 py-6 text-center text-sm text-muted-foreground">
            {ux("access.previewEmptyHint")}
          </p>
        )}
      </div>
    </section>
  );
});


