"use client";

import { AppCodeTextareaField } from "@/components/app-ui/AppCodeTextareaField";
import { AppSelectField } from "@/components/app-ui/AppSelectField";
import { AppSwitchField } from "@/components/app-ui/AppSwitchField";
import { FormDialog } from "@/components/app-ui/FormDialog";
import { Button } from "@/components/ui/button";

import type { BuilderState, PromptTranslationFn } from "./prompt-management-shared";

type StructuredOutputDialogProps = {
  builder: BuilderState;
  onBuilderChange: (updater: (current: BuilderState) => BuilderState) => void;
  open: boolean;
  setOpen: (open: boolean) => void;
  t: PromptTranslationFn;
};

export function StructuredOutputDialog(props: StructuredOutputDialogProps) {
  return (
    <FormDialog open={props.open} onOpenChange={props.setOpen} title={props.t("dialogs.structured.title")} description={props.t("dialogs.structured.description")} size="lg" footer={<Button type="button" variant="outline" onClick={() => props.setOpen(false)}>{props.t("actions.close")}</Button>}>
          <AppSwitchField label={props.t("dialogs.structured.enableTitle")} description={props.t("dialogs.structured.enableDescription")} checked={props.builder.structuredEnabled} onCheckedChange={(checked) => props.onBuilderChange((current) => ({ ...current, structuredEnabled: checked }))} />
          <div className="grid gap-3 sm:grid-cols-2">
            <AppSelectField
              label={props.t("dialogs.structured.fields.format")}
              value={props.builder.structuredFormat}
              onValueChange={(value) => props.onBuilderChange((current) => ({ ...current, structuredFormat: value as "json_schema" | "json_object" }))}
              options={[
                { value: "json_schema", label: props.t("common.jsonSchema") },
                { value: "json_object", label: props.t("common.jsonObject") },
              ]}
            />
            <AppSwitchField label={props.t("dialogs.structured.fields.strict")} description={props.builder.structuredStrict ? props.t("common.enabled") : props.t("common.disabled")} checked={props.builder.structuredStrict} onCheckedChange={(checked) => props.onBuilderChange((current) => ({ ...current, structuredStrict: checked }))} />
          </div>
          <AppCodeTextareaField label={props.t("dialogs.structured.fields.schema")} value={props.builder.structuredSchema} onChange={(event) => props.onBuilderChange((current) => ({ ...current, structuredSchema: event.target.value }))} className="min-h-72" />
    </FormDialog>
  );
}
