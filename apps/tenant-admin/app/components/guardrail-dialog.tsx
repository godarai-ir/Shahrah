"use client";

import { AlertCircle } from "lucide-react";

import { AppSelectField, CheckboxOptionList, FormDialog } from "@/components/app-ui";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

import type { BuilderState, GuardrailItem, PromptTranslationFn } from "./prompt-management-shared";

type GuardrailDialogProps = {
  builder: BuilderState;
  guardrails: GuardrailItem[];
  isLoading: boolean;
  mode: "input" | "output" | null;
  onBuilderChange: (updater: (current: BuilderState) => BuilderState) => void;
  setMode: (mode: "input" | "output" | null) => void;
  t: PromptTranslationFn;
};

export function GuardrailDialog(props: GuardrailDialogProps) {
  const selectedGuardrailMode = props.mode === "input" ? props.builder.inputGuardrails : props.builder.outputGuardrails;

  return (
    <FormDialog
      open={props.mode !== null}
      onOpenChange={(open) => { if (!open) props.setMode(null); }}
      title={props.mode === "input" ? props.t("dialogs.guardrails.inputTitle") : props.t("dialogs.guardrails.outputTitle")}
      description={props.mode === "input" ? props.t("dialogs.guardrails.inputDescription") : props.t("dialogs.guardrails.outputDescription")}
      size="md"
      footer={<Button type="button" variant="outline" onClick={() => props.setMode(null)}>{props.t("actions.close")}</Button>}
    >
          <div className="flex flex-col gap-3">
            {props.mode === "output" ? (
              <Alert variant={props.builder.stream ? "destructive" : "default"}>
                <AlertCircle className="size-4" />
                <AlertTitle>
                  {props.builder.stream ? props.t("dialogs.guardrails.streamingDisabledTitle") : props.t("dialogs.guardrails.streamingAlreadyDisabledTitle")}
                </AlertTitle>
                <AlertDescription>
                  {props.builder.stream
                    ? props.t("dialogs.guardrails.streamingDisabledDescription")
                    : props.t("dialogs.guardrails.streamingAlreadyDisabledDescription")}
                </AlertDescription>
              </Alert>
            ) : null}
            <AppSelectField
              label={props.t("dialogs.guardrails.scope")}
              value={props.builder.guardrailScope}
              onValueChange={(value) => props.onBuilderChange((current) => ({ ...current, guardrailScope: value as "all" | "last" }))}
              options={[
                { value: "last", label: props.t("dialogs.guardrails.scopeLast") },
                { value: "all", label: props.t("dialogs.guardrails.scopeAll") },
              ]}
            />
            <CheckboxOptionList
              label={props.t("dialogs.guardrails.available")}
              options={props.guardrails.map((guardrail) => ({ value: guardrail.selector || `guardrails/${guardrail.slug}`, label: guardrail.name, description: guardrail.description || guardrail.type || guardrail.selector }))}
              value={selectedGuardrailMode}
              onValueChange={(nextValues) => props.onBuilderChange((current) => ({ ...current, [props.mode === "input" ? "inputGuardrails" : "outputGuardrails"]: nextValues }))}
              emptyText={props.t("dialogs.guardrails.empty")}
              loading={props.isLoading}
            />
          </div>
    </FormDialog>
  );
}
