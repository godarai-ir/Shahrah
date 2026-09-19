import { describe, expect, it } from "vitest";

import type { OrgUnit } from "@/lib/api/org-units";
import type { Team } from "@/lib/api/teams";

import { buildOrgTree, TENANT_ROOT_ID } from "./org-tree";

function makeUnit(id: string, alias: string, displayName?: string): OrgUnit {
  return {
    id,
    tenant_id: "tenant-1",
    alias,
    display_name: displayName,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  };
}

function makeTeam(id: string, alias: string, orgUnitID: string | null, displayName?: string): Team {
  return {
    id,
    tenant_id: "tenant-1",
    alias,
    display_name: displayName,
    org_unit_id: orgUnitID,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  };
}

describe("buildOrgTree", () => {
  const unitA = makeUnit("unit-a", "unit-a", "واحد الف");
  const unitB = makeUnit("unit-b", "unit-b", "واحد ب");
  const teamA1 = makeTeam("team-a1", "team-a1", "unit-a", "تیم یک");
  const teamA2 = makeTeam("team-a2", "team-a2", "unit-a", "تیم دو");
  const teamRoot = makeTeam("team-root", "team-root", null, "تیم بی‌واحد");
  const units = [unitA, unitB];
  const teams = [teamA1, teamA2, teamRoot];

  it("renders every unit as a collapsed row with its team count, and no root row when nothing is unassigned", () => {
    const rows = buildOrgTree(units, [teamA1, teamA2], new Set());
    expect(rows).toEqual([
      { kind: "unit", id: "unit-a", unit: unitA, teamCount: 2, expanded: false },
      { kind: "unit", id: "unit-b", unit: unitB, teamCount: 0, expanded: false },
    ]);
  });

  it("expands a unit's teams only when its id is in the expanded set", () => {
    const rows = buildOrgTree(units, teams, new Set(["unit-a"]));
    expect(rows.map((row) => row.id)).toEqual(["unit-a", "team-a1", "team-a2", "unit-b", TENANT_ROOT_ID]);
  });

  it("adds a virtual tenant-root row only when there is at least one unassigned team", () => {
    const withRoot = buildOrgTree(units, teams, new Set());
    expect(withRoot.some((row) => row.kind === "root")).toBe(true);

    const withoutRoot = buildOrgTree(units, [teamA1, teamA2], new Set());
    expect(withoutRoot.some((row) => row.kind === "root")).toBe(false);
  });

  it("expands the tenant-root row only when TENANT_ROOT_ID is in the expanded set", () => {
    const collapsed = buildOrgTree(units, teams, new Set());
    expect(collapsed.find((row) => row.kind === "root")).toEqual({ kind: "root", id: TENANT_ROOT_ID, teamCount: 1, expanded: false });

    const expanded = buildOrgTree(units, teams, new Set([TENANT_ROOT_ID]));
    const rootIndex = expanded.findIndex((row) => row.kind === "root");
    expect(expanded[rootIndex + 1]).toEqual({ kind: "team", id: "team-root", team: teamRoot, unitId: null, unitName: null });
  });

  it("collapses to a flat, matching list of teams when a search query is given, with each team's unit name attached", () => {
    const rows = buildOrgTree(units, teams, new Set(), "team-a1");
    expect(rows).toEqual([{ kind: "team", id: "team-a1", team: teamA1, unitId: "unit-a", unitName: "واحد الف" }]);
  });

  it("matches search against alias, display name, and id, case-insensitively", () => {
    expect(buildOrgTree(units, teams, new Set(), "تیم دو").map((row) => row.id)).toEqual(["team-a2"]);
    expect(buildOrgTree(units, teams, new Set(), "TEAM-A2").map((row) => row.id)).toEqual(["team-a2"]);
    expect(buildOrgTree(units, teams, new Set(), "nomatch")).toEqual([]);
  });

  it("gives an unassigned team a null unit name in search results", () => {
    const rows = buildOrgTree(units, teams, new Set(), "team-root");
    expect(rows).toEqual([{ kind: "team", id: "team-root", team: teamRoot, unitId: null, unitName: null }]);
  });

  it("falls back to the unit's alias when it has no display name", () => {
    const unitNoName = makeUnit("unit-c", "unit-c");
    const teamC = makeTeam("team-c", "team-c", "unit-c");
    const rows = buildOrgTree([unitNoName], [teamC], new Set(), "team-c");
    expect(rows).toEqual([{ kind: "team", id: "team-c", team: teamC, unitId: "unit-c", unitName: "unit-c" }]);
  });
});
