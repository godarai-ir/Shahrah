import { type GuardrailCategory, defaultOperationForType } from "../catalog";
import { emptyForm, type GuardrailDraft, slugFromName, slugFromType, type WizardStep } from "../form";
import type { GuardrailFormInput, GuardrailType } from "../types";

export type WizardMode = "create" | "edit";
export type TypeCategoryFilter = "all" | GuardrailCategory;

export type WizardState = {
  mode: WizardMode;
  step: WizardStep;
  form: GuardrailDraft;
  /** Once the slug is edited by hand it no longer follows the type or name. */
  slugTouched: boolean;
  /** The sample text, or null while it should follow the selected type's default sample. */
  sample: string | null;
  /** Set when the user tried to leave the current step; shows that step's error. */
  attempted: boolean;
  dirty: boolean;
  typeQuery: string;
  category: TypeCategoryFilter;
};

export type WizardAction =
  | { type: "selectType"; guardrailType: GuardrailType }
  | { type: "update"; patch: Partial<Omit<GuardrailDraft, "type" | "name" | "slug">> }
  | { type: "setName"; name: string }
  | { type: "setSlug"; slug: string }
  | { type: "setSample"; sample: string }
  | { type: "goTo"; step: WizardStep }
  | { type: "attempt" }
  | { type: "setTypeQuery"; query: string }
  | { type: "setCategory"; category: TypeCategoryFilter }
  | { type: "reset"; state: WizardState };

export function createWizardState(): WizardState {
  return {
    mode: "create",
    step: 0,
    form: { ...emptyForm, type: null },
    slugTouched: false,
    sample: null,
    attempted: false,
    dirty: false,
    typeQuery: "",
    category: "all",
  };
}

/** Editing starts on the settings step; the type and slug are fixed after creation. */
export function editWizardState(form: GuardrailFormInput): WizardState {
  return { ...createWizardState(), mode: "edit", step: 1, form, slugTouched: true };
}

export function wizardReducer(state: WizardState, action: WizardAction): WizardState {
  switch (action.type) {
    case "selectType": {
      if (state.mode === "edit" || state.form.type === action.guardrailType) return state;
      const slug = state.slugTouched ? state.form.slug : slugFromName(state.form.name) || slugFromType(action.guardrailType);
      // Config fields of the previous type stay in the form; buildPayload sends only the current type's fields.
      // A new semantic policy starts in audit so its thresholds can be tuned on real traffic before it blocks.
      const enforcement = action.guardrailType === "semantic_policy" ? "audit" : state.form.enforcement;
      const form = { ...state.form, type: action.guardrailType, operation: defaultOperationForType(action.guardrailType), enforcement, slug };

      return { ...state, form, attempted: false, dirty: true };
    }
    case "update":
      return { ...state, form: { ...state.form, ...action.patch }, dirty: true };
    case "setName": {
      const derived = state.slugTouched ? null : slugFromName(action.name);
      const slug = derived || state.form.slug;

      return { ...state, form: { ...state.form, name: action.name, slug }, dirty: true };
    }
    case "setSlug":
      if (state.mode === "edit") return state;
      return { ...state, form: { ...state.form, slug: action.slug }, slugTouched: true, dirty: true };
    case "setSample":
      return { ...state, sample: action.sample };
    case "goTo":
      if (state.mode === "edit" && action.step === 0) return state;
      return { ...state, step: action.step, attempted: false };
    case "attempt":
      return { ...state, attempted: true };
    case "setTypeQuery":
      return { ...state, typeQuery: action.query };
    case "setCategory":
      return { ...state, category: action.category };
    case "reset":
      return action.state;
  }
}
