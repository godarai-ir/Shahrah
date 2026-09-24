import { describe, expect, it } from "vitest";

import { emptyForm } from "../form";
import { createWizardState, editWizardState, wizardReducer, type WizardAction, type WizardState } from "./state";

function run(state: WizardState, ...actions: WizardAction[]) {
  return actions.reduce(wizardReducer, state);
}

describe("wizardReducer", () => {
  it("presets the operation and slug when a type is picked", () => {
    const state = run(createWizardState(), { type: "selectType", guardrailType: "secrets_detection" });

    expect(state.form).toMatchObject({ type: "secrets_detection", operation: "mutate", slug: "secrets-detection" });
    expect(state.dirty).toBe(true);
    expect(run(state, { type: "selectType", guardrailType: "regex" }).form.operation).toBe("validate");
  });

  it("derives the slug from a Latin name until the slug is edited", () => {
    const named = run(createWizardState(), { type: "selectType", guardrailType: "not_null" }, { type: "setName", name: "Empty Output" });
    expect(named.form.slug).toBe("empty-output");

    const persian = run(named, { type: "setName", name: "خروجی خالی" });
    expect(persian.form.slug).toBe("empty-output");

    const edited = run(named, { type: "setSlug", slug: "custom" }, { type: "setName", name: "Another Name" }, { type: "selectType", guardrailType: "word_count" });
    expect(edited.form.slug).toBe("custom");
  });

  it("keeps the previous type's config when the type changes", () => {
    const state = run(
      createWizardState(),
      { type: "selectType", guardrailType: "keyword_blocklist" },
      { type: "update", patch: { terms: "jailbreak" } },
      { type: "selectType", guardrailType: "regex" },
    );

    expect(state.form).toMatchObject({ type: "regex", terms: "jailbreak" });
  });

  it("does not change the type, slug, or first step while editing", () => {
    const state = editWizardState({ ...emptyForm, type: "not_null", slug: "keep", name: "Keep" });

    expect(state.step).toBe(1);
    expect(run(state, { type: "selectType", guardrailType: "regex" }).form.type).toBe("not_null");
    expect(run(state, { type: "setSlug", slug: "other" }).form.slug).toBe("keep");
    expect(run(state, { type: "setName", name: "New Name" }).form.slug).toBe("keep");
    expect(run(state, { type: "goTo", step: 0 }).step).toBe(1);
  });

  it("clears the attempted flag when the step changes", () => {
    const state = run(createWizardState(), { type: "attempt" }, { type: "goTo", step: 1 });

    expect(state.attempted).toBe(false);
  });
});
