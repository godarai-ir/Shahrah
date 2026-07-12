"use client";

import * as React from "react";
import { List, ListFilter } from "lucide-react";

import { EmptyState } from "@/components/app-ui/EmptyState";
import { SearchField } from "@/components/app-ui/SearchField";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Field, FieldContent, FieldDescription, FieldGroup, FieldLabel, FieldLegend, FieldSet } from "@/components/ui/field";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

export type PermissionOption = {
  key: string;
  group: string;
  description: string;
};

export type PermissionEditorMode = "selected" | "all";

export type PermissionEditorLabels = {
  selectedTab: string;
  allTab: string;
  searchPlaceholder: string;
  searchAriaLabel: string;
  emptyTitle: string;
  selectGroup: string;
  clearGroup: string;
};

export type PermissionEditorProps = {
  permissions: PermissionOption[];
  value: string[];
  onValueChange: (value: string[]) => void;
  query: string;
  onQueryChange: (query: string) => void;
  mode: PermissionEditorMode;
  onModeChange: (mode: PermissionEditorMode) => void;
  labels: PermissionEditorLabels;
  formatGroupLabel?: (group: string) => React.ReactNode;
  formatPermissionDescription?: (permission: PermissionOption) => React.ReactNode;
  formatNumber?: (value: number) => string;
  dir?: "rtl" | "ltr";
  disabled?: boolean;
  className?: string;
};

export function PermissionEditor({
  permissions,
  value,
  onValueChange,
  query,
  onQueryChange,
  mode,
  onModeChange,
  labels,
  formatGroupLabel = defaultGroupLabel,
  formatPermissionDescription = (permission) => permission.description || permission.key,
  formatNumber = String,
  dir = "rtl",
  disabled,
  className,
}: PermissionEditorProps) {
  const generatedID = React.useId();
  const selected = React.useMemo(() => new Set(value), [value]);
  const groupedPermissions = React.useMemo(
    () => groupPermissions(permissions, query, formatGroupLabel, formatPermissionDescription),
    [formatGroupLabel, formatPermissionDescription, permissions, query],
  );
  const visibleGroups = React.useMemo(
    () => groupedPermissions
      .map(([group, items]) => [group, mode === "selected" ? items.filter((item) => selected.has(item.key)) : items] as const)
      .filter(([, items]) => items.length > 0),
    [groupedPermissions, mode, selected],
  );

  function setPermission(permissionKey: string, checked: boolean) {
    if (checked) {
      onValueChange(Array.from(new Set([...value, permissionKey])));
      return;
    }
    onValueChange(value.filter((item) => item !== permissionKey));
  }

  function setGroup(permissionKeys: string[], checked: boolean) {
    if (checked) {
      onValueChange(Array.from(new Set([...value, ...permissionKeys])));
      return;
    }
    const keys = new Set(permissionKeys);
    onValueChange(value.filter((item) => !keys.has(item)));
  }

  return (
    <Tabs
      value={mode}
      onValueChange={(nextMode) => onModeChange(nextMode as PermissionEditorMode)}
      dir={dir}
      className={cn("min-w-0", className)}
    >
      <TabsList className="w-full sm:w-fit">
        <TabsTrigger value="selected">
          <ListFilter aria-hidden="true" />
          {labels.selectedTab}
        </TabsTrigger>
        <TabsTrigger value="all">
          <List aria-hidden="true" />
          {labels.allTab}
        </TabsTrigger>
      </TabsList>
      <TabsContent value={mode} className="flex min-w-0 flex-col gap-3">
        <SearchField
          containerClassName="w-full"
          value={query}
          onChange={(event) => onQueryChange(event.target.value)}
          placeholder={labels.searchPlaceholder}
          aria-label={labels.searchAriaLabel}
          dir={dir}
          disabled={disabled}
        />

        <div className="max-h-96 overflow-y-auto pe-1" aria-live="polite">
          <div className="flex flex-col gap-3">
            {visibleGroups.length === 0 ? (
              <EmptyState title={labels.emptyTitle} variant="inline" density="compact" />
            ) : null}

            {visibleGroups.map(([group, items]) => {
              const permissionKeys = items.map((item) => item.key);
              const selectedCount = permissionKeys.filter((key) => selected.has(key)).length;
              const allSelected = selectedCount === permissionKeys.length;

              return (
                <FieldSet key={group} className="grid grid-cols-[minmax(0,1fr)_auto] gap-2 rounded-lg border border-border">
                  <FieldLegend variant="label" className="mb-0 flex min-w-0 items-center gap-2 px-3 pt-3">
                    <span className="truncate">{formatGroupLabel(group)}</span>
                    <Badge variant="outline">
                      {formatNumber(selectedCount)}/{formatNumber(permissionKeys.length)}
                    </Badge>
                  </FieldLegend>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    className="me-3 mt-2"
                    disabled={disabled}
                    onClick={() => setGroup(permissionKeys, !allSelected)}
                  >
                    {allSelected ? labels.clearGroup : labels.selectGroup}
                  </Button>
                  <FieldGroup className="col-span-2 gap-1 border-t border-border p-2">
                    {items.map((permission) => {
                      const checkboxID = `${generatedID}-${permission.key.replaceAll(":", "-")}`;
                      const checked = selected.has(permission.key);
                      return (
                        <Field
                          key={permission.key}
                          orientation="horizontal"
                          data-disabled={disabled || undefined}
                          className="rounded-md p-2 hover:bg-muted/40"
                        >
                          <Checkbox
                            id={checkboxID}
                            checked={checked}
                            disabled={disabled}
                            onCheckedChange={(nextChecked) => setPermission(permission.key, nextChecked === true)}
                          />
                          <FieldContent className="min-w-0">
                            <FieldLabel htmlFor={checkboxID} className="cursor-pointer font-normal">
                              {formatPermissionDescription(permission)}
                            </FieldLabel>
                            <FieldDescription dir="ltr" className="truncate text-start font-mono" title={permission.key}>
                              {permission.key}
                            </FieldDescription>
                          </FieldContent>
                        </Field>
                      );
                    })}
                  </FieldGroup>
                </FieldSet>
              );
            })}
          </div>
        </div>
      </TabsContent>
    </Tabs>
  );
}

function groupPermissions(
  permissions: PermissionOption[],
  query: string,
  formatGroupLabel: (group: string) => React.ReactNode,
  formatPermissionDescription: (permission: PermissionOption) => React.ReactNode,
) {
  const normalizedQuery = query.trim().toLocaleLowerCase();
  const filtered = normalizedQuery
    ? permissions.filter((permission) => {
        const searchableText = [
          permission.key,
          permission.description,
          textContent(formatGroupLabel(permission.group)),
          textContent(formatPermissionDescription(permission)),
        ].join(" ").toLocaleLowerCase();
        return searchableText.includes(normalizedQuery);
      })
    : permissions;
  const groups = new Map<string, PermissionOption[]>();

  for (const permission of filtered) {
    groups.set(permission.group, [...(groups.get(permission.group) ?? []), permission]);
  }

  return Array.from(groups.entries())
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([group, items]) => [group, [...items].sort((left, right) => left.key.localeCompare(right.key))] as const);
}

function defaultGroupLabel(group: string) {
  return group.replaceAll("_", " ");
}

function textContent(value: React.ReactNode) {
  return typeof value === "string" || typeof value === "number" ? String(value) : "";
}
