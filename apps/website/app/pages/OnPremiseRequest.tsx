'use client';

import Link from 'next/link';
import {
  ArrowLeft,
  ArrowRight,
  Check,
  CheckCircle2,
  Clock3,
  Loader2,
  Pencil,
  Save,
  Send,
  ServerCog,
  ShieldCheck,
} from 'lucide-react';
import { useEffect, useReducer, useRef, useState, type FormEvent, type ReactNode, type RefObject } from 'react';

import {
  contactFields,
  detailKey,
  detailMax,
  isDetailVisible,
  isQuestionVisible,
  isValidEmail,
  isValidPhone,
  notesMax,
  questionnaireSteps,
  stepQuestions,
  type Answers,
  type ContactFieldId,
  type Question,
} from '@/app/lib/on-premise-questionnaire';

const draftStorageKey = 'godarai-on-premise-request-draft-v1';

const wizardSteps = [
  {
    id: 'contact',
    title: 'سازمان و اطلاعات تماس',
    shortTitle: 'سازمان',
    description: 'برای هماهنگی‌های بعدی، مسئول فنی پروژه را معرفی کنید.',
  },
  ...questionnaireSteps.map(({ id, title, shortTitle, description }) => ({ id, title, shortTitle, description })),
  {
    id: 'review',
    title: 'مرور و ارسال',
    shortTitle: 'ارسال',
    description: 'پاسخ‌ها را یک نگاه مرور کنید؛ هر بخش را می‌توانید مستقیم ویرایش کنید.',
  },
];

const reviewStepIndex = wizardSteps.length - 1;

type ContactValues = Record<ContactFieldId, string>;

type FormState = {
  step: number;
  furthestStep: number;
  contact: ContactValues;
  answers: Answers;
  notes: string;
};

type WizardState = FormState & {
  /** True once any saved draft has been read, so the empty initial state never overwrites it. */
  hydrated: boolean;
  restoredDraft: boolean;
};

type Action =
  | { type: 'hydrate'; draft: FormState | null }
  | { type: 'goTo'; step: number }
  | { type: 'contact'; id: ContactFieldId; value: string }
  | { type: 'answer'; id: string; value: string }
  | { type: 'notes'; value: string }
  | { type: 'reset' };

const initialState: WizardState = {
  hydrated: true,
  restoredDraft: false,
  step: 0,
  furthestStep: 0,
  contact: { organization: '', contactName: '', contactRole: '', email: '', phone: '' },
  answers: {},
  notes: '',
};

function reducer(state: WizardState, action: Action): WizardState {
  switch (action.type) {
    case 'hydrate':
      return action.draft
        ? { ...action.draft, hydrated: true, restoredDraft: action.draft.furthestStep > 0 || !!action.draft.contact.organization }
        : { ...state, hydrated: true };
    case 'goTo':
      return { ...state, restoredDraft: false, step: action.step, furthestStep: Math.max(state.furthestStep, action.step) };
    case 'contact':
      return { ...state, contact: { ...state.contact, [action.id]: action.value } };
    case 'answer':
      return { ...state, answers: { ...state.answers, [action.id]: action.value } };
    case 'notes':
      return { ...state, notes: action.value };
    case 'reset':
      return initialState;
  }
}

function toFa(value: number) {
  return value.toLocaleString('fa-IR');
}

function validateContact(contact: ContactValues) {
  const errors: Partial<Record<ContactFieldId, string>> = {};
  if (!contact.organization.trim()) errors.organization = 'نام سازمان را وارد کنید.';
  if (!contact.contactName.trim()) errors.contactName = 'نام مسئول فنی را وارد کنید.';
  if (!contact.email.trim()) errors.email = 'ایمیل را وارد کنید.';
  else if (!isValidEmail(contact.email.trim())) errors.email = 'ایمیل معتبر نیست؛ مثلاً name@company.ir';
  if (!contact.phone.trim()) errors.phone = 'شماره تماس را وارد کنید.';
  else if (!isValidPhone(contact.phone.trim())) errors.phone = 'شماره تماس معتبر نیست.';
  return errors;
}

function answeredCount(questions: Question[], answers: Answers) {
  const visible = questions.filter((question) => isQuestionVisible(question, answers));
  return { answered: visible.filter((question) => answers[question.id]).length, total: visible.length };
}

function readDraft(): FormState | null {
  try {
    const raw = window.localStorage.getItem(draftStorageKey);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Partial<FormState>;
    const step = Number.isInteger(parsed.step) ? Math.min(Math.max(parsed.step ?? 0, 0), reviewStepIndex) : 0;
    return {
      step,
      furthestStep: Number.isInteger(parsed.furthestStep) ? Math.min(Math.max(parsed.furthestStep ?? 0, step), reviewStepIndex) : step,
      contact: { ...initialState.contact, ...parsed.contact },
      answers: { ...initialState.answers, ...parsed.answers },
      notes: typeof parsed.notes === 'string' ? parsed.notes : '',
    };
  } catch {
    return null;
  }
}

function writeDraft(state: FormState | null) {
  try {
    if (state) {
      const { step, furthestStep, contact, answers, notes } = state;
      window.localStorage.setItem(draftStorageKey, JSON.stringify({ step, furthestStep, contact, answers, notes }));
    }
    else window.localStorage.removeItem(draftStorageKey);
  } catch {
    // Storage can be unavailable (private mode); the form still works without drafts.
  }
}

type SubmitState = 'idle' | 'submitting' | 'success' | 'error';

export function OnPremiseRequest() {
  const [state, dispatch] = useReducer(reducer, { ...initialState, hydrated: false });
  const [contactErrors, setContactErrors] = useState<Partial<Record<ContactFieldId, string>>>({});
  const [submitState, setSubmitState] = useState<SubmitState>('idle');
  const [submitError, setSubmitError] = useState('');
  const honeypotRef = useRef<HTMLInputElement>(null);
  const topRef = useRef<HTMLDivElement>(null);
  const headingRef = useRef<HTMLHeadingElement>(null);
  const isFirstRender = useRef(true);

  useEffect(() => {
    dispatch({ type: 'hydrate', draft: readDraft() });
  }, []);

  useEffect(() => {
    if (state.hydrated && submitState !== 'success') writeDraft(state);
  }, [state, submitState]);

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      return;
    }
    const top = topRef.current;
    if (top && top.getBoundingClientRect().top < 0) {
      top.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
    headingRef.current?.focus({ preventScroll: true });
  }, [state.step, submitState]);

  const { step, contact, answers, notes } = state;
  const current = wizardSteps[step];

  function goTo(target: number) {
    if (target > 0) {
      const errors = validateContact(contact);
      setContactErrors(errors);
      if (Object.keys(errors).length > 0) {
        if (step !== 0) dispatch({ type: 'goTo', step: 0 });
        else document.getElementById(`onprem-${Object.keys(errors)[0]}`)?.focus();
        return;
      }
    }
    dispatch({ type: 'goTo', step: target });
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (step < reviewStepIndex) {
      goTo(step + 1);
      return;
    }

    setSubmitState('submitting');
    setSubmitError('');
    try {
      const response = await fetch('/api/on-premise-request', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ contact, answers, notes, website: honeypotRef.current?.value ?? '' }),
      });
      const payload = (await response.json().catch(() => null)) as { error?: string } | null;
      if (!response.ok) {
        throw new Error(payload?.error || 'ارسال درخواست انجام نشد. لطفاً دوباره تلاش کنید.');
      }
      writeDraft(null);
      setSubmitState('success');
    } catch (error) {
      setSubmitState('error');
      setSubmitError(error instanceof Error ? error.message : 'ارسال درخواست انجام نشد. لطفاً دوباره تلاش کنید.');
    }
  }

  function startOver() {
    writeDraft(null);
    dispatch({ type: 'reset' });
    setContactErrors({});
  }

  const progress = Math.round((step / reviewStepIndex) * 100);

  return (
    <main className="relative overflow-hidden px-4 py-12 sm:px-6 sm:py-16 lg:px-8">
      <div className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-96 bg-gradient-to-b from-indigo-500/10 to-transparent" />
      <div className="mx-auto max-w-6xl">
        <header className="mx-auto mb-10 max-w-3xl text-center">
          <div className="mb-5 inline-flex items-center gap-2 rounded-full border border-indigo-200 bg-indigo-50 px-4 py-2 text-sm font-semibold text-indigo-700 dark:border-indigo-500/20 dark:bg-indigo-500/10 dark:text-indigo-300">
            <ServerCog aria-hidden="true" className="size-4" />
            استقرار درون‌سازمانی
          </div>
          <h1 className="mb-4 text-3xl font-black leading-relaxed text-slate-950 dark:text-white sm:text-4xl">
            گدارAI را در زیرساخت سازمان خود مستقر کنید
          </h1>
          <p className="text-base leading-8 text-slate-600 dark:text-slate-300 sm:text-lg">
            با چند پرسش کوتاه، تصویر اولیه‌ای از زیرساخت شما می‌گیریم تا مسیر استقرار را دقیق‌تر برنامه‌ریزی کنیم.
          </p>
          <ul className="mt-6 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 text-sm text-slate-600 dark:text-slate-400">
            <li className="inline-flex items-center gap-1.5">
              <Clock3 aria-hidden="true" className="size-4 text-indigo-500" />
              حدود ۵ دقیقه
            </li>
            <li className="inline-flex items-center gap-1.5">
              <Save aria-hidden="true" className="size-4 text-indigo-500" />
              پاسخ‌ها خودکار ذخیره می‌شوند
            </li>
            <li className="inline-flex items-center gap-1.5">
              <ShieldCheck aria-hidden="true" className="size-4 text-indigo-500" />
              «نمی‌دانم» هم پاسخ قابل‌قبولی است
            </li>
          </ul>
        </header>

        <div ref={topRef} className="scroll-mt-24">
          {submitState === 'success' ? (
            <SuccessPanel headingRef={headingRef} organization={contact.organization} onStartOver={() => { startOver(); setSubmitState('idle'); }} />
          ) : (
            <div className="grid gap-6 lg:grid-cols-[15rem_1fr] lg:gap-8">
              <aside className="hidden lg:block">
                <nav aria-label="مراحل فرم" className="sticky top-24">
                  <ol className="space-y-1">
                    {wizardSteps.map((item, index) => {
                      const isCurrent = index === step;
                      const isDone = index < step || (index <= state.furthestStep && index !== step);
                      const canJump = index <= state.furthestStep && !isCurrent;
                      const counts = index > 0 && index < reviewStepIndex ? answeredCount(stepQuestions(questionnaireSteps[index - 1]), answers) : null;
                      return (
                        <li key={item.id}>
                          <button
                            type="button"
                            onClick={() => canJump && goTo(index)}
                            disabled={!canJump}
                            aria-current={isCurrent ? 'step' : undefined}
                            className={`flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-right transition-colors focus-visible:outline-2 focus-visible:outline-indigo-500 ${
                              isCurrent ? 'bg-indigo-50 dark:bg-indigo-500/10' : canJump ? 'hover:bg-slate-100 dark:hover:bg-slate-800/60' : 'cursor-default'
                            }`}
                          >
                            <span
                              className={`grid size-8 shrink-0 place-items-center rounded-full text-sm font-bold ${
                                isCurrent
                                  ? 'bg-indigo-600 text-white'
                                  : isDone
                                    ? 'bg-indigo-100 text-indigo-700 dark:bg-indigo-500/20 dark:text-indigo-300'
                                    : 'border border-slate-200 text-slate-400 dark:border-slate-700 dark:text-slate-500'
                              }`}
                            >
                              {isDone && !isCurrent ? <Check aria-hidden="true" className="size-4" /> : toFa(index + 1)}
                            </span>
                            <span className="min-w-0">
                              <span className={`block text-sm font-semibold ${isCurrent ? 'text-indigo-700 dark:text-indigo-300' : isDone ? 'text-slate-800 dark:text-slate-200' : 'text-slate-500 dark:text-slate-500'}`}>
                                {item.title}
                              </span>
                              {counts && index <= state.furthestStep && (
                                <span className="block text-xs text-slate-500 dark:text-slate-400">
                                  {toFa(counts.answered)} از {toFa(counts.total)} پاسخ
                                </span>
                              )}
                            </span>
                          </button>
                        </li>
                      );
                    })}
                  </ol>
                </nav>
              </aside>

              <form
                onSubmit={handleSubmit}
                noValidate
                className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-xl shadow-slate-950/5 dark:border-slate-800 dark:bg-slate-900"
              >
                <div className="border-b border-slate-200 px-5 pb-5 pt-5 dark:border-slate-800 sm:px-8 sm:pt-7">
                  <div className="mb-4 lg:hidden">
                    <div className="mb-2 flex items-center justify-between text-xs font-medium text-slate-500 dark:text-slate-400">
                      <span>
                        مرحله {toFa(step + 1)} از {toFa(wizardSteps.length)}
                      </span>
                      {step < reviewStepIndex && <span>بعدی: {wizardSteps[step + 1].shortTitle}</span>}
                    </div>
                    <div
                      className="h-1.5 overflow-hidden rounded-full bg-slate-100 dark:bg-slate-800"
                      role="progressbar"
                      aria-valuemin={0}
                      aria-valuemax={100}
                      aria-valuenow={progress}
                      aria-label="پیشرفت تکمیل فرم"
                    >
                      <div className="h-full rounded-full bg-indigo-600 transition-[width] duration-500" style={{ width: `${Math.max(progress, 4)}%` }} />
                    </div>
                  </div>
                  <h2 ref={headingRef} tabIndex={-1} className="text-xl font-bold text-slate-950 outline-none dark:text-white sm:text-2xl">
                    {current.title}
                  </h2>
                  <p className="mt-1.5 text-sm leading-7 text-slate-600 dark:text-slate-400">{current.description}</p>
                  {state.restoredDraft && (
                    <p className="mt-3 flex flex-wrap items-center gap-x-2 text-xs text-slate-500 dark:text-slate-400">
                      پاسخ‌های قبلی شما بازیابی شد.
                      <button type="button" onClick={startOver} className="font-semibold text-indigo-600 hover:underline dark:text-indigo-400">
                        شروع از ابتدا
                      </button>
                    </p>
                  )}
                </div>

                <div key={current.id} className="animate-in fade-in slide-in-from-left-2 px-5 py-6 duration-300 sm:px-8">
                  {step === 0 && (
                    <ContactStep
                      contact={contact}
                      errors={contactErrors}
                      onContact={(id, value) => {
                        dispatch({ type: 'contact', id, value });
                        if (contactErrors[id]) setContactErrors((errors) => ({ ...errors, [id]: undefined }));
                      }}
                    />
                  )}

                  {step > 0 && step < reviewStepIndex && (
                    <div className="space-y-8">
                      {questionnaireSteps[step - 1].groups.map((group) => (
                        <section key={group.title} aria-labelledby={`group-${group.title}`}>
                          <h3 id={`group-${group.title}`} className="mb-1 text-sm font-bold text-indigo-600 dark:text-indigo-400">
                            {group.title}
                          </h3>
                          <div className="divide-y divide-slate-100 dark:divide-slate-800">
                            {group.questions
                              .filter((question) => isQuestionVisible(question, answers))
                              .map((question) => (
                                <QuestionField
                                  key={question.id}
                                  question={question}
                                  answers={answers}
                                  onAnswer={(id, value) => dispatch({ type: 'answer', id, value })}
                                />
                              ))}
                          </div>
                        </section>
                      ))}
                    </div>
                  )}

                  {step === reviewStepIndex && (
                    <ReviewStep
                      contact={contact}
                      answers={answers}
                      notes={notes}
                      onNotes={(value) => dispatch({ type: 'notes', value })}
                      onEdit={goTo}
                    />
                  )}

                  <div className="hidden" aria-hidden="true">
                    <label htmlFor="onprem-website">وب‌سایت</label>
                    <input ref={honeypotRef} id="onprem-website" name="website" type="text" tabIndex={-1} autoComplete="off" />
                  </div>

                  {submitState === 'error' && (
                    <p role="alert" className="mt-6 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-300">
                      {submitError}
                    </p>
                  )}
                </div>

                <div className="sticky bottom-0 flex items-center justify-between gap-3 border-t border-slate-200 bg-white/95 px-5 py-4 backdrop-blur dark:border-slate-800 dark:bg-slate-900/95 sm:px-8">
                  {step > 0 ? (
                    <button
                      type="button"
                      onClick={() => goTo(step - 1)}
                      className="inline-flex min-h-11 items-center gap-2 rounded-xl px-4 text-sm font-semibold text-slate-600 transition-colors hover:bg-slate-100 focus-visible:outline-2 focus-visible:outline-indigo-500 dark:text-slate-300 dark:hover:bg-slate-800"
                    >
                      <ArrowRight aria-hidden="true" className="size-4" />
                      قبلی
                    </button>
                  ) : (
                    <span />
                  )}
                  <button
                    type="submit"
                    disabled={submitState === 'submitting'}
                    className="inline-flex min-h-11 items-center gap-2 rounded-xl bg-indigo-600 px-6 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-indigo-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500 disabled:cursor-not-allowed disabled:opacity-60"
                  >
                    {step < reviewStepIndex ? (
                      <>
                        ادامه
                        <ArrowLeft aria-hidden="true" className="size-4" />
                      </>
                    ) : submitState === 'submitting' ? (
                      <>
                        در حال ارسال...
                        <Loader2 aria-hidden="true" className="size-4 animate-spin" />
                      </>
                    ) : (
                      <>
                        ارسال درخواست
                        <Send aria-hidden="true" className="size-4 -scale-x-100" />
                      </>
                    )}
                  </button>
                </div>
              </form>
            </div>
          )}
        </div>
      </div>
    </main>
  );
}

const inputClassName =
  'w-full rounded-xl border bg-white px-4 py-3 text-sm text-slate-900 transition-colors placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:bg-slate-950 dark:text-white dark:placeholder:text-slate-500';

function ContactStep({
  contact,
  errors,
  onContact,
}: {
  contact: ContactValues;
  errors: Partial<Record<ContactFieldId, string>>;
  onContact: (id: ContactFieldId, value: string) => void;
}) {
  const fields: Array<{ id: ContactFieldId; type: string; autoComplete: string; placeholder: string; dir?: 'ltr'; wide?: boolean }> = [
    { id: 'organization', type: 'text', autoComplete: 'organization', placeholder: 'مثلاً بانک نمونه — واحد فناوری اطلاعات', wide: true },
    { id: 'contactName', type: 'text', autoComplete: 'name', placeholder: 'نام و نام خانوادگی' },
    { id: 'contactRole', type: 'text', autoComplete: 'organization-title', placeholder: 'مثلاً مدیر زیرساخت' },
    { id: 'email', type: 'email', autoComplete: 'email', placeholder: 'name@company.ir', dir: 'ltr' },
    { id: 'phone', type: 'tel', autoComplete: 'tel', placeholder: '09120000000', dir: 'ltr' },
  ];

  return (
    <div className="space-y-8">
      <div className="grid gap-5 sm:grid-cols-2">
        {fields.map((field) => {
          const definition = contactFields[field.id];
          const error = errors[field.id];
          return (
            <div key={field.id} className={field.wide ? 'sm:col-span-2' : undefined}>
              <label htmlFor={`onprem-${field.id}`} className="mb-2 block text-sm font-medium text-slate-800 dark:text-slate-200">
                {definition.label}
                {!definition.required && <span className="mr-1 text-xs font-normal text-slate-400">(اختیاری)</span>}
              </label>
              <input
                id={`onprem-${field.id}`}
                type={field.type}
                autoComplete={field.autoComplete}
                maxLength={definition.max}
                required={definition.required}
                dir={field.dir}
                value={contact[field.id]}
                onChange={(event) => onContact(field.id, event.target.value)}
                placeholder={field.placeholder}
                aria-invalid={!!error}
                aria-describedby={error ? `onprem-${field.id}-error` : undefined}
                className={`${inputClassName} ${field.dir === 'ltr' ? 'text-left' : ''} ${
                  error ? 'border-red-400 dark:border-red-500/60' : 'border-slate-200 dark:border-slate-700'
                }`}
              />
              {error && (
                <p id={`onprem-${field.id}-error`} className="mt-1.5 text-xs text-red-600 dark:text-red-400">
                  {error}
                </p>
              )}
            </div>
          );
        })}
      </div>

    </div>
  );
}

function QuestionField({
  question,
  answers,
  onAnswer,
}: {
  question: Question;
  answers: Answers;
  onAnswer: (id: string, value: string) => void;
}) {
  const labelId = `q-${question.id}-label`;
  const selected = answers[question.id] ?? '';
  const showDetail = isDetailVisible(question, answers);

  return (
    <div className="py-4">
      <div className={`flex flex-col gap-3 ${question.kind === 'choice' ? 'sm:flex-row sm:items-center sm:justify-between sm:gap-6' : ''}`}>
        <p id={labelId} className="text-sm font-medium leading-7 text-slate-800 dark:text-slate-200">
          {question.label}
        </p>
        <div role="radiogroup" aria-labelledby={labelId} className="flex shrink-0 flex-wrap gap-2">
          {question.options.map((option) => {
            const isNeutral = option.value === 'unsure' || option.value === 'needs-review';
            return (
              <label
                key={option.value}
                className={`inline-flex min-h-10 cursor-pointer select-none items-center justify-center rounded-lg border border-slate-200 px-3.5 text-sm font-medium text-slate-700 transition-colors hover:border-indigo-300 hover:bg-indigo-50/50 has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-indigo-500 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-indigo-500/10 ${
                  isNeutral
                    ? 'has-[:checked]:border-slate-500 has-[:checked]:bg-slate-600 has-[:checked]:text-white dark:has-[:checked]:bg-slate-600'
                    : 'has-[:checked]:border-indigo-600 has-[:checked]:bg-indigo-600 has-[:checked]:text-white dark:has-[:checked]:border-indigo-500 dark:has-[:checked]:bg-indigo-500'
                }`}
              >
                <input
                  type="radio"
                  name={question.id}
                  value={option.value}
                  checked={selected === option.value}
                  onChange={() => onAnswer(question.id, option.value)}
                  className="sr-only"
                />
                {option.label}
              </label>
            );
          })}
        </div>
      </div>
      {showDetail && question.kind === 'choice' && question.detail && (
        <div className="animate-in fade-in slide-in-from-top-1 mt-3 flex items-center gap-3 duration-200">
          <label htmlFor={`q-${question.id}-detail`} className="shrink-0 text-xs font-medium text-slate-500 dark:text-slate-400">
            {question.detail.label}
            <span className="mr-1 text-slate-400">(اختیاری)</span>
          </label>
          <input
            id={`q-${question.id}-detail`}
            type="text"
            maxLength={detailMax}
            dir={question.detail.dir}
            value={answers[detailKey(question.id)] ?? ''}
            onChange={(event) => onAnswer(detailKey(question.id), event.target.value)}
            placeholder={question.detail.placeholder}
            className={`${inputClassName} border-slate-200 py-2 dark:border-slate-700 sm:max-w-xs ${question.detail.dir === 'ltr' ? 'text-left' : ''}`}
          />
        </div>
      )}
    </div>
  );
}

function ReviewStep({
  contact,
  answers,
  notes,
  onNotes,
  onEdit,
}: {
  contact: ContactValues;
  answers: Answers;
  notes: string;
  onNotes: (value: string) => void;
  onEdit: (step: number) => void;
}) {
  const totals = questionnaireSteps.reduce(
    (sum, step) => {
      const counts = answeredCount(stepQuestions(step), answers);
      return { answered: sum.answered + counts.answered, total: sum.total + counts.total };
    },
    { answered: 0, total: 0 },
  );
  const unanswered = totals.total - totals.answered;

  return (
    <div className="space-y-4">
      <div
        className={`rounded-xl border px-4 py-3 text-sm leading-7 ${
          unanswered === 0
            ? 'border-emerald-200 bg-emerald-50 text-emerald-800 dark:border-emerald-500/30 dark:bg-emerald-500/10 dark:text-emerald-300'
            : 'border-amber-200 bg-amber-50 text-amber-800 dark:border-amber-500/30 dark:bg-amber-500/10 dark:text-amber-200'
        }`}
      >
        {unanswered === 0
          ? 'همه پرسش‌ها پاسخ داده شده‌اند. آماده ارسال است.'
          : `${toFa(unanswered)} پرسش بی‌پاسخ مانده است. می‌توانید همین حالا ارسال کنید؛ کارشناسان ما در گفت‌وگوی بعدی تکمیلش می‌کنند.`}
      </div>

      <ReviewCard title="سازمان و اطلاعات تماس" onEdit={() => onEdit(0)} defaultOpen>
        <dl className="grid gap-x-6 gap-y-2 text-sm sm:grid-cols-2">
          {(Object.keys(contactFields) as ContactFieldId[])
            .filter((id) => contact[id].trim())
            .map((id) => (
              <div key={id} className="flex gap-2">
                <dt className="shrink-0 text-slate-500 dark:text-slate-400">{contactFields[id].label}:</dt>
                <dd className="min-w-0 break-words font-medium text-slate-900 dark:text-white">{contact[id]}</dd>
              </div>
            ))}
        </dl>
      </ReviewCard>

      {questionnaireSteps.map((step, index) => {
        const counts = answeredCount(stepQuestions(step), answers);
        return (
          <ReviewCard
            key={step.id}
            title={step.title}
            badge={`${toFa(counts.answered)} از ${toFa(counts.total)}`}
            complete={counts.answered === counts.total}
            onEdit={() => onEdit(index + 1)}
          >
            <dl className="divide-y divide-slate-100 text-sm dark:divide-slate-800">
              {stepQuestions(step)
                .filter((question) => isQuestionVisible(question, answers))
                .map((question) => {
                  const label = question.options.find((option) => option.value === answers[question.id])?.label;
                  const detail = isDetailVisible(question, answers) ? answers[detailKey(question.id)]?.trim() : '';
                  return (
                    <div key={question.id} className="flex flex-col gap-1 py-2 sm:flex-row sm:justify-between sm:gap-6">
                      <dt className="text-slate-600 dark:text-slate-400">{question.label}</dt>
                      <dd className={`shrink-0 font-medium ${label ? 'text-slate-900 dark:text-white' : 'text-slate-400 dark:text-slate-500'}`}>
                        {label ?? 'بی‌پاسخ'}
                        {detail && <span className="mr-1 font-normal text-slate-500 dark:text-slate-400">({detail})</span>}
                      </dd>
                    </div>
                  );
                })}
            </dl>
          </ReviewCard>
        );
      })}

      <div className="pt-4">
        <label htmlFor="onprem-notes" className="mb-2 block text-sm font-medium text-slate-800 dark:text-slate-200">
          توضیحات تکمیلی
          <span className="mr-1 text-xs font-normal text-slate-400">(اختیاری)</span>
        </label>
        <textarea
          id="onprem-notes"
          rows={4}
          maxLength={notesMax}
          value={notes}
          onChange={(event) => onNotes(event.target.value)}
          placeholder="الزامات خاص، محدودیت‌های شبکه، زمان‌بندی موردنظر یا هر نکته‌ای که بهتر است بدانیم..."
          className={`${inputClassName} border-slate-200 dark:border-slate-700`}
        />
      </div>
    </div>
  );
}

function ReviewCard({
  title,
  badge,
  complete,
  defaultOpen,
  onEdit,
  children,
}: {
  title: string;
  badge?: string;
  complete?: boolean;
  defaultOpen?: boolean;
  onEdit: () => void;
  children: ReactNode;
}) {
  return (
    <details open={defaultOpen} className="group rounded-xl border border-slate-200 dark:border-slate-800">
      <summary className="flex min-h-14 cursor-pointer list-none items-center gap-3 px-4 py-3 marker:content-none focus-visible:rounded-xl focus-visible:outline-2 focus-visible:outline-indigo-500 [&::-webkit-details-marker]:hidden">
        <span className="flex-1 font-semibold text-slate-900 dark:text-white">{title}</span>
        {badge && (
          <span
            className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
              complete
                ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'
                : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300'
            }`}
          >
            {badge}
          </span>
        )}
        <button
          type="button"
          onClick={(event) => {
            event.preventDefault();
            onEdit();
          }}
          className="inline-flex min-h-9 items-center gap-1.5 rounded-lg px-2.5 text-xs font-semibold text-indigo-600 hover:bg-indigo-50 focus-visible:outline-2 focus-visible:outline-indigo-500 dark:text-indigo-400 dark:hover:bg-indigo-500/10"
        >
          <Pencil aria-hidden="true" className="size-3.5" />
          ویرایش
        </button>
      </summary>
      <div className="border-t border-slate-100 px-4 py-3 dark:border-slate-800">{children}</div>
    </details>
  );
}

function SuccessPanel({
  headingRef,
  organization,
  onStartOver,
}: {
  headingRef: RefObject<HTMLHeadingElement | null>;
  organization: string;
  onStartOver: () => void;
}) {
  return (
    <div className="animate-in fade-in zoom-in-95 mx-auto max-w-2xl rounded-2xl border border-slate-200 bg-white px-6 py-10 text-center shadow-xl shadow-slate-950/5 duration-300 dark:border-slate-800 dark:bg-slate-900 sm:px-10">
      <CheckCircle2 aria-hidden="true" className="mx-auto mb-4 size-14 text-emerald-500" />
      <h2 ref={headingRef} tabIndex={-1} className="mb-3 text-2xl font-bold text-slate-950 outline-none dark:text-white">
        درخواست شما ثبت شد
      </h2>
      <p className="leading-8 text-slate-600 dark:text-slate-300">
        از اینکه وقت گذاشتید سپاسگزاریم{organization ? `؛ درخواست ${organization} به دست ما رسید` : ''}. کارشناسان گدارAI پاسخ‌ها را بررسی می‌کنند و برای هماهنگی‌های فنی با شما تماس می‌گیرند.
      </p>
      <p className="mt-3 text-sm leading-7 text-slate-500 dark:text-slate-400">
        تکمیل این پرسشنامه به معنای تأیید یا رد آمادگی سازمان نیست؛ ارزیابی نهایی پس از بررسی پاسخ‌ها انجام می‌شود.
      </p>
      <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
        <Link
          href="/demo"
          className="inline-flex min-h-11 items-center gap-2 rounded-xl bg-indigo-600 px-5 text-sm font-semibold text-white transition-colors hover:bg-indigo-700"
        >
          در این فاصله، یک دموی خصوصی رزرو کنید
          <ArrowLeft aria-hidden="true" className="size-4" />
        </Link>
        <button type="button" onClick={onStartOver} className="min-h-11 rounded-xl px-4 text-sm font-semibold text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800">
          ثبت درخواست دیگر
        </button>
      </div>
    </div>
  );
}
