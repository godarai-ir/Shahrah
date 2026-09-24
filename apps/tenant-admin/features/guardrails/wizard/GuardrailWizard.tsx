"use client";

import { useEffect, useReducer, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { ArrowLeft, ArrowRight, Check, X } from "lucide-react";

import { OperationNotice } from "@/components/app-ui/OperationNotice";
import { UnsavedChangesDialog } from "@/components/app-ui/UnsavedChangesDialog";
import { showFeedbackToast } from "@/components/app-ui/feedback-toast";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { getLocaleDirection } from "@/lib/locale-direction";
import { cn } from "@/lib/utils";

import { createGuardrail, getGuardrail, listGuardrails, updateGuardrail } from "../api";
import { localizedError } from "../errors";
import { buildPayload, firstInvalidStep, formFromGuardrail, type GuardrailDraft, validateStep, WIZARD_STEPS, type WizardStep } from "../form";
import { guardrailSentence } from "../sentence";
import type { GuardrailFormInput } from "../types";
import { useGuardrailFormat } from "../use-guardrail-format";
import { StepBehavior } from "./StepBehavior";
import { StepConfig } from "./StepConfig";
import { StepReview } from "./StepReview";
import { StepType } from "./StepType";
import { SummaryPanel } from "./SummaryPanel";
import { TestPanel } from "./TestPanel";
import { createWizardState, editWizardState, type WizardMode, wizardReducer, type WizardState } from "./state";

const LIST_HREF = "/guardrails";

type GuardrailWizardProps = { mode: "create" } | { mode: "edit"; slug: string };

function withType(form: GuardrailDraft): GuardrailFormInput | null {
  return form.type ? { ...form, type: form.type } : null;
}

type StepperProps = {
  state: WizardState;
  reachable: (step: WizardStep) => boolean;
  onGoTo: (step: WizardStep) => void;
};

function WizardStepper({ state, reachable, onGoTo }: StepperProps) {
  const t = useTranslations("guardrails");
  const format = useGuardrailFormat();
  const hint = (step: WizardStep, done: boolean) => {
    if (step === 0) return state.form.type ? t(`types.${state.form.type}`) : t("wizard.stepHints.type");
    if (step === 1) return done ? t("wizard.stepHints.done") : t("wizard.stepHints.config");
    if (step === 2) return done ? t(`enforcement.${state.form.enforcement}`) : t("wizard.stepHints.behavior");

    return t("wizard.stepHints.review");
  };

  return (
    <nav aria-label={t("wizard.stepsLabel")} className="min-w-0 flex-1">
      <ol className="grid grid-cols-2 gap-2 md:grid-cols-4">
        {WIZARD_STEPS.map((key, index) => {
          const step = index as WizardStep;
          const current = step === state.step;
          const done = step < state.step || (state.mode === "edit" && step === 0);
          const enabled = !current && reachable(step);

          return (
            <li key={key} className="min-w-0">
              <button
                type="button"
                aria-current={current ? "step" : undefined}
                disabled={!enabled && !current}
                onClick={() => (enabled ? onGoTo(step) : undefined)}
                className={cn(
                  "flex w-full min-w-0 items-center gap-3 border-b-2 border-transparent px-1 pt-2 pb-3 text-start transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring",
                  current && "border-primary",
                  done && !current && "border-primary-subtle-border",
                  enabled ? "cursor-pointer hover:border-primary-subtle-border" : "cursor-default",
                )}
              >
                <span
                  aria-hidden="true"
                  className={cn(
                    "flex size-7 shrink-0 items-center justify-center rounded-full border text-xs font-semibold tabular-nums",
                    done ? "border-primary bg-primary text-primary-foreground" : current ? "border-primary bg-primary-soft text-primary" : "border-input-border bg-card text-muted-foreground",
                  )}
                >
                  {done ? <Check className="size-3.5" strokeWidth={3} /> : format.formatNumber(index + 1)}
                </span>
                <span className="flex min-w-0 flex-col">
                  <span className={cn("truncate text-sm", current ? "font-semibold text-foreground" : done ? "font-medium text-foreground" : "font-medium text-muted-foreground")}>
                    {t(`wizard.steps.${key}`)}
                  </span>
                  <span className="truncate text-xs text-muted-foreground">{hint(step, done)}</span>
                </span>
              </button>
            </li>
          );
        })}
      </ol>
    </nav>
  );
}

function DoneView({ name, sentence, slug, onAnother }: { name: string; sentence: string; slug: string; onAnother: () => void }) {
  const t = useTranslations("guardrails");
  const headingRef = useRef<HTMLHeadingElement>(null);

  useEffect(() => {
    headingRef.current?.focus();
  }, []);

  return (
    <Card className="mx-auto w-full max-w-lg">
      <CardContent className="flex flex-col items-center gap-4 p-8 text-center md:p-8">
        <span aria-hidden="true" className="flex size-12 items-center justify-center rounded-full border border-success-soft-border bg-success-soft text-success-soft-foreground">
          <Check className="size-6" strokeWidth={2.5} />
        </span>
        <h2 ref={headingRef} tabIndex={-1} className="text-lg font-semibold outline-none">{t("wizard.done.title", { name })}</h2>
        <p className="text-sm leading-7 text-muted-foreground">{sentence}</p>
        <div className="mt-2 flex flex-wrap justify-center gap-2">
          <Button asChild>
            <Link href={`${LIST_HREF}?created=${encodeURIComponent(slug)}`}>{t("wizard.done.back")}</Link>
          </Button>
          <Button type="button" variant="outline" onClick={onAnother}>{t("wizard.done.another")}</Button>
        </div>
      </CardContent>
    </Card>
  );
}

function WizardFlow({ initialState, existingSlugs, originalSlug }: { initialState: WizardState; existingSlugs: readonly string[]; originalSlug?: string }) {
  const t = useTranslations("guardrails");
  const locale = useLocale();
  const direction = getLocaleDirection(locale);
  const router = useRouter();
  const format = useGuardrailFormat();
  const [state, dispatch] = useReducer(wizardReducer, initialState);
  const [isSaving, setIsSaving] = useState(false);
  const [submitError, setSubmitError] = useState("");
  const [discardOpen, setDiscardOpen] = useState(false);
  const [done, setDone] = useState<{ name: string; slug: string; sentence: string } | null>(null);
  const headingRef = useRef<HTMLHeadingElement>(null);
  const previousStep = useRef(state.step);

  const mode: WizardMode = state.mode;
  const context = { mode, existingSlugs };
  const translator = format.t;
  const form = withType(state.form);
  const sentence = form ? guardrailSentence(form, format) : "";
  const stepError = state.attempted ? validateStep(state.step, state.form, translator, context) : "";
  const lastStep = state.step === 3;
  const minStep: WizardStep = mode === "edit" ? 1 : 0;
  const sample = state.sample ?? (form && t.has(`wizard.test.samples.${form.type}`) ? t(`wizard.test.samples.${form.type}`) : "");
  const notice = submitError
    ? { kind: "error" as const, message: submitError }
    : stepError ? { kind: "error" as const, message: stepError } : null;

  useEffect(() => {
    if (previousStep.current === state.step) return;
    previousStep.current = state.step;
    window.scrollTo(0, 0);
    headingRef.current?.focus({ preventScroll: true });
  }, [state.step]);

  function reachable(step: WizardStep) {
    if (step < minStep) return false;
    if (mode === "edit" || step <= state.step) return true;
    const steps = [0, 1, 2, 3].slice(0, step) as WizardStep[];

    return steps.every((previous) => !validateStep(previous, state.form, translator, context));
  }

  function goTo(step: WizardStep) {
    setSubmitError("");
    dispatch({ type: "goTo", step });
  }

  async function submit(complete: GuardrailFormInput) {
    setIsSaving(true);
    setSubmitError("");
    try {
      if (mode === "edit" && originalSlug) {
        await updateGuardrail(originalSlug, buildPayload(complete, "edit"));
        showFeedbackToast({ kind: "success", title: t("alerts.success"), description: t("messages.saveSuccess") });
        router.push(`${LIST_HREF}?updated=${encodeURIComponent(originalSlug)}`);
        return;
      }
      await createGuardrail(buildPayload(complete, "create"));
      setDone({ name: complete.name.trim(), slug: complete.slug.trim(), sentence: guardrailSentence(complete, format) });
      window.scrollTo(0, 0);
    } catch (error) {
      setSubmitError(localizedError(error, t("messages.saveError"), translator));
    } finally {
      setIsSaving(false);
    }
  }

  function next() {
    if (isSaving) return;
    if (validateStep(state.step, state.form, translator, context)) {
      dispatch({ type: "attempt" });
      return;
    }
    if (!lastStep) {
      goTo((state.step + 1) as WizardStep);
      return;
    }
    const invalid = firstInvalidStep(state.form, translator, context);
    if (invalid !== null) {
      goTo(invalid);
      dispatch({ type: "attempt" });
      return;
    }
    if (form) void submit(form);
  }

  function cancel() {
    if (state.dirty) {
      setDiscardOpen(true);
      return;
    }
    router.push(LIST_HREF);
  }

  if (done) {
    return (
      <DoneView
        name={done.name}
        slug={done.slug}
        sentence={done.sentence}
        onAnother={() => {
          setDone(null);
          dispatch({ type: "reset", state: createWizardState() });
        }}
      />
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start gap-3 border-b border-border">
        <WizardStepper state={state} reachable={reachable} onGoTo={goTo} />
        <Button type="button" variant="outline" size="sm" className="mt-2 shrink-0" onClick={cancel}>
          <X data-icon="inline-start" aria-hidden="true" />
          {t("wizard.actions.cancel")}
        </Button>
      </div>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_21.25rem]">
        <section className="flex min-w-0 flex-col gap-4">
          {state.step === 0 ? (
            <StepType
              headingRef={headingRef}
              selected={state.form.type}
              query={state.typeQuery}
              category={state.category}
              onQueryChange={(query) => dispatch({ type: "setTypeQuery", query })}
              onCategoryChange={(category) => dispatch({ type: "setCategory", category })}
              onSelect={(guardrailType) => dispatch({ type: "selectType", guardrailType })}
            />
          ) : null}
          {state.step === 1 && form ? (
            <StepConfig headingRef={headingRef} form={form} dir={direction} onChange={(patch) => dispatch({ type: "update", patch })} />
          ) : null}
          {state.step === 2 && form ? (
            <StepBehavior headingRef={headingRef} form={form} mode={mode} dir={direction} onChange={(patch) => dispatch({ type: "update", patch })} />
          ) : null}
          {state.step === 3 && form ? (
            <StepReview
              headingRef={headingRef}
              form={form}
              mode={mode}
              sentence={sentence}
              existingSlugs={existingSlugs}
              format={format}
              onNameChange={(name) => dispatch({ type: "setName", name })}
              onSlugChange={(slug) => dispatch({ type: "setSlug", slug })}
              onDescriptionChange={(description) => dispatch({ type: "update", patch: { description } })}
              onEditStep={goTo}
            />
          ) : null}

          <OperationNotice notice={notice} labels={{ success: t("alerts.success"), error: t("alerts.error"), warning: t("alerts.error"), info: t("alerts.success") }} />
        </section>

        <aside className="flex min-w-0 flex-col gap-4 lg:sticky lg:top-20">
          <SummaryPanel type={state.form.type} name={state.form.name} sentence={sentence} />
          {form && state.step >= 1 ? (
            <TestPanel form={form} sample={sample} format={format} onSampleChange={(value) => dispatch({ type: "setSample", sample: value })} />
          ) : null}
        </aside>
      </div>

      <footer className="sticky bottom-0 z-10 -mx-4 flex items-center gap-3 border-t border-border bg-background px-4 py-3 md:-mx-6 md:px-6">
        <span className="text-sm text-muted-foreground">
          {t("wizard.stepOf", { current: format.formatNumber(state.step + 1), total: format.formatNumber(WIZARD_STEPS.length) })}
        </span>
        <div className="ms-auto flex gap-2">
          {state.step > minStep ? (
            <Button type="button" variant="outline" disabled={isSaving} onClick={() => goTo((state.step - 1) as WizardStep)}>
              <ArrowRight data-icon="inline-start" aria-hidden="true" className="ltr:rotate-180" />
              {t("wizard.actions.back")}
            </Button>
          ) : null}
          <Button type="button" disabled={isSaving} aria-busy={isSaving || undefined} onClick={next}>
            {isSaving ? <Spinner data-icon="inline-start" /> : null}
            {lastStep
              ? isSaving
                ? mode === "edit" ? t("actions.saving") : t("wizard.actions.creating")
                : mode === "edit" ? t("wizard.actions.saveChanges") : t("wizard.actions.create")
              : t("wizard.actions.next")}
            {!isSaving && lastStep ? <Check data-icon="inline-end" aria-hidden="true" /> : null}
            {!lastStep ? <ArrowLeft data-icon="inline-end" aria-hidden="true" className="ltr:rotate-180" /> : null}
          </Button>
        </div>
      </footer>

      <UnsavedChangesDialog
        open={discardOpen}
        onOpenChange={setDiscardOpen}
        title={mode === "edit" ? t("wizard.discard.editTitle") : t("wizard.discard.title")}
        description={t("wizard.discard.description")}
        stayLabel={t("wizard.discard.stay")}
        discardLabel={t("wizard.discard.confirm")}
        onDiscard={() => router.push(LIST_HREF)}
      />
    </div>
  );
}

function WizardLoading({ label }: { label: string }) {
  return (
    <div role="status" aria-label={label} className="flex flex-col gap-6">
      <Skeleton className="h-14 w-full" />
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_21.25rem]">
        <Skeleton className="h-96 w-full rounded-xl" />
        <Skeleton className="h-48 w-full rounded-xl" />
      </div>
    </div>
  );
}

export function GuardrailWizard(props: GuardrailWizardProps) {
  const t = useTranslations("guardrails");
  const format = useGuardrailFormat();
  const slug = props.mode === "edit" ? props.slug : null;
  const [existingSlugs, setExistingSlugs] = useState<readonly string[]>([]);
  const [editState, setEditState] = useState<WizardState | null>(null);
  const [loadError, setLoadError] = useState("");

  useEffect(() => {
    let active = true;
    if (slug === null) {
      // Duplicate slugs are checked against the loaded list; the server still answers 409 for the rest.
      listGuardrails("")
        .then((response) => {
          if (active) setExistingSlugs(response.data.map((item) => item.slug));
        })
        .catch(() => undefined);
    } else {
      getGuardrail(slug)
        .then((item) => {
          if (active) setEditState(editWizardState(formFromGuardrail(item)));
        })
        .catch((error: unknown) => {
          if (active) setLoadError(localizedError(error, t("messages.loadError"), format.t));
        });
    }

    return () => {
      active = false;
    };
  }, [format.t, slug, t]);

  if (slug === null) return <WizardFlow initialState={createWizardState()} existingSlugs={existingSlugs} />;
  if (loadError) {
    return (
      <Alert variant="destructive" role="alert">
        <AlertDescription className="flex flex-wrap items-center justify-between gap-3">
          <span>{loadError}</span>
          <Button asChild variant="outline" size="sm">
            <Link href={LIST_HREF}>{t("wizard.done.back")}</Link>
          </Button>
        </AlertDescription>
      </Alert>
    );
  }
  if (!editState) return <WizardLoading label={t("wizard.loading")} />;

  return <WizardFlow initialState={editState} existingSlugs={existingSlugs} originalSlug={slug} />;
}
