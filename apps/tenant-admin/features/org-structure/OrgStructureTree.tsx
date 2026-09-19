"use client";

import { Building2, ChevronDown, ChevronRight, Users } from "lucide-react";
import { useTranslations } from "next-intl";

import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { EmptyState } from "@/components/app-ui/EmptyState";
import { CompactValue } from "@/components/app-ui/CompactValue";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { formatCompactIrr } from "@/lib/compact-number";
import { cn } from "@/lib/utils";

import { TENANT_ROOT_ID, type OrgTreeRow } from "./org-tree";
import { getTeamName, getOrgUnitName, getWorkspaceType, getActiveKeys } from "./helpers";

export type OrgStructureTreeSelection = { kind: "unit"; id: string } | { kind: "team"; id: string } | null;

export function OrgStructureTree({
  rows,
  isLoading,
  isSearching,
  searchQuery,
  selection,
  onSelectUnit,
  onSelectTeam,
  onToggleExpand,
  costByTeamID,
  locale,
}: {
  rows: OrgTreeRow[];
  isLoading: boolean;
  isSearching: boolean;
  searchQuery: string;
  selection: OrgStructureTreeSelection;
  onSelectUnit: (unitID: string) => void;
  onSelectTeam: (teamID: string) => void;
  onToggleExpand: (id: string) => void;
  costByTeamID: Map<string, number> | null;
  locale: string;
}) {
  const t = useTranslations("orgStructure");

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
        icon={isSearching ? Users : Building2}
        title={isSearching ? t("tree.emptySearch", { query: searchQuery }) : t("tree.emptyUnits")}
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
          {rows.map((row) => {
            if (row.kind === "unit") {
              const isSelected = selection?.kind === "unit" && selection.id === row.id;
              return (
                <TableRow key={row.id} className={cn("cursor-pointer bg-muted/20", isSelected && "bg-muted/60")} onClick={() => onSelectUnit(row.id)}>
                  <TableCell colSpan={1}>
                    <div className="flex min-w-0 items-center gap-2">
                      <button
                        type="button"
                        aria-label={row.expanded ? "collapse" : "expand"}
                        className="shrink-0 rounded p-0.5 text-muted-foreground hover:bg-muted"
                        onClick={(event) => { event.stopPropagation(); onToggleExpand(row.id); }}
                      >
                        {row.expanded ? <ChevronDown className="size-4" /> : <ChevronRight className="size-4" />}
                      </button>
                      <Building2 className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
                      <span className="truncate font-medium" title={getOrgUnitName(row.unit)}>{getOrgUnitName(row.unit)}</span>
                      <Badge variant="secondary" className="shrink-0">{t("tree.teamCount", { count: row.teamCount })}</Badge>
                    </div>
                  </TableCell>
                  <TableCell className="hidden @md:table-cell" />
                  <TableCell colSpan={2} className="text-end text-muted-foreground">—</TableCell>
                </TableRow>
              );
            }

            if (row.kind === "root") {
              const isSelected = selection?.kind === "unit" && selection.id === TENANT_ROOT_ID;
              return (
                <TableRow key={row.id} className={cn("cursor-pointer bg-muted/10", isSelected && "bg-muted/50")} onClick={() => onToggleExpand(row.id)}>
                  <TableCell colSpan={1}>
                    <div className="flex min-w-0 items-center gap-2">
                      <button
                        type="button"
                        aria-label={row.expanded ? "collapse" : "expand"}
                        className="shrink-0 rounded p-0.5 text-muted-foreground hover:bg-muted"
                        onClick={(event) => { event.stopPropagation(); onToggleExpand(row.id); }}
                      >
                        {row.expanded ? <ChevronDown className="size-4" /> : <ChevronRight className="size-4" />}
                      </button>
                      <span className="truncate text-muted-foreground">{t("tree.root")}</span>
                      <Badge variant="outline" className="shrink-0">{t("tree.teamCount", { count: row.teamCount })}</Badge>
                    </div>
                  </TableCell>
                  <TableCell className="hidden @md:table-cell" />
                  <TableCell colSpan={2} />
                </TableRow>
              );
            }

            const isSelected = selection?.kind === "team" && selection.id === row.id;
            const costIRR = costByTeamID?.get(row.team.id) ?? null;
            return (
              <TableRow key={row.id} className={cn("cursor-pointer", isSelected && "bg-muted/50")} onClick={() => onSelectTeam(row.id)}>
                <TableCell>
                  <div className={cn("flex min-w-0 items-center gap-2", !isSearching && "ps-7")}>
                    <span className="truncate font-medium" title={getTeamName(row.team)}>{getTeamName(row.team)}</span>
                    {isSearching && row.unitName ? <span className="shrink-0 truncate text-xs text-muted-foreground">/ {row.unitName}</span> : null}
                    {isSearching && !row.unitId ? <span className="shrink-0 truncate text-xs text-muted-foreground">/ {t("tree.root")}</span> : null}
                  </div>
                </TableCell>
                <TableCell className="hidden @md:table-cell">
                  <Badge variant="outline">{getWorkspaceType(row.team)}</Badge>
                </TableCell>
                <TableCell className="text-end tabular-nums">{getActiveKeys(row.team)}</TableCell>
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
