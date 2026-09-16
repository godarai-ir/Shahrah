'use client';

import { Mail, Phone, MapPin } from 'lucide-react';
import { FormEvent, useState } from 'react';

import { submitForm } from '@/app/lib/submit-form';

type SubmitState = 'idle' | 'submitting' | 'success' | 'error';

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
    <div className="py-24 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <h1 className="text-5xl font-bold mb-6 bg-gradient-to-l from-indigo-400 to-purple-500 bg-clip-text text-transparent">
            تماس با ما
          </h1>
          <p className="text-xl text-slate-600 dark:text-gray-400 max-w-2xl mx-auto">
            سوال، پیشنهاد یا بازخوردی دارید؟ دوست داریم از شما بشنویم
          </p>
        </div>

        <div className="grid md:grid-cols-2 gap-12">
          <div>
            <h2 className="text-2xl font-bold mb-8">راه‌های ارتباطی</h2>
            <div className="space-y-6">
              {[
                {
                  icon: Mail,
                  title: 'ایمیل',
                  value: 'support@godarai.ir',
                  description: 'پاسخگویی در کمتر از ۲۴ ساعت',
                },
                {
                  icon: Phone,
                  title: 'تلفن',
                  value: '021-4415-9743',
                  description: 'شنبه تا چهارشنبه، ۹ صبح تا ۶ عصر',
                },
                {
                  icon: MapPin,
                  title: 'آدرس',
                  value: 'تهران، کارخانه نوآوری آزادی',
                  description: 'بزرگراه شهید لشکری (جاده مخصوص کرج) نبش خیابان شهید آزمون نیا، پلاک 31',
                },
              ].map((contact) => (
                <div key={contact.title} className="flex gap-4 p-6 rounded-xl border border-gray-200 dark:border-gray-800">
                  <contact.icon className="w-6 h-6 text-indigo-500 shrink-0 mt-1" />
                  <div>
                    <h3 className="font-bold mb-1">{contact.title}</h3>
                    <p className="text-slate-900 dark:text-white mb-1">{contact.value}</p>
                    <p className="text-sm text-slate-600 dark:text-gray-400">{contact.description}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div>
            <h2 className="text-2xl font-bold mb-8">پیام بگذارید</h2>
            <form className="space-y-6" onSubmit={handleSubmit}>
              <div>
                <label className="block text-sm font-medium mb-2" htmlFor="contact-name">نام و نام خانوادگی</label>
                <input
                  id="contact-name"
                  name="name"
                  type="text"
                  autoComplete="name"
                  maxLength={120}
                  required
                  className="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="نام شما"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2" htmlFor="contact-email">ایمیل</label>
                <input
                  id="contact-email"
                  name="email"
                  type="email"
                  autoComplete="email"
                  maxLength={254}
                  required
                  className="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="email@example.com"
                  dir="ltr"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2" htmlFor="contact-subject">موضوع</label>
                <input
                  id="contact-subject"
                  name="subject"
                  type="text"
                  maxLength={200}
                  required
                  className="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="موضوع پیام شما"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-2" htmlFor="contact-message">پیام</label>
                <textarea
                  id="contact-message"
                  name="message"
                  rows={5}
                  maxLength={5000}
                  required
                  className="w-full px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  placeholder="پیام خود را بنویسید..."
                />
              </div>
              <div className="hidden" aria-hidden="true">
                <label htmlFor="contact-website">وب‌سایت</label>
                <input id="contact-website" name="website" type="text" tabIndex={-1} autoComplete="off" />
              </div>
              <button
                type="submit"
                disabled={submitState === 'submitting'}
                className="w-full py-3 rounded-lg bg-indigo-500 hover:bg-indigo-600 disabled:cursor-not-allowed disabled:opacity-60 text-white font-medium transition-all shadow-[0_0_15px_rgba(99,102,241,0.3)]"
              >
                {submitState === 'submitting' ? 'در حال ارسال...' : 'ارسال پیام'}
              </button>
              {submitMessage && (
                <p
                  role="status"
                  aria-live="polite"
                  className={submitState === 'success' ? 'text-sm text-emerald-600' : 'text-sm text-red-600'}
                >
                  {submitMessage}
                </p>
              )}
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
