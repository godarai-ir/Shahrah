'use client';

import { useId, useState } from 'react';
import { Check } from 'lucide-react';
import { Eyebrow, sectionHeading } from './Eyebrow';
import { MediaSlot } from './MediaSlot';
import { Reveal } from './Reveal';
import type { HomeMediaId } from '@/app/lib/home-media';

type Task = { title: string; text: string };

type Team = {
  id: string;
  label: string;
  tasks: Task[];
  media: HomeMediaId;
};

const teams: Team[] = [
  {
    id: 'sales',
    label: 'فروش',
    tasks: [
      { title: 'آماده‌سازی جلسه', text: 'خلاصه‌ی مذاکره‌های قبلی و قول‌های داده‌شده را از CRM بیرون می‌کشد.' },
      { title: 'ثبت بعد از تماس', text: 'یادداشت‌ها را مرتب می‌کند و کار بعدی را در CRM ثبت می‌کند.' },
      { title: 'پیگیری مشتری', text: 'برای مشتری‌هایی که مدتی جواب نداده‌اند پیش‌نویس پیام آماده می‌کند.' },
    ],
    media: 'teamSales',
  },
  {
    id: 'support',
    label: 'پشتیبانی',
    tasks: [
      { title: 'پیش‌نویس پاسخ', text: 'از روی مستندات و سابقه‌ی خود مشتری جواب تیکت را می‌نویسد.' },
      { title: 'دسته‌بندی پیام‌ها', text: 'پیام‌های رسیده به بله، ایتا، واتساپ یا تلگرام را دسته‌بندی می‌کند و فوری‌ها را جلو می‌اندازد.' },
      { title: 'تأیید قبل از ارسال', text: 'اگر بخواهید، هیچ جوابی بدون تأیید همکار پشتیبانی برای مشتری نمی‌رود.' },
    ],
    media: 'teamSupport',
  },
  {
    id: 'finance',
    label: 'مالی',
    tasks: [
      { title: 'تطبیق فاکتور', text: 'فاکتورها را با سفارش‌ها مقایسه می‌کند و هر جا نخواند علامت می‌زند.' },
      { title: 'گزارش آخر ماه', text: 'از روی جدول‌های خودتان پیش‌نویس گزارش را آماده می‌کند.' },
      { title: 'کنترل مصرف', text: 'مصرف هوش مصنوعی هر تیم را کنار سقفش نشان می‌دهد.' },
    ],
    media: 'teamFinance',
  },
  {
    id: 'legal',
    label: 'حقوقی',
    tasks: [
      { title: 'مرور اولیه', text: 'بندهایی را که با سیاست شرکت نمی‌خواند علامت می‌زند و دلیلش را می‌نویسد.' },
      { title: 'مقایسه نسخه‌ها', text: 'تفاوت نسخه‌های مختلف یک قرارداد را فهرست می‌کند.' },
      { title: 'دستورالعمل خودتان', text: 'مرور بر اساس چک‌لیست و قاعده‌های خودتان انجام می‌شود.' },
    ],
    media: 'teamLegal',
  },
  {
    id: 'hr',
    label: 'منابع انسانی',
    tasks: [
      { title: 'سؤال‌های تکراری', text: 'درباره‌ی مرخصی، بیمه و ماموریت از روی آیین‌نامه‌ی خودتان جواب می‌دهد.' },
      { title: 'ورود کارمند جدید', text: 'فهرست کارهای هفته‌ی اول را برای هر کارمند آماده می‌کند.' },
      { title: 'دسترسی محدود', text: 'خودتان تعیین می‌کنید هر کس به چه چیزی دسترسی داشته باشد.' },
    ],
    media: 'teamHr',
  },
  {
    id: 'engineering',
    label: 'فنی',
    tasks: [
      { title: 'تسک تا کد', text: 'تسک‌های Jira را با تغییرهای Git و سندهای Notion کنار هم می‌گذارد.' },
      { title: 'ریشه‌یابی خطا', text: 'گزارش رخدادها و مستندات داخلی را می‌گردد و علت احتمالی را پیشنهاد می‌دهد.' },
      { title: 'API سازگار با OpenAI', text: 'اگر خودتان محصول می‌سازید، همان درگاه را مستقیم هم می‌توانید بردارید.' },
    ],
    media: 'teamEngineering',
  },
];

export function TeamTabs() {
  const baseId = useId();
  const [activeId, setActiveId] = useState(teams[0].id);
  const active = teams.find((team) => team.id === activeId) ?? teams[0];

  return (
    <section id="teams" className="w-full bg-white py-14 dark:bg-slate-950 lg:py-24">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-12 px-6">
        <div className="flex flex-col items-center gap-6 text-center">
          <Reveal>
            <Eyebrow>کار هر تیم</Eyebrow>
          </Reveal>
          <Reveal>
            <h2 className={`max-w-3xl ${sectionHeading}`}>هر تیم، یک دستیار</h2>
          </Reveal>
          <Reveal>
            <p className="m-0 max-w-xl text-base leading-8 text-slate-600 dark:text-gray-400">
              چند کاری که معمولاً اول به دستیار سپرده می‌شود.
            </p>
          </Reveal>
        </div>

        <Reveal>
          <div
            role="tablist"
            aria-label="تیم‌ها"
            className="relative flex w-full overflow-x-auto overflow-y-hidden border-b border-gray-200 dark:border-gray-800"
          >
            {teams.map((team) => {
              const selected = team.id === activeId;

              return (
                <button
                  key={team.id}
                  type="button"
                  role="tab"
                  id={`${baseId}-tab-${team.id}`}
                  aria-selected={selected}
                  aria-controls={`${baseId}-panel`}
                  onClick={() => setActiveId(team.id)}
                  className={`-mb-px flex-1 whitespace-nowrap border-b-2 px-5 py-5 text-base transition-colors duration-200 md:text-lg ${
                    selected
                      ? 'border-slate-950 font-semibold text-slate-950 dark:border-white dark:text-white'
                      : 'border-transparent font-medium text-slate-500 hover:text-slate-950 dark:text-gray-400 dark:hover:text-white'
                  }`}
                >
                  {team.label}
                </button>
              );
            })}
          </div>
        </Reveal>

        <div
          role="tabpanel"
          id={`${baseId}-panel`}
          aria-labelledby={`${baseId}-tab-${active.id}`}
          className="grid grid-cols-1 gap-10 lg:grid-cols-2 lg:gap-16"
        >
          <div className="relative">
            <MediaSlot key={active.media} id={active.media} className="rounded-2xl lg:aspect-auto lg:h-full lg:min-h-96" />
          </div>
          <div className="flex flex-col gap-10 lg:gap-12 lg:py-4">
            <h3 className="m-0 whitespace-pre-line text-3xl font-semibold leading-tight text-slate-950 dark:text-white md:text-4xl lg:text-5xl">
              {`تیم ${active.label}\nبا دستیار گدارAI`}
            </h3>
            <ul className="m-0 flex list-none flex-col p-0">
              {active.tasks.map((task, index) => (
                <li
                  key={task.title}
                  className={`flex items-start gap-5 py-5 ${index === 0 ? '' : 'border-t border-gray-200 dark:border-gray-800'}`}
                >
                  <span className="flex h-7 items-center pt-px">
                    <Check className="h-5 w-5 text-blue-500" />
                  </span>
                  <div className="flex flex-col gap-1.5">
                    <span className="text-base font-semibold text-slate-950 dark:text-white">{task.title}</span>
                    <span className="text-sm leading-7 text-slate-600 dark:text-gray-400">{task.text}</span>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </section>
  );
}
