"use client";

import { useState } from "react";
import { Plus, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/app-ui/ConfirmDialog";
import { cn } from "@/lib/utils";
import {
  PrincipalEntitySelect,
  type PrincipalEntitySelection,
  type PrincipalEntitySelectLabels,
  type PrincipalSearchKind,
  type PrincipalSearchOption,
} from "@/app/components/principal-search-input";
import { hideMobileIdentity } from "@/lib/auth/mobile-auth-ui";

export type PrincipalPickerItem = {
  type: "user" | "team" | "org_unit" | "virtual_account" | "everyone";
  principal_id?: string;
  display_name?: string;
};

export function PrincipalPicker<TItem extends PrincipalPickerItem = PrincipalPickerItem>(props: {
  id: string;
  title: string;
  onSearchPrincipal: (kind: PrincipalSearchKind, query: string) => Promise<PrincipalSearchOption[]>;
  principalSelectorLabels: PrincipalEntitySelectLabels;
  selected: TItem[];
  setSelected: (value: TItem[]) => void;
  allowEveryone: boolean;
  addLabel: string;
  removeLabel: string;
  kinds?: PrincipalSearchKind[];
  disabled?: boolean;
  variant?: "default" | "filter";
  className?: string;
  description?: string;
  searchLabel?: string;
  getRemoveLabel?: (item: TItem) => string;
  everyoneConfirmation?: {
    title: string;
    description: string;
    confirmLabel: string;
    cancelLabel: string;
  };
}) {
  const [candidate, setCandidate] = useState<PrincipalEntitySelection | null>(null);
  const [everyoneConfirmOpen, setEveryoneConfirmOpen] = useState(false);
  const isFilterVariant = props.variant === "filter";

  function addPrincipal(item: TItem) {
    const exists = props.selected.some(
      (current) =>
        current.type === item.type &&
        (current.principal_id ?? "") === (item.principal_id ?? ""),
    );
    if (exists) return;
    props.setSelected([...props.selected, item]);
  }

  function removePrincipal(item: TItem) {
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
    if (candidate.type !== "user" && candidate.type !== "team" && candidate.type !== "org_unit" && candidate.type !== "virtual_account" && candidate.type !== "everyone") return;
    if (candidate.type === "everyone" && props.everyoneConfirmation) {
      setEveryoneConfirmOpen(true);
      return;
    }
    commitCandidate(candidate);
  }

  function commitCandidate(nextCandidate: PrincipalEntitySelection) {
    if (nextCandidate.type === "external_identity") return;
    addPrincipal({
      type: nextCandidate.type,
      principal_id: nextCandidate.principal_id,
      display_name: nextCandidate.display_name,
    } as TItem);
    setCandidate(null);
  }

  return (
    <div
      className={cn(
        "flex min-h-full flex-col",
        isFilterVariant
          ? "gap-3 rounded-md bg-transparent px-3 py-3 md:grid md:min-h-0 md:grid-cols-[auto_minmax(0,1fr)_minmax(8rem,auto)] md:items-start"
          : "gap-4 rounded-2xl bg-background/85 px-4 py-4 ring-1 ring-border/35",
        props.className,
      )}
    >
      <div className={cn("flex items-start justify-between gap-3", isFilterVariant && "md:min-w-32")}>
        <div>
          <h4 className={cn(isFilterVariant ? "text-xs font-normal leading-8 text-muted-foreground" : "text-base font-semibold")}>{props.title}</h4>
          {props.description ? <p className="mt-1 text-sm leading-6 text-muted-foreground">{props.description}</p> : null}
        </div>
      </div>

      <div className={cn(isFilterVariant ? "flex min-w-0 flex-row-reverse items-start gap-2" : "flex w-full min-w-0 items-start gap-2")}>
        {isFilterVariant ? (
          <Button
            type="button"
            variant="outline"
            size="icon-sm"
            onClick={addSelectedCandidate}
            disabled={props.disabled || !candidate || candidate.type === "external_identity"}
            className="shrink-0"
            aria-label={props.addLabel}
            title={props.addLabel}
          >
            <Plus data-icon="inline-start" />
            <span className="sr-only">{props.addLabel}</span>
          </Button>
        ) : null}
        <PrincipalEntitySelect
          id={props.id}
          selected={candidate}
          onSelectedChange={props.disabled ? () => {} : setCandidate}
          onSearch={props.onSearchPrincipal}
          labels={props.principalSelectorLabels}
          kinds={props.kinds ?? ["user", "team", "org_unit"]}
          includeEveryone={props.allowEveryone}
          className={cn("min-w-0 flex-1", isFilterVariant && "flex-1")}
          compact={isFilterVariant}
          ariaLabel={props.searchLabel}
        />
        {isFilterVariant ? null : (
          <Button
            type="button"
            variant="outline"
            size="icon-sm"
            onClick={addSelectedCandidate}
            disabled={props.disabled || !candidate || candidate.type === "external_identity"}
            className="shrink-0"
            aria-label={props.addLabel}
            title={props.addLabel}
          >
            <Plus data-icon="inline-start" />
            <span className="sr-only">{props.addLabel}</span>
          </Button>
        )}
      </div>

      <div className={cn("flex min-h-9 flex-wrap items-center gap-2", isFilterVariant && "md:justify-end")}>
        {props.selected.map((item) => (
          <button
            key={`${item.type}-${item.principal_id ?? "everyone"}`}
            type="button"
            className="inline-flex max-w-full items-center gap-2 rounded-full bg-background px-3 py-1.5 text-xs ring-1 ring-border/35 transition hover:bg-destructive/5 hover:ring-destructive/30"
            onClick={() => {
              if (props.disabled) return;
              removePrincipal(item);
            }}
            title={props.getRemoveLabel?.(item) ?? props.removeLabel}
            aria-label={props.getRemoveLabel?.(item) ?? props.removeLabel}
            disabled={props.disabled}
          >
            <span className="max-w-[240px] truncate">{hideMobileIdentity(item.display_name) || item.principal_id || item.type}</span>
            <X aria-hidden="true" className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          </button>
        ))}
      </div>
      {props.everyoneConfirmation ? (
        <ConfirmDialog
          open={everyoneConfirmOpen}
          onOpenChange={setEveryoneConfirmOpen}
          title={props.everyoneConfirmation.title}
          description={props.everyoneConfirmation.description}
          confirmLabel={props.everyoneConfirmation.confirmLabel}
          cancelLabel={props.everyoneConfirmation.cancelLabel}
          destructive={false}
          onConfirm={() => {
            if (candidate?.type === "everyone") commitCandidate(candidate);
          }}
        />
      ) : null}
    </div>
  );
}

