'use client';

import { useState, type ReactNode } from 'react';
import Link from 'next/link';
import { ArrowDown, Check, Minus, Search } from 'lucide-react';
import { ClosingBand, ComparisonTable, FAQ, Section } from '@/app/components/patterns';
import { buttonPrimary, buttonSecondary, sectionCopy, sectionHeading, textLink } from '@/app/components/sectionType';

// Workplace pricing: a monthly seat per colleague, each with its own monthly credit. All counts are estimates
// derived from the same credit formulas the owner's pricing reference uses.

const fa = (value: number) => value.toLocaleString('fa-IR');

type Model = 'light' | 'mid' | 'heavy';
type Task = 'msg' | 'text' | 'office' | 'image' | 'kb';

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

const giftCredits = 500;
const seats = [
  {
    id: 'regular',
    name: 'کاربر عادی',
    tagline: 'برای کار روزمره',
    credits: 4000,
    price: 2250000,
    popular: true,
    roles: ['مدیر پروژه', 'کارشناس فروش و بازاریابی', 'پشتیبانی', 'منابع انسانی', 'تحلیل‌گر'],
    sign: 'بیشتر کارهای روزانه‌اش را با کمک گدار انجام می‌دهد، ولی کارش پر از فایل حجیم یا مدل‌های سنگین نیست.',
    extra: undefined,
  },
  {
    id: 'pro',
    name: 'کاربر حرفه‌ای',
    tagline: 'برای کسی که کارش را با گدار پیش می‌برد',
    credits: 20000,
    price: 11250000,
    popular: false,
    roles: ['تحلیل‌گر ارشد', 'تولیدکننده‌ی محتوا', 'توسعه‌دهنده', 'پژوهشگر', 'مدیر محصول'],
    sign: 'کارش را عملاً با گدار پیش می‌برد: ساعت‌ها در روز، فایل‌های حجیم یا عامل‌های هوش مصنوعی.',
    extra: 'و اجرای عامل‌های هوش مصنوعی چندمرحله‌ای، گفتگوهای طولانی با مدل‌های سنگین و کار روی فایل‌های حجیم.',
  },
] as const;

const cardTasks: { task: Task; label: string }[] = [
  { task: 'msg', label: 'پیام گفتگو' },
  { task: 'text', label: 'خلاصه‌ی گزارش ۲۰ صفحه‌ای' },
  { task: 'office', label: 'تحلیل اکسل و ساخت گزارش Word' },
  { task: 'image', label: 'استخراج اطلاعات از عکس و فاکتور' },
];

const quiz = [
  { id: 'free', label: 'فقط می‌خواهد گدار را امتحان کند', seat: 'trial', text: 'اعتبار هدیه کافی است؛ ۵۰۰ اعتبار برای امتحان. بعد از آن به کاربر عادی یا حرفه‌ای می‌رود.' },
  { id: 'regular', label: 'چت، ترجمه و خلاصه‌سازی', seat: 'regular', text: 'کاربر عادی مناسب است؛ مصرف روزمره و سبک تا متوسط.' },
  { id: 'daily', label: 'تحلیل فایل و گزارش، هر روز', seat: 'pro', text: 'کاربر حرفه‌ای مناسب است؛ تحلیل روزانه‌ی فایل و گزارش.' },
  { id: 'heavy', label: 'فایل‌های سنگین یا عامل هوش مصنوعی', seat: 'pro', text: 'کاربر حرفه‌ای شروع خوبی است. اگر مصرفش بیشتر شد، ادمین اعتبار اضافه می‌خرد یا با فروش صحبت می‌کنید.' },
] as const;

const exampleTasks: { task: Task; category: string; label: string }[] = [
  { task: 'text', category: 'متن', label: 'خلاصه‌ی یک گزارش ۲۰ صفحه‌ای' },
  { task: 'image', category: 'عکس', label: 'استخراج اطلاعات از ۱۰ عکس یا فاکتور اسکن‌شده' },
  { task: 'office', category: 'Word و Excel', label: 'تحلیل یک فایل اکسل و ساخت گزارش Word از آن' },
  { task: 'kb', category: 'منابع دانشی سازمان', label: 'پرسش از اسناد سازمان و گرفتن پاسخ همراه با منبع' },
];

type Cell = 'yes' | 'no' | string;
const comparison: { title: string; rows: { label: string; note?: string; business: Cell; enterprise: Cell }[] }[] = [
  {
    title: 'ویژگی‌ها',
    rows: [
      { label: 'دسترسی به بیش از ۱۳۰ مدل', note: 'GPT، Claude، Gemini، Grok، Kimi و…', business: 'yes', enterprise: 'yes' },
      { label: 'ورودی و خروجی فایل چندرسانه‌ای', note: 'تصویر، PDF، ساخت و ویرایش فایل Word و Excel، اجرای کد', business: 'yes', enterprise: 'yes' },
      { label: 'عامل هوش مصنوعی اختصاصی، چندمرحله‌ای و خودکار', note: 'زمان‌بندی‌شده یا با رویداد', business: 'yes', enterprise: 'yes' },
      { label: 'تعامل با ابزارهای بیرون سازمان', note: 'MCP', business: 'yes', enterprise: 'yes' },
    ],
  },
  { title: 'داده‌ی سازمان', rows: [{ label: 'تعامل با منابع دانشی', business: 'yes', enterprise: 'yes' }] },
  {
    title: 'مدیریت و امنیت',
    rows: [
      { label: 'نقش‌های ادمین سازمان، ادمین تیم و عضو', business: 'yes', enterprise: 'yes' },
      { label: 'سقف و گزارش مصرف برای هر نفر، تیم، مدل یا ابزار', business: 'yes', enterprise: 'yes' },
      { label: 'ثبت وقایع مدیریتی', note: 'Audit Log', business: 'yes', enterprise: 'yes' },
      { label: 'گاردریل‌های معنایی', business: 'yes', enterprise: 'yes' },
      { label: 'ورود یکپارچه', note: 'SSO', business: 'با درخواست', enterprise: 'yes' },
      { label: 'استقرار در محل یا ابر خصوصی', note: 'On-premise / VPC', business: 'no', enterprise: 'yes' },
      { label: 'استفاده از کلید API خودتان', note: 'BYOK', business: 'no', enterprise: 'yes' },
    ],
  },
  {
    title: 'پشتیبانی',
    rows: [
      { label: 'نوع پشتیبانی', business: 'تیکت', enterprise: 'مدیر مشتری اختصاصی' },
      { label: 'توافق‌نامه‌ی سطح خدمات', note: 'SLA', business: 'no', enterprise: 'yes' },
    ],
  },
  { title: 'توسعه‌دهندگان', rows: [{ label: 'دسترسی به API مدل‌های هوش مصنوعی', business: 'yes', enterprise: 'yes' }] },
];

const faqs = [
  {
    question: 'اعتبار چیست؟',
    answer: 'اعتبار واحد مصرف در گدار است. هر کاربر هر ماه مقداری اعتبار می‌گیرد و هر پیام، فایل یا اجرای عامل هوش مصنوعی مقداری از آن را مصرف می‌کند. مدل قوی‌تر یا متن طولانی‌تر اعتبار بیشتری می‌برد.',
  },
  {
    question: 'مصرف اعتبار چطور حساب می‌شود؟',
    answer: 'اعتبار هر پاسخ از دو بخش می‌آید: پردازش مدل، و ابزارهایی که گدار برای جواب دادن اجرا می‌کند، مثل جستجوی وب. یک پیام معمولی با مدل سبک حدود ۰٫۷۵ اعتبار، با مدل میان‌رده حدود ۴٫۵ و با مدل سنگین حدود ۹ اعتبار می‌برد. اعتبار هر پاسخ زیر همان پاسخ نشان داده می‌شود.',
  },
  {
    question: 'تعداد پیام روی کارت‌ها دقیق است؟',
    answer: 'نه، تقریبی است و فرض می‌کند بخشی از اعتبار صرف ابزارها می‌شود. با مدل سبک تعداد پیام چند برابر می‌شود و با گفتگوی طولانی یا عامل هوش مصنوعی سنگین کمتر.',
  },
  {
    question: 'اعتبار هدیه چطور کار می‌کند؟',
    answer: 'هر حساب یک بار ۵۰۰ اعتبار هدیه می‌گیرد تا وارد پنل شود و گدار را امتحان کند. این اعتبار تاریخ انقضا ندارد و تجدید نمی‌شود. برای ادامه، نوع کاربر عادی یا حرفه‌ای را انتخاب می‌کنید.',
  },
  { question: 'اعتبار مصرف‌نشده به ماه بعد می‌رسد؟', answer: 'نه. اعتبار ماهانه‌ی کاربران در پایان هر ماه صفر می‌شود و منتقل نمی‌شود.' },
  {
    question: 'اگر اعتبار تمام شود چه می‌شود؟',
    answer: 'پاسخی که در حال تولید است کامل می‌شود. بعد از آن، اگر اعتبار کاربر تمام شده باشد، از اعتبار اضافه‌ی مشترک سازمان برداشت می‌شود. اگر آن هم نباشد، استفاده تا خرید اعتبار اضافه یا شروع ماه بعد متوقف می‌شود. هیچ هزینه‌ای بیش از آنچه خریده‌اید ثبت نمی‌شود.',
  },
  {
    question: 'فرق نوع کاربر عادی و حرفه‌ای چیست؟',
    answer: 'قابلیت‌ها یکی است و فقط اعتبار ماهانه فرق می‌کند: ۴٬۰۰۰ و ۲۰٬۰۰۰ اعتبار. مدیر تیم یا ادمین پنل برای هر همکار نوع کاربر جدا انتخاب می‌کند و هر وقت لازم شد آن را تغییر می‌دهد.',
  },
  {
    question: 'اعتبار کاربرها بین هم مشترک است؟',
    answer: 'نه. اعتبار هر کاربر مال خود اوست و به دیگری منتقل نمی‌شود. مشترک بودن فقط برای اعتبار اضافه است که سازمان می‌خرد و ادمین می‌تواند سقف برداشت هر نفر از آن را تعیین کند.',
  },
  {
    question: 'API و اتوماسیون از کجا اعتبار می‌گیرند؟',
    answer: 'درخواست‌هایی که به هیچ همکاری وصل نیستند، مثل API و عامل‌های هوش مصنوعی خودکار، فقط از اعتبار اضافه‌ی مشترک سازمان کم می‌کنند. ادمین می‌تواند برای آن‌ها سقف ماهانه بگذارد.',
  },
  {
    question: 'وقتی کاربر اضافه یا حذف می‌کنم صورتحساب چطور می‌شود؟',
    answer: 'کاربر تازه فوراً فعال می‌شود و هزینه و اعتبارش به نسبت روزهای باقی‌مانده‌ی ماه حساب می‌شود. حذف کاربر از ماه بعد اعمال می‌شود.',
  },
  {
    question: 'کدام مدل‌ها در دسترس‌اند؟',
    answer: 'بیش از ۱۳۰ مدل از ارائه‌دهنده‌های مختلف، با اعتبار یکسان برای همه‌ی انواع کاربر. اعتبار به مدل یا ارائه‌دهنده‌ی خاصی وابسته نیست.',
  },
  {
    question: 'اگر دلار گران شود، قیمت تغییر می‌کند؟',
    answer: 'نرخ مدل‌ها بر حسب اعتبار ثابت است و بودجه‌های شما جابه‌جا نمی‌شود. قیمت تومانی تا نوسان ۱۰٪ دلار تغییر نمی‌کند. هر تغییر قیمت ۳۰ روز قبل اعلام می‌شود و از ماه بعد اعمال می‌شود.',
  },
  {
    question: 'چه زمانی گدار برای سازمان‌ها لازم است؟',
    answer: 'وقتی بیش از ۱۰۰ نفر هستید، یا به استقرار در محل یا ابر خصوصی، ورود یکپارچه (SSO)، پشتیبانی اولویت‌دار با SLA، مدیر مشتری اختصاصی یا کلید API خودتان (BYOK) نیاز دارید. با تیم فروش صحبت کنید تا با هم قیمت را تعیین کنیم.',
    link: { href: '/demo', label: 'تماس با فروش' },
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
          <h1 className="m-0 text-balance text-h1 font-extrabold text-site-ink lg:text-display">به اندازه‌ی کار هر نفر. به تومان.</h1>
          <p className={`m-0 max-w-2xl ${sectionCopy}`}>
            هر همکار با یک اشتراک ماهانه به همه‌ی مدل‌ها و ابزارهای گدار دسترسی دارد. شما فقط نوع کاربر هر نفر را انتخاب می‌کنید.
          </p>
        </header>

        <div className="grid gap-6 lg:grid-cols-2">
          <article className="flex flex-col gap-6 rounded-card border border-site-line-strong bg-site-surface p-6 lg:p-8">
            <div className="flex flex-col gap-2">
              <h2 className="m-0 text-h3 font-semibold text-site-ink">گدار برای کسب‌وکارها</h2>
              <p className="m-0 text-body text-site-ink-soft">برای شرکت‌ها و تیم‌ها تا ۱۰۰ نفر.</p>
            </div>
            <div className="flex flex-col gap-1 border-y border-site-line py-5">
              <span className="text-caption text-site-ink-soft">شروع از</span>
              <span className="text-h2 font-extrabold text-site-ink">{fa(seats[0].price)} تومان</span>
              <span className="text-label text-site-ink-soft">به ازای هر کاربر در ماه</span>
            </div>
            <p className="m-0 text-label text-site-ink">
              می‌خواهید اول امتحان کنید؟ هر حساب یک بار {fa(giftCredits)} اعتبار هدیه می‌گیرد؛ بدون تاریخ انقضا.{' '}
              <a href="#seats" className={textLink}>
                نوع‌های کاربر
                <ArrowDown aria-hidden="true" className="size-4" />
              </a>
            </p>
            <ul className="m-0 flex list-none flex-col gap-3 p-0">
              <Bullet>دسترسی به بیش از ۱۳۰ مدل، از جمله GPT، Claude، Gemini و Grok</Bullet>
              <Bullet>ساخت عامل هوش مصنوعی اختصاصی برای کارهای تکراری تیم، مثل پاسخ به مشتری یا تهیه‌ی گزارش</Bullet>
              <Bullet>تعامل با ابزارهای بیرونی سازمان (MCP)</Bullet>
              <Bullet>گاردریل معنایی برای درخواست‌ها و موضوع‌هایی که سازمان مجاز نمی‌داند</Bullet>
              <Bullet>گزارش مصرف و مدیریت کاربران</Bullet>
            </ul>
            <div className="mt-auto flex flex-col gap-3">
              <Link href="/onboarding" className={buttonPrimary}>
                شروع با اعتبار هدیه
              </Link>
              <p className="m-0 text-caption text-site-ink-soft">قیمت‌ها بدون احتساب مالیات بر ارزش افزوده است.</p>
            </div>
          </article>

          <article className="flex flex-col gap-6 rounded-card border border-site-line bg-site-sand p-6 lg:p-8">
            <div className="flex flex-col gap-2">
              <h2 className="m-0 text-h3 font-semibold text-site-ink">گدار برای سازمان‌ها</h2>
              <p className="m-0 text-body text-site-ink-soft">
                برای سازمان‌هایی که به استقرار اختصاصی، ورود یکپارچه، پشتیبانی با تعهد سطح خدمات و مدیر مشتری اختصاصی نیاز دارند.
              </p>
            </div>
            <div className="flex flex-col gap-1 border-y border-site-line-strong py-5">
              <span className="text-h2 font-extrabold text-site-ink">قیمت سفارشی</span>
              <span className="text-label text-site-ink-soft">بر اساس تعداد کاربر و حجم مصرف</span>
            </div>
            <p className="m-0 text-label font-medium text-site-ink">همه‌ی امکانات گدار برای کسب‌وکارها، به‌علاوه:</p>
            <ul className="m-0 flex list-none flex-col gap-3 p-0">
              <Bullet>ورود یکپارچه (SSO)</Bullet>
              <Bullet>استقرار در محل یا ابر خصوصی</Bullet>
              <Bullet>پشتیبانی اولویت‌دار با تعهد سطح خدمات (SLA)</Bullet>
              <Bullet>مدیر مشتری اختصاصی</Bullet>
              <Bullet>استفاده از کلید API خودتان (BYOK)</Bullet>
            </ul>
            <div className="mt-auto">
              <Link href="/demo" className={`${buttonSecondary} w-full`}>
                تماس با فروش
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
  const highlight = (id: string) => (picked?.seat === id ? 'border-site-ink ring-2 ring-site-ink' : 'border-site-line');

  return (
    <Section id="seats" tone="panel" layout="flex flex-col gap-8">
      <div className="flex flex-col gap-4">
        <h2 className={`max-w-3xl ${sectionHeading}`}>هر همکار، کاربر خودش.</h2>
        <p className={`m-0 max-w-2xl ${sectionCopy}`}>
          هر کاربر یکی از نوع‌های زیر را دارد. مدیر تیم یا ادمین پنل هر وقت لازم شد، نوع کاربر هر نفر را تغییر می‌دهد.
        </p>
      </div>

      <div className="flex flex-col gap-3">
        <p className="m-0 text-label font-medium text-site-ink">این همکار بیشتر چه کار می‌کند؟</p>
        <div className="flex flex-wrap gap-2">
          {quiz.map((option) => (
            <button key={option.id} type="button" aria-pressed={picked?.id === option.id} onClick={() => setPicked(option)} className={segmentButton(picked?.id === option.id)}>
              {option.label}
            </button>
          ))}
        </div>
        <p aria-live="polite" className="m-0 min-h-7 text-body text-site-ink">
          {picked ? picked.text : 'یکی از گزینه‌ها را بزنید تا نوع کاربر مناسب را نشان بدهیم.'}
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {seats.map((seat) => (
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
              <span className="text-label text-site-ink-soft">{fa(seat.credits)} اعتبار هر ماه</span>
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
              {seat.extra ? <p className="m-0 text-label text-site-ink-soft">{seat.extra}</p> : null}
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
            <p className="m-0 mt-auto text-label text-site-ink-soft">نشانه: {seat.sign}</p>
          </article>
        ))}

        <article className={`flex flex-col gap-5 rounded-card border bg-site-surface p-6 lg:p-8 ${highlight('trial')}`}>
          <div className="flex flex-col gap-1">
            <h3 className="m-0 text-h3 font-semibold text-site-ink">اعتبار هدیه</h3>
            <p className="m-0 text-label text-site-ink-soft">برای دمو و آشنایی با گدار</p>
          </div>
          <div className="flex flex-col gap-1 border-y border-site-line py-4">
            <span className="text-h3 font-extrabold text-site-ink">رایگان</span>
            <span className="text-label text-site-ink-soft">{fa(giftCredits)} اعتبار، یک بار برای هر حساب</span>
          </div>
          <p className="m-0 text-body text-site-ink">
            حدود {fa(monthlyCount(giftCredits, 'msg', 'light'))} پیام با مدل سبک. این اعتبار تاریخ انقضا ندارد و تجدید نمی‌شود.
          </p>
          <Link href="/onboarding" className={`${buttonSecondary} mt-auto`}>
            شروع با اعتبار هدیه
          </Link>
        </article>
      </div>
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
          تعداد تقریبی تکرار هر کار در ماه، برای هر نوع کاربر. عددها برای همان یک کار است؛ اگر چند کار را ترکیب کنید، تقسیم می‌شود.
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
        columns={['کاربر عادی', 'کاربر حرفه‌ای']}
        rows={exampleTasks.map((item) => ({
          label: item.label,
          note: item.category,
          values: seats.map((seat) => `حدود ${fa(monthlyCount(seat.credits, item.task, model))} بار`),
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
        <h2 className={`max-w-3xl ${sectionHeading}`}>کسب‌وکارها یا سازمان‌ها؟ ردیف به ردیف.</h2>
        <label className="relative flex max-w-md items-center">
          <span className="sr-only">جستجوی قابلیت</span>
          <Search aria-hidden="true" className="pointer-events-none absolute start-4 size-4 text-site-ink-soft" />
          <input
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="جستجوی قابلیت، مثلاً SSO"
            className="w-full rounded-field border border-site-line-strong bg-site-surface py-3 pe-4 ps-11 text-body text-site-ink placeholder:text-site-ink-soft focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus"
          />
        </label>
      </div>

      <div className="overflow-hidden rounded-card border border-site-line bg-site-surface">
        <table className="w-full border-collapse text-start">
          <caption className="sr-only">مقایسه‌ی گدار برای کسب‌وکارها و سازمان‌ها</caption>
          <thead>
            <tr className="bg-site-panel">
              <th scope="col" className="px-4 py-4 text-start text-label font-medium text-site-ink-soft sm:px-6">قابلیت</th>
              <th scope="col" className="w-24 px-2 py-4 text-center text-label font-semibold text-site-ink sm:w-40">کسب‌وکارها</th>
              <th scope="col" className="w-24 px-2 py-4 text-center text-label font-semibold text-site-ink sm:w-40">سازمان‌ها</th>
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
                        <CellValue value={row.business} />
                      </span>
                    </td>
                    <td className="px-2 py-4 align-top">
                      <span className="flex justify-center text-center">
                        <CellValue value={row.enterprise} />
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

export function CommercialPricing() {
  return (
    <main className="bg-site-ground">
      <Plans />
      <Seats />
      <Examples />
      <Compare />
      <FAQ id="pricing-faq" title="پیش از انتخاب بدانید." items={faqs} />
      <ClosingBand
        title="هوش مصنوعی را در کار روزمره‌ی همه‌ی تیم‌ها جا بیندازید."
        copy="با چند کاربر شروع کنید، مصرف را ببینید و بر اساس کار واقعی تیم، نوع کاربر هر نفر را تنظیم کنید. اگر کمک خواستید، یک جلسه‌ی ۳۰ دقیقه‌ای با ما بگذارید."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'رزرو دمو' }}
      />
    </main>
  );
}
