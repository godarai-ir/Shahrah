"use client";

import { Users } from "lucide-react";
import { useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { formatCompactIrr } from "@/lib/compact-number";
import { hideMobileIdentity } from "@/lib/auth/mobile-auth-ui";
import type { Team, TeamMember } from "@/lib/api/teams";

import { getActiveKeys, getAdminCount, getTeamName, getVirtualUsers, getWorkspaceType } from "./helpers";

const MEMBER_PREVIEW_LIMIT = 4;

export function TeamDetail({
  team,
  members,
  isLoading,
  costIRR,
  locale,
  canManage,
  canDelete,
  onEdit,
  onManageMembers,
  onDelete,
}: {
  team: Team;
  members: TeamMember[];
  isLoading: boolean;
  costIRR: number | null;
  locale: string;
  canManage: boolean;
  canDelete: boolean;
  onEdit: () => void;
  onManageMembers: () => void;
  onDelete: () => void;
}) {
  const t = useTranslations("orgStructure");
  const tTeams = useTranslations("teams");

  const activeKeys = getActiveKeys(team);
  const virtualUsers = getVirtualUsers(team);
  const adminCount = members.filter((member) => member.role === "team_admin").length || getAdminCount(team) || 0;
  const deleteBlocked = activeKeys > 0 || members.length > 0 || virtualUsers > 0;

  return (
    <div className="flex flex-col gap-4">
      <AppCard
        title={
          <div className="flex flex-wrap items-center gap-2">
            <span>{getTeamName(team)}</span>
            <Badge variant="outline">{tTeams(`workspace.${getWorkspaceType(team)}`)}</Badge>
          </div>
        }
        description={team.alias}
        actions={canManage ? <Button type="button" variant="outline" size="sm" onClick={onEdit}>{t("actions.edit")}</Button> : undefined}
        contentClassName="grid grid-cols-1 gap-px overflow-hidden rounded-lg border border-border bg-border sm:grid-cols-3"
      >
        <div className="bg-card p-4">
          <p className="text-xs text-muted-foreground">{t("detail.team.members")}</p>
          <p className="mt-1 text-xl font-semibold tabular-nums">{members.length}</p>
          <p className="text-xs text-muted-foreground">{t("detail.team.membersSubtitle", { count: adminCount })}</p>
        </div>
        <div className="bg-card p-4">
          <p className="text-xs text-muted-foreground">{t("detail.team.activeKeys")}</p>
          <p className="mt-1 text-xl font-semibold tabular-nums">{activeKeys}</p>
          <p className="text-xs text-muted-foreground">{t("detail.team.activeKeysSubtitle", { count: virtualUsers })}</p>
        </div>
        <div className="bg-card p-4">
          <p className="text-xs text-muted-foreground">{t("detail.team.costMonth")}</p>
          <p className="mt-1 text-xl font-semibold tabular-nums">
            {costIRR === null ? <span className="text-muted-foreground">—</span> : <CompactValue dir="ltr" display={formatCompactIrr(costIRR, { locale })} />}
          </p>
        </div>
      </AppCard>


      <AppCard
        title={
          <span className="flex items-center gap-2">
            <Users aria-hidden="true" />
            <span>{t("detail.team.membersPreview")}</span>
            <Badge variant="secondary">{members.length}</Badge>
          </span>
        }
        actions={canManage ? <Button type="button" size="sm" onClick={onManageMembers}>{t("actions.manageMembers")}</Button> : undefined}
        contentClassName="flex flex-col gap-3 text-sm"
      >
        {isLoading ? (
          <div className="flex flex-col gap-2">
            <Skeleton className="h-14 w-full" />
            <Skeleton className="h-14 w-full" />
          </div>
        ) : members.length === 0 ? (
          <p className="text-muted-foreground">{t("detail.team.noMembers")}</p>
        ) : (
          <>
            {members.slice(0, MEMBER_PREVIEW_LIMIT).map((member) => {
              const label = hideMobileIdentity(member.user_display_name) || member.user_email?.trim() || member.user_id;
              return (
                <div key={member.user_id} className="flex items-center justify-between gap-3 rounded-lg border border-border bg-muted/20 p-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <Avatar className="size-8">
                      <AvatarFallback dir="ltr" className="text-xs">{getAvatarInitials(label)}</AvatarFallback>
                    </Avatar>
                    <p className="truncate font-medium" dir="auto">{label}</p>
                  </div>
                  <Badge variant={member.role === "team_admin" ? "default" : "secondary"} className="shrink-0">
                    {member.role === "team_admin" ? tTeams("members.teamAdmin") : tTeams("members.teamMember")}
                  </Badge>
                </div>
              );
            })}
            {members.length > MEMBER_PREVIEW_LIMIT ? (
              <p className="text-xs text-muted-foreground">{t("detail.team.andMore", { count: members.length - MEMBER_PREVIEW_LIMIT })}</p>
            ) : null}
          </>
        )}
      </AppCard>

      {canDelete ? (
        <AppCard contentClassName="flex flex-col gap-2 text-sm">
          {deleteBlocked ? (
            <div className="flex flex-col gap-1">
              <p className="font-medium">{t("detail.team.deleteGuardTitle")}</p>
              <p className="text-muted-foreground">{t("detail.team.deleteGuardDescription")}</p>
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

function getAvatarInitials(value: string) {
  const parts = value.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  const first = Array.from(parts[0])[0] ?? "";
  const last = parts.length > 1 ? Array.from(parts.at(-1) ?? "")[0] ?? "" : "";
  return `${first}${last}`.toUpperCase();
}
