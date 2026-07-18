"use client";

import { memo } from "react";
import { ShieldCheck } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import {
  type PrincipalEntitySelectLabels,
  type PrincipalSearchKind,
} from "@/app/components/principal-search-input";
import { PrincipalPicker } from "@/app/components/principal-picker";
import type { AccessPrincipal, AccessPrincipalCandidate } from "./types";
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
  showSaveAction?: boolean;
}) {
  const ux = useTranslations("providerAccounts.ux");

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
    </section>
  );
});


