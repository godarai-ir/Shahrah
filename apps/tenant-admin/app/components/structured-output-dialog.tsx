"use client";

import { AppSelectField, AppSwitchField } from "@/components/app-ui";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";

import { Field } from "./prompt-ui";
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
    <Dialog open={props.open} onOpenChange={props.setOpen}>
      <DialogContent className="w-[min(960px,95vw)] max-w-[960px]">
        <DialogHeader>
          <DialogTitle>{props.t("dialogs.structured.title")}</DialogTitle>
          <DialogDescription>{props.t("dialogs.structured.description")}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
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
          <Field label={props.t("dialogs.structured.fields.schema")}>
            <Textarea
              value={props.builder.structuredSchema}
              onChange={(event) => props.onBuilderChange((current) => ({ ...current, structuredSchema: event.target.value }))}
              className="min-h-[300px] font-mono text-xs"
              dir="ltr"
              spellCheck={false}
            />
          </Field>
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
