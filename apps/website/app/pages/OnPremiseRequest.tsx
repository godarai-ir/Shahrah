'use client';

import Link from 'next/link';
import {
  ArrowLeft,
  ArrowRight,
  Check,
  CheckCircle2,
  Loader2,
  Pencil,
  Send,
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
import { submitForm } from '@/app/lib/submit-form';

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

// Only confirmed On-Premise facts (docs/site-architecture/04-open-questions.md): local models and GodarAI-managed
// updates are not claimed.
const onPremisePoints = [
  { title: 'داده و گزارش‌ها روی زیرساخت شما', text: 'گدارAI داخل سرورهای سازمان اجرا می‌شود.' },
  { title: 'همان فضای کاری و کنترل‌ها', text: 'دستیارها، گاردریل معنایی و سقف هزینه، با ورود سازمانی و همگام‌سازی کاربران.' },
  { title: 'قرارداد سالانه و پشتیبانی', text: 'با فاکتور رسمی.' },
];

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
      await submitForm(
        '/api/on-premise-request',
        { contact, answers, notes, website: honeypotRef.current?.value ?? '' },
        'ارسال درخواست انجام نشد. لطفاً دوباره تلاش کنید.',
      );
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
    <main className="bg-site-ground px-4 py-10 sm:px-6 lg:py-20">
      <div className="mx-auto max-w-content">
        <header className="mx-auto mb-8 flex max-w-3xl flex-col items-center gap-4 text-center">
          <h1 className="m-0 text-balance text-h1 font-extrabold text-site-ink lg:text-display">گدارAI، داخل سرورهای خودتان.</h1>
          <p className="m-0 text-lead font-light text-site-ink-soft">
            برای سازمانی که داده‌اش نباید از زیرساخت خودش بیرون برود؛ با قرارداد سالانه و پشتیبانی.
          </p>
        </header>

        <ul className="m-0 mb-6 grid list-none gap-3 p-0 sm:grid-cols-3 sm:gap-6">
          {onPremisePoints.map((point) => (
            <li key={point.title} className="flex flex-col gap-2 rounded-card border border-site-line bg-site-surface p-5">
              <span className="text-label font-semibold text-site-ink">{point.title}</span>
              <span className="text-caption text-site-ink-soft">{point.text}</span>
            </li>
          ))}
        </ul>

        <p className="mx-auto mb-10 max-w-3xl text-center text-label text-site-ink-soft">
          روند کار: پرسشنامه، جلسه‌ی فنی، نصب. زیرساخت لازم: Kubernetes. حدود ۵ دقیقه وقت می‌گیرد، پاسخ‌ها خودکار ذخیره می‌شوند و «نمی‌دانم» هم پاسخ قابل‌قبولی است.
        </p>

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
                            className={`flex w-full items-center gap-3 rounded-card px-3 py-2.5 text-right transition-colors focus-visible:outline-2 focus-visible:outline-site-focus ${
                              isCurrent ? 'bg-site-sand' : canJump ? 'hover:bg-site-panel' : 'cursor-default'
                            }`}
                          >
                            <span
                              className={`grid size-8 shrink-0 place-items-center rounded-full text-sm font-bold ${
                                isCurrent
                                  ? 'bg-site-ink text-site-surface'
                                  : isDone
                                    ? 'bg-site-sand text-site-accent-text'
                                    : 'border border-site-line text-site-ink-soft'
                              }`}
                            >
                              {isDone && !isCurrent ? <Check aria-hidden="true" className="size-4" /> : toFa(index + 1)}
                            </span>
                            <span className="min-w-0">
                              <span className={`block text-sm font-semibold ${isCurrent ? 'text-site-accent-text' : isDone ? 'text-site-ink' : 'text-site-ink-soft'}`}>
                                {item.title}
                              </span>
                              {counts && index <= state.furthestStep && (
                                <span className="block text-xs text-site-ink-soft">
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
                className="overflow-hidden rounded-card border border-site-line bg-site-surface"
              >
                <div className="border-b border-site-line px-5 pb-5 pt-5 sm:px-8 sm:pt-7">
                  <div className="mb-4 lg:hidden">
                    <div className="mb-2 flex items-center justify-between text-xs font-medium text-site-ink-soft">
                      <span>
                        مرحله {toFa(step + 1)} از {toFa(wizardSteps.length)}
                      </span>
                      {step < reviewStepIndex && <span>بعدی: {wizardSteps[step + 1].shortTitle}</span>}
                    </div>
                    <div
                      className="h-1.5 overflow-hidden rounded-full bg-site-panel"
                      role="progressbar"
                      aria-valuemin={0}
                      aria-valuemax={100}
                      aria-valuenow={progress}
                      aria-label="پیشرفت تکمیل فرم"
                    >
                      <div className="h-full rounded-full bg-site-accent transition-[width] duration-500" style={{ width: `${Math.max(progress, 4)}%` }} />
                    </div>
                  </div>
                  <h2 ref={headingRef} tabIndex={-1} className="text-xl font-bold text-site-ink outline-none sm:text-2xl">
                    {current.title}
                  </h2>
                  <p className="mt-1.5 text-sm leading-7 text-site-ink-soft">{current.description}</p>
                  {state.restoredDraft && (
                    <p className="mt-3 flex flex-wrap items-center gap-x-2 text-xs text-site-ink-soft">
                      پاسخ‌های قبلی شما بازیابی شد.
                      <button type="button" onClick={startOver} className="font-semibold text-site-accent-text hover:underline">
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
                          <h3 id={`group-${group.title}`} className="mb-1 text-sm font-bold text-site-accent-text">
                            {group.title}
                          </h3>
                          <div className="divide-y divide-site-line">
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
                    <p role="alert" className="mt-6 rounded-card border border-site-danger bg-site-panel px-4 py-3 text-sm text-site-danger">
                      {submitError}
                    </p>
                  )}
                </div>

                <div className="sticky bottom-0 flex items-center justify-between gap-3 border-t border-site-line bg-site-surface px-5 py-4 sm:px-8">
                  {step > 0 ? (
                    <button
                      type="button"
                      onClick={() => goTo(step - 1)}
                      className="inline-flex min-h-11 items-center gap-2 rounded-card px-4 text-sm font-semibold text-site-ink-soft transition-colors hover:bg-site-panel focus-visible:outline-2 focus-visible:outline-site-focus"
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
                    className="inline-flex min-h-11 items-center gap-2 rounded-card bg-site-accent px-6 text-sm font-semibold text-site-on-accent transition-colors hover:bg-site-accent-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus disabled:cursor-not-allowed disabled:opacity-60"
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
  'w-full rounded-card border bg-site-surface px-4 py-3 text-sm text-site-ink transition-colors placeholder:text-site-ink-soft focus:outline-none focus:ring-2 focus:ring-site-focus';

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
              <label htmlFor={`onprem-${field.id}`} className="mb-2 block text-sm font-medium text-site-ink">
                {definition.label}
                {!definition.required && <span className="mr-1 text-xs font-normal text-site-ink-soft">(اختیاری)</span>}
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
                  error ? 'border-site-danger' : 'border-site-line'
                }`}
              />
              {error && (
                <p id={`onprem-${field.id}-error`} className="mt-1.5 text-xs text-site-danger">
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
        <p id={labelId} className="text-sm font-medium leading-7 text-site-ink">
          {question.label}
        </p>
        <div role="radiogroup" aria-labelledby={labelId} className="flex shrink-0 flex-wrap gap-2">
          {question.options.map((option) => {
            const isNeutral = option.value === 'unsure' || option.value === 'needs-review';
            return (
              <label
                key={option.value}
                className={`inline-flex min-h-10 cursor-pointer select-none items-center justify-center rounded-lg border border-site-line px-3.5 text-sm font-medium text-site-ink transition-colors hover:border-site-line hover:bg-site-sand has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-site-focus ${
                  isNeutral
                    ? 'has-[:checked]:border-site-line has-[:checked]:bg-site-line-strong has-[:checked]:text-site-on-accent'
                    : 'has-[:checked]:border-site-line-strong has-[:checked]:bg-site-accent has-[:checked]:text-site-on-accent'
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
          <label htmlFor={`q-${question.id}-detail`} className="shrink-0 text-xs font-medium text-site-ink-soft">
            {question.detail.label}
            <span className="mr-1 text-site-ink-soft">(اختیاری)</span>
          </label>
          <input
            id={`q-${question.id}-detail`}
            type="text"
            maxLength={detailMax}
            dir={question.detail.dir}
            value={answers[detailKey(question.id)] ?? ''}
            onChange={(event) => onAnswer(detailKey(question.id), event.target.value)}
            placeholder={question.detail.placeholder}
            className={`${inputClassName} border-site-line py-2 sm:max-w-xs ${question.detail.dir === 'ltr' ? 'text-left' : ''}`}
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
        className={`rounded-card border px-4 py-3 text-sm leading-7 ${
          unanswered === 0
            ? 'border-site-success bg-site-panel text-site-success'
            : 'border-site-line bg-site-panel text-site-accent-text'
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
                <dt className="shrink-0 text-site-ink-soft">{contactFields[id].label}:</dt>
                <dd className="min-w-0 break-words font-medium text-site-ink">{contact[id]}</dd>
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
            <dl className="divide-y divide-site-line text-sm">
              {stepQuestions(step)
                .filter((question) => isQuestionVisible(question, answers))
                .map((question) => {
                  const label = question.options.find((option) => option.value === answers[question.id])?.label;
                  const detail = isDetailVisible(question, answers) ? answers[detailKey(question.id)]?.trim() : '';
                  return (
                    <div key={question.id} className="flex flex-col gap-1 py-2 sm:flex-row sm:justify-between sm:gap-6">
                      <dt className="text-site-ink-soft">{question.label}</dt>
                      <dd className={`shrink-0 font-medium ${label ? 'text-site-ink' : 'text-site-ink-soft'}`}>
                        {label ?? 'بی‌پاسخ'}
                        {detail && <span className="mr-1 font-normal text-site-ink-soft">({detail})</span>}
                      </dd>
                    </div>
                  );
                })}
            </dl>
          </ReviewCard>
        );
      })}

      <div className="pt-4">
        <label htmlFor="onprem-notes" className="mb-2 block text-sm font-medium text-site-ink">
          توضیحات تکمیلی
          <span className="mr-1 text-xs font-normal text-site-ink-soft">(اختیاری)</span>
        </label>
        <textarea
          id="onprem-notes"
          rows={4}
          maxLength={notesMax}
          value={notes}
          onChange={(event) => onNotes(event.target.value)}
          placeholder="الزامات خاص، محدودیت‌های شبکه، زمان‌بندی موردنظر یا هر نکته‌ای که بهتر است بدانیم..."
          className={`${inputClassName} border-site-line`}
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
    <details open={defaultOpen} className="group rounded-card border border-site-line">
      <summary className="flex min-h-14 cursor-pointer list-none items-center gap-3 px-4 py-3 marker:content-none focus-visible:rounded-card focus-visible:outline-2 focus-visible:outline-site-focus [&::-webkit-details-marker]:hidden">
        <span className="flex-1 font-semibold text-site-ink">{title}</span>
        {badge && (
          <span
            className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${
              complete
                ? 'bg-site-panel text-site-success'
                : 'bg-site-panel text-site-ink-soft'
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
          className="inline-flex min-h-9 items-center gap-1.5 rounded-lg px-2.5 text-xs font-semibold text-site-accent-text hover:bg-site-sand focus-visible:outline-2 focus-visible:outline-site-focus"
        >
          <Pencil aria-hidden="true" className="size-3.5" />
          ویرایش
        </button>
      </summary>
      <div className="border-t border-site-line px-4 py-3">{children}</div>
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
    <div className="animate-in fade-in zoom-in-95 mx-auto max-w-2xl rounded-card border border-site-line bg-site-surface px-6 py-10 text-center duration-300 sm:px-10">
      <CheckCircle2 aria-hidden="true" className="mx-auto mb-4 size-14 text-site-success" />
      <h2 ref={headingRef} tabIndex={-1} className="mb-3 text-2xl font-bold text-site-ink outline-none">
        درخواست شما ثبت شد
      </h2>
      <p className="leading-8 text-site-ink-soft">
        از اینکه وقت گذاشتید سپاسگزاریم{organization ? `؛ درخواست ${organization} به دست ما رسید` : ''}. کارشناسان گدارAI پاسخ‌ها را بررسی می‌کنند و برای هماهنگی‌های فنی با شما تماس می‌گیرند.
      </p>
      <p className="mt-3 text-sm leading-7 text-site-ink-soft">
        تکمیل این پرسشنامه به معنای تأیید یا رد آمادگی سازمان نیست؛ ارزیابی نهایی پس از بررسی پاسخ‌ها انجام می‌شود.
      </p>
      <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
        <Link
          href="/demo"
          className="inline-flex min-h-11 items-center gap-2 rounded-card bg-site-accent px-5 text-sm font-semibold text-site-on-accent transition-colors hover:bg-site-accent-hover"
        >
          در این فاصله، یک دموی خصوصی رزرو کنید
          <ArrowLeft aria-hidden="true" className="size-4" />
        </Link>
        <button type="button" onClick={onStartOver} className="min-h-11 rounded-card px-4 text-sm font-semibold text-site-ink-soft hover:bg-site-panel">
          ثبت درخواست دیگر
        </button>
      </div>
    </div>
  );
}
