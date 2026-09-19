import type { OrgUnit } from "@/lib/api/org-units";
import type { Team } from "@/lib/api/teams";

export const TENANT_ROOT_ID = "__tenant_root__";

export type OrgTreeRow =
  | { kind: "unit"; id: string; unit: OrgUnit; teamCount: number; expanded: boolean }
  | { kind: "root"; id: typeof TENANT_ROOT_ID; teamCount: number; expanded: boolean }
  | { kind: "team"; id: string; team: Team; unitId: string | null; unitName: string | null };

function matchesQuery(team: Team, query: string): boolean {
  const needle = query.trim().toLowerCase();
  if (needle === "") return true;
  const haystacks = [team.alias, team.display_name ?? "", team.id];
  return haystacks.some((value) => value.toLowerCase().includes(needle));
}

export function buildOrgTree(units: readonly OrgUnit[], teams: readonly Team[], expanded: ReadonlySet<string>, query = ""): OrgTreeRow[] {
  const trimmedQuery = query.trim();
  const unitNameByID = new Map(units.map((unit) => [unit.id, unit.display_name?.trim() || unit.alias]));

  if (trimmedQuery !== "") {
    return teams
      .filter((team) => matchesQuery(team, trimmedQuery))
      .map((team) => ({
        kind: "team" as const,
        id: team.id,
        team,
        unitId: team.org_unit_id ?? null,
        unitName: team.org_unit_id ? (unitNameByID.get(team.org_unit_id) ?? null) : null,
      }));
  }

  const teamsByUnitID = new Map<string, Team[]>();
  const rootTeams: Team[] = [];
  for (const team of teams) {
    if (team.org_unit_id) {
      const bucket = teamsByUnitID.get(team.org_unit_id);
      if (bucket) bucket.push(team);
      else teamsByUnitID.set(team.org_unit_id, [team]);
    } else {
      rootTeams.push(team);
    }
  }

  const rows: OrgTreeRow[] = [];
  for (const unit of units) {
    const unitTeams = teamsByUnitID.get(unit.id) ?? [];
    const isExpanded = expanded.has(unit.id);
    rows.push({ kind: "unit", id: unit.id, unit, teamCount: unitTeams.length, expanded: isExpanded });
    if (isExpanded) {
      for (const team of unitTeams) {
        rows.push({ kind: "team", id: team.id, team, unitId: unit.id, unitName: unitNameByID.get(unit.id) ?? null });
      }
    }
  }

  if (rootTeams.length > 0) {
    const isExpanded = expanded.has(TENANT_ROOT_ID);
    rows.push({ kind: "root", id: TENANT_ROOT_ID, teamCount: rootTeams.length, expanded: isExpanded });
    if (isExpanded) {
      for (const team of rootTeams) {
        rows.push({ kind: "team", id: team.id, team, unitId: null, unitName: null });
      }
    }
  }

  return rows;
}
