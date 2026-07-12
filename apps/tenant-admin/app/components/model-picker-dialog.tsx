"use client";

import { Check } from "lucide-react";
import { useMemo, useState } from "react";

import type { ProviderCatalogModel } from "@/app/components/provider-account-wizard/types";
import { SearchPickerDialog } from "@/components/app-ui";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

import { formatNumber, type BuilderState, type GatewayModel, type PromptTranslationFn } from "./prompt-management-shared";

type ModelPickerDialogProps = {
  builder: BuilderState;
  locale: string;
  models: GatewayModel[];
  onBuilderChange: (updater: (current: BuilderState) => BuilderState) => void;
  open: boolean;
  providerCatalog: Array<{ id: string; label: string; description: string }>;
  providerModels: ProviderCatalogModel[];
  setOpen: (open: boolean) => void;
  t: PromptTranslationFn;
};

export function ModelPickerDialog(props: ModelPickerDialogProps) {
  const [modelSearch, setModelSearch] = useState("");
  const [selectedProviderFilter, setSelectedProviderFilter] = useState("");

  const actualProviderGroups = useMemo(() => {
    const grouped = new Map<string, GatewayModel[]>();
    for (const model of props.models) {
      const key = (model.provider || "other").trim().toLowerCase();
      const list = grouped.get(key) ?? [];
      list.push(model);
      grouped.set(key, list);
    }
    return grouped;
  }, [props.models]);

  const visibleProviderID = selectedProviderFilter || props.builder.selectedProvider || props.providerCatalog[0]?.id || "";

  const visibleModels = useMemo(() => {
    const source = actualProviderGroups.get(visibleProviderID) ?? props.models;
    const q = modelSearch.trim().toLowerCase();
    if (!q) return source;
    return source.filter((item) => {
      const aliases = item.aliases?.join(" ") ?? "";
      return `${item.id} ${aliases}`.toLowerCase().includes(q);
    });
  }, [actualProviderGroups, modelSearch, props.models, visibleProviderID]);

  const visibleProviderCatalogModels = useMemo(() => {
    const source = props.providerModels.filter((item) => item.provider.trim().toLowerCase() === visibleProviderID);
    const byModel = new Map(source.map((item) => [item.model, item]));
    return visibleModels.map((item) => byModel.get(item.id) ?? byModel.get(item.aliases?.[0] ?? "") ?? null);
  }, [props.providerModels, visibleModels, visibleProviderID]);

  return (
    <SearchPickerDialog open={props.open} onOpenChange={props.setOpen} title={props.t("dialogs.model.title")} description={props.t("dialogs.model.description")} searchValue={modelSearch} onSearchValueChange={setModelSearch} searchPlaceholder={props.t("dialogs.model.searchPlaceholder")} size="xl">
          <div className="grid min-h-[380px] gap-3 md:grid-cols-[220px_minmax(0,1fr)]">
            <div className="space-y-1 rounded-lg border bg-muted/20 p-2">
              {props.providerCatalog.map((provider) => (
                <button
                  key={provider.id}
                  type="button"
                  className={cn(
                    "w-full rounded-md px-3 py-2 text-start text-sm transition hover:bg-background",
                    visibleProviderID === provider.id && "bg-primary/10 text-primary ring-1 ring-primary/30",
                  )}
                  onClick={() => setSelectedProviderFilter(provider.id)}
                >
                  <div className="font-medium">{provider.label}</div>
                  {provider.description ? <div className="line-clamp-2 text-xs text-muted-foreground">{provider.description}</div> : null}
                </button>
              ))}
            </div>
            <div className="space-y-2 rounded-lg border p-2">
              {visibleModels.length === 0 ? (
                <div className="flex h-full items-center justify-center text-sm text-muted-foreground">{props.t("dialogs.model.empty")}</div>
              ) : (
                visibleModels.map((model, index) => {
                  const catalog = visibleProviderCatalogModels[index];
                  const isSelected = props.builder.selectedModel === model.id;
                  return (
                    <button
                      key={model.id}
                      type="button"
                      className={cn(
                        "w-full rounded-md border p-3 text-start transition hover:bg-accent",
                        isSelected && "border-primary/40 bg-primary/5",
                      )}
                      onClick={() => {
                        props.onBuilderChange((current) => ({
                          ...current,
                          selectedModel: model.id,
                          selectedProvider: visibleProviderID,
                        }));
                        props.setOpen(false);
                      }}
                    >
                      <div className="flex items-center justify-between gap-3">
                        <div className="min-w-0">
                          <div className="truncate font-medium">{catalog?.display_name || model.id}</div>
                          <div className="truncate text-xs text-muted-foreground" dir="ltr">
                            {model.id}
                          </div>
                        </div>
                        {isSelected ? <Check className="size-4 shrink-0 text-primary" /> : null}
                      </div>
                      {catalog ? (
                        <div className="mt-2 flex flex-wrap gap-2">
                          {catalog.context_window ? (
                            <Badge variant="outline">{props.t("config.model.contextBadge", { count: formatNumber(catalog.context_window, props.locale) })}</Badge>
                          ) : null}
                          {(catalog.capabilities ?? []).slice(0, 4).map((capability) => (
                            <Badge key={`${model.id}-${capability}`} variant="secondary">
                              {capability}
                            </Badge>
                          ))}
                        </div>
                      ) : (
                        <div className="mt-2 flex flex-wrap gap-2">
                          {(model.aliases ?? []).slice(0, 3).map((alias) => (
                            <Badge key={`${model.id}-${alias}`} variant="outline">
                              {alias}
                            </Badge>
                          ))}
                        </div>
                      )}
                    </button>
                  );
                })
              )}
            </div>
          </div>
    </SearchPickerDialog>
  );
}
