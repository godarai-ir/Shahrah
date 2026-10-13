'use client';

import { createContext, useContext, useEffect, useId, useState, type ReactNode } from 'react';
import { createPortal } from 'react-dom';
import Link from 'next/link';
import { ArrowLeft, Check, ChevronDown, Info, Minus, Sparkles } from 'lucide-react';

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
    features: ['۱۰ هزار درخواست در ماه', '۳ کاربر', 'دسترسی یکپارچه به بیش از ۱۲۰۰ مدل (Unified API)', 'ثبت لاگ تا ۳ روز', 'ردیابی درخواست تا ۳۰ روز', 'تا ۳ قالب پرامپت'],
  },
  {
    slug: 'team',
    name: 'تیم',
    description: 'برای تیم‌هایی که محصول را به محیط عملیاتی می‌برند',
    features: ['۱۰۰ هزار درخواست در ماه', '۱۰ کاربر', 'کش معنایی (Semantic Caching)', 'قواعد مدیریت هزینه و نرخ درخواست (Budget & Rate Limits)', 'ثبت لاگ تا ۳۰ روز', 'ردیابی درخواست تا ۹۰ روز'],
  },
  {
    slug: 'growth',
    name: 'رشد',
    description: 'برای کسب‌وکارهای در حال رشد با نیازهای پیشرفته',
    features: ['۱ میلیون درخواست در ماه', '۳۰ کاربر', 'هشدارهای نظارتی و مدیریتی', 'خروجی OpenTelemetry', 'ورود یکپارچه (SSO)', 'وب‌هوک حفاظت هوشمند اختصاصی (Guardrail Webhook)'],
  },
  {
    slug: 'enterprise',
    name: 'سازمانی',
    description: 'برای مقیاس بالا، استقرار اختصاصی و الزامات سازمانی',
    features: ['بیش از ۱۰ میلیون درخواست در ماه', 'تعداد کاربر سفارشی', 'استقرار در محل یا VPC', 'نقش‌ها و نگهداری داده سفارشی', 'مدیریت واحدهای سازمان', 'آنبوردینگ اختصاصی و پشتیبانی با اولویت بالا'],
  },
] as const;

const comparisonSections: Array<{
  title: string;
  rows: Array<{ label: string; values: [ComparisonValue, ComparisonValue, ComparisonValue, ComparisonValue] }>;
}> = [
  {
    title: 'محدودیت‌های اصلی',
    rows: [
      { label: 'درخواست در ماه', values: ['۱۰ هزار', '۱۰۰ هزار', '۱ میلیون', 'سفارشی؛ بیش از ۱۰ میلیون'] },
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
    answer: 'گدارAI یک درگاه یکپارچه برای اتصال، مدیریت و پایش مدل‌های هوش مصنوعی است. هزینه پلن، امکانات پلتفرم و سقف استفاده درج‌شده در جدول مقایسه را پوشش می‌دهد؛ هزینه مصرف مدل‌ها بر اساس سرویس‌دهنده و میزان استفاده شما جداگانه محاسبه می‌شود.',
  },
  {
    question: 'کدام پلن برای تیم ما مناسب‌تر است؟',
    answer: 'برای نمونه‌سازی و پروژه‌های کوچک از پلن توسعه‌دهنده شروع کنید. پلن تیم برای محصول‌های عملیاتی، پلن رشد برای حجم بالاتر و کنترل‌های پیشرفته، و پلن سازمانی برای استقرار اختصاصی، ساختارهای بزرگ و نیازهای سفارشی طراحی شده است.',
  },
  {
    question: 'آیا می‌توانیم قبل از خرید، گدارAI را آزمایش کنیم؟',
    answer: 'بله. پلن‌های غیرسازمانی با دوره آزمایشی ۱۵روزه ارائه می‌شوند تا اتصال API، مدل‌ها و جریان کاری خود را پیش از تصمیم نهایی ارزیابی کنید.',
  },
  {
    question: 'پرداخت ماهانه، سه‌ماهه و سالانه چه تفاوتی دارند؟',
    answer: 'امکانات هر پلن ثابت است و تفاوت اصلی در دوره صورتحساب و تخفیف احتمالی همان دوره است. با انتخاب گزینه‌های بالای کارت‌ها، قیمت به‌روز هر دوره را می‌بینید.',
  },
  {
    question: 'اگر از سقف درخواست پلن عبور کنیم چه می‌شود؟',
    answer: 'سقف درخواست هر پلن در جدول مشخص است. پیش از رسیدن به سقف می‌توانید مصرف را پایش کنید و برای ارتقای پلن یا تنظیم ظرفیت مناسب با تیم گدارAI هماهنگ شوید؛ شرایط ادامه سرویس و هزینه مازاد بر اساس قرارداد پلن شما تعیین می‌شود.',
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
    answer: 'بله. با تغییر حجم درخواست، تعداد کاربران یا نیازهای امنیتی می‌توانید برای ارتقا یا تغییر پلن اقدام کنید. زمان اعمال تغییر و نحوه محاسبه صورتحساب بر اساس دوره پرداخت فعلی شما مشخص می‌شود.',
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
      <main className="overflow-hidden">
      <section className="relative px-4 pb-16 pt-20 sm:px-6 sm:pt-24 lg:px-8">
        <div className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-[34rem] bg-[radial-gradient(circle_at_50%_0%,rgba(99,102,241,0.16),transparent_60%)]" />
        <div className="mx-auto max-w-7xl">
          <header className="mx-auto flex max-w-3xl flex-col items-center gap-5 text-center">
            <div className="inline-flex items-center gap-2 rounded-full border border-indigo-200 bg-indigo-50 px-3 py-1.5 text-sm font-semibold text-indigo-700 dark:border-indigo-500/30 dark:bg-indigo-500/10 dark:text-indigo-300">
              <Sparkles aria-hidden="true" className="size-4" />
              پلن‌های گدارAI
            </div>
            <h1 className="text-4xl font-bold leading-[1.35] tracking-tight text-slate-950 dark:text-white sm:text-5xl lg:text-6xl">
              در هر نقطه از مسیر رشد، همراهتانیم
            </h1>
            <p className="max-w-2xl text-base leading-8 text-slate-600 dark:text-slate-300 sm:text-lg">
              از ساخت اولین نمونه تا پردازش میلیون‌ها درخواست؛ امکانات هر پلن را شفاف مقایسه کنید و متناسب با نیاز تیمتان انتخاب کنید.
            </p>
          </header>

          <div className="mt-10 flex justify-center">
            <div aria-label="دوره پرداخت" className="inline-grid grid-cols-3 rounded-xl border border-slate-200 bg-white p-1 shadow-sm dark:border-slate-700 dark:bg-slate-900" role="group">
              {(['monthly', 'quarterly', 'yearly'] as BillingInterval[]).map((item) => (
                <button
                  key={item}
                  type="button"
                  aria-pressed={interval === item}
                  onClick={() => setInterval(item)}
                  className={`min-h-11 rounded-lg px-3 text-sm font-semibold transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500 sm:px-5 ${interval === item ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800'}`}
                >
                  {intervalLabel(item)}
                </button>
              ))}
            </div>
          </div>

          <div className="mt-12 grid items-stretch gap-5 md:grid-cols-2 xl:grid-cols-4">
            {planDefinitions.map((definition) => (
              <PlanCard
                key={definition.slug}
                definition={definition}
                plan={plansBySlug.get(definition.slug)}
                interval={interval}
              />
            ))}
          </div>

          {plans.length === 0 ? (
            <p className="mt-5 text-center text-sm leading-6 text-slate-500 dark:text-slate-400">
              قیمت‌های آنلاین موقتاً در دسترس نیستند؛ جزئیات پلن‌ها همچنان قابل مقایسه است.
            </p>
          ) : null}
        </div>
      </section>

      <section aria-labelledby="compare-plans" className="border-y border-slate-200 bg-slate-50/70 px-4 py-16 dark:border-slate-800 dark:bg-slate-950/40 sm:px-6 sm:py-20 lg:px-8">
        <div className="mx-auto max-w-7xl">
          <div className="max-w-2xl">
            <p className="text-sm font-semibold text-indigo-600 dark:text-indigo-400">مقایسه کامل</p>
            <h2 id="compare-plans" className="mt-2 text-3xl font-bold text-slate-950 dark:text-white sm:text-4xl">همه قابلیت‌ها، کنار هم</h2>
            <p className="mt-3 leading-7 text-slate-600 dark:text-slate-300">جدول را بررسی کنید تا تفاوت محدودیت‌ها، امنیت، مشاهده‌پذیری و سطح پشتیبانی هر پلن را ببینید.</p>
          </div>

          <div className="mt-8 max-w-full overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <div className="max-w-full overflow-x-auto" dir="ltr" tabIndex={0}>
              <table className="w-full min-w-[920px] border-collapse text-sm" dir="rtl">
              <caption className="sr-only">مقایسه قابلیت‌های پلن‌های توسعه‌دهنده، تیم، رشد و سازمانی</caption>
              <thead>
                <tr className="border-b border-slate-200 dark:border-slate-700">
                  <th className="sticky right-0 bg-white px-5 py-5 text-right font-semibold text-slate-500 dark:bg-slate-900 dark:text-slate-400" scope="col">قابلیت</th>
                  {planDefinitions.map((plan) => (
                    <th key={plan.slug} className={`min-w-40 px-4 py-5 text-center text-base font-bold ${plan.slug === 'growth' ? 'bg-indigo-50 text-indigo-700 dark:bg-indigo-500/10 dark:text-indigo-300' : 'text-slate-900 dark:text-white'}`} scope="col">
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
          <p className="mt-3 text-sm text-slate-500 dark:text-slate-400 md:hidden">برای مشاهده همه ستون‌ها، جدول را به طرفین بکشید.</p>
        </div>
      </section>

      <section aria-labelledby="pricing-faq" className="relative px-4 py-16 sm:px-6 sm:py-20 lg:px-8">
        <div className="pointer-events-none absolute inset-x-0 top-0 -z-10 mx-auto h-96 max-w-5xl bg-[radial-gradient(circle_at_50%_0%,rgba(99,102,241,0.09),transparent_65%)]" />
        <div className="mx-auto grid max-w-7xl gap-10 lg:grid-cols-[0.72fr_1.28fr] lg:gap-16">
          <div className="lg:sticky lg:top-24 lg:self-start">
            <p className="text-sm font-semibold text-indigo-600 dark:text-indigo-400">پرسش‌های پرتکرار</p>
            <h2 id="pricing-faq" className="mt-2 text-3xl font-bold leading-[1.45] text-slate-950 dark:text-white sm:text-4xl">
              پیش از انتخاب پلن بدانید
            </h2>
            <p className="mt-4 max-w-md leading-8 text-slate-600 dark:text-slate-300">
              پاسخ کوتاه به پرسش‌های رایج درباره هزینه، محدودیت‌ها، امنیت و استقرار گدارAI.
            </p>
            <Link href="/contact" className="mt-6 inline-flex min-h-11 items-center gap-2 text-sm font-semibold text-indigo-600 transition-colors hover:text-indigo-700 focus-visible:rounded-md focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-indigo-500 dark:text-indigo-400 dark:hover:text-indigo-300">
              پاسخ پرسشتان را پیدا نکردید؟
              <ArrowLeft aria-hidden="true" className="size-4" />
            </Link>
          </div>

          <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white px-5 shadow-sm dark:border-slate-800 dark:bg-slate-900 sm:px-7">
            {pricingFaqs.map((item, index) => (
              <details key={item.question} className="group border-b border-slate-200 last:border-b-0 dark:border-slate-800" open={index === 0}>
                <summary className="flex min-h-20 cursor-pointer list-none items-center justify-between gap-5 py-5 text-right font-bold leading-7 text-slate-950 marker:content-none focus-visible:rounded-lg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500 dark:text-white [&::-webkit-details-marker]:hidden">
                  <span>{item.question}</span>
                  <span className="grid size-8 shrink-0 place-items-center rounded-full bg-slate-100 text-slate-500 transition-colors group-open:bg-indigo-50 group-open:text-indigo-600 dark:bg-slate-800 dark:text-slate-400 dark:group-open:bg-indigo-500/15 dark:group-open:text-indigo-300">
                    <ChevronDown aria-hidden="true" className="size-4 transition-transform duration-200 group-open:rotate-180" />
                  </span>
                </summary>
                <p className="max-w-3xl pb-6 pl-12 text-sm leading-8 text-slate-600 dark:text-slate-300 sm:text-base">
                  {item.answer}
                </p>
              </details>
            ))}
          </div>
        </div>
      </section>

      <section className="px-4 py-16 sm:px-6 sm:py-20 lg:px-8">
        <div className="mx-auto flex max-w-5xl flex-col items-center gap-6 rounded-3xl border border-indigo-200 bg-indigo-50 px-6 py-10 text-center dark:border-indigo-500/30 dark:bg-indigo-500/10 sm:px-10">
          <h2 className="text-2xl font-bold text-slate-950 dark:text-white sm:text-3xl">هنوز برای انتخاب پلن مطمئن نیستید؟</h2>
          <p className="max-w-2xl leading-7 text-slate-600 dark:text-slate-300">نیازهای فنی، حجم درخواست و الزامات استقرار تیمتان را با ما در میان بگذارید تا انتخاب دقیق‌تری داشته باشید.</p>
          <Link href="/contact" className="inline-flex min-h-12 items-center justify-center gap-2 rounded-xl bg-indigo-600 px-6 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-indigo-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500">
            مشاوره برای انتخاب پلن
            <ArrowLeft aria-hidden="true" className="size-4" />
          </Link>
          <Link href="/pricing/changelog" className="text-sm font-medium text-indigo-700 hover:underline dark:text-indigo-300">
            تاریخچه‌ی تغییرات قیمت و نرخ‌نامه
          </Link>
        </div>
      </section>
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
    <article className={`relative flex min-h-[34rem] flex-col rounded-2xl border bg-white p-6 shadow-sm transition-transform duration-200 hover:-translate-y-1 dark:bg-slate-900 ${isGrowth ? 'border-indigo-500 ring-2 ring-indigo-500/15' : 'border-slate-200 dark:border-slate-800'}`}>
      {isGrowth ? (
        <span className="absolute -top-3 right-5 rounded-full bg-indigo-600 px-3 py-1 text-xs font-bold text-white shadow-sm">پیشنهاد ما</span>
      ) : null}
      <div className="flex flex-col gap-2">
        <h2 className="text-xl font-bold text-slate-950 dark:text-white">{definition.name}</h2>
        <p className="min-h-14 text-sm leading-7 text-slate-600 dark:text-slate-400">{definition.description}</p>
      </div>

      <div className="mt-5 min-h-24 border-b border-slate-200 pb-5 dark:border-slate-700">
        {isDeveloper ? (
          <>
            <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
              <span className="text-3xl font-black tracking-tight text-slate-950 dark:text-white">{formatAmount(0)}</span>
              <span className="text-sm font-medium text-slate-500 dark:text-slate-400">رایگان</span>
            </div>
            <p className="mt-2 text-sm text-slate-500 dark:text-slate-400">بدون نیاز به پرداخت</p>
          </>
        ) : price?.amount_minor != null ? (
          <>
            <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
              <span className="text-3xl font-black tracking-tight text-slate-950 dark:text-white">{formatAmount(price.amount_minor)}</span>
              <span className="text-sm font-medium text-slate-500 dark:text-slate-400">{currencyLabel(price.currency)} / {intervalUnit(interval)}</span>
            </div>
            {price.discount_percent > 0 ? <p className="mt-2 text-sm font-semibold text-emerald-700 dark:text-emerald-400">{formatAmount(price.discount_percent)}٪ تخفیف</p> : null}
          </>
        ) : (
          <>
            <p className="text-2xl font-bold text-slate-950 dark:text-white">{isEnterprise ? 'قیمت‌گذاری سفارشی' : 'استعلام قیمت'}</p>
            <p className="mt-2 text-sm text-slate-500 dark:text-slate-400">متناسب با نیاز و دوره پرداخت</p>
          </>
        )}
      </div>

      <p className="mt-5 text-sm font-bold text-slate-900 dark:text-white">امکانات کلیدی</p>
      <ul className="mt-4 flex flex-1 flex-col gap-3 text-sm leading-6 text-slate-700 dark:text-slate-300">
        {definition.features.map((feature) => (
          <li key={feature} className="flex items-start gap-2.5">
            <span className="mt-0.5 grid size-5 shrink-0 place-items-center rounded-full bg-indigo-50 text-indigo-600 dark:bg-indigo-500/15 dark:text-indigo-300">
              <Check aria-hidden="true" className="size-3.5" />
            </span>
            <FeatureLabel label={feature} description={featureExplanations[feature]} />
          </li>
        ))}
      </ul>

      <Link href={`/onboarding?plan=${encodeURIComponent(definition.slug)}`} className={`mt-6 inline-flex min-h-12 items-center justify-center gap-2 rounded-xl px-4 text-sm font-semibold transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500 ${isGrowth ? 'bg-indigo-600 text-white hover:bg-indigo-700' : 'bg-slate-100 text-slate-950 hover:bg-slate-200 dark:bg-slate-800 dark:text-white dark:hover:bg-slate-700'}`}>
        {isEnterprise ? 'تماس با فروش' : '۱۵ روز رایگان شروع کنید'}
        <ArrowLeft aria-hidden="true" className="size-4" />
      </Link>
    </article>
  );
}

function ComparisonSection({ section }: { section: (typeof comparisonSections)[number] }) {
  return (
    <>
      <tr>
        <th className="bg-slate-100 px-5 py-3 text-right text-sm font-bold text-slate-900 dark:bg-slate-800 dark:text-white" colSpan={5} scope="colgroup">{section.title}</th>
      </tr>
      {section.rows.map((row) => (
        <tr key={row.label} className="border-b border-slate-100 last:border-b-0 dark:border-slate-800">
          <th className="sticky right-0 z-10 w-[28%] bg-white px-5 py-4 text-right font-medium leading-6 text-slate-700 dark:bg-slate-900 dark:text-slate-300" scope="row">
            <FeatureLabel label={row.label} description={featureExplanations[row.label]} />
          </th>
          {row.values.map((value, index) => (
            <td key={`${row.label}-${planDefinitions[index].slug}`} className={`px-4 py-4 text-center leading-6 ${index === 2 ? 'bg-indigo-50/70 dark:bg-indigo-500/5' : ''}`}>
              <ComparisonCell value={value} />
            </td>
          ))}
        </tr>
      ))}
    </>
  );
}

function ComparisonCell({ value }: { value: ComparisonValue }) {
  if (value === true) return <><Check aria-hidden="true" className="mx-auto size-5 text-emerald-600 dark:text-emerald-400" /><span className="sr-only">دارد</span></>;
  if (value === false) return <><Minus aria-hidden="true" className="mx-auto size-5 text-slate-300 dark:text-slate-600" /><span className="sr-only">ندارد</span></>;
  return <span className="text-slate-700 dark:text-slate-300">{value}</span>;
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
      className="inline-grid size-6 shrink-0 place-items-center rounded-full text-indigo-600 transition-colors hover:bg-indigo-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-500 dark:text-indigo-300 dark:hover:bg-indigo-500/15"
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
          className="pointer-events-none fixed z-50 max-h-[calc(100vh-1.5rem)] w-72 max-w-[calc(100vw-1.5rem)] overflow-y-auto rounded-xl border border-slate-200 bg-white p-3 text-right text-xs font-normal leading-6 text-slate-700 shadow-xl dark:border-slate-700 dark:bg-slate-950 dark:text-slate-200"
          style={{
            left: active.left,
            top: active.top,
            transform: active.placement === 'top' ? 'translateY(-100%)' : undefined,
          }}
        >
          <span className="mb-1 block font-bold text-slate-950 dark:text-white">{active.label}</span>
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
