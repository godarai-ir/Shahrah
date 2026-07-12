"use client";

import { AppSelectField, AppSwitchField, AppTextField, FormDialog } from "@/components/app-ui";
import { Button } from "@/components/ui/button";

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
    <FormDialog open={props.open} onOpenChange={props.setOpen} title={props.t("dialogs.cache.title")} description={props.t("dialogs.cache.description")} footer={<Button type="button" variant="outline" onClick={() => props.setOpen(false)}>{props.t("actions.close")}</Button>}>
          <AppSwitchField label={props.t("dialogs.cache.enableTitle")} description={props.t("dialogs.cache.enableDescription")} checked={props.builder.cacheEnabled} onCheckedChange={(checked) => props.onBuilderChange((current) => ({ ...current, cacheEnabled: checked }))} />
          <AppSelectField
            label={props.t("dialogs.cache.fields.type")}
            value={props.builder.cacheType}
            onValueChange={(value) => props.onBuilderChange((current) => ({ ...current, cacheType: value as "simple" | "smart" }))}
            options={[
              { value: "simple", label: props.t("common.simple") },
              { value: "smart", label: props.t("common.smart") },
            ]}
          />
          <AppTextField label={props.t("dialogs.cache.fields.ttl")} value={props.builder.cacheTTL} onChange={(event) => props.onBuilderChange((current) => ({ ...current, cacheTTL: event.target.value }))} />
          <AppTextField label={props.t("dialogs.cache.fields.namespace")} value={props.builder.cacheNamespace} onChange={(event) => props.onBuilderChange((current) => ({ ...current, cacheNamespace: event.target.value }))} />
          {props.builder.cacheType === "smart" ? (
            <AppTextField label={props.t("dialogs.cache.fields.similarity")} value={props.builder.cacheSimilarityThreshold} onChange={(event) => props.onBuilderChange((current) => ({ ...current, cacheSimilarityThreshold: event.target.value }))} />
          ) : null}
    </FormDialog>
  );
}
