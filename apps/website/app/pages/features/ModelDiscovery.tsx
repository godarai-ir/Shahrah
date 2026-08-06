'use client';

import { motion } from 'motion/react';
import Image from 'next/image';
import Link from 'next/link';
import {
  ArrowLeft,
  BarChart3,
  BookOpen,
  Check,
  Columns3,
  Database,
  Filter,
  Gauge,
  GitCompare,
  Globe2,
  LineChart,
  ListChecks,
  Search,
  ShieldCheck,
  Sparkles,
  Workflow,
} from 'lucide-react';
import { MagicCard } from '../../components/ui/MagicCard';
import { FAQAccordion } from '../../components/ui/FAQAccordion';

const fadeIn = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.5 },
};

const docsLinks = [
  {
    title: 'مفاهیم مدل و ارائه‌دهنده',
    description: 'با تعریف ارائه‌دهنده، مدل، قابلیت‌ها و داده‌های کاتالوگ آشنا شوید.',
    href: '/docs/provider-models-concepts',
  },
  {
    title: 'مرور کاتالوگ مدل‌ها',
    description: 'جزئیات کاتالوگ، فیلترها و روش پیدا کردن مدل مناسب برای هر سناریو.',
    href: '/docs/provider-models-overview',
  },
  {
    title: 'مقایسه مدل‌ها',
    description: 'راهنمای مقایسه مدل‌ها بر اساس هزینه، ورودی/خروجی، قابلیت و محدودیت‌ها.',
    href: '/docs/provider-models-compare',
  },
  {
    title: 'مدل مجازی در درگاه',
    description: 'مدل انتخاب‌شده را پشت یک شناسه پایدار در محیط عملیاتی قرار دهید.',
    href: '/docs/virtual-model',
  },
];

const providerGroups = [
  {
    name: 'AI21',
    description: 'مدل‌های زبانی سازمانی برای تولید متن، خلاصه‌سازی، پاسخ‌گویی و کاربردهای تجاری.',
  },
  {
    name: 'Anthropic',
    description: 'خانواده Claude برای استدلال، ایمنی و وظایف پیچیده در تیم‌های محصول و فنی.',
  },
  {
    name: 'AWS Bedrock و Bedrock Mantle',
    description: 'دسترسی ابری و ارکستراسیون مدل‌ها در بستر AWS برای استقرار سازمانی.',
  },
  {
    name: 'Azure AI Foundry',
    description: 'پلتفرم مایکروسافت برای توسعه، ارزیابی و استقرار راهکارهای هوش مصنوعی در مقیاس سازمانی.',
  },
  {
    name: 'Azure OpenAI',
    description: 'دسترسی سازمانی به مدل‌های OpenAI با امنیت، انطباق و یکپارچگی با Azure.',
  },
  {
    name: 'Cerebras',
    description: 'زیرساخت و مدل‌های هوش مصنوعی با تمرکز بر سرعت استنتاج و پردازش مدل‌های بزرگ.',
  },
  {
    name: 'Cohere',
    description: 'مدل‌های زبانی سازمانی برای تولید متن، جست‌وجوی معنایی، امبدینگ و کاربردهای سازمانی.',
  },
  {
    name: 'Databricks',
    description: 'پلتفرم داده و هوش مصنوعی برای ساخت، آموزش و ارائه مدل‌های زبانی کنار داده‌های سازمان.',
  },
  {
    name: 'Deepgram',
    description: 'مدل‌های گفتار به متن، متن به گفتار و پردازش صوت برای تجربه‌های صوتی.',
  },
  {
    name: 'DeepInfra',
    description: 'اجرای مدل‌های متن‌باز و تجاری با API سازگار و هزینه مناسب.',
  },
  {
    name: 'ElevenLabs',
    description: 'تولید صدای طبیعی و تبدیل متن به گفتار با کیفیت بالا.',
  },
  {
    name: 'Google Gemini و Vertex AI',
    description: 'مدل‌های چندرسانه‌ای گوگل و بستر ابری Vertex برای توسعه و مدیریت مدل‌ها.',
  },
  {
    name: 'Groq',
    description: 'زیرساخت اجرای فوق‌سریع مدل‌های زبانی برای کاهش زمان پاسخ‌گویی.',
  },
  {
    name: 'Mistral AI',
    description: 'مدل‌های متن‌باز و تجاری با تمرکز بر کارایی، عملکرد و استقرار انعطاف‌پذیر.',
  },
  {
    name: 'OpenAI',
    description: 'مدل‌های GPT، o-series و مدل‌های پیشرفته برای متن، تصویر، صوت و استدلال.',
  },
  {
    name: 'OpenRouter',
    description: 'دسترسی یکپارچه به مدل‌های متنوع ارائه‌دهندگان مختلف از طریق یک API.',
  },
  {
    name: 'Perplexity AI',
    description: 'مدل‌ها و سرویس‌های جستجوی هوشمند مبتنی بر LLM برای پاسخ‌های مستند و به‌روز.',
  },
  {
    name: 'SambaNova',
    description: 'زیرساخت و مدل‌های هوش مصنوعی سازمانی برای اجرای سریع مدل‌های مولد در مقیاس بزرگ.',
  },
  {
    name: 'Together AI',
    description: 'دسترسی، آموزش و استقرار مدل‌های متن‌باز با API سازگار و تنوع گسترده.',
  },
  {
    name: 'Wafer',
    description: 'زیرساخت پردازشی هوش مصنوعی برای اجرای مدل‌های زبانی و بارهای پردازشی سنگین.',
  },
  {
    name: 'xAI',
    description: 'خانواده مدل‌های Grok برای مکالمه، استدلال و تولید محتوا.',
  },
];

const capabilities = [
  {
    icon: Search,
    title: 'جست‌وجوی سریع مدل مناسب',
    description:
      'مدل‌ها را بر اساس نام، ارائه‌دهنده، توضیح، دسته‌بندی و کاربرد جست‌وجو کنید و فهرست گزینه‌ها را محدود کنید.',
  },
  {
    icon: Filter,
    title: 'فیلتر بر اساس معیارهای فنی',
    description:
      'نوع ورودی و خروجی، پنجره زمینه، قیمت، قابلیت‌ها و پارامترهای پشتیبانی‌شده را برای هر مدل بررسی کنید.',
  },
  {
    icon: GitCompare,
    title: 'مقایسه مدل‌ها کنار هم',
    description:
      'قیمت، پنجره زمینه، نوع ورودی و خروجی و قابلیت‌های چند مدل را در یک نما مقایسه کنید.',
  },
  {
    icon: Database,
    title: 'جزئیات ارائه‌دهنده و مدل',
    description:
      'وضعیت، مستندات، سیاست داده، منطقه میزبانی، قابلیت‌ها و محدودیت‌های هر مدل را یک‌جا ببینید.',
  },
  {
    icon: ShieldCheck,
    title: 'ارزیابی پیش از اتصال',
    description:
      'پیش از اتصال مدل به درگاه، هزینه، قابلیت و سازگاری آن با نیاز محصول را بسنجید.',
  },
  {
    icon: Workflow,
    title: 'اتصال به درگاه مدل‌های زبانی',
    description:
      'پس از انتخاب، مدل را در یک حساب ارائه‌دهنده فعال کنید یا پشت مدل مجازی و مسیر جایگزین قرار دهید.',
  },
];

const faqItems = [
  {
    question: 'کاتالوگ گدارAI چه تفاوتی با یک فهرست ساده دارد؟',
    answer:
      'کاتالوگ علاوه بر نام مدل، ارائه‌دهنده، نوع ورودی و خروجی، قیمت، پنجره زمینه، قابلیت‌ها، پارامترهای پشتیبانی‌شده و پیوند مستندات را نشان می‌دهد.',
  },
  {
    question: 'آیا می‌توانم مدل‌های چند ارائه‌دهنده را مقایسه کنم؟',
    answer:
      'بله. داده مدل‌های ارائه‌دهنده‌های مختلف با معیارهای مشترک نمایش داده می‌شود تا بتوانید آن‌ها را فیلتر و مقایسه کنید.',
  },
  {
    question: 'بعد از انتخاب مدل چه کاری باید انجام دهم؟',
    answer:
      'مدل را در حساب ارائه‌دهنده فضای کاری فعال کنید. سپس می‌توانید شناسه آن را مستقیم به درگاه بفرستید یا آن را در مدل مجازی و مسیر جایگزین به کار ببرید.',
  },
  {
    question: 'چرا این قابلیت برای تیم‌های ایرانی مهم است؟',
    answer:
      'چون علاوه بر کیفیت، باید هزینه، دسترسی، ارائه‌دهنده جایگزین و پایداری مسیر عملیاتی را نیز در تصمیم خود لحاظ کنید.',
  },
];

function CatalogHeroVisual() {
  const rows = [
    { model: 'gpt-4.1', provider: 'OpenAI', context: '1M', price: '$$', tone: 'bg-emerald-500' },
    { model: 'claude-sonnet-4', provider: 'Anthropic', context: '200K', price: '$$$', tone: 'bg-purple-500' },
    { model: 'gemini-2.5-pro', provider: 'Google Gemini', context: '1M', price: '$$', tone: 'bg-sky-500' },
    { model: 'mistral-large', provider: 'Mistral AI', context: '128K', price: '$$', tone: 'bg-amber-500' },
  ];

  return (
    <div className="relative">
      <div className="absolute -inset-4 rounded-3xl bg-indigo-500/10 blur-3xl dark:bg-indigo-500/20" />
      <div className="relative overflow-hidden rounded-2xl border border-slate-200 bg-slate-50 shadow-2xl dark:border-slate-800 dark:bg-[#121827]">
        <div className="flex items-center justify-between border-b border-slate-200 bg-white/70 px-4 py-3 dark:border-slate-800 dark:bg-slate-900/60">
          <div className="flex items-center gap-2">
            <div className="h-3 w-3 rounded-full bg-rose-500/80" />
            <div className="h-3 w-3 rounded-full bg-amber-500/80" />
            <div className="h-3 w-3 rounded-full bg-emerald-500/80" />
          </div>
          <div className="flex items-center gap-2 rounded-lg bg-slate-100 px-3 py-1 text-xs font-medium text-slate-500 dark:bg-slate-800 dark:text-slate-400">
            <Sparkles className="h-3.5 w-3.5" />
            Model Discovery
          </div>
        </div>

        <div className="grid gap-4 p-4 lg:grid-cols-[280px_1fr]">
          <aside className="space-y-3 rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-800 dark:bg-slate-900">
            <div>
              <div className="mb-2 text-xs font-bold text-slate-900 dark:text-white">جست‌وجوی کاتالوگ</div>
              <div className="flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 dark:border-slate-800 dark:bg-[#0B0F19]">
                <Search className="h-4 w-4 text-indigo-500" />
                <span className="text-xs text-slate-500 dark:text-slate-400">reasoning, vision, audio...</span>
              </div>
            </div>
            {[
              ['Provider', 'OpenAI, Anthropic, Google'],
              ['Modality', 'text, image, audio'],
              ['Context', '128K تا 1M'],
              ['قیمت', 'ورودی و خروجی جداگانه'],
            ].map(([label, value]) => (
              <div key={label} className="rounded-lg bg-slate-50 p-3 dark:bg-slate-800/50">
                <div className="text-[10px] font-medium text-slate-500 dark:text-slate-400">{label}</div>
                <div className="mt-1 text-xs font-bold text-slate-800 dark:text-slate-200" dir="ltr">
                  {value}
                </div>
              </div>
            ))}
          </aside>

          <div className="space-y-3">
            <div className="grid grid-cols-3 gap-3">
              {[
                { label: 'ارائه‌دهنده فعال', value: '۲۱', icon: Globe2 },
                { label: 'مدل قابل جست‌وجو', value: 'صدها', icon: Database },
                { label: 'فیلتر تصمیم‌ساز', value: '۱۰+', icon: Filter },
              ].map((item) => (
                <div key={item.label} className="rounded-xl border border-slate-200 bg-white p-3 dark:border-slate-800 dark:bg-slate-900">
                  <item.icon className="mb-2 h-4 w-4 text-indigo-500" />
                  <div className="text-lg font-black text-slate-900 dark:text-white">{item.value}</div>
                  <div className="text-[10px] text-slate-500 dark:text-slate-400">{item.label}</div>
                </div>
              ))}
            </div>

            <div className="overflow-hidden rounded-xl border border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-900">
              <div className="grid grid-cols-[1.4fr_1fr_80px_70px] gap-3 border-b border-slate-100 px-4 py-2 text-[10px] font-bold text-slate-400 dark:border-slate-800">
                <span>مدل</span>
                <span>Provider</span>
                <span>Context</span>
                <span>Cost</span>
              </div>
              {rows.map((row) => (
                <div
                  key={row.model}
                  className="grid grid-cols-[1.4fr_1fr_80px_70px] items-center gap-3 border-b border-slate-50 px-4 py-3 text-xs last:border-0 dark:border-slate-800/60"
                >
                  <div className="flex min-w-0 items-center gap-2">
                    <span className={`h-2.5 w-2.5 shrink-0 rounded-full ${row.tone}`} />
                    <span className="truncate font-bold text-slate-900 dark:text-white" dir="ltr">
                      {row.model}
                    </span>
                  </div>
                  <span className="truncate text-slate-500 dark:text-slate-400" dir="ltr">
                    {row.provider}
                  </span>
                  <span className="text-slate-700 dark:text-slate-300" dir="ltr">
                    {row.context}
                  </span>
                  <span className="text-slate-500 dark:text-slate-400" dir="ltr">
                    {row.price}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function ProviderLogoStrip() {
  const logos = [
    { src: '/providers/openai.svg', alt: 'OpenAI' },
    { src: '/providers/anthropic.svg', alt: 'Anthropic' },
    { src: '/providers/google.svg', alt: 'Google Gemini' },
  ];

  return (
    <div className="flex flex-wrap items-center justify-center gap-3">
      {logos.map((logo) => (
        <div key={logo.alt} className="flex items-center gap-2 rounded-full border border-slate-200 bg-white px-4 py-2 shadow-sm dark:border-slate-800 dark:bg-slate-900">
          <Image src={logo.src} alt={logo.alt} width={22} height={22} className="h-5 w-5 object-contain" />
          <span className="text-sm font-bold text-slate-700 dark:text-slate-200" dir="ltr">
            {logo.alt}
          </span>
        </div>
      ))}
      <span className="rounded-full border border-dashed border-indigo-200 bg-indigo-50 px-4 py-2 text-sm font-bold text-indigo-600 dark:border-indigo-500/30 dark:bg-indigo-500/10 dark:text-indigo-300">
        و ۱۸ ارائه‌دهنده دیگر
      </span>
    </div>
  );
}

export function ModelDiscovery() {
  return (
    <main className="flex min-h-screen flex-col bg-white text-slate-900 dark:bg-[#0B0F19] dark:text-white" dir="rtl">
      <section className="relative overflow-hidden pb-20 pt-32">
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top,#4F46E5_0%,transparent_52%)] opacity-[0.08] dark:opacity-[0.16]" />
        <div className="relative z-10 mx-auto grid max-w-7xl items-center gap-12 px-4 sm:px-6 lg:grid-cols-[0.92fr_1.08fr] lg:px-8">
          <div>
            <motion.div
              initial={{ opacity: 0, scale: 0.92 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5 }}
              className="mb-8 inline-flex items-center gap-2 rounded-full border border-indigo-100 bg-indigo-50 px-4 py-2 text-sm font-bold text-indigo-600 dark:border-indigo-500/20 dark:bg-indigo-500/10 dark:text-indigo-400"
              style={{ fontFamily: 'var(--font-heading)' }}
            >
              <Database className="h-4 w-4" />
              کاتالوگ و مقایسه مدل‌ها
            </motion.div>

            <motion.h1
              className="mb-8 text-4xl font-bold leading-tight tracking-tight text-slate-900 dark:text-white md:text-5xl lg:text-6xl"
              style={{ fontFamily: 'var(--font-heading)' }}
              {...fadeIn}
            >
              مدل مناسب را سریع‌تر پیدا کنید؛{' '}
              <span className="bg-gradient-to-l from-indigo-500 to-teal-500 bg-clip-text text-transparent">
                با داده، نه حدس
              </span>
            </motion.h1>

            <motion.p
              className="mb-10 max-w-2xl text-lg leading-relaxed text-slate-600 dark:text-slate-400 md:text-xl"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >
              کاتالوگ گدارAI اطلاعات مدل‌ها و ارائه‌دهنده‌ها را کنار هم قرار می‌دهد تا پیش از اتصال به درگاه،
              مدل‌ها را جست‌وجو، فیلتر و با معیارهای مشترک مقایسه کنید.
            </motion.p>

            <motion.div
              className="flex flex-col gap-4 sm:flex-row"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
            >
              <Link
                href="/catalog"
                className="flex min-h-12 w-full items-center justify-center gap-2 rounded-xl bg-indigo-600 px-8 py-4 text-lg font-bold text-white shadow-[0_0_20px_rgba(79,70,229,0.3)] transition-all hover:bg-indigo-700 hover:shadow-[0_0_30px_rgba(79,70,229,0.5)] active:scale-95 sm:w-auto"
              >
                مشاهده کاتالوگ مدل‌ها
                <ArrowLeft className="h-5 w-5" />
              </Link>
              <Link
                href="/docs/provider-models-overview"
                className="flex min-h-12 w-full items-center justify-center rounded-xl border-2 border-slate-200 bg-white px-8 py-4 text-lg font-bold text-slate-900 transition-all hover:border-indigo-600 active:scale-95 dark:border-slate-800 dark:bg-slate-900 dark:text-white dark:hover:border-indigo-500 sm:w-auto"
              >
                مستندات کاتالوگ
              </Link>
            </motion.div>
          </div>

          <motion.div
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, delay: 0.25 }}
          >
            <CatalogHeroVisual />
          </motion.div>
        </div>
      </section>

      <section className="border-y border-slate-100 bg-slate-50/60 py-12 dark:border-slate-800/50 dark:bg-slate-900/20">
        <div className="mx-auto max-w-7xl px-4 text-center sm:px-6 lg:px-8">
          <p className="mb-6 text-sm font-bold text-slate-500 dark:text-slate-400">
            مدل‌های ۲۱ ارائه‌دهنده فعال را در یک نمای قابل جست‌وجو بررسی و مقایسه کنید
          </p>
          <ProviderLogoStrip />
        </div>
      </section>

      <section className="py-24">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="mx-auto mb-16 max-w-3xl text-center">
            <h2
              className="mb-4 text-3xl font-bold text-slate-900 dark:text-white md:text-4xl"
              style={{ fontFamily: 'var(--font-heading)' }}
            >
              از «کدام مدل؟» تا «چطور در محصول استفاده کنیم؟»
            </h2>
            <p className="text-lg leading-relaxed text-slate-600 dark:text-slate-400">
              انتخاب مدل بر هزینه، کیفیت پاسخ، تأخیر و تجربه کاربر اثر می‌گذارد. کاتالوگ گدارAI داده‌های لازم را
              برای مقایسه فراهم می‌کند و مسیر اتصال مدل منتخب به درگاه را نشان می‌دهد.
            </p>
          </div>

          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
            {capabilities.map((item, index) => (
              <motion.div
                key={item.title}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.4, delay: (index % 3) * 0.08 }}
              >
                <MagicCard
                  className="h-full rounded-2xl border border-slate-200 bg-white p-6 transition-colors hover:border-indigo-200 dark:border-slate-800 dark:bg-slate-900/60 dark:hover:border-indigo-800"
                  gradientColor="rgba(20,184,166,0.1)"
                >
                  <div className="mb-5 flex h-11 w-11 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400">
                    <item.icon className="h-5 w-5" />
                  </div>
                  <h3 className="mb-3 text-lg font-bold text-slate-900 dark:text-white">{item.title}</h3>
                  <p className="text-sm leading-relaxed text-slate-600 dark:text-slate-400">{item.description}</p>
                </MagicCard>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      <section className="border-y border-slate-100 bg-slate-50 py-24 dark:border-slate-800/50 dark:bg-slate-900/30">
        <div className="mx-auto grid max-w-7xl gap-12 px-4 sm:px-6 lg:grid-cols-[0.9fr_1.1fr] lg:px-8">
          <div>
            <div className="mb-4 inline-flex items-center gap-2 rounded-full bg-teal-50 px-3 py-1 text-xs font-bold text-teal-700 dark:bg-teal-500/10 dark:text-teal-300">
              <Columns3 className="h-3.5 w-3.5" />
              مقایسه بدون پراکندگی
            </div>
            <h2
              className="mb-5 text-3xl font-bold leading-tight text-slate-900 dark:text-white md:text-4xl"
              style={{ fontFamily: 'var(--font-heading)' }}
            >
              پیش از انتخاب، بده‌بستان‌های هر مدل را بسنجید
            </h2>
            <p className="mb-8 text-lg leading-relaxed text-slate-600 dark:text-slate-400">
              بعضی مدل‌ها برای استدلال مناسب‌ترند، بعضی تأخیر کمتری دارند و بعضی از صوت یا تصویر پشتیبانی می‌کنند.
              مقایسه این تفاوت‌ها کمک می‌کند تیم‌های محصول، فنی و مالی با معیارهای مشترک تصمیم بگیرند.
            </p>

            <div className="grid gap-3 sm:grid-cols-2">
              {[
                'مقایسه قیمت ورودی و خروجی',
                'بررسی نوع ورودی و خروجی',
                'بررسی پنجره زمینه و قابلیت‌ها',
                'یافتن ارائه‌دهنده جایگزین برای زمان خطا',
              ].map((item) => (
                <div key={item} className="flex items-center gap-2 rounded-xl border border-slate-200 bg-white p-3 text-sm font-bold text-slate-700 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-200">
                  <Check className="h-4 w-4 shrink-0 text-emerald-500" />
                  {item}
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-xl dark:border-slate-800 dark:bg-[#121827]">
            <div className="mb-5 flex items-center justify-between">
              <div>
                <h3 className="text-sm font-bold text-slate-900 dark:text-white">مقایسه مدل‌های منتخب</h3>
                <p className="text-xs text-slate-500 dark:text-slate-400">سه مدل منتخب برای سناریوی پشتیبانی مشتری</p>
              </div>
              <GitCompare className="h-5 w-5 text-indigo-500" />
            </div>

            <div className="grid gap-3 md:grid-cols-3">
              {[
                { model: 'GPT', provider: 'OpenAI', score: '۹۲', accent: 'text-emerald-500' },
                { model: 'Claude', provider: 'Anthropic', score: '۹۰', accent: 'text-purple-500' },
                { model: 'Gemini', provider: 'Google', score: '۸۸', accent: 'text-sky-500' },
              ].map((item) => (
                <div key={item.model} className="rounded-xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-900">
                  <div className="mb-4 text-xs text-slate-500 dark:text-slate-400" dir="ltr">
                    {item.provider}
                  </div>
                  <div className="mb-1 text-xl font-black text-slate-900 dark:text-white" dir="ltr">
                    {item.model}
                  </div>
                  <div className={`mb-4 text-3xl font-black ${item.accent}`}>{item.score}</div>
                  <div className="space-y-2 text-xs text-slate-600 dark:text-slate-400">
                    <div className="flex items-center justify-between">
                      <span>تأخیر</span>
                      <span>کم</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span>پنجره زمینه</span>
                      <span>زیاد</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span>هزینه</span>
                      <span>متوسط</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      <section className="py-24">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="mx-auto mb-12 max-w-3xl text-center">
            <h2
              className="mb-4 text-3xl font-bold text-slate-900 dark:text-white md:text-4xl"
              style={{ fontFamily: 'var(--font-heading)' }}
            >
              ۲۱ ارائه‌دهنده فعال در کاتالوگ گدارAI
            </h2>
            <p className="text-lg leading-relaxed text-slate-600 dark:text-slate-400">
              مدل‌های متنی، چندرسانه‌ای، صوتی و امبدینگ را در یک نمای یکدست بررسی کنید و سپس مدل منتخب را به
              مسیر عملیاتی خود متصل کنید.
            </p>
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {providerGroups.map((provider) => (
              <div
                key={provider.name}
                className="rounded-2xl border border-slate-200 bg-slate-50 p-5 transition-colors hover:border-indigo-200 dark:border-slate-800 dark:bg-slate-900/60 dark:hover:border-indigo-800"
              >
                <div className="mb-3 flex items-center gap-2">
                  <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-white text-xs font-black text-indigo-600 shadow-sm dark:bg-slate-800 dark:text-indigo-300">
                    {provider.name.slice(0, 1)}
                  </span>
                  <h3 className="font-bold text-slate-900 dark:text-white" dir="ltr">
                    {provider.name}
                  </h3>
                </div>
                <p className="text-sm leading-relaxed text-slate-600 dark:text-slate-400">{provider.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="border-y border-slate-100 bg-slate-50 py-24 dark:border-slate-800/50 dark:bg-slate-900/30">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="mx-auto mb-12 max-w-3xl text-center">
            <h2
              className="mb-4 text-3xl font-bold text-slate-900 dark:text-white md:text-4xl"
              style={{ fontFamily: 'var(--font-heading)' }}
            >
              جزئیات فنی هر مرحله را در مستندات بخوانید
            </h2>
            <p className="text-lg text-slate-600 dark:text-slate-400">
              برای شناخت داده‌های کاتالوگ، مقایسه مدل‌ها و ساخت مدل مجازی، راهنمای مرتبط را دنبال کنید.
            </p>
          </div>

          <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-4">
            {docsLinks.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className="group flex h-full flex-col rounded-2xl border border-slate-200 bg-white p-5 transition-all hover:-translate-y-1 hover:border-indigo-300 hover:shadow-lg dark:border-slate-800 dark:bg-slate-900 dark:hover:border-indigo-700"
              >
                <BookOpen className="mb-4 h-5 w-5 text-indigo-500" />
                <h3 className="mb-2 font-bold text-slate-900 transition-colors group-hover:text-indigo-600 dark:text-white dark:group-hover:text-indigo-400">
                  {item.title}
                </h3>
                <p className="mb-4 flex-1 text-sm leading-relaxed text-slate-600 dark:text-slate-400">{item.description}</p>
                <span className="inline-flex items-center gap-1 text-sm font-bold text-indigo-600 dark:text-indigo-400">
                  مطالعه بیشتر
                  <ArrowLeft className="h-4 w-4 transition-transform group-hover:-translate-x-1" />
                </span>
              </Link>
            ))}
          </div>
        </div>
      </section>

      <section className="py-24">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="grid gap-6 md:grid-cols-3">
            {[
              {
                icon: ListChecks,
                title: 'برای تیم محصول',
                description: 'مدل‌هایی را انتخاب کنید که با تجربه کاربر، کیفیت پاسخ و سناریوی محصول هماهنگ‌اند.',
              },
              {
                icon: Gauge,
                title: 'برای تیم فنی',
                description: 'پیش از اتصال به درگاه، محدودیت‌ها، قابلیت‌ها و مقصدهای جایگزین را بررسی کنید.',
              },
              {
                icon: LineChart,
                title: 'برای تیم مالی',
                description: 'هزینه و مصرف مدل منتخب را در کنار ارائه‌دهنده‌های جایگزین بسنجید.',
              },
            ].map((item) => (
              <div key={item.title} className="rounded-2xl border border-slate-200 bg-white p-7 dark:border-slate-800 dark:bg-slate-900/60">
                <item.icon className="mb-5 h-8 w-8 text-teal-500" />
                <h3 className="mb-3 text-xl font-bold text-slate-900 dark:text-white">{item.title}</h3>
                <p className="text-sm leading-relaxed text-slate-600 dark:text-slate-400">{item.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="border-t border-slate-100 bg-slate-50 py-24 dark:border-slate-800/50 dark:bg-slate-900/30">
        <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8">
          <div className="mb-12 text-center">
            <h2
              className="text-3xl font-bold text-slate-900 dark:text-white"
              style={{ fontFamily: 'var(--font-heading)' }}
            >
              پرسش‌های پرتکرار
            </h2>
          </div>
          <FAQAccordion items={faqItems} />
        </div>
      </section>

      <section className="relative overflow-hidden py-24">
        <div className="absolute inset-0 bg-indigo-600 dark:bg-indigo-900/40" />
        <div className="relative z-10 mx-auto max-w-4xl px-4 text-center sm:px-6 lg:px-8">
          <BarChart3 className="mx-auto mb-6 h-10 w-10 text-indigo-100" />
          <h2
            className="mb-6 text-3xl font-bold leading-tight text-white md:text-5xl"
            style={{ fontFamily: 'var(--font-heading)' }}
          >
            انتخاب مدل را از یک تصمیم مبهم به یک جریان قابل مدیریت تبدیل کنید
          </h2>
          <p className="mx-auto mb-10 max-w-3xl text-lg leading-relaxed text-indigo-100 md:text-xl">
            مدل‌ها را با داده‌های یکدست مقایسه کنید و سپس مدل منتخب را به درگاه یا یک مدل مجازی متصل کنید.
          </p>
          <div className="flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              href="/catalog"
              className="flex min-h-12 w-full items-center justify-center gap-2 rounded-xl bg-white px-8 py-4 text-lg font-bold text-indigo-700 shadow-xl transition-all hover:bg-indigo-50 active:scale-95 sm:w-auto"
            >
              رفتن به کاتالوگ
              <ArrowLeft className="h-5 w-5" />
            </Link>
            <Link
              href="/demo"
              className="flex min-h-12 w-full items-center justify-center rounded-xl border-2 border-white/30 bg-transparent px-8 py-4 text-lg font-bold text-white transition-all hover:border-white active:scale-95 sm:w-auto"
            >
              درخواست دمو
            </Link>
          </div>
        </div>
      </section>
    </main>
  );
}
