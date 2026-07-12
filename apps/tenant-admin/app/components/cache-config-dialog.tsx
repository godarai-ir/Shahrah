"use client";

import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";

import { Field } from "./prompt-ui";
import type { BuilderState, PromptTranslationFn } from "./prompt-management-shared";

type CacheConfigDialogProps = {
  builder: BuilderState;
  onBuilderChange: (updater: (current: BuilderState) => BuilderState) => void;
  open: boolean;
  setOpen: (open: boolean) => void;
  t: PromptTranslationFn;
};

export function CacheConfigDialog(props: CacheConfigDialogProps) {
  return (
    <Dialog open={props.open} onOpenChange={props.setOpen}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{props.t("dialogs.cache.title")}</DialogTitle>
          <DialogDescription>{props.t("dialogs.cache.description")}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="flex items-center justify-between rounded-md border px-3 py-2">
            <div>
              <div className="text-sm font-medium">{props.t("dialogs.cache.enableTitle")}</div>
              <div className="text-xs text-muted-foreground">{props.t("dialogs.cache.enableDescription")}</div>
            </div>
            <Switch
              checked={props.builder.cacheEnabled}
              onCheckedChange={(checked) => props.onBuilderChange((current) => ({ ...current, cacheEnabled: checked }))}
            />
          </div>
          <AppSelectField
            label={props.t("dialogs.cache.fields.type")}
            value={props.builder.cacheType}
            onValueChange={(value) => props.onBuilderChange((current) => ({ ...current, cacheType: value as "simple" | "smart" }))}
            options={[
              { value: "simple", label: props.t("common.simple") },
              { value: "smart", label: props.t("common.smart") },
            ]}
          />
          <Field label={props.t("dialogs.cache.fields.ttl")}>
            <Input value={props.builder.cacheTTL} onChange={(event) => props.onBuilderChange((current) => ({ ...current, cacheTTL: event.target.value }))} />
          </Field>
          <Field label={props.t("dialogs.cache.fields.namespace")}>
            <Input value={props.builder.cacheNamespace} onChange={(event) => props.onBuilderChange((current) => ({ ...current, cacheNamespace: event.target.value }))} />
          </Field>
          {props.builder.cacheType === "smart" ? (
            <Field label={props.t("dialogs.cache.fields.similarity")}>
              <Input
                value={props.builder.cacheSimilarityThreshold}
                onChange={(event) => props.onBuilderChange((current) => ({ ...current, cacheSimilarityThreshold: event.target.value }))}
              />
            </Field>
          ) : null}
        </div>
        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => props.setOpen(false)}>
            {props.t("actions.close")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
