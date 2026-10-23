'use client';

import { createContext, useContext, useEffect, useId, useState, type ReactNode } from 'react';
import { createPortal } from 'react-dom';
import Link from 'next/link';
import { ArrowLeft, Check, Info, Minus } from 'lucide-react';
import { ClosingBand, FAQ } from '@/app/components/patterns';
import { buttonOnNavy, buttonPrimary, buttonSecondary, sectionCopy, sectionHeading, textLink } from '@/app/components/sectionType';

export type PublicPlanPrice = {
  billing_interval: 'monthly' | 'quarterly' | 'yearly';
  currency: string;
  amount_minor?: number | null;
  discount_percent: number;
};

export type PublicCommercialPlan = {
  slug: string;
  display_name: string;
  description?: string;
  prices?: PublicPlanPrice[];
  features?: Record<string, unknown>;
};

type BillingInterval = PublicPlanPrice['billing_interval'];
type ComparisonValue = boolean | string;
type TooltipState = {
  id: string;
  label: string;
  description: string;
  left: number;
  top: number;
  placement: 'top' | 'bottom';
  pinned: boolean;
};

type FeatureTooltipContextValue = {
  activeId?: string;
  show: (id: string, label: string, description: string, trigger: HTMLElement) => void;
  hide: (id: string) => void;
  togglePinned: (id: string, label: string, description: string, trigger: HTMLElement) => void;
};

const FeatureTooltipContext = createContext<FeatureTooltipContextValue | null>(null);

const featureExplanations: Record<string, string> = {
  'کش معنایی (Semantic Caching)': 'پاسخ درخواست‌هایی که از نظر معنایی شبیه‌اند ذخیره و دوباره استفاده می‌شود تا زمان پاسخ و هزینه مصرف مدل کاهش پیدا کند.',
  'قواعد مدیریت هزینه و نرخ درخواست (Budget & Rate Limits)': 'برای هزینه مصرفی و تعداد درخواست‌ها در یک بازه زمانی سقف تعیین می‌کنید تا مصرف سرویس قابل‌کنترل بماند.',
  'خروجی OpenTelemetry': 'داده‌های پایش را با استاندارد OpenTelemetry به ابزارهای مانیتورینگ و مشاهده‌پذیری سازمان خود ارسال می‌کنید.',
  'استقرار در محل یا VPC': 'گدارAI در زیرساخت اختصاصی یا ابر مجازی خصوصی سازمان شما اجرا می‌شود تا کنترل بیشتری روی شبکه و داده‌ها داشته باشید.',
  'API یکسان برای بیش از ۱۲۰۰ مدل (Unified API)': 'با یک API ثابت به مدل‌های مختلف دسترسی دارید و می‌توانید بدون تغییر گسترده کد، مدل یا سرویس‌دهنده را عوض کنید.',
  'مدل مجازی (Virtual Model)': 'یک نام ثابت برای یک یا چند مدل واقعی تعریف می‌کنید تا مسیریابی، جابه‌جایی و سیاست‌های دسترسی بدون تغییر کد مصرف‌کننده مدیریت شوند.',
  'محیط تست زنده (Live Playground)': 'پرامپت‌ها، مدل‌ها و تنظیمات را پیش از اتصال اپلیکیشن، مستقیماً در محیط وب آزمایش و مقایسه می‌کنید.',
  'کش درخواست‌های مشابه (Simple Caching)': 'پاسخ درخواست‌های کاملاً یکسان برای مدت مشخص ذخیره می‌شود تا درخواست تکراری دوباره به مدل ارسال نشود.',
  'کش درخواست‌های هم‌معنی (Semantic Caching)': 'درخواست‌های دارای مفهوم مشابه، حتی با متن متفاوت، می‌توانند از پاسخ ذخیره‌شده استفاده کنند.',
  'توزیع بار وزنی بین مدل‌ها (Weighted Load Balancing)': 'درصدی از ترافیک را بر اساس وزن دلخواه بین چند مدل یا سرویس‌دهنده تقسیم می‌کنید.',
  'مسیریابی بر مبنای سرعت (Latency-based Routing)': 'هر درخواست به مسیری هدایت می‌شود که در آن لحظه زمان پاسخ بهتری دارد.',
  'مسیر جایگزین خودکار هنگام خطا (Fallback)': 'اگر مدل یا سرویس‌دهنده اصلی پاسخ ندهد، درخواست به‌صورت خودکار به مسیر جایگزین ارسال می‌شود.',
  'فراداده سفارشی (Custom Metadata)': 'اطلاعات دلخواه مانند شناسه مشتری، پروژه یا محیط اجرا را به درخواست‌ها اضافه می‌کنید تا گزارش‌گیری و جست‌وجو دقیق‌تر شود.',
  'ردیابی درخواست (Tracing)': 'مسیر کامل پردازش هر درخواست، از ورود تا پاسخ مدل، ثبت می‌شود تا خطا و کندی را دقیق‌تر بررسی کنید.',
  'خروجی سازگار با OpenTelemetry': 'لاگ‌ها و ردگیری‌ها را با استاندارد OpenTelemetry به ابزار مانیتورینگ مورد استفاده سازمان می‌فرستید.',
  'نسخه‌بندی پرامپت و متغیرها (Prompt Versioning)': 'تغییرات پرامپت را نسخه‌بندی می‌کنید و متغیرهای آن را بدون تغییر مستقیم در کد مدیریت می‌کنید.',
  'اتصال سرویس حفاظت هوشمند ثالث (Third-party Guardrails)': 'سرویس‌های تخصصی کنترل محتوا و ایمنی را به مسیر پردازش درخواست و پاسخ مدل متصل می‌کنید.',
  'وب‌هوک حفاظت هوشمند اختصاصی (Guardrail Webhook)': 'منطق اختصاصی کنترل محتوا را در سرویس خود پیاده می‌کنید و گدارAI آن را هنگام پردازش فراخوانی می‌کند.',
  'کلید دسترسی شخصی (Private Access Token)': 'این کلید به یک کاربر تعلق دارد و برای کارهای شخصی، ابزارهای توسعه و دسترسی برنامه‌نویسی همان کاربر استفاده می‌شود.',
  'کلید دسترسی مجازی (Virtual Access Token)': 'از این کلید برای اتصال اپلیکیشن‌ها و سرویس‌های نرم‌افزاری به گدارAI می‌توانید استفاده کنید، بدون اینکه کلید اصلی کاربر را در اختیار آن‌ها قرار دهید.',
  'کنترل دسترسی مبتنی بر نقش (RBAC)': 'سطح دسترسی افراد را با نقش‌هایی مانند مدیر، توسعه‌دهنده یا مشاهده‌گر کنترل می‌کنید.',
  'ورود یکپارچه (SSO)': 'کاربران با حساب سازمانی و سرویس هویت مرکزی مجموعه وارد می‌شوند و نیازی به حساب جداگانه ندارند.',
  'ثبت وقایع مدیریتی (Audit Log)': 'تغییرات حساس و فعالیت‌های مدیریتی ثبت می‌شوند تا امکان بررسی امنیتی و حسابرسی وجود داشته باشد.',
  'استقرار ابری (SaaS)': 'سرویس روی زیرساخت مدیریت‌شده گدارAI ارائه می‌شود و نیازی به نصب و نگهداری توسط تیم شما ندارد.',
  'استقرار در محل یا ابر خصوصی (On-premise / VPC)': 'سرویس در زیرساخت خود سازمان یا شبکه ابری خصوصی آن مستقر می‌شود.',
  'توافقنامه سطح خدمات (SLA)': 'تعهد رسمی درباره سطح دسترس‌پذیری، کیفیت پشتیبانی و زمان پاسخ‌گویی سرویس است.',
};

const planDefinitions = [
  {
    slug: 'developer',
    name: 'توسعه‌دهنده',
    description: 'برای شروع، نمونه‌سازی و پروژه‌های کوچک',
    features: ['۳ کاربر', 'دسترسی یکپارچه به بیش از ۱۲۰۰ مدل (Unified API)', 'ثبت لاگ تا ۳ روز', 'ردیابی درخواست تا ۳۰ روز', 'تا ۳ قالب پرامپت'],
  },
  {
    slug: 'team',
    name: 'تیم',
    description: 'برای تیم‌هایی که محصول را به محیط عملیاتی می‌برند',
    features: ['۱۰ کاربر', 'کش معنایی (Semantic Caching)', 'قواعد مدیریت هزینه و نرخ درخواست (Budget & Rate Limits)', 'ثبت لاگ تا ۳۰ روز', 'ردیابی درخواست تا ۹۰ روز'],
  },
  {
    slug: 'growth',
    name: 'رشد',
    description: 'برای کسب‌وکارهای در حال رشد با نیازهای پیشرفته',
    features: ['۳۰ کاربر', 'هشدارهای نظارتی و مدیریتی', 'خروجی OpenTelemetry', 'ورود یکپارچه (SSO)', 'وب‌هوک حفاظت هوشمند اختصاصی (Guardrail Webhook)'],
  },
  {
    slug: 'enterprise',
    name: 'سازمانی',
    description: 'برای مقیاس بالا، استقرار اختصاصی و الزامات سازمانی',
    features: ['تعداد کاربر سفارشی', 'استقرار در محل یا VPC', 'نقش‌ها و نگهداری داده سفارشی', 'مدیریت واحدهای سازمان', 'آنبوردینگ اختصاصی و پشتیبانی با اولویت بالا'],
  },
] as const;

const comparisonSections: Array<{
  title: string;
  rows: Array<{ label: string; values: [ComparisonValue, ComparisonValue, ComparisonValue, ComparisonValue] }>;
}> = [
  {
    title: 'محدودیت‌های اصلی',
    rows: [
      { label: 'تعداد کاربران', values: ['۳', '۱۰', '۳۰', 'سفارشی'] },
    ],
  },
  {
    title: 'درگاه هوش مصنوعی',
    rows: [
      { label: 'API یکسان برای بیش از ۱۲۰۰ مدل (Unified API)', values: [true, true, true, true] },
      { label: 'مدل مجازی (Virtual Model)', values: [true, true, true, true] },
      { label: 'محیط تست زنده (Live Playground)', values: [true, true, true, true] },
      { label: 'کش درخواست‌های مشابه (Simple Caching)', values: ['TTL یک‌روزه', true, true, true] },
      { label: 'کش درخواست‌های هم‌معنی (Semantic Caching)', values: [false, true, true, true] },
    ],
  },
  {
    title: 'کنترل و نظارت',
    rows: [
      { label: 'توزیع بار وزنی بین مدل‌ها (Weighted Load Balancing)', values: [false, true, true, true] },
      { label: 'مسیریابی بر مبنای سرعت (Latency-based Routing)', values: [false, true, true, true] },
      { label: 'مسیریابی بر مبنای اولویت', values: [true, true, true, true] },
      { label: 'مسیر جایگزین خودکار هنگام خطا (Fallback)', values: [true, true, true, 'قابلیت‌های پیشرفته'] },
      { label: 'قواعد مدیریت هزینه ریالی و دلاری', values: [false, true, true, true] },
      { label: 'قواعد نرخ درخواست', values: [false, true, true, true] },
    ],
  },
  {
    title: 'مشاهده‌پذیری',
    rows: [
      { label: 'دوره نگهداری لاگ', values: ['۳ روز', '۳۰ روز', '۳۰ روز', 'سفارشی'] },
      { label: 'ردیابی درخواست (Tracing)', values: ['۳۰ روز', '۹۰ روز', '۹۰ روز', 'خروجی S3 سازگار'] },
      { label: 'فراداده سفارشی (Custom Metadata)', values: [false, true, true, true] },
      { label: 'قیمت‌گذاری سفارشی مدل‌ها', values: [false, true, true, true] },
      { label: 'گزارش شاخص‌های عملیاتی و هزینه‌ای', values: [true, true, true, true] },
      { label: 'فیلتر با فراداده سفارشی', values: [false, true, true, true] },
      { label: 'هشدارهای نظارتی و مدیریتی', values: [false, false, true, true] },
      { label: 'خروجی سازگار با OpenTelemetry', values: [false, false, true, true] },
    ],
  },
  {
    title: 'مدیریت پرامپت و حفاظت هوشمند',
    rows: [
      { label: 'تعداد قالب پرامپت', values: ['تا ۳', 'نامحدود', 'نامحدود', 'نامحدود'] },
      { label: 'نسخه‌بندی پرامپت و متغیرها (Prompt Versioning)', values: [true, true, true, true] },
      { label: 'اتصال سرویس حفاظت هوشمند ثالث (Third-party Guardrails)', values: [false, true, true, true] },
      { label: 'وب‌هوک حفاظت هوشمند اختصاصی (Guardrail Webhook)', values: [false, false, true, true] },
    ],
  },
  {
    title: 'امنیت و استقرار',
    rows: [
      { label: 'کلید دسترسی شخصی (Private Access Token)', values: [true, true, true, true] },
      { label: 'کلید دسترسی مجازی (Virtual Access Token)', values: [false, true, true, true] },
      { label: 'کنترل دسترسی مبتنی بر نقش (RBAC)', values: [false, true, true, 'نقش‌های سفارشی'] },
      { label: 'ورود یکپارچه (SSO)', values: [false, false, true, true] },
      { label: 'مدیریت تیم‌ها', values: [false, true, true, true] },
      { label: 'مدیریت واحدهای سازمان', values: [false, false, false, true] },
      { label: 'ثبت وقایع مدیریتی (Audit Log)', values: [false, false, true, true] },
      { label: 'استقرار ابری (SaaS)', values: [true, true, true, true] },
      { label: 'استقرار در محل یا ابر خصوصی (On-premise / VPC)', values: [false, false, false, true] },
    ],
  },
  {
    title: 'پشتیبانی',
    rows: [
      { label: 'نوع پشتیبانی', values: ['چت‌بات', 'محصول عملیاتی', 'محصول عملیاتی', 'اولویت بالا و آنبوردینگ اختصاصی'] },
      { label: 'توافقنامه سطح خدمات (SLA)', values: [false, 'SLA استاندارد', 'SLA سازمانی', 'SLA سازمانی'] },
    ],
  },
];

const pricingFaqs = [
  {
    question: 'گدارAI چیست و هزینه آن شامل چه چیزی می‌شود؟',
    answer: 'گدارAI فضای کاری هوش مصنوعی برای همه‌ی تیم‌ها و درگاه مدل‌های زبانی برای تیم فنی است. هزینه‌ی پلن، امکانات و سقف استفاده‌ی جدول مقایسه را پوشش می‌دهد؛ مصرف مدل‌ها جداگانه و با کردیت از کیف پول تومانی کم می‌شود.',
  },
  {
    question: 'کدام پلن برای تیم ما مناسب‌تر است؟',
    answer: 'برای نمونه‌سازی و پروژه‌های کوچک از پلن توسعه‌دهنده شروع کنید. پلن تیم برای محصول‌های عملیاتی، پلن رشد برای حجم بالاتر و کنترل‌های پیشرفته، و پلن سازمانی برای استقرار اختصاصی، ساختارهای بزرگ و نیازهای سفارشی طراحی شده است.',
  },
  {
    question: 'آیا می‌توانیم قبل از خرید، گدارAI را آزمایش کنیم؟',
    answer: 'بله. با ثبت‌نام، اعتبار هدیه‌ی اولیه می‌گیرید تا اتصال API، مدل‌ها و جریان کاری خود را پیش از تصمیم نهایی ارزیابی کنید. این اعتبار تاریخ انقضا ندارد.',
  },
  {
    question: 'پرداخت ماهانه، سه‌ماهه و سالانه چه تفاوتی دارند؟',
    answer: 'امکانات هر پلن ثابت است و تفاوت اصلی در دوره صورتحساب و تخفیف احتمالی همان دوره است. با انتخاب گزینه‌های بالای کارت‌ها، قیمت به‌روز هر دوره را می‌بینید.',
  },
  {
    question: 'آیا برای تیم‌ها و محیط‌های مختلف می‌توان محدودیت جداگانه تعریف کرد؟',
    answer: 'بله. از پلن تیم به بالا می‌توانید برای کنترل هزینه و نرخ درخواست قانون تعریف کنید. امکانات دقیق مدیریت تیم، نقش‌ها و واحدهای سازمانی بسته به پلن متفاوت است و در جدول مقایسه آمده است.',
  },
  {
    question: 'کلیدهای سرویس‌دهنده‌های مدل چگونه محافظت می‌شوند؟',
    answer: 'اپلیکیشن‌های شما با توکن دسترسی گدارAI کار می‌کنند و لازم نیست کلید اصلی سرویس‌دهنده را در اختیار هر کاربر یا سرویس بگذارید. امکان مدیریت توکن‌های شخصی و مجازی و کنترل دسترسی مبتنی بر نقش، مطابق پلن انتخابی در دسترس است.',
  },
  {
    question: 'آیا امکان استقرار در زیرساخت خودمان وجود دارد؟',
    answer: 'بله. در پلن سازمانی می‌توانید درباره استقرار در محل یا ابر خصوصی (VPC)، نگهداری داده و الزامات شبکه با تیم ما هماهنگ کنید. جزئیات معماری و تعهدات اجرایی متناسب با نیاز سازمان تعیین می‌شود.',
  },
  {
    question: 'آیا بعداً می‌توانیم پلن را تغییر دهیم؟',
    answer: 'بله. تغییر پلن وسط دوره به‌تناسب محاسبه می‌شود.',
  },
  {
    question: 'فاکتور رسمی می‌دهید؟',
    answer: 'بله. برای هر پرداخت فاکتور رسمی صادر می‌شود.',
  },
  {
    question: 'پرداخت چطور است؟',
    answer: 'با کیف پول تومانی و کردیت پیش‌پرداخت، بدون کارت بین‌المللی.',
  },
  {
    question: 'برای پرسش‌های فنی یا انتخاب پلن از کجا کمک بگیریم؟',
    answer: 'از صفحه تماس با ما یا ایمیل support@godarai.ir با تیم گدارAI در ارتباط باشید. برای پلن سازمانی، بررسی نیازهای فنی، ظرفیت، استقرار و آنبوردینگ به‌صورت اختصاصی انجام می‌شود.',
  },
] as const;

type PlanDefinition = (typeof planDefinitions)[number];

export function CommercialPricing({ plans }: { plans: PublicCommercialPlan[] }) {
  const [interval, setInterval] = useState<BillingInterval>('monthly');
  const plansBySlug = new Map(plans.map((plan) => [plan.slug.toLowerCase(), plan]));

  return (
    <FeatureTooltipProvider>
      <main className="bg-site-ground">
        <section className="px-4 pb-10 pt-10 sm:px-6 lg:pb-16 lg:pt-20">
          <div className="mx-auto max-w-content">
            <header className="mx-auto flex max-w-3xl flex-col items-center gap-4 text-center">
              <h1 className="text-balance text-h1 font-extrabold text-site-ink lg:text-display">قیمت روشن. به تومان.</h1>
              <p className="max-w-2xl text-lead font-light text-site-ink-soft">
                اول با اعتبار هدیه امتحان کنید. بعد پلن ماهانه را انتخاب کنید، یا گدارAI را روی سرورهای خودتان نصب کنید.
              </p>
            </header>

            <div className="mx-auto mt-8 flex max-w-3xl flex-col items-center justify-between gap-4 rounded-card border border-site-line bg-site-sand px-6 py-5 sm:flex-row">
              <p className="m-0 text-body font-medium text-site-ink">اعتبار هدیه برای شروع؛ بدون تاریخ انقضا.</p>
              <Link href="/onboarding" className={buttonPrimary}>
                شروع با اعتبار هدیه
              </Link>
            </div>

            <div className="mt-10 flex justify-center">
              <div aria-label="دوره پرداخت" className="inline-grid grid-cols-3 rounded-full border border-site-line bg-site-surface p-1" role="group">
                {(['monthly', 'quarterly', 'yearly'] as BillingInterval[]).map((item) => (
                  <button
                    key={item}
                    type="button"
                    aria-pressed={interval === item}
                    onClick={() => setInterval(item)}
                    className={`min-h-11 rounded-full px-4 text-label font-medium transition-colors duration-150 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus sm:px-6 ${interval === item ? 'bg-site-ink text-site-surface' : 'text-site-ink-soft hover:bg-site-panel'}`}
                  >
                    {intervalLabel(item)}
                  </button>
                ))}
              </div>
            </div>

            <div className="mt-8 grid items-stretch gap-3 sm:gap-6 md:grid-cols-2 xl:grid-cols-4">
              {planDefinitions.map((definition) => (
                <PlanCard key={definition.slug} definition={definition} plan={plansBySlug.get(definition.slug)} interval={interval} />
              ))}
            </div>

            {plans.length === 0 ? (
              <p className="mt-5 text-center text-label text-site-ink-soft">
                فهرست قیمت‌ها الان بارگیری نشد؛ جزئیات پلن‌ها همچنان قابل مقایسه است. برای قیمت با ما تماس بگیرید.
              </p>
            ) : null}
          </div>
        </section>

        <section className="bg-site-panel px-4 py-10 sm:px-6 lg:py-20">
          <div className="mx-auto grid max-w-content gap-3 sm:gap-6 lg:grid-cols-3">
            <div className="flex flex-col gap-4 rounded-card bg-site-night p-6 text-site-night-ink lg:col-span-2 lg:p-10">
              <h2 className="m-0 text-h3 font-bold lg:text-h2">روی سرورهای خودتان.</h2>
              <p className="m-0 max-w-xl text-body text-site-night-soft">
                نصب داخل زیرساخت سازمان (Kubernetes) با قرارداد سالانه و پشتیبانی؛ همان فضای کاری، گاردریل و کنترل هزینه، با ورود سازمانی.
              </p>
              <div>
                <Link href="/on-premise" className={buttonOnNavy}>
                  درخواست استقرار در سازمان
                  <ArrowLeft aria-hidden="true" className="size-4" />
                </Link>
              </div>
            </div>
            <div className="flex flex-col gap-3 rounded-card border border-site-line bg-site-surface p-6 lg:p-8">
              <h2 className="m-0 text-h3 font-semibold text-site-ink">مصرف مدل‌ها</h2>
              <p className="m-0 text-body text-site-ink-soft">
                مصرف مدل با کردیت از کیف پول تومانی کم می‌شود و برای هر تیم و هر کلید سقف می‌گذارید.
              </p>
              <Link href="/features/cost-control" className={textLink}>
                مدیریت هزینه
                <ArrowLeft aria-hidden="true" className="size-4" />
              </Link>
            </div>
          </div>
        </section>

        <section aria-labelledby="compare-plans" className="px-4 py-10 sm:px-6 lg:py-20">
          <div className="mx-auto max-w-content">
            <div className="flex max-w-2xl flex-col gap-4">
              <h2 id="compare-plans" className={sectionHeading}>همه‌ی قابلیت‌ها، کنار هم.</h2>
              <p className={`m-0 ${sectionCopy}`}>تفاوت محدودیت‌ها، امنیت، مشاهده‌پذیری و پشتیبانی هر پلن.</p>
            </div>

            <div className="mt-8 max-w-full overflow-hidden rounded-card border border-site-line bg-site-surface">
              <div className="max-w-full overflow-x-auto" dir="ltr" tabIndex={0}>
                <table className="w-full min-w-230 border-collapse text-label" dir="rtl">
                  <caption className="sr-only">مقایسه‌ی قابلیت‌های پلن‌های توسعه‌دهنده، تیم، رشد و سازمانی</caption>
                  <thead>
                    <tr className="border-b border-site-line">
                      <th className="sticky right-0 bg-site-surface px-5 py-5 text-right font-medium text-site-ink-soft" scope="col">قابلیت</th>
                      {planDefinitions.map((plan) => (
                        <th key={plan.slug} className={`min-w-40 px-4 py-5 text-center text-body font-bold text-site-ink ${plan.slug === 'growth' ? 'bg-site-sand' : ''}`} scope="col">
                          {plan.name}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {comparisonSections.map((section) => (
                      <ComparisonSection key={section.title} section={section} />
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
            <p className="mt-3 text-label text-site-ink-soft md:hidden">برای دیدن همه‌ی ستون‌ها، جدول را به طرفین بکشید.</p>
          </div>
        </section>

        <FAQ title="پیش از انتخاب پلن بدانید." id="pricing-faq" items={pricingFaqs.map((item) => ({ ...item }))} />

        <ClosingBand
          title="هنوز مطمئن نیستید؟"
          titleMuted="با هم انتخاب می‌کنیم."
          copy="نیازهای تیم، حجم درخواست و الزامات استقرار را بگویید تا پلن درست را پیدا کنیم."
          primary={{ href: '/demo', label: 'درخواست دمو' }}
          secondary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        />
      </main>
    </FeatureTooltipProvider>
  );
}

function PlanCard({ definition, plan, interval }: { definition: PlanDefinition; plan?: PublicCommercialPlan; interval: BillingInterval }) {
  const price = plan?.prices?.find((item) => item.billing_interval === interval);
  const isDeveloper = definition.slug === 'developer';
  const isGrowth = definition.slug === 'growth';
  const isEnterprise = definition.slug === 'enterprise';

  return (
    <article className={`relative flex flex-col rounded-card border bg-site-surface p-6 ${isGrowth ? 'border-site-ink' : 'border-site-line'}`}>
      {isGrowth ? (
        <span className="absolute -top-3 right-5 rounded-full bg-site-ink px-3 py-0.5 text-caption font-medium text-site-surface">پیشنهاد ما</span>
      ) : null}
      <div className="flex flex-col gap-2">
        <h2 className="m-0 text-h3 font-bold text-site-ink">{definition.name}</h2>
        <p className="m-0 min-h-14 text-label text-site-ink-soft">{definition.description}</p>
      </div>

      <div className="mt-5 min-h-24 border-b border-site-line pb-5">
        {isDeveloper ? (
          <>
            <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
              <span className="text-h2 font-extrabold text-site-ink">{formatAmount(0)}</span>
              <span className="text-label font-medium text-site-ink-soft">رایگان</span>
            </div>
            <p className="mt-2 text-label text-site-ink-soft">بدون نیاز به پرداخت</p>
          </>
        ) : price?.amount_minor != null ? (
          <>
            <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
              <span className="text-h2 font-extrabold text-site-ink">{formatAmount(price.amount_minor)}</span>
              <span className="text-label font-medium text-site-ink-soft">{currencyLabel(price.currency)} / {intervalUnit(interval)}</span>
            </div>
            {price.discount_percent > 0 ? <p className="mt-2 text-label font-medium text-site-success">{formatAmount(price.discount_percent)}٪ تخفیف</p> : null}
          </>
        ) : (
          <>
            <p className="m-0 text-h3 font-bold text-site-ink">{isEnterprise ? 'قیمت‌گذاری سفارشی' : 'استعلام قیمت'}</p>
            <p className="mt-2 text-label text-site-ink-soft">متناسب با نیاز و دوره‌ی پرداخت</p>
          </>
        )}
      </div>

      <p className="mt-5 text-label font-semibold text-site-ink">امکانات کلیدی</p>
      <ul className="mt-3 flex flex-1 flex-col gap-3 text-label text-site-ink">
        {definition.features.map((feature) => (
          <li key={feature} className="flex items-start gap-2.5">
            <Check aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-site-success" />
            <FeatureLabel label={feature} description={featureExplanations[feature]} />
          </li>
        ))}
      </ul>

      <Link
        href={isEnterprise ? '/demo' : `/onboarding?plan=${encodeURIComponent(definition.slug)}`}
        className={`mt-6 ${isGrowth ? buttonPrimary : buttonSecondary}`}
      >
        {isEnterprise ? 'تماس با فروش' : 'شروع با اعتبار هدیه'}
        <ArrowLeft aria-hidden="true" className="size-4" />
      </Link>
    </article>
  );
}

function ComparisonSection({ section }: { section: (typeof comparisonSections)[number] }) {
  return (
    <>
      <tr>
        <th className="bg-site-panel px-5 py-3 text-right text-label font-bold text-site-ink" colSpan={5} scope="colgroup">{section.title}</th>
      </tr>
      {section.rows.map((row) => (
        <tr key={row.label} className="border-b border-site-line last:border-b-0">
          <th className="sticky right-0 z-10 w-1/4 bg-site-surface px-5 py-4 text-right font-medium text-site-ink" scope="row">
            <FeatureLabel label={row.label} description={featureExplanations[row.label]} />
          </th>
          {row.values.map((value, index) => (
            <td key={`${row.label}-${planDefinitions[index].slug}`} className={`px-4 py-4 text-center ${index === 2 ? 'bg-site-ground' : ''}`}>
              <ComparisonCell value={value} />
            </td>
          ))}
        </tr>
      ))}
    </>
  );
}

function ComparisonCell({ value }: { value: ComparisonValue }) {
  if (value === true) return <><Check aria-hidden="true" className="mx-auto size-5 text-site-success" /><span className="sr-only">دارد</span></>;
  if (value === false) return <><Minus aria-hidden="true" className="mx-auto size-5 text-site-line-strong" /><span className="sr-only">ندارد</span></>;
  return <span className="text-site-ink">{value}</span>;
}

function FeatureLabel({ label, description }: { label: string; description?: string }) {
  if (!description) return <span>{label}</span>;

  const { leading, trailing, trailingDirection } = splitFeatureLabel(label);

  return (
    <span>
      {leading ? <>{leading}{' '}</> : null}
      <span className="inline-flex whitespace-nowrap items-center gap-1.5" dir={trailingDirection}>
        <span>{trailing}</span>
        <FeatureHelp label={label} description={description} />
      </span>
    </span>
  );
}

function splitFeatureLabel(label: string) {
  const parentheticalStart = label.lastIndexOf(' (');
  if (parentheticalStart > 0) {
    return {
      leading: label.slice(0, parentheticalStart),
      trailing: label.slice(parentheticalStart + 1),
      trailingDirection: 'rtl' as const,
    };
  }

  const lastSpace = label.lastIndexOf(' ');
  return {
    leading: lastSpace > 0 ? label.slice(0, lastSpace) : '',
    trailing: lastSpace > 0 ? label.slice(lastSpace + 1) : label,
    trailingDirection: 'rtl' as const,
  };
}

function FeatureHelp({ label, description }: { label: string; description: string }) {
  const id = useId();
  const tooltip = useContext(FeatureTooltipContext);
  const isOpen = tooltip?.activeId === id;

  return (
    <button
      type="button"
      aria-describedby={isOpen ? id : undefined}
      aria-label={`راهنمای ${label}`}
      data-feature-tooltip-trigger="true"
      onBlur={() => tooltip?.hide(id)}
      onClick={(event) => tooltip?.togglePinned(id, label, description, event.currentTarget)}
      onFocus={(event) => tooltip?.show(id, label, description, event.currentTarget)}
      onMouseEnter={(event) => tooltip?.show(id, label, description, event.currentTarget)}
      onMouseLeave={() => tooltip?.hide(id)}
      className="inline-grid size-6 shrink-0 place-items-center rounded-full text-site-ink-soft transition-colors hover:bg-site-panel hover:text-site-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-site-focus"
    >
      <Info aria-hidden="true" className="size-3.5" />
    </button>
  );
}

function FeatureTooltipProvider({ children }: { children: ReactNode }) {
  const [active, setActive] = useState<TooltipState | null>(null);

  useEffect(() => {
    if (!active) return;

    const close = () => setActive(null);
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') close();
    };
    const handlePointerDown = (event: PointerEvent) => {
      if (event.target instanceof Element && event.target.closest('[data-feature-tooltip-trigger]')) return;
      close();
    };

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('resize', close);
    window.addEventListener('scroll', close, true);
    document.addEventListener('pointerdown', handlePointerDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('resize', close);
      window.removeEventListener('scroll', close, true);
      document.removeEventListener('pointerdown', handlePointerDown);
    };
  }, [active]);

  const createState = (id: string, label: string, description: string, trigger: HTMLElement, pinned: boolean): TooltipState => {
    const rect = trigger.getBoundingClientRect();
    const tooltipWidth = Math.min(288, window.innerWidth - 24);
    const left = Math.min(
      Math.max(rect.left + rect.width / 2 - tooltipWidth / 2, 12),
      window.innerWidth - tooltipWidth - 12,
    );
    const placement = rect.bottom + 168 < window.innerHeight ? 'bottom' : 'top';

    return {
      id,
      label,
      description,
      left,
      top: placement === 'bottom' ? rect.bottom + 8 : rect.top - 8,
      placement,
      pinned,
    };
  };

  const value: FeatureTooltipContextValue = {
    activeId: active?.id,
    show: (id, label, description, trigger) => {
      setActive((current) => {
        if (current?.id === id && current.pinned) return current;
        return createState(id, label, description, trigger, false);
      });
    },
    hide: (id) => {
      setActive((current) => current?.id === id && !current.pinned ? null : current);
    },
    togglePinned: (id, label, description, trigger) => {
      setActive((current) => current?.id === id && current.pinned ? null : createState(id, label, description, trigger, true));
    },
  };

  return (
    <FeatureTooltipContext.Provider value={value}>
      {children}
      {active ? createPortal(
        <div
          id={active.id}
          role="tooltip"
          dir="rtl"
          className="pointer-events-none fixed z-50 max-h-svh w-72 overflow-y-auto rounded-card border border-site-line bg-site-surface p-3 text-right text-caption font-normal text-site-ink"
          style={{
            left: active.left,
            top: active.top,
            transform: active.placement === 'top' ? 'translateY(-100%)' : undefined,
          }}
        >
          <span className="mb-1 block font-bold text-site-ink">{active.label}</span>
          {active.description}
        </div>,
        document.body,
      ) : null}
    </FeatureTooltipContext.Provider>
  );
}

function intervalLabel(value: BillingInterval) {
  if (value === 'quarterly') return 'سه‌ماهه';
  if (value === 'yearly') return 'سالانه';
  return 'ماهانه';
}

function intervalUnit(value: BillingInterval) {
  if (value === 'quarterly') return 'سه ماه';
  if (value === 'yearly') return 'سال';
  return 'ماه';
}

function formatAmount(value: number) {
  return new Intl.NumberFormat('fa-IR').format(value);
}

function currencyLabel(currency: string) {
  const labels: Record<string, string> = { IRR: 'ریال', IRT: 'تومان', USD: 'دلار' };
  return labels[currency.toUpperCase()] ?? currency;
}
