'use client';

import { ArrowLeft, Mail, MapPin, Phone } from 'lucide-react';
import Link from 'next/link';
import { FormEvent, useState } from 'react';

import { PageHero, Section } from '@/app/components/patterns';
import { buttonPrimary, iconWell, sectionHeading, textLink } from '@/app/components/sectionType';
import { submitForm } from '@/app/lib/submit-form';

type SubmitState = 'idle' | 'submitting' | 'success' | 'error';

const fieldClass =
  'w-full rounded-field border border-site-line-strong bg-site-surface px-4 py-3 text-body text-site-ink placeholder:text-site-ink-soft focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus';
const labelClass = 'text-label font-medium text-site-ink';

const contacts = [
  { icon: Mail, title: 'ایمیل', value: 'support@godarai.ir', description: 'پاسخگویی در کمتر از ۲۴ ساعت', dir: 'ltr' as const },
  { icon: Phone, title: 'تلفن', value: '021-4415-9743', description: 'شنبه تا چهارشنبه، ۹ صبح تا ۶ عصر', dir: 'ltr' as const },
  {
    icon: MapPin,
    title: 'آدرس',
    value: 'تهران، کارخانه نوآوری آزادی',
    description: 'بزرگراه شهید لشکری (جاده مخصوص کرج) نبش خیابان شهید آزمون نیا، پلاک 31',
    dir: undefined,
  },
];

export function Contact() {
  const [submitState, setSubmitState] = useState<SubmitState>('idle');
  const [submitMessage, setSubmitMessage] = useState('');

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const formData = new FormData(form);

    setSubmitState('submitting');
    setSubmitMessage('');

    try {
      await submitForm('/api/contact', Object.fromEntries(formData.entries()), 'ارسال پیام انجام نشد. لطفاً دوباره تلاش کنید.');

      form.reset();
      setSubmitState('success');
      setSubmitMessage('پیام شما با موفقیت ارسال شد.');
    } catch (error) {
      setSubmitState('error');
      setSubmitMessage(
        error instanceof Error ? error.message : 'ارسال پیام انجام نشد. لطفاً دوباره تلاش کنید.',
      );
    }
  }

  return (
    <main className="bg-site-ground">
      <PageHero
        title="با ما در تماس باشید."
        lead="سؤال، پیشنهاد یا بازخوردی دارید؟ دوست داریم از شما بشنویم."
        primary={{ href: '#contact-form', label: 'نوشتن پیام' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
      />

      <Section id="contact" tone="panel" layout="grid gap-12 lg:grid-cols-2 lg:gap-16">
        <div className="flex flex-col gap-8">
          <h2 className={sectionHeading}>راه‌های ارتباطی</h2>
          <ul className="m-0 flex list-none flex-col divide-y divide-site-line border-y border-site-line p-0">
            {contacts.map((contact) => (
              <li key={contact.title} className="flex gap-4 py-5">
                <span className={`size-10 ${iconWell}`}>
                  <contact.icon aria-hidden="true" className="size-5" />
                </span>
                <div className="flex min-w-0 flex-col gap-1">
                  <h3 className="m-0 text-label font-semibold text-site-ink">{contact.title}</h3>
                  <p dir={contact.dir} className="m-0 text-body text-site-ink">{contact.value}</p>
                  <p className="m-0 text-caption text-site-ink-soft">{contact.description}</p>
                </div>
              </li>
            ))}
          </ul>
          <Link href="/demo" className={textLink}>
            گفت‌وگو درباره‌ی خرید؟ درخواست دمو
            <ArrowLeft aria-hidden="true" className="size-4" />
          </Link>
        </div>

        <div className="flex flex-col gap-8 rounded-card border border-site-line bg-site-surface p-6 lg:p-8">
          <h2 className={sectionHeading}>پیام بگذارید</h2>
          <form id="contact-form" className="flex flex-col gap-6" onSubmit={handleSubmit}>
            <div className="flex flex-col gap-2">
              <label className={labelClass} htmlFor="contact-name">نام و نام خانوادگی</label>
              <input id="contact-name" name="name" type="text" autoComplete="name" maxLength={120} required className={fieldClass} placeholder="نام شما" />
            </div>
            <div className="flex flex-col gap-2">
              <label className={labelClass} htmlFor="contact-email">ایمیل</label>
              <input
                id="contact-email"
                name="email"
                type="email"
                autoComplete="email"
                maxLength={254}
                required
                className={`${fieldClass} text-left`}
                placeholder="email@example.com"
                dir="ltr"
              />
            </div>
            <div className="flex flex-col gap-2">
              <label className={labelClass} htmlFor="contact-subject">موضوع</label>
              <input id="contact-subject" name="subject" type="text" maxLength={200} required className={fieldClass} placeholder="موضوع پیام شما" />
            </div>
            <div className="flex flex-col gap-2">
              <label className={labelClass} htmlFor="contact-message">پیام</label>
              <textarea id="contact-message" name="message" rows={5} maxLength={5000} required className={fieldClass} placeholder="پیام خود را بنویسید…" />
            </div>
            <div className="hidden" aria-hidden="true">
              <label htmlFor="contact-website">وب‌سایت</label>
              <input id="contact-website" name="website" type="text" tabIndex={-1} autoComplete="off" />
            </div>
            <button type="submit" disabled={submitState === 'submitting'} className={`${buttonPrimary} w-full disabled:cursor-not-allowed disabled:opacity-60`}>
              {submitState === 'submitting' ? 'در حال ارسال…' : 'ارسال پیام'}
            </button>
            {submitMessage && (
              <p role="status" aria-live="polite" className={`m-0 text-label ${submitState === 'success' ? 'text-site-success' : 'text-site-danger'}`}>
                {submitMessage}
              </p>
            )}
          </form>
        </div>
      </Section>
    </main>
  );
}
