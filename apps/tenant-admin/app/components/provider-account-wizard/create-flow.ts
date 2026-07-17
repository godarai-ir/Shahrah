import type { CreateWizardSectionID } from "./wizard-layout";

export type CreatePersistenceReason = "continue-details" | "continue-models" | "continue-access" | "credential-test" | "activate";

export function shouldPersistCreateAccount(reason: CreatePersistenceReason): boolean {
  return reason === "credential-test" || reason === "activate";
}

export function getCreateStepState({
  activeStep,
  step,
  completedThrough,
}: {
  activeStep: CreateWizardSectionID;
  step: CreateWizardSectionID;
  completedThrough: number;
}): "active" | "complete" | "idle" {
  if (activeStep === step) return "active";
  const order: CreateWizardSectionID[] = ["details", "models", "access", "review"];
  const index = order.indexOf(step);
  return index >= 0 && index < completedThrough ? "complete" : "idle";
}

export function buildCreatedConnectionHref(provider: string, accountName: string): string {
  const params = new URLSearchParams({
    tab: "connected",
    provider: provider.trim().toLowerCase(),
    created: accountName.trim(),
  });
  return `/provider-accounts?${params.toString()}`;
}
