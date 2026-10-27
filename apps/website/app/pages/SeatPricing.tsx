'use client';

import { useState, type ReactNode } from 'react';
import Link from 'next/link';
import { ArrowDown, Check, Minus, Search } from 'lucide-react';
import { ClosingBand, ComparisonTable, FAQ, Section } from '@/app/components/patterns';
import { buttonPrimary, buttonSecondary, sectionCopy, sectionHeading, textLink } from '@/app/components/sectionType';

// Public seat pricing, following the owner's pricing reference: free seats, then Pro and Max at fixed monthly
// prices. Task counts are estimates from the same credit formulas the reference uses.

const fa = (value: number) => value.toLocaleString('fa-IR');

type Model = 'light' | 'mid' | 'heavy';
type Task = 'msg' | 'text' | 'office' | 'image' | 'kb';
type SeatId = 'free' | 'pro' | 'max';

const modelRates: Record<Model, [number, number]> = { light: [0.125, 1], mid: [1, 5], heavy: [2, 10] };
const modelNames: Record<Model, string> = { light: 'سبک', mid: 'میان‌رده', heavy: 'سنگین' };

function taskCredits(task: Task, model: Model) {
  const [input, output] = modelRates[model];

  return {
    msg: (2 * input + 0.5 * output) / 0.8,
    text: 15 * input + 0.6 * output,
    office: 15 * input + 3 * output + 6,
    image: 10 * (1 + 0.5 * input + 0.3 * output),
    kb: 8 + 8 * input + 0.6 * output,
  }[task];
}

// Rounds a monthly count to a readable step, the way the pricing reference does.
function roundCount(value: number) {
  const step = value < 100 ? 10 : value < 1000 ? 50 : value < 3000 ? 100 : value < 10000 ? 500 : 1000;
  const whole = Math.floor(value / step);

  return (whole + ((value - whole * step) / step >= 0.3 ? 1 : 0)) * step;
}

const monthlyCount = (credits: number, task: Task, model: Model) => roundCount(credits / taskCredits(task, model));

const freeCredits = 500;
const paidSeats = [
  {
    id: 'pro' as const,
    name: 'Pro',
    tagline: 'برای اعضایی که هر روز با هوش مصنوعی کار می‌کنند',
    credits: 4000,
    price: 2250000,
    popular: true,
    roles: ['مدیر پروژه', 'کارشناس فروش و بازاریابی', 'پشتیبانی', 'منابع انسانی'],
    sign: 'بیشتر کارهای روزانه‌اش را با گدار انجام می‌دهد، ولی با فایل‌های حجیم و مدل‌های سنگین زیاد سروکار ندارد.',
  },
  {
    id: 'max' as const,
    name: 'Max',
    tagline: 'برای تحلیل‌گران، تولید محتوا و مصرف سنگین',
    credits: 20000,
    price: 11250000,
    popular: false,
    roles: ['تحلیل‌گر', 'تولیدکننده‌ی محتوا', 'توسعه‌دهنده', 'پژوهشگر', 'مدیر محصول'],
    sign: 'کارش را عملاً با گدار پیش می‌برد: ساعت‌ها در روز، با فایل‌های حجیم یا عامل‌های هوش مصنوعی.',
  },
];

const cardTasks: { task: Task; label: string }[] = [
  { task: 'msg', label: 'پیام گفت‌وگو' },
  { task: 'text', label: 'خلاصه‌ی گزارش ۲۰ صفحه‌ای' },
  { task: 'office', label: 'تحلیل اکسل و ساخت گزارش Word' },
  { task: 'image', label: 'استخراج اطلاعات از عکس و فاکتور' },
];

const quiz: { id: string; label: string; seat: SeatId; text: string }[] = [
  { id: 'try', label: 'فقط می‌خواهد گدار را امتحان کند', seat: 'free', text: 'صندلی رایگان کافی است: ۵۰۰ اعتبار برای شروع. بعد از آن Pro یا Max می‌گیرد.' },
  { id: 'chat', label: 'چت، ترجمه و خلاصه‌سازی', seat: 'pro', text: 'Pro مناسب است؛ مصرف روزمره و سبک تا متوسط.' },
  { id: 'daily', label: 'تحلیل فایل و گزارش، هر روز', seat: 'max', text: 'Max مناسب است؛ تحلیل روزانه‌ی فایل و گزارش.' },
  { id: 'heavy', label: 'فایل‌های سنگین یا عامل هوش مصنوعی', seat: 'max', text: 'Max شروع خوبی است. اگر مصرفش از این هم بیشتر شد، تا سقفی که مدیر تعیین می‌کند از اعتبار مشترک سازمان برمی‌دارد.' },
];

const exampleTasks: { task: Task; category: string; label: string }[] = [
  { task: 'text', category: 'متن', label: 'خلاصه‌ی یک گزارش ۲۰ صفحه‌ای' },
  { task: 'image', category: 'عکس', label: 'استخراج اطلاعات از ۱۰ عکس یا فاکتور اسکن‌شده' },
  { task: 'office', category: 'Word و Excel', label: 'تحلیل یک فایل اکسل و ساخت گزارش Word از آن' },
  { task: 'kb', category: 'منابع دانشی سازمان', label: 'پرسش از اسناد سازمان و گرفتن پاسخ همراه با منبع' },
];

type Cell = 'yes' | 'no' | string;
const comparison: { title: string; rows: { label: string; note?: string; cloud: Cell; dedicated: Cell }[] }[] = [
  {
    title: 'ویژگی‌ها',
    rows: [
      { label: 'دسترسی به بیش از ۱۳۰ مدل', note: 'GPT، Claude، Gemini، Grok، Kimi و…', cloud: 'yes', dedicated: 'yes' },
      { label: 'ورودی و خروجی فایل چندرسانه‌ای', note: 'تصویر، PDF، ساخت و ویرایش فایل Word و Excel، اجرای کد', cloud: 'yes', dedicated: 'yes' },
      { label: 'عامل هوش مصنوعی اختصاصی، چندمرحله‌ای و خودکار', note: 'زمان‌بندی‌شده یا با رویداد', cloud: 'yes', dedicated: 'yes' },
      { label: 'تعامل با ابزارهای بیرون سازمان', note: 'MCP', cloud: 'yes', dedicated: 'yes' },
      { label: 'تعامل با منابع دانشی سازمان', cloud: 'yes', dedicated: 'yes' },
    ],
  },
  {
    title: 'مدیریت و امنیت',
    rows: [
      { label: 'نقش‌های مدیر سازمان، مدیر تیم و عضو', cloud: 'yes', dedicated: 'yes' },
      { label: 'سقف و گزارش مصرف برای هر نفر و تیم', cloud: 'yes', dedicated: 'yes' },
      { label: 'ثبت وقایع مدیریتی', note: 'Audit Log', cloud: 'yes', dedicated: 'yes' },
      { label: 'گاردریل‌های معنایی', cloud: 'yes', dedicated: 'yes' },
      { label: 'ورود یکپارچه', note: 'SSO', cloud: 'no', dedicated: 'yes' },
      { label: 'نصب در محل یا ابر خصوصی', note: 'On-premise / VPC', cloud: 'no', dedicated: 'yes' },
    ],
  },
  {
    title: 'پشتیبانی و توسعه',
    rows: [
      { label: 'نوع پشتیبانی', cloud: 'تیکت', dedicated: 'مدیر مشتری اختصاصی' },
      { label: 'توافق‌نامه‌ی سطح خدمات', note: 'SLA', cloud: 'no', dedicated: 'yes' },
      { label: 'دسترسی به API مدل‌ها', note: 'از اعتبار مشترک سازمان', cloud: 'yes', dedicated: 'yes' },
    ],
  },
];

const faqs = [
  {
    question: 'اعتبار چیست؟',
    answer: 'اعتبار واحد مصرف در گدار است. هر عضو با صندلی خودش اعتبار شخصی می‌گیرد و هر پیام، فایل یا اجرای عامل هوش مصنوعی مقداری از آن را مصرف می‌کند. مدل قوی‌تر یا متن طولانی‌تر اعتبار بیشتری مصرف می‌کند.',
  },
  {
    question: 'مصرف اعتبار چطور حساب می‌شود؟',
    answer: 'اعتبار هر پاسخ از دو بخش می‌آید: پردازش مدل، و ابزارهایی که گدار برای جواب دادن اجرا می‌کند، مثل جست‌وجوی وب. یک پیام معمولی با مدل سبک حدود ۰٫۷۵ اعتبار، با مدل میان‌رده حدود ۴٫۵ و با مدل سنگین حدود ۹ اعتبار مصرف می‌کند. اعتبار هر پاسخ زیر همان پاسخ نشان داده می‌شود.',
  },
  {
    question: 'عددهای روی کارت‌ها دقیق است؟',
    answer: 'نه، تقریبی است و فرض می‌کند بخشی از اعتبار صرف ابزارها می‌شود. با مدل سبک تعداد بیشتر می‌شود و با گفت‌وگوی طولانی یا عامل هوش مصنوعی سنگین کمتر.',
  },
  {
    question: 'دوره‌ی آزمایشی دارید؟',
    answer: 'لازم نیست: تا ۵ عضو هر سازمان صندلی رایگان با ۵۰۰ اعتبار می‌گیرند و بدون پرداخت با همه‌ی امکانات کار می‌کنند.',
  },
  {
    question: 'Pro و Max چقدر هزینه دارند؟',
    answer: 'Pro ماهی ۲٬۲۵۰٬۰۰۰ تومان با ۴٬۰۰۰ اعتبار و Max ماهی ۱۱٬۲۵۰٬۰۰۰ تومان با ۲۰٬۰۰۰ اعتبار است. قیمت‌ها بدون مالیات بر ارزش افزوده است و صورتحساب سالانه تخفیف دارد؛ اعتبار در هر حال ماه‌به‌ماه داده می‌شود.',
  },
  {
    question: 'اعتبار مصرف‌نشده به ماه بعد می‌رسد؟',
    answer: 'نه. اعتبار Pro و Max هر ماه از نو داده می‌شود و مانده‌ی ماه قبل منتقل نمی‌شود.',
  },
  {
    question: 'اگر اعتبار یک عضو تمام شود چه می‌شود؟',
    answer: 'پاسخی که در حال نوشته شدن است کامل می‌شود. بعد از آن، اعضای Pro و Max تا سقفی که مدیر تعیین می‌کند از اعتبار مشترک سازمان برمی‌دارند. اگر آن هم تمام شده باشد، کار تا خرید بسته‌ی اعتبار یا شروع ماه بعد متوقف می‌شود. هیچ‌وقت بیش از آنچه خریده‌اید هزینه ثبت نمی‌شود.',
  },
  {
    question: 'اعتبار مشترک سازمان و بسته‌های اعتبار برای چیست؟',
    answer: 'اعتبار مشترک سازمان فقط با خرید بسته پر می‌شود. مصرف API، کلیدهای دسترسی، اجراهای خودکار و گاردریل‌ها از همین اعتبار کم می‌شود، نه از اعتبار شخصی اعضا. بسته‌های بزرگ‌تر برای هر اعتبار ارزان‌ترند و با پایان اشتراک صندلی‌ها از بین نمی‌روند.',
  },
  {
    question: 'اگر عضوی را حذف کنیم، صندلی پولی او چه می‌شود؟',
    answer: 'صندلی به سازمان برمی‌گردد و تا پایان دوره‌ی پرداخت‌شده می‌توانید آن را به عضو دیگری بدهید. هزینه‌ی صندلی برگردانده نمی‌شود.',
  },
  {
    question: 'یادم برود تمدید کنم چه می‌شود؟',
    answer: 'از ۷ روز مانده به پایان دوره، در پنل یادآوری می‌بینید. اگر تمدید نکنید فقط صندلی‌های پولی غیرفعال می‌شوند و هر وقت دوباره صندلی بخرید، اشتراک تازه‌ای شروع می‌شود.',
  },
  {
    question: 'اگر قیمت‌ها تغییر کند چه می‌شود؟',
    answer: 'هر افزایش قیمت دست‌کم ۳۰ روز پیش از اعمال اعلام می‌شود و کاهش‌ها فوراً اعمال می‌شوند. اعتباری که خریده‌اید ارزشش را از دست نمی‌دهد.',
    link: { href: '/pricing/changelog', label: 'تغییرات قیمت' },
  },
  {
    question: 'کی به استقرار اختصاصی نیاز داریم؟',
    answer: 'وقتی داده نباید از شبکه‌ی سازمان بیرون برود، یا به ورود یکپارچه (SSO)، پشتیبانی با SLA و مدیر مشتری اختصاصی نیاز دارید. نیازها و قیمت را با هم بررسی می‌کنیم.',
    link: { href: '/on-premise', label: 'درخواست استقرار اختصاصی' },
  },
  {
    question: 'پلن‌های قبلی (توسعه‌دهنده، تیم، رشد و سازمانی) چه شدند؟',
    answer: 'دیگر پلنی در کار نیست و همه‌ی امکانات برای همه است. هزینه فقط از دو جا می‌آید: صندلی‌های پولی اعضا و بسته‌های اعتبار. اشتراک‌های فعلی به همین مدل منتقل شده‌اند.',
  },
];

const segmentButton = (pressed: boolean) =>
  `inline-flex min-h-11 items-center justify-center rounded-full border px-4 py-2 text-label font-medium transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus ${
    pressed ? 'border-site-ink bg-site-night text-site-night-ink' : 'border-site-line-strong text-site-ink hover:bg-site-panel'
  }`;

function Bullet({ children }: { children: ReactNode }) {
  return (
    <li className="flex items-start gap-3">
      <Check aria-hidden="true" className="mt-1 size-4 shrink-0 text-site-ink" />
      <span className="text-body text-site-ink">{children}</span>
    </li>
  );
}

function Plans() {
  return (
    <section className="bg-site-ground px-4 pb-10 pt-12 sm:px-6 lg:pb-20 lg:pt-20">
      <div className="mx-auto flex w-full max-w-content flex-col gap-10">
        <header className="mx-auto flex max-w-3xl flex-col items-center gap-4 text-center">
          <h1 className="m-0 text-balance text-h1 font-extrabold text-site-ink lg:text-display">برای هر عضو یک صندلی. به اندازه‌ی کارش.</h1>
          <p className={`m-0 max-w-2xl ${sectionCopy}`}>
            پلنی برای انتخاب نیست و همه‌ی امکانات برای همه است. فقط صندلی هر عضو را انتخاب می‌کنید؛ مصرف API هم از اعتبار مشترک سازمان کم می‌شود.
          </p>
        </header>

        <div className="grid gap-6 lg:grid-cols-2">
          <article className="flex flex-col gap-6 rounded-card border border-site-line-strong bg-site-surface p-6 lg:p-8">
            <div className="flex flex-col gap-2">
              <h2 className="m-0 text-h3 font-semibold text-site-ink">روی ابر گدارAI</h2>
              <p className="m-0 text-body text-site-ink-soft">برای شرکت‌ها و تیم‌ها. با چند عضو شروع کنید و هر وقت لازم شد صندلی اضافه کنید.</p>
            </div>
            <div className="flex flex-col gap-1 border-y border-site-line py-5">
              <span className="text-caption text-site-ink-soft">صندلی Pro</span>
              <span className="text-h2 font-extrabold text-site-ink">{fa(paidSeats[0].price)} تومان</span>
              <span className="text-label text-site-ink-soft">برای هر عضو در ماه</span>
            </div>
            <p className="m-0 text-label text-site-ink">
              تا ۵ عضو هر سازمان صندلی رایگان با {fa(freeCredits)} اعتبار می‌گیرند.{' '}
              <a href="#seats" className={textLink}>
                صندلی‌ها
                <ArrowDown aria-hidden="true" className="size-4" />
              </a>
            </p>
            <ul className="m-0 flex list-none flex-col gap-3 p-0">
              <Bullet>دسترسی به بیش از ۱۳۰ مدل، از جمله GPT، Claude، Gemini و Grok</Bullet>
              <Bullet>ساخت عامل هوش مصنوعی اختصاصی برای کارهای تکراری تیم</Bullet>
              <Bullet>تعامل با ابزارهای بیرونی سازمان (MCP)</Bullet>
              <Bullet>گاردریل معنایی برای موضوع‌هایی که سازمان مجاز نمی‌داند</Bullet>
              <Bullet>سقف و گزارش مصرف برای هر عضو و تیم</Bullet>
            </ul>
            <div className="mt-auto flex flex-col gap-3">
              <Link href="/onboarding" className={buttonPrimary}>
                رایگان شروع کنید
              </Link>
              <p className="m-0 text-caption text-site-ink-soft">قیمت‌ها بدون مالیات بر ارزش افزوده است.</p>
            </div>
          </article>

          <article className="flex flex-col gap-6 rounded-card border border-site-line bg-site-sand p-6 lg:p-8">
            <div className="flex flex-col gap-2">
              <h2 className="m-0 text-h3 font-semibold text-site-ink">استقرار اختصاصی</h2>
              <p className="m-0 text-body text-site-ink-soft">
                برای سازمان‌هایی که داده‌شان نباید از شبکه بیرون برود، یا به ورود یکپارچه، تعهد سطح خدمات و مدیر مشتری اختصاصی نیاز دارند.
              </p>
            </div>
            <div className="flex flex-col gap-1 border-y border-site-line-strong py-5">
              <span className="text-h2 font-extrabold text-site-ink">قیمت سفارشی</span>
              <span className="text-label text-site-ink-soft">بر اساس تعداد عضو و حجم مصرف</span>
            </div>
            <p className="m-0 text-label font-medium text-site-ink">همه‌ی امکانات ابر گدارAI، به‌علاوه:</p>
            <ul className="m-0 flex list-none flex-col gap-3 p-0">
              <Bullet>نصب در محل یا ابر خصوصی</Bullet>
              <Bullet>ورود یکپارچه (SSO)</Bullet>
              <Bullet>پشتیبانی اولویت‌دار با تعهد سطح خدمات (SLA)</Bullet>
              <Bullet>مدیر مشتری اختصاصی</Bullet>
            </ul>
            <div className="mt-auto">
              <Link href="/on-premise" className={`${buttonSecondary} w-full`}>
                درخواست استقرار اختصاصی
              </Link>
            </div>
          </article>
        </div>
      </div>
    </section>
  );
}

function Seats() {
  const [picked, setPicked] = useState<(typeof quiz)[number] | undefined>();
  const highlight = (id: SeatId) => (picked?.seat === id ? 'border-site-ink ring-2 ring-site-ink' : 'border-site-line');

  return (
    <Section id="seats" tone="panel" layout="flex flex-col gap-8">
      <div className="flex flex-col gap-4">
        <h2 className={`max-w-3xl ${sectionHeading}`}>هر عضو، صندلی خودش.</h2>
        <p className={`m-0 max-w-2xl ${sectionCopy}`}>
          امکانات برای همه یکی است؛ فقط اعتبار ماهانه فرق می‌کند. مدیر سازمان هر وقت لازم شد صندلی هر عضو را عوض می‌کند.
        </p>
      </div>

      <div className="flex flex-col gap-3">
        <p className="m-0 text-label font-medium text-site-ink">این همکار بیشتر با گدار چه کار می‌کند؟</p>
        <div className="flex flex-wrap gap-2">
          {quiz.map((option) => (
            <button key={option.id} type="button" aria-pressed={picked?.id === option.id} onClick={() => setPicked(option)} className={segmentButton(picked?.id === option.id)}>
              {option.label}
            </button>
          ))}
        </div>
        <p aria-live="polite" className="m-0 min-h-7 text-body text-site-ink">
          {picked ? picked.text : 'یکی را انتخاب کنید تا صندلی مناسبش را نشان بدهیم.'}
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <article className={`flex flex-col gap-5 rounded-card border bg-site-surface p-6 lg:p-8 ${highlight('free')}`}>
          <div className="flex flex-col gap-1">
            <h3 className="m-0 text-h3 font-semibold text-site-ink">رایگان</h3>
            <p className="m-0 text-label text-site-ink-soft">برای شروع و آشنایی اعضای تیم</p>
          </div>
          <div className="flex flex-col gap-1 border-y border-site-line py-4">
            <span className="text-h3 font-extrabold text-site-ink">۰ تومان</span>
            <span className="text-label text-site-ink-soft">{fa(freeCredits)} اعتبار، یک بار برای هر عضو</span>
          </div>
          <ul className="m-0 flex list-none flex-col gap-3 p-0">
            <Bullet>تا ۵ عضو در هر سازمان</Bullet>
            <Bullet>همه‌ی امکانات پلتفرم</Bullet>
            <Bullet>حدود {fa(monthlyCount(freeCredits, 'msg', 'light'))} پیام با مدل سبک</Bullet>
          </ul>
          <Link href="/onboarding" className={`${buttonSecondary} mt-auto`}>
            شروع رایگان
          </Link>
        </article>

        {paidSeats.map((seat) => {
          return (
            <article key={seat.id} className={`flex flex-col gap-5 rounded-card border bg-site-surface p-6 lg:p-8 ${highlight(seat.id)}`}>
              <div className="flex items-start justify-between gap-3">
                <div className="flex flex-col gap-1">
                  <h3 className="m-0 text-h3 font-semibold text-site-ink">{seat.name}</h3>
                  <p className="m-0 text-label text-site-ink-soft">{seat.tagline}</p>
                </div>
                {seat.popular ? (
                  <span className="shrink-0 rounded-full bg-site-accent px-3 py-1 text-caption font-medium text-site-on-accent">پرکاربردترین</span>
                ) : null}
              </div>
              <div className="flex flex-col gap-1 border-y border-site-line py-4">
                <span className="text-h3 font-extrabold text-site-ink">{fa(seat.price)} تومان در ماه</span>
                <span className="text-label text-site-ink-soft">{fa(seat.credits)} اعتبار در ماه</span>
              </div>
              <div className="flex flex-col gap-2">
                <p className="m-0 text-label text-site-ink-soft">در یک ماه، با مدل سبک، حدود:</p>
                <ul className="m-0 flex list-none flex-col gap-1.5 p-0">
                  {cardTasks.map((item) => (
                    <li key={item.task} className="text-body text-site-ink">
                      <span className="font-semibold">{fa(monthlyCount(seat.credits, item.task, 'light'))}</span> {item.label}
                    </li>
                  ))}
                </ul>
              </div>
              <div className="flex flex-col gap-2">
                <p className="m-0 text-caption font-medium text-site-ink-soft">مناسب برای</p>
                <ul className="m-0 flex list-none flex-wrap gap-2 p-0">
                  {seat.roles.map((role) => (
                    <li key={role} className="rounded-full border border-site-line px-3 py-1 text-caption text-site-ink">
                      {role}
                    </li>
                  ))}
                </ul>
              </div>
              <p className="m-0 mt-auto text-label text-site-ink-soft">برای کسی که {seat.sign}</p>
            </article>
          );
        })}
      </div>

      <p className="m-0 text-label text-site-ink-soft">قیمت‌ها بدون مالیات بر ارزش افزوده است. صورتحساب سالانه تخفیف دارد.</p>
    </Section>
  );
}

function Examples() {
  const [model, setModel] = useState<Model>('light');

  return (
    <Section id="examples" layout="flex flex-col gap-8">
      <div className="flex flex-col gap-4">
        <h2 className={`max-w-3xl ${sectionHeading}`}>هر کار را چند بار در ماه می‌شود انجام داد؟</h2>
        <p className={`m-0 max-w-2xl ${sectionCopy}`}>
          تعداد تقریبی هر کار در یک ماه، با اعتبار هر صندلی. عددها برای انجام فقط همان یک کار است؛ اگر چند کار انجام دهید، اعتبار بینشان تقسیم می‌شود.
        </p>
      </div>

      <div className="flex flex-col gap-3">
        <p id="model-label" className="m-0 text-label font-medium text-site-ink">با کدام مدل کار می‌کنید؟</p>
        <div role="group" aria-labelledby="model-label" className="flex flex-wrap gap-2">
          {(Object.keys(modelNames) as Model[]).map((key) => (
            <button key={key} type="button" aria-pressed={model === key} onClick={() => setModel(key)} className={segmentButton(model === key)}>
              {modelNames[key]}
            </button>
          ))}
        </div>
      </div>

      <ComparisonTable
        caption={`تعداد تقریبی تکرار هر کار در ماه با مدل ${modelNames[model]}`}
        columns={paidSeats.map((seat) => seat.name)}
        rows={exampleTasks.map((item) => ({
          label: item.label,
          note: item.category,
          values: paidSeats.map((seat) => `حدود ${fa(monthlyCount(seat.credits, item.task, model))} بار`),
        }))}
      />

      <p className="m-0 max-w-3xl text-label text-site-ink-soft">
        مدل سبک، سریع و کم‌مصرف است؛ مدل میان‌رده برای بیشتر کارها کافی است؛ و مدل سنگین، دقیق‌تر ولی پرمصرف‌تر.
      </p>
    </Section>
  );
}

function CellValue({ value }: { value: Cell }) {
  if (value === 'yes') {
    return (
      <>
        <Check aria-hidden="true" className="size-5 text-site-ink" />
        <span className="sr-only">دارد</span>
      </>
    );
  }

  if (value === 'no') {
    return (
      <>
        <Minus aria-hidden="true" className="size-5 text-site-ink-soft" />
        <span className="sr-only">ندارد</span>
      </>
    );
  }

  return <span className="text-label text-site-ink">{value}</span>;
}

function Compare() {
  const [query, setQuery] = useState('');
  const term = query.trim().toLowerCase();
  const groups = comparison
    .map((group) => ({ ...group, rows: group.rows.filter((row) => !term || `${row.label} ${row.note ?? ''}`.toLowerCase().includes(term)) }))
    .filter((group) => group.rows.length > 0);

  return (
    <Section id="compare" tone="panel" layout="flex flex-col gap-8">
      <div className="flex flex-col gap-4">
        <h2 className={`max-w-3xl ${sectionHeading}`}>ابر گدارAI یا استقرار اختصاصی؟ ردیف به ردیف.</h2>
        <label className="relative flex max-w-md items-center">
          <span className="sr-only">جست‌وجوی قابلیت</span>
          <Search aria-hidden="true" className="pointer-events-none absolute start-4 size-4 text-site-ink-soft" />
          <input
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="جست‌وجوی قابلیت، مثلاً SSO"
            className="w-full rounded-field border border-site-line-strong bg-site-surface py-3 pe-4 ps-11 text-body text-site-ink placeholder:text-site-ink-soft focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus"
          />
        </label>
      </div>

      <div className="overflow-hidden rounded-card border border-site-line bg-site-surface">
        <table className="w-full border-collapse text-start">
          <caption className="sr-only">مقایسه‌ی ابر گدارAI و استقرار اختصاصی</caption>
          <thead>
            <tr className="bg-site-panel">
              <th scope="col" className="px-4 py-4 text-start text-label font-medium text-site-ink-soft sm:px-6">قابلیت</th>
              <th scope="col" className="w-24 px-2 py-4 text-center text-label font-semibold text-site-ink sm:w-40">ابر گدارAI</th>
              <th scope="col" className="w-24 px-2 py-4 text-center text-label font-semibold text-site-ink sm:w-40">اختصاصی</th>
            </tr>
          </thead>
          {groups.length ? (
            groups.map((group) => (
              <tbody key={group.title}>
                <tr className="border-t border-site-line bg-site-ground">
                  <th scope="rowgroup" colSpan={3} className="px-4 py-3 text-start text-caption font-semibold text-site-ink-soft sm:px-6">
                    {group.title}
                  </th>
                </tr>
                {group.rows.map((row) => (
                  <tr key={row.label} className="border-t border-site-line">
                    <th scope="row" className="px-4 py-4 text-start align-top font-normal sm:px-6">
                      <span className="block text-body text-site-ink">{row.label}</span>
                      {row.note ? <span className="block text-caption text-site-ink-soft">{row.note}</span> : null}
                    </th>
                    <td className="px-2 py-4 align-top">
                      <span className="flex justify-center text-center">
                        <CellValue value={row.cloud} />
                      </span>
                    </td>
                    <td className="px-2 py-4 align-top">
                      <span className="flex justify-center text-center">
                        <CellValue value={row.dedicated} />
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            ))
          ) : (
            <tbody>
              <tr className="border-t border-site-line">
                <td colSpan={3} className="px-6 py-6 text-center text-body text-site-ink-soft">
                  قابلیتی با این عبارت پیدا نشد.
                </td>
              </tr>
            </tbody>
          )}
        </table>
      </div>
    </Section>
  );
}

export function SeatPricing() {
  return (
    <main className="bg-site-ground">
      <Plans />
      <Seats />
      <Examples />
      <Compare />
      <FAQ id="pricing-faq" title="پیش از خرید صندلی بدانید." items={faqs} />
      <ClosingBand
        title="با صندلی‌های رایگان شروع کنید."
        titleMuted="بعد تصمیم بگیرید."
        copy="فضای کاری سازمان را بسازید، تا ۵ عضو را رایگان دعوت کنید و هر وقت لازم شد Pro، Max یا بسته‌ی اعتبار بخرید."
        primary={{ href: '/onboarding', label: 'رایگان شروع کنید' }}
        secondary={{ href: '/pricing/changelog', label: 'تغییرات قیمت' }}
      />
    </main>
  );
}
