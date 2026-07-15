import Link from 'next/link';
import {
  ArrowLeft,
  Braces,
  CheckCircle2,
  ClipboardCheck,
  Code2,
  GitBranch,
  History,
  PlayCircle,
  ShieldCheck,
  SlidersHorizontal,
} from 'lucide-react';

const outcomes = [
  {
    icon: GitBranch,
    title: 'نسخه ثابت برای هر اجرا',
    desc: 'هر تغییر مهم در پرامپت به نسخه جدا تبدیل می‌شود تا تیم بداند کدام نسخه در محصول استفاده شده است.',
  },
  {
    icon: Braces,
    title: 'متغیرهای قابل کنترل',
    desc: 'مقادیر زمان اجرا مثل نام مشتری، شماره سفارش یا متن تیکت را جدا از قالب پرامپت ارسال کنید.',
  },
  {
    icon: ClipboardCheck,
    title: 'Spec قابل بازبینی',
    desc: 'پیام‌ها، مدل، تنظیمات و متادیتای نسخه در یک مشخصات فنی نگهداری می‌شود و برای بازبینی تیمی قابل پیگیری است.',
  },
  {
    icon: ShieldCheck,
    title: 'همراه با گاردریل و پایش',
    desc: 'اجرای نسخه از مسیر درگاه می‌تواند کنار گاردریل، گزارش رخداد، رد درخواست و کنترل هزینه بررسی شود.',
  },
];

const workflow = [
  'قالب پرامپت را با پیام راهنما، قالب گفتگو و متغیرهای ورودی بسازید.',
  'در تست آزمایشی، نمونه داده واقعی اما غیرحساس وارد کنید؛ مثلاً پیگیری سفارش یک فروشگاه.',
  'نسخه را با یادداشت روشن ذخیره کنید تا تغییرات بعدی قابل مقایسه باشد.',
  'در اپلیکیشن، نسخه ذخیره‌شده را با `prompt_version_fqn` و `prompt_variables` اجرا کنید.',
];

const faqs = [
  {
    q: 'رجیستری پرامپت چه فرقی با نگهداری متن پرامپت در کد دارد؟',
    a: 'وقتی پرامپت در رجیستری نسخه‌بندی می‌شود، تغییرات آن بدون جست‌وجو در چند مخزن کد قابل بازبینی، مقایسه و اجرای دوباره است.',
  },
  {
    q: 'آیا می‌توانم یک نسخه مشخص را همیشه در محصول نگه دارم؟',
    a: 'بله. FQN نسخه، یک نسخه مشخص را هدف می‌گیرد. برای انتشار تغییرات، نسخه جدید بسازید و سپس اپلیکیشن را به همان نسخه وصل کنید.',
  },
  {
    q: 'برای داده‌های حساس کاربران چه کار کنم؟',
    a: 'در نمونه‌ها از داده واقعی حساس استفاده نکنید. هنگام اجرا، گاردریل و سیاست‌های گزارش رخداد را متناسب با داده‌هایی مثل شماره تماس، کد ملی یا متن تیکت تنظیم کنید.',
  },
];

export function PromptEngineering() {
  return (
    <main className="overflow-hidden">
      <section className="relative px-4 py-20 sm:px-6 lg:px-8 lg:py-28">
        <div className="absolute inset-x-0 top-0 -z-10 h-72 bg-gradient-to-b from-indigo-500/10 to-transparent" />
        <div className="mx-auto grid max-w-7xl items-center gap-12 lg:grid-cols-[1.05fr_0.95fr]">
          <div>
            <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-indigo-100 bg-indigo-50 px-3 py-1 text-xs font-medium text-indigo-600 dark:border-indigo-500/20 dark:bg-indigo-500/10 dark:text-indigo-300">
              <History className="h-4 w-4" />
              رجیستری پرامپت برای تیم‌های محصول و فنی
            </div>
            <h1 className="mb-6 text-3xl font-bold leading-[1.35] text-slate-900 dark:text-white sm:text-5xl">
              پرامپت‌های محصول را از متن پراکنده به نسخه قابل اجرا تبدیل کنید
            </h1>
            <p className="mb-8 max-w-2xl text-base leading-8 text-slate-600 dark:text-gray-400 sm:text-lg">
              در گدارAI، قالب پرامپت، متغیرها، مدل، Spec و FQN در یک مسیر نگهداری می‌شوند. نتیجه این است که اجرای پرامپت برای پشتیبانی، فروش، عملیات یا تحلیل داده فارسی قابل بازبینی و قابل تکرار می‌ماند.
            </p>
            <div className="flex flex-col gap-3 sm:flex-row">
              <Link
                href="/pricing"
                className="inline-flex min-h-12 items-center justify-center gap-2 rounded-xl bg-indigo-500 px-6 py-3 font-medium text-white shadow-[0_0_20px_rgba(99,102,241,0.25)] transition-colors hover:bg-indigo-600"
              >
                شروع کنید
                <ArrowLeft className="h-5 w-5" />
              </Link>
              <Link
                href="/docs"
                className="inline-flex min-h-12 items-center justify-center gap-2 rounded-xl border border-gray-200 bg-white px-6 py-3 font-medium text-slate-700 transition-colors hover:bg-gray-50 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-200 dark:hover:bg-gray-800"
              >
                مشاهده مستندات
                <Code2 className="h-5 w-5" />
              </Link>
            </div>
          </div>

          <div className="rounded-2xl border border-gray-200 bg-white p-4 shadow-xl dark:border-gray-800 dark:bg-[#0B0F19]">
            <div className="mb-4 flex items-center justify-between border-b border-gray-100 pb-4 dark:border-gray-800">
              <div>
                <p className="text-sm font-bold text-slate-900 dark:text-white">پیگیری سفارش</p>
                <p className="mt-1 text-xs text-slate-500 dark:text-gray-400">نسخه ۴، آماده اجرا</p>
              </div>
              <span className="rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-600 dark:text-emerald-300">
                منتشر شده
              </span>
            </div>
            <div className="space-y-3 text-sm">
              <div className="rounded-xl bg-slate-50 p-4 dark:bg-slate-900/70">
                <p className="mb-2 text-xs font-medium text-slate-500 dark:text-gray-400">پیام راهنما</p>
                <p className="leading-7 text-slate-700 dark:text-gray-200">
                  پاسخ را کوتاه، محترمانه و مناسب پیامک بنویس. اگر وضعیت سفارش نامشخص است، کاربر را به پشتیبانی ارجاع بده.
                </p>
              </div>
              <div className="rounded-xl bg-slate-50 p-4 dark:bg-slate-900/70">
                <p className="mb-2 text-xs font-medium text-slate-500 dark:text-gray-400">متغیرهای ورودی</p>
                <div className="flex flex-wrap gap-2" dir="ltr">
                  {['{{customer_name}}', '{{order_id}}', '{{delivery_status}}'].map((item) => (
                    <code key={item} className="rounded-md border border-gray-200 bg-white px-2 py-1 text-xs text-slate-700 dark:border-gray-700 dark:bg-[#111827] dark:text-gray-200">
                      {item}
                    </code>
                  ))}
                </div>
              </div>
              <div className="rounded-xl border border-indigo-100 bg-indigo-50 p-4 dark:border-indigo-500/20 dark:bg-indigo-500/10">
                <p className="mb-2 text-xs font-medium text-indigo-600 dark:text-indigo-300">FQN نسخه</p>
                <code className="block overflow-x-auto text-xs text-slate-800 dark:text-gray-100" dir="ltr">
                  chat_prompt:shop-support/order-status:v4
                </code>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section className="border-y border-gray-200 bg-slate-50 px-4 py-16 dark:border-gray-800 dark:bg-[#111827]/40 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-7xl">
          <div className="mb-10 max-w-3xl">
            <h2 className="mb-4 text-2xl font-bold text-slate-900 dark:text-white sm:text-4xl">
              چه چیزی در رجیستری پرامپت کنترل می‌شود؟
            </h2>
            <p className="text-sm leading-7 text-slate-600 dark:text-gray-400 sm:text-base">
              هدف این صفحه وعده کیفیت خودکار نیست. رجیستری کمک می‌کند تغییرات پرامپت، اجرای نسخه و اثر آن روی محصول شفاف‌تر شود.
            </p>
          </div>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {outcomes.map((item) => (
              <div key={item.title} className="rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-[#0B0F19]">
                <item.icon className="mb-5 h-7 w-7 text-indigo-500" />
                <h3 className="mb-3 text-lg font-bold text-slate-900 dark:text-white">{item.title}</h3>
                <p className="text-sm leading-7 text-slate-600 dark:text-gray-400">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="px-4 py-16 sm:px-6 lg:px-8">
        <div className="mx-auto grid max-w-7xl gap-10 lg:grid-cols-[0.9fr_1.1fr]">
          <div>
            <h2 className="mb-4 text-2xl font-bold text-slate-900 dark:text-white sm:text-3xl">
              مسیر پیشنهادی انتشار پرامپت
            </h2>
            <p className="text-sm leading-7 text-slate-600 dark:text-gray-400 sm:text-base">
              برای تیم‌هایی که پرامپت پشتیبانی، فروش یا تحلیل متن را در محصول واقعی استفاده می‌کنند، بهتر است هر تغییر قابل پیگیری باشد.
            </p>
          </div>
          <div className="space-y-4">
            {workflow.map((item, index) => (
              <div key={item} className="flex gap-4 rounded-2xl border border-gray-200 p-5 dark:border-gray-800">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-indigo-500 text-sm font-bold text-white">
                  {(index + 1).toLocaleString('fa-IR')}
                </div>
                <p className="text-sm leading-7 text-slate-700 dark:text-gray-300">{item}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="border-t border-gray-200 bg-slate-50 px-4 py-16 dark:border-gray-800 dark:bg-[#111827]/40 sm:px-6 lg:px-8">
        <div className="mx-auto grid max-w-7xl gap-8 lg:grid-cols-3">
          <div className="lg:col-span-1">
            <div className="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-indigo-500/10 text-indigo-500">
              <PlayCircle className="h-6 w-6" />
            </div>
            <h2 className="mb-3 text-2xl font-bold text-slate-900 dark:text-white">پرسش‌های رایج</h2>
            <p className="text-sm leading-7 text-slate-600 dark:text-gray-400">
              چند نکته کوتاه برای تصمیم‌گیری درباره استفاده از رجیستری پرامپت در محصول.
            </p>
          </div>
          <div className="space-y-4 lg:col-span-2">
            {faqs.map((faq) => (
              <div key={faq.q} className="rounded-2xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-[#0B0F19]">
                <div className="mb-3 flex items-start gap-3">
                  <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0 text-emerald-500" />
                  <h3 className="font-bold text-slate-900 dark:text-white">{faq.q}</h3>
                </div>
                <p className="text-sm leading-7 text-slate-600 dark:text-gray-400">{faq.a}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="px-4 py-16 text-center sm:px-6 lg:px-8">
        <SlidersHorizontal className="mx-auto mb-5 h-10 w-10 text-indigo-500" />
        <h2 className="mb-4 text-2xl font-bold text-slate-900 dark:text-white sm:text-3xl">
          پرامپت را مثل بخشی از زیرساخت محصول مدیریت کنید
        </h2>
        <p className="mx-auto mb-8 max-w-2xl text-sm leading-7 text-slate-600 dark:text-gray-400 sm:text-base">
          وقتی پرامپت در تجربه کاربر اثر مستقیم دارد، نسخه، متغیر، مدل، گاردریل و پایش آن باید کنار هم دیده شوند.
        </p>
        <Link
          href="/pricing"
          className="inline-flex min-h-12 items-center justify-center gap-2 rounded-xl bg-indigo-500 px-6 py-3 font-medium text-white transition-colors hover:bg-indigo-600"
        >
          ساخت فضای کاری
          <ArrowLeft className="h-5 w-5" />
        </Link>
      </section>
    </main>
  );
}
