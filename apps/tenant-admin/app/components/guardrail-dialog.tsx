"use client";

import { AlertCircle } from "lucide-react";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";

import { Field } from "./prompt-ui";
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
    <Dialog open={props.mode !== null} onOpenChange={(open) => { if (!open) props.setMode(null); }}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{props.mode === "input" ? props.t("dialogs.guardrails.inputTitle") : props.t("dialogs.guardrails.outputTitle")}</DialogTitle>
          <DialogDescription>
            {props.mode === "input" ? props.t("dialogs.guardrails.inputDescription") : props.t("dialogs.guardrails.outputDescription")}
          </DialogDescription>
        </DialogHeader>
        {props.isLoading ? (
          <div className="space-y-2">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        ) : (
          <div className="space-y-3">
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
            <Field label={props.t("dialogs.guardrails.scope")}>
              <select
                value={props.builder.guardrailScope}
                onChange={(event) => props.onBuilderChange((current) => ({ ...current, guardrailScope: event.target.value as "all" | "last" }))}
                className="h-10 rounded-md border bg-background px-3 text-sm"
              >
                <option value="last">{props.t("dialogs.guardrails.scopeLast")}</option>
                <option value="all">{props.t("dialogs.guardrails.scopeAll")}</option>
              </select>
            </Field>
            <div className="max-h-[360px] space-y-2 overflow-y-auto rounded-lg border p-2">
              {props.guardrails.map((guardrail) => {
                const selector = guardrail.selector || `guardrails/${guardrail.slug}`;
                const checked = selectedGuardrailMode.includes(selector);
                return (
                  <label key={guardrail.id || guardrail.slug} className="flex items-start gap-3 rounded-md border p-3">
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={(event) => {
                        props.onBuilderChange((current) => {
                          const targetKey = props.mode === "input" ? "inputGuardrails" : "outputGuardrails";
                          const currentValues = current[targetKey];
                          const nextValues = event.target.checked
                            ? [...currentValues, selector]
                            : currentValues.filter((value) => value !== selector);
                          return { ...current, [targetKey]: nextValues };
                        });
                      }}
                      className="mt-1"
                    />
                    <span className="min-w-0 flex-1">
                      <span className="block font-medium">{guardrail.name}</span>
                      <span className="block text-xs text-muted-foreground">{guardrail.description || guardrail.type || selector}</span>
                    </span>
                  </label>
                );
              })}
            </div>
          </div>
        )}
        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => props.setMode(null)}>
            {props.t("actions.close")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
