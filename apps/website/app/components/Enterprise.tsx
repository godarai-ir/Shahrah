import Link from 'next/link';
import { Eyebrow, sectionCopy, sectionHeading } from './Eyebrow';
import { Reveal } from './Reveal';

type Column = {
  number: string;
  title: string;
  items: string[];
  panel: string;
  hover: string;
  dot: string;
  numberColor: string;
};

const columns: Column[] = [
  {
    number: '۰۱',
    title: 'امنیت و حریم خصوصی',
    items: [
      'گاردریل‌ها داده‌ی حساس را قبل از رسیدن به مدل می‌گیرند',
      'هر دستیار فقط به ابزارهای مجاز خودش دسترسی دارد',
      'هر درخواست در گزارش رخدادها ثبت می‌شود',
    ],
    panel: 'bg-rose-50 dark:bg-rose-500/10',
    hover: 'hover:bg-rose-100/50 dark:hover:bg-rose-500/10',
    dot: 'bg-rose-500',
    numberColor: 'text-rose-500',
  },
  {
    number: '۰۲',
    title: 'مدیریت و مصرف',
    items: [
      'مشخص می‌کنید چه کسی از کدام دستیار استفاده کند',
      'برای هر تیم سقف مصرف می‌گذارید',
      'کیف پول تومانی و گزارش مالی روشن',
    ],
    panel: 'bg-green-50 dark:bg-green-500/10',
    hover: 'hover:bg-green-100/50 dark:hover:bg-green-500/10',
    dot: 'bg-green-500',
    numberColor: 'text-green-600 dark:text-green-400',
  },
  {
    number: '۰۳',
    title: 'استقرار و پشتیبانی',
    items: [
      'روی ابر گدارAI، همین امروز و بدون نصب',
      'داخل سرور خودتان، برای داده‌ی محرمانه',
      'رابط راست‌به‌چپ و پشتیبانی فارسی',
    ],
    panel: 'bg-blue-50 dark:bg-blue-500/10',
    hover: 'hover:bg-blue-100/50 dark:hover:bg-blue-500/10',
    dot: 'bg-blue-500',
    numberColor: 'text-blue-500',
  },
];

export function Enterprise() {
  return (
    <section id="terms" className="w-full bg-white py-14 dark:bg-slate-950 lg:py-24">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-12 px-6">
        <div className="flex flex-col gap-6">
          <Reveal>
            <Eyebrow>برای سازمان‌ها</Eyebrow>
          </Reveal>
          <Reveal>
            <h2 className={`max-w-3xl ${sectionHeading}`}>داده و دسترسی‌ها همیشه دست خودتان است</h2>
          </Reveal>
          <Reveal>
            <p className={`max-w-3xl ${sectionCopy}`}>
              دستیاری که به CRM و اسناد شما سر می‌زند باید قاعده داشته باشد. قاعده را شما می‌نویسید؛ اجرایش با ماست.
            </p>
          </Reveal>
        </div>

        <Reveal className="flex flex-col gap-3">
          <div className="flex flex-wrap items-baseline justify-between gap-3 px-1">
            <span className="text-sm text-slate-500 dark:text-gray-400">کجا اجرا شود</span>
            <Link
              href="/on-premise"
              className="group inline-flex items-center gap-2 text-sm font-medium text-slate-950 transition-colors hover:text-blue-500 dark:text-white"
            >
              درخواست استقرار درون‌سازمانی
              <span className="transition-transform duration-200 group-hover:-translate-x-0.5">←</span>
            </Link>
          </div>

          <div className="grid grid-cols-1 overflow-hidden rounded-3xl md:grid-cols-3">
            {columns.map((column) => (
              <div key={column.number} className={`flex flex-col ${column.panel}`}>
                <header className="flex items-center gap-4 px-7 pb-6 pt-7">
                  <div className="flex flex-col">
                    <span className={`text-base font-semibold leading-tight md:text-lg ${column.numberColor}`}>{column.number}</span>
                    <span className="text-base font-semibold leading-tight text-slate-950 dark:text-white">{column.title}</span>
                  </div>
                </header>
                <ul className="m-0 flex list-none flex-col p-0 pb-3">
                  {column.items.map((item) => (
                    <li
                      key={item}
                      className={`group flex items-start gap-3 px-7 py-3 transition-colors duration-150 ${column.hover}`}
                    >
                      <span className={`mt-2.5 block h-1.5 w-1.5 shrink-0 rounded-full ${column.dot}`} />
                      <span className="text-sm leading-7 text-slate-950 dark:text-gray-200">{item}</span>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </Reveal>
      </div>
    </section>
  );
}
