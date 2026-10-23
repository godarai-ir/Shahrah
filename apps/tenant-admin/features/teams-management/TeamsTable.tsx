"use client";

import { Users } from "lucide-react";
import { useTranslations } from "next-intl";

import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { formatCompactIrr } from "@/lib/compact-number";
import { cn } from "@/lib/utils";

import type { Team } from "@/lib/api/teams";

import { getTeamName, getWorkspaceType, getActiveKeys } from "./helpers";

export type TeamsTableSelection = { kind: "team"; id: string } | null;

export function TeamsTable({
  rows,
  isLoading,
  isSearching,
  searchQuery,
  selection,
  onSelectTeam,
  costByTeamID,
  locale,
}: {
  rows: Team[];
  isLoading: boolean;
  isSearching: boolean;
  searchQuery: string;
  selection: TeamsTableSelection;
  onSelectTeam: (teamID: string) => void;
  costByTeamID: Map<string, number> | null;
  locale: string;
}) {
  const t = useTranslations("teamsPage");

  if (isLoading) {
    return (
      <div className="flex flex-col gap-2 p-4" aria-label={t("tree.loading")}>
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-10 w-full" />
      </div>
    );
  }

  if (rows.length === 0) {
    return (
      <EmptyState
        icon={Users}
        title={isSearching ? t("tree.emptySearch", { query: searchQuery }) : t("tree.emptyTeams")}
        className="min-h-40"
      />
    );
  }

  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <Table className="table-fixed text-xs lg:text-sm">
        <TableHeader>
          <TableRow>
            <TableHead className="w-2/5">{t("tree.name")}</TableHead>
            <TableHead className="hidden w-1/5 @md:table-cell">{t("tree.environment")}</TableHead>
            <TableHead className="w-1/5">{t("tree.activeKeys")}</TableHead>
            <TableHead className="w-1/5">{t("tree.monthCost")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((team) => {
            const isSelected = selection?.kind === "team" && selection.id === team.id;
            const costIRR = costByTeamID?.get(team.id) ?? null;
            return (
              <TableRow key={team.id} className={cn("cursor-pointer", isSelected && "bg-muted/50")} onClick={() => onSelectTeam(team.id)}>
                <TableCell>
                  <div className="flex min-w-0 items-center gap-2">
                    <span className="truncate font-medium" title={getTeamName(team)}>{getTeamName(team)}</span>
                  </div>
                </TableCell>
                <TableCell className="hidden @md:table-cell">
                  <Badge variant="outline">{getWorkspaceType(team)}</Badge>
                </TableCell>
                <TableCell className="text-end tabular-nums">{getActiveKeys(team)}</TableCell>
                <TableCell className="text-end tabular-nums">
                  {costIRR === null ? <span className="text-muted-foreground">—</span> : <CompactValue dir="ltr" display={formatCompactIrr(costIRR, { locale })} />}
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}
