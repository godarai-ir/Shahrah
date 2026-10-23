'use client';

import { useRouter } from 'next/navigation';
import { ArrowLeft, ArrowRight, Check, Loader2 } from 'lucide-react';
import { useEffect, useRef, useState, type FormEvent, type ReactNode, type RefObject } from 'react';

import { BuildGuildGiftNote } from '@/app/components/BuildGuildGiftNote';
import {
  emptyAnswers,
  isValidMobile,
  maxChallenges,
  otherValue,
  surveyQuestions as q,
  textMax,
  writeStoredSurvey,
  type BuildGuildAnswers,
  type SurveyOption,
} from '@/app/lib/build-guild-survey';

const totalSteps = 5;

/** Fields each step owns, so "skip" clears exactly that question's answer. */
const stepFields: Array<Array<keyof BuildGuildAnswers>> = [
  ['fullName', 'mobile'],
  ['role', 'roleOther'],
  ['orgSize'],
  ['challenges', 'challengesOther'],
  ['aiManagement', 'aiManagementOther'],
];

function toFa(value: number) {
  return value.toLocaleString('fa-IR');
}

export function BuildGuildSurvey() {
  const router = useRouter();
  const [step, setStep] = useState(0);
  const [answers, setAnswers] = useState<BuildGuildAnswers>(emptyAnswers);
  const [mobileError, setMobileError] = useState('');
  const [finishing, setFinishing] = useState(false);
  const honeypotRef = useRef<HTMLInputElement>(null);
  const headingRef = useRef<HTMLHeadingElement>(null);
  const isFirstRender = useRef(true);

  useEffect(() => {
    router.prefetch('/onboarding');
  }, [router]);

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      return;
    }
    headingRef.current?.focus({ preventScroll: true });
  }, [step]);

  function update<K extends keyof BuildGuildAnswers>(key: K, value: BuildGuildAnswers[K]) {
    setAnswers((current) => ({ ...current, [key]: value }));
  }

  async function finish(finalAnswers: BuildGuildAnswers) {
    setFinishing(true);
    // Kept for the onboarding page, which shows the gift note and carries the answers as a hidden input.
    writeStoredSurvey(finalAnswers);
    try {
      await fetch('/api/build-guild-survey', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ answers: finalAnswers, website: honeypotRef.current?.value ?? '' }),
        keepalive: true,
      });
    } catch {
      // The gift flow must not stall on email delivery.
    }
    router.push('/onboarding?ref=build-guild');
  }

  function advance(nextAnswers: BuildGuildAnswers) {
    if (step < totalSteps - 1) setStep(step + 1);
    else void finish(nextAnswers);
  }

  function handleContinue(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (finishing) return;
    if (step === 0 && answers.mobile.trim() && !isValidMobile(answers.mobile)) {
      setMobileError('شماره همراه معتبر نیست؛ مثلاً ۰۹۱۲۰۰۰۰۰۰۰');
      document.getElementById('bg-mobile')?.focus();
      return;
    }
    advance(answers);
  }

  function handleSkip() {
    if (finishing) return;
    const cleared = { ...answers };
    for (const field of stepFields[step]) {
      (cleared as Record<string, unknown>)[field] = emptyAnswers[field];
    }
    setAnswers(cleared);
    setMobileError('');
    advance(cleared);
  }

  const progress = ((step + 1) / totalSteps) * 100;

  return (
    <main className="relative overflow-hidden px-4 py-12 sm:px-6 sm:py-16 lg:px-8">
      <div className="mx-auto max-w-2xl">
        <header className="mb-8 text-center">
          <h1 className="mb-4 text-2xl font-extrabold leading-relaxed text-site-ink sm:text-3xl">
            🎁 هدیهٔ ما به شرکت‌کنندگان Build-Guild: ۱ میلیون تومان اعتبار رایگان API هوش مصنوعی
          </h1>
          <p className="text-base leading-8 text-site-ink-soft">
            پاسخ‌های شما به ما کمک می‌کند تجربهٔ گدارAI را برای تیم‌ها و سازمان‌ها، بهتر طراحی کنیم. پاسخ به همهٔ سؤال‌ها اختیاری است.
          </p>
        </header>

        <BuildGuildGiftNote className="mb-6" />

        <form
          onSubmit={handleContinue}
          noValidate
          className="overflow-hidden rounded-card border border-site-line bg-site-surface"
        >
          <div className="px-5 pt-5 sm:px-8 sm:pt-7">
            <div className="mb-2 text-xs font-medium text-site-ink-soft">
              سؤال {toFa(step + 1)} از {toFa(totalSteps)}
            </div>
            <div
              className="h-1.5 overflow-hidden rounded-full bg-site-panel"
              role="progressbar"
              aria-valuemin={1}
              aria-valuemax={totalSteps}
              aria-valuenow={step + 1}
              aria-valuetext={`سؤال ${toFa(step + 1)} از ${toFa(totalSteps)}`}
            >
              <div className="h-full rounded-full bg-site-accent transition-[width] duration-500" style={{ width: `${progress}%` }} />
            </div>
          </div>

          <div key={step} className="animate-in fade-in slide-in-from-left-2 px-5 py-6 duration-300 sm:px-8">
            {step === 0 && (
              <QuestionShell headingRef={headingRef} title={q.contact.label}>
                <div className="grid gap-5 sm:grid-cols-2">
                  <TextField id="bg-fullName" label={q.contact.fullName} value={answers.fullName} autoComplete="name" placeholder="مثلاً سارا محمدی" onChange={(value) => update('fullName', value)} />
                  <TextField
                    id="bg-mobile"
                    label={q.contact.mobile}
                    value={answers.mobile}
                    type="tel"
                    inputMode="tel"
                    dir="ltr"
                    autoComplete="tel"
                    placeholder="09120000000"
                    error={mobileError}
                    onChange={(value) => {
                      update('mobile', value.replace(/[^\d۰-۹+\s-]/g, ''));
                      if (mobileError) setMobileError('');
                    }}
                  />
                </div>
              </QuestionShell>
            )}

            {step === 1 && (
              <QuestionShell headingRef={headingRef} title={q.role.label}>
                <OptionList name="role" options={q.role.options} selected={[answers.role]} onToggle={(value) => update('role', value)} />
                {answers.role === otherValue && (
                  <OtherInput id="bg-roleOther" label="نقش خود را وارد کنید" value={answers.roleOther} onChange={(value) => update('roleOther', value)} />
                )}
              </QuestionShell>
            )}

            {step === 2 && (
              <QuestionShell headingRef={headingRef} title={q.orgSize.label}>
                <OptionList name="orgSize" options={q.orgSize.options} selected={[answers.orgSize]} onToggle={(value) => update('orgSize', value)} />
              </QuestionShell>
            )}

            {step === 3 && (
              <QuestionShell
                headingRef={headingRef}
                title={q.challenges.label}
                hint={`حداکثر ${toFa(maxChallenges)} گزینه — ${toFa(answers.challenges.length)} از ${toFa(maxChallenges)} انتخاب شده`}
              >
                <OptionList
                  name="challenges"
                  multiple
                  options={q.challenges.options}
                  selected={answers.challenges}
                  maxReached={answers.challenges.length >= maxChallenges}
                  onToggle={(value) =>
                    update(
                      'challenges',
                      answers.challenges.includes(value)
                        ? answers.challenges.filter((item) => item !== value)
                        : answers.challenges.length < maxChallenges
                          ? [...answers.challenges, value]
                          : answers.challenges,
                    )
                  }
                />
                {answers.challenges.includes(otherValue) && (
                  <OtherInput id="bg-challengesOther" label="چالش دیگر را بنویسید" value={answers.challengesOther} onChange={(value) => update('challengesOther', value)} />
                )}
              </QuestionShell>
            )}

            {step === 4 && (
              <QuestionShell headingRef={headingRef} title={q.aiManagement.label}>
                <OptionList name="aiManagement" options={q.aiManagement.options} selected={[answers.aiManagement]} onToggle={(value) => update('aiManagement', value)} />
                {answers.aiManagement === otherValue && (
                  <OtherInput id="bg-aiManagementOther" label="روش مدیریت را بنویسید" value={answers.aiManagementOther} onChange={(value) => update('aiManagementOther', value)} />
                )}
              </QuestionShell>
            )}

            <div className="hidden" aria-hidden="true">
              <label htmlFor="bg-website">وب‌سایت</label>
              <input ref={honeypotRef} id="bg-website" name="website" type="text" tabIndex={-1} autoComplete="off" />
            </div>
          </div>

          <div className="flex flex-wrap items-center justify-between gap-3 border-t border-site-line px-5 py-4 sm:px-8">
            {step > 0 ? (
              <button
                type="button"
                onClick={() => setStep(step - 1)}
                disabled={finishing}
                className="inline-flex min-h-11 items-center gap-2 rounded-card px-3 text-sm font-semibold text-site-ink-soft transition-colors hover:bg-site-panel focus-visible:outline-2 focus-visible:outline-site-focus disabled:opacity-50"
              >
                <ArrowRight aria-hidden="true" className="size-4" />
                بازگشت
              </button>
            ) : (
              <span />
            )}
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={handleSkip}
                disabled={finishing}
                className="min-h-11 rounded-card px-3 text-sm font-medium text-site-ink-soft transition-colors hover:bg-site-panel hover:text-site-ink focus-visible:outline-2 focus-visible:outline-site-focus disabled:opacity-50"
              >
                رد کردن این سؤال
              </button>
              <button
                type="submit"
                disabled={finishing}
                className="inline-flex min-h-11 items-center gap-2 rounded-card bg-site-accent px-6 text-sm font-semibold text-site-on-accent transition-colors hover:bg-site-accent-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus disabled:cursor-not-allowed disabled:opacity-60"
              >
                ادامه
                {finishing ? <Loader2 aria-hidden="true" className="size-4 animate-spin" /> : <ArrowLeft aria-hidden="true" className="size-4" />}
              </button>
            </div>
          </div>
        </form>
      </div>
    </main>
  );
}

const inputClassName =
  'w-full rounded-card border bg-site-surface px-4 py-3 text-sm text-site-ink transition-colors placeholder:text-site-ink-soft focus:outline-none focus:ring-2 focus:ring-site-focus';

function QuestionShell({
  headingRef,
  title,
  hint,
  children,
}: {
  headingRef: RefObject<HTMLHeadingElement | null>;
  title: string;
  hint?: string;
  children: ReactNode;
}) {
  return (
    <fieldset>
      <legend className="contents">
        <h2 ref={headingRef} tabIndex={-1} className="text-lg font-bold leading-8 text-site-ink outline-none sm:text-xl">
          {title}
        </h2>
      </legend>
      {hint && <p className="mt-1 text-sm text-site-ink-soft">{hint}</p>}
      <div className="mt-5">{children}</div>
    </fieldset>
  );
}

function TextField({
  id,
  label,
  value,
  onChange,
  error,
  dir,
  type = 'text',
  ...inputProps
}: {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  error?: string;
  dir?: 'ltr';
  type?: string;
  autoComplete?: string;
  placeholder?: string;
  inputMode?: 'tel';
}) {
  return (
    <div>
      <label htmlFor={id} className="mb-2 block text-sm font-medium text-site-ink">
        {label}
      </label>
      <input
        id={id}
        type={type}
        dir={dir}
        maxLength={textMax}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        aria-invalid={!!error}
        aria-describedby={error ? `${id}-error` : undefined}
        className={`${inputClassName} ${dir === 'ltr' ? 'text-left' : ''} ${error ? 'border-site-danger' : 'border-site-line'}`}
        {...inputProps}
      />
      {error && (
        <p id={`${id}-error`} className="mt-1.5 text-xs text-site-danger">
          {error}
        </p>
      )}
    </div>
  );
}

function OptionList({
  name,
  options,
  selected,
  onToggle,
  multiple = false,
  maxReached = false,
}: {
  name: string;
  options: SurveyOption[];
  selected: string[];
  onToggle: (value: string) => void;
  multiple?: boolean;
  maxReached?: boolean;
}) {
  return (
    <div className="grid gap-2">
      {options.map((option) => {
        const isChecked = selected.includes(option.value);
        const isDisabled = multiple && maxReached && !isChecked;
        return (
          <label
            key={option.value}
            className={`flex min-h-12 select-none items-center gap-3 rounded-card border px-4 py-2.5 text-sm leading-6 transition-colors has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-site-focus ${
              isChecked
                ? 'border-site-line-strong bg-site-sand font-medium text-site-accent-text'
                : isDisabled
                  ? 'cursor-not-allowed border-site-line text-site-ink-soft'
                  : 'cursor-pointer border-site-line text-site-ink hover:border-site-line hover:bg-site-sand'
            }`}
          >
            <input
              type={multiple ? 'checkbox' : 'radio'}
              name={name}
              value={option.value}
              checked={isChecked}
              disabled={isDisabled}
              onChange={() => onToggle(option.value)}
              className="sr-only"
            />
            <span
              aria-hidden="true"
              className={`grid size-5 shrink-0 place-items-center border ${multiple ? 'rounded-md' : 'rounded-full'} ${
                isChecked ? 'border-site-line-strong bg-site-accent text-site-on-accent' : 'border-site-line'
              }`}
            >
              {isChecked && <Check className="size-3.5" strokeWidth={3} />}
            </span>
            <span>{option.label}</span>
          </label>
        );
      })}
    </div>
  );
}

function OtherInput({ id, label, value, onChange }: { id: string; label: string; value: string; onChange: (value: string) => void }) {
  return (
    <div className="animate-in fade-in slide-in-from-top-1 mt-3 duration-200">
      <label htmlFor={id} className="sr-only">
        {label}
      </label>
      <input
        id={id}
        type="text"
        autoFocus
        maxLength={textMax}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={label}
        className={`${inputClassName} border-site-line`}
      />
    </div>
  );
}
