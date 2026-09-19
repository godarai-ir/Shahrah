"use client";

import { ChevronRight } from "lucide-react";
import { useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { DateTimeValue } from "@/components/app-ui/DateTimeValue";
import { Skeleton } from "@/components/ui/skeleton";
import { formatCompactIrr } from "@/lib/compact-number";
import type { OrgUnit, OrgUnitMember } from "@/lib/api/org-units";
import type { Team } from "@/lib/api/teams";

import { getOrgUnitName, getTeamName, getWorkspaceType } from "./helpers";

export function OrgUnitDetail({
  orgUnit,
  teams,
  members,
  isLoading,
  costIRR,
  locale,
  canManage,
  canDelete,
  onEdit,
  onManageMembers,
  onSelectTeam,
  onDelete,
}: {
  orgUnit: OrgUnit;
  teams: Team[];
  members: OrgUnitMember[];
  isLoading: boolean;
  costIRR: number | null;
  locale: string;
  canManage: boolean;
  canDelete: boolean;
  onEdit: () => void;
  onManageMembers: () => void;
  onSelectTeam: (teamID: string) => void;
  onDelete: () => void;
}) {
  const t = useTranslations("orgStructure");
  const tOrgUnits = useTranslations("orgUnits");
  const admins = members.filter((member) => member.role === "org_admin");
  const deleteBlocked = teams.length > 0;

  return (
    <div className="flex flex-col gap-4">
      <AppCard
        title={getOrgUnitName(orgUnit)}
        description={orgUnit.alias}
        actions={canManage ? <Button type="button" variant="outline" size="sm" onClick={onEdit}>{t("actions.edit")}</Button> : undefined}
        contentClassName="grid grid-cols-1 gap-px overflow-hidden rounded-lg border border-border bg-border sm:grid-cols-3"
      >
        <div className="bg-card p-4">
          <p className="text-xs text-muted-foreground">{t("detail.unit.teams")}</p>
          <p className="mt-1 text-xl font-semibold tabular-nums">{teams.length}</p>
        </div>
        <div className="bg-card p-4">
          <p className="text-xs text-muted-foreground">{t("detail.unit.members")}</p>
          <p className="mt-1 text-xl font-semibold tabular-nums">{members.length}</p>
        </div>
        <div className="bg-card p-4">
          <p className="text-xs text-muted-foreground">{t("detail.unit.costMonth")}</p>
          <p className="mt-1 text-xl font-semibold tabular-nums">
            {costIRR === null ? <span className="text-muted-foreground">—</span> : <CompactValue dir="ltr" display={formatCompactIrr(costIRR, { locale })} />}
          </p>
        </div>
      </AppCard>

      <AppCard title={t("detail.unit.childTeams")} contentClassName="flex flex-col gap-2 text-sm">
        {teams.length === 0 ? (
          <p className="text-muted-foreground">{t("detail.unit.noChildTeams")}</p>
        ) : (
          teams.map((team) => (
            <button
              key={team.id}
              type="button"
              onClick={() => onSelectTeam(team.id)}
              className="flex items-center justify-between gap-2 rounded-lg border border-border bg-muted/20 p-3 text-start hover:bg-muted/40"
            >
              <div className="min-w-0">
                <p className="truncate font-medium">{getTeamName(team)}</p>
                <p className="truncate text-xs text-muted-foreground">{team.alias}</p>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <Badge variant="outline">{getWorkspaceType(team)}</Badge>
                <ChevronRight className="size-4 text-muted-foreground rtl:rotate-180" aria-hidden="true" />
              </div>
            </button>
          ))
        )}
      </AppCard>

      <AppCard
        title={t("detail.unit.admins")}
        actions={canManage ? <Button type="button" size="sm" onClick={onManageMembers}>{t("actions.manageMembers")}</Button> : undefined}
        contentClassName="flex flex-col gap-2 text-sm"
      >
        {isLoading ? (
          <div className="flex flex-col gap-2">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        ) : admins.length === 0 ? (
          <p className="text-muted-foreground">{t("detail.unit.noAdmins")}</p>
        ) : (
          admins.map((member) => (
            <div key={member.id} className="flex items-center justify-between gap-3 rounded-lg border border-border bg-muted/20 p-3">
              <span className="truncate text-sm font-medium" dir="auto">{member.user_id}</span>
              <Badge>{tOrgUnits("roles.orgAdmin")}</Badge>
            </div>
          ))
        )}
      </AppCard>

      <AppCard contentClassName="flex flex-col gap-3 text-sm">
        <p className="text-xs text-muted-foreground">{t("detail.unit.lastChange")}: <DateTimeValue value={orgUnit.updated_at} locale={locale} format={{ year: "numeric", month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" }} /></p>
        <Collapsible>
          <CollapsibleTrigger asChild>
            <Button type="button" variant="ghost" size="sm" className="w-fit">{t("detail.unit.rawMetadata")}</Button>
          </CollapsibleTrigger>
          <CollapsibleContent>
            <pre dir="ltr" className="mt-2 max-h-40 overflow-auto rounded-lg bg-muted p-3 text-start text-xs text-muted-foreground">
              {JSON.stringify(orgUnit.metadata ?? {}, null, 2)}
            </pre>
          </CollapsibleContent>
        </Collapsible>
      </AppCard>

      {canDelete ? (
        <AppCard contentClassName="flex flex-col gap-2 text-sm">
          {deleteBlocked ? (
            <div className="flex flex-col gap-1">
              <p className="font-medium">{t("detail.unit.deleteGuardTitle")}</p>
              <p className="text-muted-foreground">{t("detail.unit.deleteGuardDescription", { count: teams.length })}</p>
            </div>
          ) : null}
          <div>
            <Button type="button" variant="outline" className="text-destructive" disabled={deleteBlocked} onClick={onDelete}>
              {t("actions.delete")}
            </Button>
          </div>
        </AppCard>
      ) : null}
    </div>
  );
}
