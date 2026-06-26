"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Switch } from "@/components/ui/switch";
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
          <div className="flex items-center justify-between rounded-md border px-3 py-2">
            <div>
              <div className="text-sm font-medium">{props.t("dialogs.structured.enableTitle")}</div>
              <div className="text-xs text-muted-foreground">{props.t("dialogs.structured.enableDescription")}</div>
            </div>
            <Switch
              checked={props.builder.structuredEnabled}
              onCheckedChange={(checked) => props.onBuilderChange((current) => ({ ...current, structuredEnabled: checked }))}
            />
          </div>
          <div className="grid gap-3 sm:grid-cols-2">
            <Field label={props.t("dialogs.structured.fields.format")}>
              <select
                value={props.builder.structuredFormat}
                onChange={(event) => props.onBuilderChange((current) => ({ ...current, structuredFormat: event.target.value as "json_schema" | "json_object" }))}
                className="h-10 rounded-md border bg-background px-3 text-sm"
              >
                <option value="json_schema">{props.t("common.jsonSchema")}</option>
                <option value="json_object">{props.t("common.jsonObject")}</option>
              </select>
            </Field>
            <Field label={props.t("dialogs.structured.fields.strict")}>
              <div className="flex h-10 items-center justify-between rounded-md border px-3">
                <span className="text-sm text-muted-foreground">{props.builder.structuredStrict ? props.t("common.enabled") : props.t("common.disabled")}</span>
                <Switch
                  checked={props.builder.structuredStrict}
                  onCheckedChange={(checked) => props.onBuilderChange((current) => ({ ...current, structuredStrict: checked }))}
                />
              </div>
            </Field>
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
