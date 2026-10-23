"use client";

import { Boxes, CheckCircle2, ServerCog, Waypoints } from "lucide-react";
import { useState, type KeyboardEvent as ReactKeyboardEvent } from "react";

import { EmptyState } from "@/components/app-ui/EmptyState";
import { SearchPickerDialog } from "@/components/app-ui/SearchPickerDialog";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { cn } from "@/lib/utils";
import type { GatewayModel } from "@/lib/prompt-builder";

import type { PlaygroundMode } from "./domain";

type PlaygroundTranslation = (key: string, values?: Record<string, string | number>) => string;

type ModelPickerFilter = "all" | "virtual" | "provider";

export function ModelPickerDialog(props: { open: boolean; onOpenChange: (open: boolean) => void; search: string; onSearchChange: (value: string) => void; models: GatewayModel[]; mode: PlaygroundMode; selectedModel: string; onSelect: (model: GatewayModel) => void; t: PlaygroundTranslation }) {
  const [filter, setFilter] = useState<ModelPickerFilter>("all");
  const q = props.search.trim().toLowerCase();
  const matchingModels = props.models.filter((model) => `${model.id} ${model.provider ?? ""} ${model.owned_by} ${(model.aliases ?? []).join(" ")}`.toLowerCase().includes(q));
  const models = matchingModels.filter((model) => filter === "all" || (filter === "virtual" ? model.provider === "virtual" : model.provider !== "virtual"));
  const virtualModels = pinSelectedModel(models.filter((model) => model.provider === "virtual"), props.selectedModel);
  const providerModels = pinSelectedModel(models.filter((model) => model.provider !== "virtual"), props.selectedModel);
  const selectedModel = props.models.find((model) => model.id === props.selectedModel);
  const selectedIsVirtual = !selectedModel || selectedModel.provider === "virtual";
  const groups = selectedIsVirtual
    ? [
        { key: "virtual", title: props.t("model.virtualGroup"), models: virtualModels, virtual: true },
        { key: "provider", title: props.t("model.providerGroup"), models: providerModels, virtual: false },
      ]
    : [
        { key: "provider", title: props.t("model.providerGroup"), models: providerModels, virtual: false },
        { key: "virtual", title: props.t("model.virtualGroup"), models: virtualModels, virtual: true },
      ];
  const modeLabel = props.mode === "image_generation" ? props.t("mode.imageGeneration") : props.mode === "layout_parsing" ? props.t("mode.ocr") : props.t("mode.chat");

  function changeOpen(open: boolean) {
    props.onOpenChange(open);
    if (!open) {
      props.onSearchChange("");
      setFilter("all");
    }
  }

  function selectModel(model: GatewayModel) {
    props.onSearchChange("");
    setFilter("all");
    props.onSelect(model);
  }

  return (
    <SearchPickerDialog
      open={props.open}
      onOpenChange={changeOpen}
      title={<span className="flex items-center gap-2"><span className="flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary"><Boxes aria-hidden="true" className="size-4" /></span>{props.t("model.select")}</span>}
      description={props.t("model.pickerDescription", { mode: modeLabel })}
      searchValue={props.search}
      onSearchValueChange={props.onSearchChange}
      searchPlaceholder={props.t("model.search")}
      contentClassName="overflow-hidden"
      bodyClassName="min-h-0"
    >
      <ToggleGroup
        type="single"
        value={filter}
        onValueChange={(value) => value && setFilter(value as ModelPickerFilter)}
        aria-label={props.t("model.filterLabel")}
        className="w-full gap-1 rounded-lg border border-border bg-muted/30 p-1"
      >
        <ToggleGroupItem value="all" className="h-8 flex-1 px-3 data-[state=on]:bg-background data-[state=on]:shadow-xs">
          {props.t("model.filterAll")}
        </ToggleGroupItem>
        <ToggleGroupItem value="virtual" className="h-8 flex-1 px-3 data-[state=on]:bg-background data-[state=on]:shadow-xs">
          {props.t("model.filterVirtual")}
        </ToggleGroupItem>
        <ToggleGroupItem value="provider" className="h-8 flex-1 px-3 data-[state=on]:bg-background data-[state=on]:shadow-xs">
          {props.t("model.filterProvider")}
        </ToggleGroupItem>
      </ToggleGroup>
      <ScrollArea className="h-80 pe-2">
        <div data-model-picker-list className="flex flex-col gap-4">
          {groups.map((group) => (
            <ModelPickerGroup
              key={group.key}
              title={group.title}
              models={group.models}
              selectedModel={props.selectedModel}
              onSelect={selectModel}
              t={props.t}
              virtual={group.virtual}
            />
          ))}
          {models.length === 0 ? <EmptyState variant="inline" density="compact" icon={Boxes} title={props.t("model.empty")} /> : null}
        </div>
      </ScrollArea>
    </SearchPickerDialog>
  );
}

function pinSelectedModel(models: GatewayModel[], selectedModel: string): GatewayModel[] {
  return [...models].sort((first, second) => {
    if (first.id === selectedModel) return -1;
    if (second.id === selectedModel) return 1;
    return first.id.localeCompare(second.id);
  });
}

function focusAdjacentModel(event: ReactKeyboardEvent<HTMLButtonElement>) {
  if (!["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key)) return;
  const list = event.currentTarget.closest("[data-model-picker-list]");
  const options = list ? [...list.querySelectorAll<HTMLButtonElement>("[data-model-picker-option]")] : [];
  const currentIndex = options.indexOf(event.currentTarget);
  if (currentIndex < 0 || options.length === 0) return;

  event.preventDefault();
  const nextIndex = event.key === "Home"
    ? 0
    : event.key === "End"
      ? options.length - 1
      : (currentIndex + (event.key === "ArrowDown" ? 1 : -1) + options.length) % options.length;
  options[nextIndex]?.focus();
}

function ModelPickerGroup(props: { title: string; models: GatewayModel[]; selectedModel: string; onSelect: (model: GatewayModel) => void; t: PlaygroundTranslation; virtual?: boolean }) {
  if (props.models.length === 0) return null;
  const GroupIcon = props.virtual ? Waypoints : ServerCog;
  return (
    <section className="flex flex-col gap-2" aria-label={props.title}>
      <div className="flex items-center justify-between gap-3 px-1">
        <h3 className="flex items-center gap-2 text-sm font-semibold"><GroupIcon aria-hidden="true" className={cn("size-4", props.virtual ? "text-info-soft-foreground" : "text-primary")} />{props.title}</h3>
        <Badge variant="secondary" className="tabular-nums">{props.models.length}</Badge>
      </div>
      <div className="overflow-hidden rounded-lg border border-border bg-card divide-y divide-border">
        {props.models.map((model) => <ModelPickerOption key={model.id} model={model} selected={model.id === props.selectedModel} onSelect={props.onSelect} t={props.t} />)}
      </div>
    </section>
  );
}

function ModelPickerOption(props: { model: GatewayModel; selected: boolean; onSelect: (model: GatewayModel) => void; t: PlaygroundTranslation }) {
  const virtual = props.model.provider === "virtual";
  const TypeIcon = virtual ? Waypoints : ServerCog;
  const identity = modelPickerIdentity(props.model);
  return (
    <button
      type="button"
      aria-pressed={props.selected}
      data-model-picker-option
      className={cn(
        "group flex w-full items-center gap-3 px-3 py-2.5 text-start transition-colors hover:bg-accent/50 focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring focus-visible:outline-none",
        props.selected && "bg-primary-soft",
      )}
      onClick={() => props.onSelect(props.model)}
      onKeyDown={focusAdjacentModel}
    >
      <span className={cn("flex size-9 shrink-0 items-center justify-center rounded-lg border", virtual ? "border-info-soft-border bg-info-soft text-info-soft-foreground" : "border-primary/20 bg-primary/10 text-primary")}>
        <TypeIcon aria-hidden="true" className="size-4" />
      </span>
      <span className="min-w-0 flex-1">
        <code className="block truncate text-end font-mono text-sm font-semibold text-foreground" dir="ltr" title={props.model.id}>{identity.modelName}</code>
        {identity.metadata ? <span className="mt-1 block truncate text-end text-xs text-muted-foreground" dir="ltr">{identity.metadata}</span> : null}
      </span>
      {props.selected ? <span className="flex shrink-0 items-center gap-1 text-xs font-medium text-primary"><CheckCircle2 aria-hidden="true" className="size-4" />{props.t("model.selected")}</span> : null}
    </button>
  );
}

function modelPickerIdentity(model: GatewayModel): { modelName: string; metadata: string } {
  if (model.provider === "virtual") {
    const alias = model.aliases?.find((item) => item && item !== model.id) ?? "";
    return { modelName: model.id, metadata: alias };
  }

  const parts = model.id.split(":");
  const provider = model.provider || (parts.length === 3 ? parts[0] : model.owned_by);
  const account = parts.length === 3 ? parts[1] : "";
  let modelName = parts.length === 3 ? parts[2] : model.id;
  const providerPrefix = `${provider}/`;
  if (modelName.toLocaleLowerCase().startsWith(providerPrefix.toLocaleLowerCase())) {
    modelName = modelName.slice(providerPrefix.length);
  }

  return { modelName, metadata: [provider, account].filter(Boolean).join(" · ") };
}

