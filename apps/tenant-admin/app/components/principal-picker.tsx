"use client";

import { useState } from "react";
import { Plus, X } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  PrincipalEntitySelect,
  type PrincipalEntitySelection,
  type PrincipalEntitySelectLabels,
  type PrincipalSearchKind,
  type PrincipalSearchOption,
} from "@/app/components/principal-search-input";

export type PrincipalPickerItem = {
  type: "user" | "team" | "org_unit" | "everyone";
  principal_id?: string;
  display_name?: string;
};

export function PrincipalPicker(props: {
  id: string;
  title: string;
  onSearchPrincipal: (kind: PrincipalSearchKind, query: string) => Promise<PrincipalSearchOption[]>;
  principalSelectorLabels: PrincipalEntitySelectLabels;
  selected: PrincipalPickerItem[];
  setSelected: (value: PrincipalPickerItem[]) => void;
  allowEveryone: boolean;
  addLabel: string;
  removeLabel: string;
}) {
  const [candidate, setCandidate] = useState<PrincipalEntitySelection | null>(null);

  function addPrincipal(item: PrincipalPickerItem) {
    const exists = props.selected.some(
      (current) =>
        current.type === item.type &&
        (current.principal_id ?? "") === (item.principal_id ?? ""),
    );
    if (exists) return;
    props.setSelected([...props.selected, item]);
  }

  function removePrincipal(item: PrincipalPickerItem) {
    props.setSelected(
      props.selected.filter(
        (current) =>
          !(
            current.type === item.type &&
            (current.principal_id ?? "") === (item.principal_id ?? "")
          ),
      ),
    );
  }

  function addSelectedCandidate() {
    if (!candidate) return;
    if (candidate.type === "external_identity") return;
    if (candidate.type !== "user" && candidate.type !== "team" && candidate.type !== "org_unit" && candidate.type !== "everyone") return;
    addPrincipal({
      type: candidate.type,
      principal_id: candidate.principal_id,
      display_name: candidate.display_name,
    });
    setCandidate(null);
  }

  return (
    <div className="flex min-h-full flex-col gap-4 rounded-lg border border-border/70 bg-card p-4">
      <div className="flex items-start justify-between gap-3">
        <h4 className="text-base font-semibold">{props.title}</h4>
        <Badge variant="secondary" className="shrink-0">
          {props.selected.length}
        </Badge>
      </div>

      <div className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto]">
        <PrincipalEntitySelect
          id={props.id}
          selected={candidate}
          onSelectedChange={setCandidate}
          onSearch={props.onSearchPrincipal}
          labels={props.principalSelectorLabels}
          kinds={["user", "team", "org_unit"]}
          includeEveryone={props.allowEveryone}
          className="min-w-0"
        />
        <Button
          type="button"
          variant="outline"
          onClick={addSelectedCandidate}
          disabled={!candidate || candidate.type === "external_identity"}
          className="w-full sm:w-auto"
        >
          <Plus className="h-4 w-4" />
          {props.addLabel}
        </Button>
      </div>

      <div className="flex min-h-9 flex-wrap items-center gap-2">
        {props.selected.map((item) => (
          <button
            key={`${item.type}-${item.principal_id ?? "everyone"}`}
            type="button"
            className="inline-flex max-w-full items-center gap-2 rounded-full border bg-background px-3 py-1.5 text-xs transition hover:border-destructive/40 hover:bg-destructive/5"
            onClick={() => removePrincipal(item)}
            title={props.removeLabel}
          >
            <span className="max-w-[240px] truncate">{item.display_name || item.principal_id || item.type}</span>
            <X className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          </button>
        ))}
      </div>
    </div>
  );
}

