'use client';

import { motion } from 'motion/react';
import Link from 'next/link';
import { FAQAccordion } from '../../components/ui/FAQAccordion';
import {
  AlertTriangle, ArrowLeft, BarChart3, CheckCircle, Code2, CreditCard, FileText,
  GitBranch, Globe, Headphones, Layers, Network, Scale, Server, Shield, ShieldCheck,
  Wallet, Wrench, Zap,
} from 'lucide-react';

const fadeIn = {
  initial: { opacity: 0, y: 20 },
  whileInView: { opacity: 1, y: 0 },
  viewport: { once: true },
  transition: { duration: 0.4 },
};

// ─── Key differentiators ─────────────────────────────────────────────────────
const differentiators = [
  {
    icon: Zap,
    category: 'معماری و عملکرد درگاه',
    godar: 'درگاه با Go نوشته شده و برای ترافیک هم‌زمان و مصرف پایدار حافظه طراحی شده است. نسخه ابری را تیم گدارAI نگهداری و مقیاس‌دهی می‌کند.',
    litellm: 'پراکسی متن‌باز Python که برای شروع سریع و ترافیک کم تا متوسط مناسب است؛ در حجم بالا، تنظیم worker، حافظه و مقیاس افقی بر عهده تیم شماست.',
  },
  {
    icon: GitBranch,
    category: 'مسیریابی و توزیع بار',
    godar: 'مدل مجازی با مسیریابی اولویتی، وزنی و مبتنی بر تأخیر، تلاش دوباره، مهلت پاسخ و مسیر جایگزین؛ همه از رابط وب و بدون تغییر شناسه مدل در کد.',
    litellm: 'Router با fallback و استراتژی‌های توزیع بار دارد؛ پیکربندی عمدتاً در فایل YAML است و برای حالت چندنمونه‌ای به Redis و Postgres نیاز دارد.',
  },
  {
    icon: Server,
    category: 'محل استقرار و مالکیت داده',
    godar: 'استقرار ابری (SaaS) یا استقرار درون‌سازمانی و VPC در پلن سازمانی؛ دسترسی به محتوای لاگ با مجوز جداگانه از فراداده محدود می‌شود.',
    litellm: 'خودمیزبان است و داده در زیرساخت شما می‌ماند؛ اما امنیت، به‌روزرسانی، پشتیبان‌گیری و پایداری همه اجزا بر عهده تیم خودتان است.',
  },
  {
    icon: Network,
    category: 'درگاه MCP و ایجنت‌ها',
    godar: 'درگاه MCP با کنترل دسترسی در سطح ابزار، روش‌های مختلف احراز هویت سرور، ثبت رخداد هر فراخوانی و رد درخواست یکپارچه با فراخوانی مدل.',
    litellm: 'پشتیبانی MCP و قابلیت‌های ایجنت در حال توسعه فعال است؛ آمادگی آن را برای سیاست‌گذاری و حسابرسی موردنیاز خود پیش از اتکا ارزیابی کنید.',
  },
  {
    icon: ShieldCheck,
    category: 'گاردریل‌ها',
    godar: 'گاردریل داخلی برای Regex، تشخیص Secret، PII پایه، JSON Schema و محدودیت طول؛ با دو شیوه audit و enforce، به‌علاوه وب‌هوک اختصاصی و سرویس ثالث.',
    litellm: 'گاردریل عمدتاً از طریق اتصال به سرویس‌های بیرونی مانند Presidio یا ارائه‌دهندگان تجاری فراهم می‌شود که هرکدام استقرار و نگهداری جداگانه دارند.',
  },
  {
    icon: BarChart3,
    category: 'مشاهده‌پذیری',
    godar: 'لاگ، رد درخواست با جزئیات تلاش‌ها و مسیر جایگزین، فراداده سفارشی، گزارش هزینه و خروجی OpenTelemetry در یک رابط فارسی.',
    litellm: 'callbackهای متنوع برای ارسال داده به ابزارهای بیرونی دارد؛ برای تحلیل عمیق معمولاً یک پلتفرم مشاهده‌پذیری جداگانه باید راه‌اندازی شود.',
  },
  {
    icon: FileText,
    category: 'مدیریت پرامپت',
    godar: 'قالب پرامپت با نسخه‌بندی و متغیر، به‌همراه محیط تست زنده برای مقایسه مدل‌ها و نسخه‌ها پیش از اتصال اپلیکیشن.',
    litellm: 'مدیریت پرامپت به‌صورت محدود و اغلب با اتکا به ابزارهای بیرونی ارائه می‌شود؛ برای جریان‌های حساس، بلوغ آن را بررسی کنید.',
  },
  {
    icon: Wallet,
    category: 'کنترل هزینه',
    godar: 'سقف هزینه ریالی و دلاری و محدودیت نرخ برای تیم، کاربر و حساب مجازی؛ قیمت‌گذاری سفارشی مدل و گزارش هزینه به تفکیک مدل و ارائه‌دهنده.',
    litellm: 'بودجه و محدودیت برای کلید، کاربر و تیم دارد؛ هزینه‌ها به دلار محاسبه می‌شوند و گزارش مالی متناسب با بازار ایران باید جداگانه ساخته شود.',
  },
  {
    icon: CreditCard,
    category: 'دسترسی و پرداخت',
    godar: 'دسترسی به بیش از ۱۲۰۰ مدل بدون نیاز به پرداخت ارزی مستقیم یا کارت اعتباری خارجی، با صورتحساب ریالی.',
    litellm: 'نرم‌افزار رایگان است، اما حساب و پرداخت هر ارائه‌دهنده مدل را خودتان باید تهیه و مدیریت کنید.',
  },
  {
    icon: Headphones,
    category: 'پشتیبانی',
    godar: 'پشتیبانی فارسی و هم‌وقت با تیم؛ آنبوردینگ اختصاصی و پشتیبانی با اولویت بالا در پلن سازمانی.',
    litellm: 'پشتیبانی جامعه از طریق GitHub و Discord؛ پشتیبانی تجاری در نسخه Enterprise و با قرارداد ارزی ارائه می‌شود.',
  },
];

// ─── Evaluation questions ────────────────────────────────────────────────────
const evaluationQuestions = [
  {
    question: 'آیا به استقرار درون‌سازمانی و کنترل کامل داده نیاز داریم؟',
    godar: 'نسخه ابری برای شروع سریع و استقرار On-Premise یا VPC برای سازمان‌هایی که داده نباید از شبکه داخلی خارج شود؛ با نقش‌ها و نگهداری داده سفارشی.',
    litellm: 'خودمیزبانی ممکن است، اما گاردریل، مشاهده‌پذیری و پشتیبان‌گیری را باید با چند سرویس جداگانه کامل کنید و مسئولیت امنیت همه آن‌ها با شماست.',
  },
  {
    question: 'امروز از LiteLLM استفاده می‌کنیم. آیا باید مهاجرت کنیم؟',
    godar: 'اگر نگهداری Redis، Postgres و callbackها زمان تیم را می‌گیرد، یا به کنترل هزینه ریالی، RBAC و گاردریل داخلی نیاز دارید، مهاجرت ارزش ارزیابی دارد.',
    litellm: 'برای نمونه‌سازی یا یک سرویس داخلی با ترافیک محدود، اگر تیم شما از نگهداری آن راضی است، ادامه استفاده منطقی است.',
  },
  {
    question: 'حاکمیت ایجنت‌ها و ابزارهای MCP چقدر برای ما فوری است؟',
    godar: 'دسترسی ابزار به ابزار، ثبت هویت و علت تصمیم برای هر فراخوانی و رد درخواست مشترک با فراخوانی مدل در همان فضای کاری.',
    litellm: 'قابلیت‌های MCP وجود دارد و به‌سرعت تغییر می‌کند؛ برای حسابرسی و سیاست‌گذاری دقیق، پوشش فعلی را با سناریوی خودتان آزمایش کنید.',
  },
  {
    question: 'هزینه AI را در کل سازمان چگونه کنترل کنیم؟',
    godar: 'سقف هزینه پیش از مصرف، محدودیت نرخ، حافظه نهان معنایی و گزارش هزینه ریالی به تفکیک تیم، کاربر، مدل و فراداده سفارشی.',
    litellm: 'بودجه برای کلید و تیم وجود دارد، اما تبدیل هزینه دلاری به گزارش قابل استفاده برای واحد مالی داخلی به کار اضافه نیاز دارد.',
  },
  {
    question: 'به مشاهده‌پذیری کامل نیاز داریم یا فقط شمارش توکن؟',
    godar: 'لاگ، رد درخواست، نتیجه گاردریل، هزینه و هشدار در یک رابط؛ و در صورت نیاز خروجی OpenTelemetry به ابزار فعلی سازمان.',
    litellm: 'داده خام از طریق callback ارسال می‌شود و داشبورد، نگهداری و جست‌وجوی آن به ابزار بیرونی وابسته است.',
  },
  {
    question: 'آیا تیم ما باید زیرساخت درگاه را نگهداری کند؟',
    godar: 'در نسخه ابری خیر؛ ارتقا، مقیاس‌دهی و پایداری با گدارAI است و تیم روی محصول تمرکز می‌کند.',
    litellm: 'بله؛ به‌روزرسانی نسخه‌ها، پایش سلامت پراکسی، پایگاه‌داده و Redis و پاسخ به حادثه با تیم شماست.',
  },
];

// ─── Pain points ─────────────────────────────────────────────────────────────
const painPoints = [
  {
    pain: 'رسیدن به سقف معماری پراکسی',
    benefit: 'درگاه Go با مصرف منابع پیش‌بینی‌پذیر و مقیاس‌دهی مدیریت‌شده در نسخه ابری',
    impact: 'به‌جای تنظیم worker و حافظه پراکسی، ظرفیت را بر اساس داده واقعی مصرف برنامه‌ریزی می‌کنید.',
  },
  {
    pain: 'نگهداری زیرساخت به‌جای ساخت محصول',
    benefit: 'پلتفرم مدیریت‌شده؛ بدون نیاز به نگهداری Redis، Postgres و چند سرویس جانبی',
    impact: 'زمان مهندسی که صرف ارتقا و عیب‌یابی پراکسی می‌شد به توسعه قابلیت‌های محصول برمی‌گردد.',
  },
  {
    pain: 'الزامات امنیتی و انطباق سازمانی',
    benefit: 'RBAC، SSO، ثبت وقایع مدیریتی و استقرار On-Premise یا VPC',
    impact: 'پاسخ به پرسش‌های واحد امنیت و حراست سازمان پیش از تأیید نهایی پروژه ساده‌تر می‌شود.',
  },
  {
    pain: 'پرداخت ارزی و مدیریت چند حساب خارجی',
    benefit: 'یک قرارداد و صورتحساب ریالی برای دسترسی به بیش از ۱۲۰۰ مدل',
    impact: 'ریسک مسدودشدن کارت یا حساب و هماهنگی مالی با چند ارائه‌دهنده خارجی کمتر می‌شود.',
  },
  {
    pain: 'وابستگی گاردریل به سرویس‌های بیرونی',
    benefit: 'گاردریل داخلی در مسیر درخواست با شیوه audit برای آزمایش بدون قطع سرویس',
    impact: 'برای بررسی‌های رایج نیازی به استقرار و نگهداری سرویس جداگانه ندارید.',
  },
  {
    pain: 'پراکندگی داده مصرف و هزینه',
    benefit: 'گزارش هزینه و عملکرد به تفکیک تیم، مدل و فراداده سفارشی در رابط فارسی',
    impact: 'تیم فنی و مالی بر اساس یک منبع داده مشترک درباره مدل و بودجه تصمیم می‌گیرند.',
  },
];

// ─── Pitfalls ────────────────────────────────────────────────────────────────
const pitfalls = [
  'سقف مقیاس‌پذیری را مسئله آینده فرض کنید؛ تغییر معماری درگاه پس از رشد ترافیک پرهزینه‌تر است.',
  'برای حادثه‌های بحرانی محیط عملیاتی فقط به پاسخ جامعه متن‌باز در GitHub یا Discord اتکا کنید.',
  'هزینه «رایگان» را فقط با لایسنس بسنجید و زمان مهندسی نگهداری، پایش و ارتقا را نادیده بگیرید.',
  'بدون ثبت رخداد و کنترل دسترسی در سطح ابزار، ایجنت‌ها را به ابزارهای MCP داخلی متصل کنید.',
  'کلیدهای خام ارائه‌دهنده را در سرویس‌های مختلف پخش کنید به‌جای صدور توکن مستقل برای هر سرویس.',
  'اثر گاردریل‌ها را پیش از فعال‌سازی enforce روی داده واقعی فارسی با شیوه audit نسنجید.',
];

// ─── Migration scenarios ─────────────────────────────────────────────────────
const scenarios = [
  {
    icon: Layers,
    title: 'تیم پلتفرم با چند سرویس داخلی',
    desc: 'اتصال‌های پراکنده به پراکسی را با مدل‌های مجازی جایگزین می‌کند و برای هر سرویس توکن مستقل با مدل‌های مجاز صادر می‌کند.',
  },
  {
    icon: Wallet,
    title: 'محصول SaaS با مصرف رو‌به‌رشد',
    desc: 'با سقف هزینه برای هر تیم و حافظه نهان معنایی، رشد مصرف را قابل پیش‌بینی و گزارش هزینه را برای واحد مالی قابل استفاده می‌کند.',
  },
  {
    icon: Shield,
    title: 'سازمان با الزامات حراست و امنیت',
    desc: 'درگاه را در زیرساخت داخلی مستقر می‌کند و دسترسی به محتوای لاگ را فقط به نقش‌های مجاز می‌دهد.',
  },
  {
    icon: Wrench,
    title: 'تیم ایجنت و اتوماسیون',
    desc: 'فراخوانی مدل و ابزارهای MCP را از یک مسیر مشترک عبور می‌دهد تا خطا، تلاش دوباره و مجوز هر ابزار قابل پیگیری باشد.',
  },
];

// ─── FAQ ─────────────────────────────────────────────────────────────────────
const faqItems = [
  {
    question: 'تفاوت اصلی گدارAI و LiteLLM چیست؟',
    answer: 'LiteLLM یک پراکسی متن‌باز Python است که دسترسی سریع به چند ارائه‌دهنده را ممکن می‌کند و نگهداری آن با تیم شماست. گدارAI یک پلتفرم مدیریت‌شده برای اتصال، کنترل و پایش مدل‌ها و ابزارهای AI است که درگاه، گاردریل، مدیریت پرامپت، کنترل هزینه و مشاهده‌پذیری را یکجا ارائه می‌کند و برای تیم‌های ایرانی طراحی شده است.',
  },
  {
    question: 'LiteLLM رایگان است؛ هزینه گدارAI چگونه توجیه می‌شود؟',
    answer: 'هزینه واقعی درگاه فقط لایسنس نیست. زمان مهندسی برای استقرار، ارتقا، پایش، پشتیبان‌گیری، راه‌اندازی سرویس‌های گاردریل و مشاهده‌پذیری و پاسخ به حادثه را هم محاسبه کنید. در کنار آن، تهیه حساب و پرداخت ارزی برای هر ارائه‌دهنده هزینه و ریسک جداگانه دارد.',
  },
  {
    question: 'LiteLLM را در محیط عملیاتی داریم. چه نشانه‌هایی می‌گوید باید مهاجرت را بررسی کنیم؟',
    answer: 'کندی یا ناپایداری پراکسی با افزایش ترافیک، زمان زیاد تیم برای نگهداری Redis و Postgres، نیاز به RBAC و SSO، نیاز به گزارش هزینه ریالی، الزام استقرار درون‌سازمانی با پشتیبانی رسمی، یا نیاز به حاکمیت دسترسی ایجنت‌ها به ابزارهای MCP.',
  },
  {
    question: 'مهاجرت از LiteLLM به گدارAI چقدر کار دارد؟',
    answer: 'هر دو رابط سازگار با OpenAI ارائه می‌کنند؛ در اغلب موارد نشانی پایه، توکن و شناسه مدل تغییر می‌کند. مدل‌های تعریف‌شده در پیکربندی LiteLLM را به مدل مجازی تبدیل کنید و قراردادهای اختصاصی هر ارائه‌دهنده را پیش از مهاجرت کامل در محیط خود آزمایش کنید.',
  },
  {
    question: 'حاکمیت MCP و ایجنت‌ها در دو پلتفرم چه تفاوتی دارد؟',
    answer: 'در گدارAI دسترسی ایجنت به هر ابزار جداگانه تعریف می‌شود، درخواست غیرمجاز پیش از رسیدن به سرور رد می‌شود و رخداد آن با هویت، ابزار و علت تصمیم ثبت می‌شود. در LiteLLM قابلیت‌های MCP وجود دارد و به‌سرعت در حال تغییر است؛ پوشش آن را با سناریوی حسابرسی خودتان بسنجید.',
  },
  {
    question: 'داده‌های ما کجا نگهداری می‌شود؟',
    answer: 'در نسخه ابری، داده در زیرساخت مدیریت‌شده گدارAI نگهداری می‌شود و مدت نگهداری لاگ و رد درخواست به پلن شما بستگی دارد. در پلن سازمانی می‌توانید گدارAI را در زیرساخت خود یا VPC مستقر کنید و سیاست نگهداری داده سفارشی تعریف کنید.',
  },
  {
    question: 'آیا گدارAI برای تیم‌های کوچک بیش از حد لازم نیست؟',
    answer: 'خیر. پلن توسعه‌دهنده برای نمونه‌سازی و پروژه‌های کوچک طراحی شده و می‌توانید فقط از درگاه و مدل مجازی شروع کنید. قابلیت‌هایی مانند گاردریل، بودجه و SSO را هر زمان که لازم شد فعال کنید.',
  },
  {
    question: 'تیم ما به Python مسلط است؛ چرا روی LiteLLM نمانیم؟',
    answer: 'تسلط به Python برای توسعه اپلیکیشن ارزشمند است، اما پرسش اصلی این است که آیا تیم باید زمان خود را صرف نگهداری زیرساخت درگاه کند یا ساخت محصول. SDKهای سازگار با OpenAI در Python بدون تغییر با گدارAI کار می‌کنند.',
  },
];

// ─── Section heading ─────────────────────────────────────────────────────────
function SectionHeading({ title, subtitle, dark = false }: { title: string; subtitle?: string; dark?: boolean }) {
  return (
    <div className="text-center mb-14 max-w-3xl mx-auto">
      <h2
        className={`text-3xl md:text-4xl font-bold mb-4 ${dark ? 'text-white' : 'text-slate-900 dark:text-white'}`}
        style={{ fontFamily: 'var(--font-heading)' }}
      >
        {title}
      </h2>
      {subtitle && (
        <p className={`text-lg leading-relaxed ${dark ? 'text-slate-400' : 'text-slate-600 dark:text-slate-400'}`}>{subtitle}</p>
      )}
    </div>
  );
}

// ─── Main Component ──────────────────────────────────────────────────────────
export function GodarAIvsLiteLLM() {
  return (
    <main className="flex min-h-screen flex-col bg-white dark:bg-[#0B0F19] text-slate-900 dark:text-white" dir="rtl">

      {/* ═══ 1. Hero ══════════════════════════════════════════════════════════ */}
      <section className="relative pt-32 pb-20 overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,#4F46E5_0%,transparent_50%)] opacity-[0.08] dark:opacity-[0.15] pointer-events-none" />

        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10">
          <div className="text-center max-w-4xl mx-auto">
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.5 }}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-indigo-50 dark:bg-indigo-500/10 border border-indigo-100 dark:border-indigo-500/20 text-indigo-600 dark:text-indigo-400 text-sm font-bold mb-8"
              style={{ fontFamily: 'var(--font-heading)' }}
            >
              <Scale className="w-4 h-4" />
              مقایسه گدارAI و LiteLLM
            </motion.div>

            <motion.h1
              className="text-4xl md:text-5xl lg:text-6xl font-bold tracking-tight mb-8 leading-tight text-slate-900 dark:text-white"
              style={{ fontFamily: 'var(--font-heading)' }}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
            >
              گدارAI{' '}
              <span className="text-transparent bg-clip-text bg-gradient-to-l from-indigo-500 to-purple-500">چه زمانی</span>{' '}
              انتخاب مناسب‌تری است؟
            </motion.h1>

            <motion.p
              className="text-lg md:text-xl text-slate-600 dark:text-slate-400 mb-6 leading-relaxed max-w-3xl mx-auto"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
            >
              LiteLLM برای نمونه‌سازی و تیم‌های کوچکی که نگهداری پراکسی را خودشان بر عهده می‌گیرند گزینه خوبی است. وقتی مدل‌ها به محصول عملیاتی می‌رسند، گدارAI درگاه، گاردریل، کنترل هزینه، درگاه MCP و مشاهده‌پذیری را در یک پلتفرم مدیریت‌شده و متناسب با نیاز تیم‌های ایرانی ارائه می‌کند.
            </motion.p>

            <motion.p
              className="text-xl md:text-2xl font-bold text-indigo-600 dark:text-indigo-400 mb-10"
              style={{ fontFamily: 'var(--font-heading)' }}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.15 }}
            >
              لایسنس LiteLLM رایگان است؛ گدارAI نگهداری را از دوش تیم شما برمی‌دارد.
            </motion.p>

            <motion.div
              className="flex flex-col sm:flex-row items-center justify-center gap-4"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
            >
              <Link
                href="/onboarding"
                className="w-full sm:w-auto px-8 py-4 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-lg shadow-[0_0_20px_rgba(79,70,229,0.3)] hover:shadow-[0_0_30px_rgba(79,70,229,0.5)] transition-all active:scale-95 flex items-center justify-center gap-2 group"
              >
                ۱۵ روز رایگان شروع کنید
                <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
              </Link>
              <Link
                href="/demo"
                className="w-full sm:w-auto px-8 py-4 bg-white dark:bg-slate-900 text-slate-900 dark:text-white border-2 border-slate-200 dark:border-slate-800 hover:border-indigo-600 dark:hover:border-indigo-500 rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center"
              >
                گفت‌وگو با کارشناس
              </Link>
            </motion.div>

            <div className="mt-10 flex flex-wrap justify-center gap-3">
              {['بیش از ۱۲۰۰ مدل با یک API', 'بدون پرداخت ارزی', 'رابط فارسی', 'استقرار ابری یا درون‌سازمانی'].map((chip) => (
                <span key={chip} className="flex items-center gap-1.5 px-4 py-2 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-full text-sm font-medium text-slate-700 dark:text-slate-300 shadow-sm">
                  <CheckCircle className="w-4 h-4 text-emerald-500" />
                  {chip}
                </span>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* ═══ 2. Key differentiators ══════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading
            title="تفاوت‌های کلیدی"
            subtitle="مقایسه بر اساس نیازهایی که معمولاً هنگام انتقال مدل‌ها از نمونه آزمایشی به محیط عملیاتی مطرح می‌شوند."
          />

          <div className="hidden md:grid grid-cols-[1fr_200px_1fr] gap-4 px-6 mb-3 text-sm font-bold" style={{ fontFamily: 'var(--font-heading)' }}>
            <div className="text-indigo-600 dark:text-indigo-400">گدارAI</div>
            <div className="text-center text-slate-500">حوزه</div>
            <div className="text-slate-500">LiteLLM</div>
          </div>

          <div className="space-y-3">
            {differentiators.map((row, i) => (
              <motion.div
                key={row.category}
                {...fadeIn}
                transition={{ duration: 0.4, delay: (i % 4) * 0.05 }}
                className="grid grid-cols-1 md:grid-cols-[1fr_200px_1fr] gap-4 bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-5 md:p-6 md:items-center"
              >
                <div className="md:order-2 flex md:flex-col items-center gap-3 md:gap-2 md:text-center">
                  <div className="w-10 h-10 bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center shrink-0">
                    <row.icon className="w-5 h-5" />
                  </div>
                  <h3 className="font-bold text-slate-900 dark:text-white" style={{ fontFamily: 'var(--font-heading)' }}>{row.category}</h3>
                </div>
                <div className="md:order-1 rounded-xl bg-indigo-50/60 dark:bg-indigo-500/5 border border-indigo-100 dark:border-indigo-500/20 p-4">
                  <div className="md:hidden text-xs font-bold text-indigo-600 dark:text-indigo-400 mb-1">گدارAI</div>
                  <p className="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">{row.godar}</p>
                </div>
                <div className="md:order-3 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-100 dark:border-slate-700/50 p-4">
                  <div className="md:hidden text-xs font-bold text-slate-500 mb-1">LiteLLM</div>
                  <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{row.litellm}</p>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 3. Evaluation questions ═════════════════════════════════════════ */}
      <section className="py-24">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading
            title="پرسش‌های کلیدی برای ارزیابی"
            subtitle="پیش از انتخاب درگاه، پاسخ تیم خود را به این پرسش‌ها روشن کنید."
          />

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {evaluationQuestions.map((item, i) => (
              <motion.div
                key={item.question}
                {...fadeIn}
                transition={{ duration: 0.4, delay: (i % 2) * 0.1 }}
                className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-3xl p-6 md:p-8"
              >
                <div className="flex items-start gap-3 mb-5">
                  <span className="w-8 h-8 rounded-lg bg-indigo-600 text-white flex items-center justify-center text-sm font-bold shrink-0">
                    {(i + 1).toLocaleString('fa-IR')}
                  </span>
                  <h3 className="text-lg font-bold text-slate-900 dark:text-white leading-relaxed" style={{ fontFamily: 'var(--font-heading)' }}>
                    {item.question}
                  </h3>
                </div>
                <div className="space-y-3">
                  <div className="rounded-xl bg-white dark:bg-slate-900 border border-indigo-100 dark:border-indigo-500/20 p-4">
                    <div className="text-xs font-bold text-indigo-600 dark:text-indigo-400 mb-1">گدارAI</div>
                    <p className="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">{item.godar}</p>
                  </div>
                  <div className="rounded-xl bg-white/60 dark:bg-slate-900/40 border border-slate-200 dark:border-slate-800 p-4">
                    <div className="text-xs font-bold text-slate-500 mb-1">LiteLLM</div>
                    <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{item.litellm}</p>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 4. Try before you migrate ═══════════════════════════════════════ */}
      <section className="py-24 bg-slate-900 dark:bg-[#060913] text-white border-y border-slate-800">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <span className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-white/10 border border-white/15 text-indigo-300 text-sm font-bold mb-6">
            <Code2 className="w-4 h-4" />
            پیش از مهاجرت آزمایش کنید
          </span>
          <h2 className="text-3xl md:text-4xl font-bold mb-5" style={{ fontFamily: 'var(--font-heading)' }}>
            دو درگاه را کنار هم اجرا کنید
          </h2>
          <p className="text-lg text-slate-400 leading-relaxed max-w-3xl mx-auto mb-10">
            لازم نیست یک‌باره مهاجرت کنید. بخشی از ترافیک یا یک سرویس غیرحساس را به گدارAI بفرستید و تأخیر، نرخ خطا، هزینه و تجربه تیم را با LiteLLM مقایسه کنید. چون هر دو رابط سازگار با OpenAI دارند، تغییر در کد به نشانی پایه، توکن و شناسه مدل محدود است.
          </p>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-10 text-right">
            {[
              { icon: Zap, title: '۱۵ روز آزمایش رایگان', desc: 'برای پلن‌های غیرسازمانی' },
              { icon: Globe, title: 'بیش از ۱۲۰۰ مدل', desc: 'از یک نشانی پایه و یک توکن' },
              { icon: BarChart3, title: 'مقایسه با داده واقعی', desc: 'لاگ، رد درخواست و گزارش هزینه' },
            ].map((item) => (
              <div key={item.title} className="flex items-center gap-3 bg-slate-800/50 border border-slate-700/50 rounded-2xl p-5">
                <div className="w-10 h-10 bg-indigo-500/20 text-indigo-300 rounded-xl flex items-center justify-center shrink-0">
                  <item.icon className="w-5 h-5" />
                </div>
                <div>
                  <div className="font-bold text-white">{item.title}</div>
                  <div className="text-sm text-slate-400">{item.desc}</div>
                </div>
              </div>
            ))}
          </div>

          <Link
            href="/onboarding"
            className="inline-flex w-full sm:w-auto px-8 py-4 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-bold text-lg transition-all active:scale-95 items-center justify-center gap-2 group"
          >
            کنار LiteLLM اجرا کنید
            <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
          </Link>
        </div>
      </section>

      {/* ═══ 5. Pain points ══════════════════════════════════════════════════ */}
      <section className="py-24">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading
            title="گدارAI کدام مشکل‌ها را حل می‌کند؟"
            subtitle="چالش‌هایی که تیم‌ها معمولاً پس از رشد مصرف روی یک پراکسی خودمیزبان با آن روبه‌رو می‌شوند."
          />

          <div className="hidden lg:grid grid-cols-3 gap-4 px-6 mb-3 text-sm font-bold text-slate-500" style={{ fontFamily: 'var(--font-heading)' }}>
            <div>چالش</div>
            <div>راهکار گدارAI</div>
            <div>اثر برای تیم</div>
          </div>

          <div className="space-y-3">
            {painPoints.map((row, i) => (
              <motion.div
                key={row.pain}
                {...fadeIn}
                transition={{ duration: 0.4, delay: (i % 3) * 0.05 }}
                className="grid grid-cols-1 lg:grid-cols-3 gap-3 lg:gap-4 bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-2xl p-5 lg:p-6 lg:items-center"
              >
                <div className="flex items-center gap-2">
                  <AlertTriangle className="w-5 h-5 text-amber-500 shrink-0" />
                  <h3 className="font-bold text-slate-900 dark:text-white" style={{ fontFamily: 'var(--font-heading)' }}>{row.pain}</h3>
                </div>
                <div className="flex items-start gap-2">
                  <CheckCircle className="w-4 h-4 text-emerald-500 shrink-0 mt-1" />
                  <p className="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">{row.benefit}</p>
                </div>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed lg:border-r lg:border-slate-200 lg:dark:border-slate-700 lg:pr-4">
                  {row.impact}
                </p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 6. Pitfalls ═════════════════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading
            title="اشتباه‌های رایج هنگام انتخاب درگاه"
            subtitle="این خطاها را در ارزیابی گدارAI، LiteLLM یا هر درگاه دیگری تکرار نکنید."
          />
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {pitfalls.map((item, i) => (
              <motion.div
                key={item}
                {...fadeIn}
                transition={{ duration: 0.4, delay: (i % 2) * 0.08 }}
                className="flex items-start gap-3 bg-white dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 rounded-2xl p-5"
              >
                <div className="w-8 h-8 rounded-lg bg-rose-50 dark:bg-rose-500/10 text-rose-600 dark:text-rose-400 flex items-center justify-center shrink-0">
                  <AlertTriangle className="w-4 h-4" />
                </div>
                <p className="text-slate-700 dark:text-slate-300 leading-relaxed">{item}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 7. Migration scenarios ══════════════════════════════════════════ */}
      <section className="py-24">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading
            title="سناریوهای رایج مهاجرت به گدارAI"
            subtitle="الگوهایی که تیم‌ها هنگام انتقال از پراکسی خودمیزبان به یک پلتفرم مدیریت‌شده دنبال می‌کنند."
          />
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
            {scenarios.map((item, i) => (
              <motion.div
                key={item.title}
                {...fadeIn}
                transition={{ duration: 0.4, delay: i * 0.08 }}
                className="bg-slate-50 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 hover:border-indigo-200 dark:hover:border-indigo-800 transition-colors group"
              >
                <div className="w-10 h-10 bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 rounded-xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                  <item.icon className="w-5 h-5" />
                </div>
                <h3 className="font-bold text-slate-900 dark:text-white mb-2" style={{ fontFamily: 'var(--font-heading)' }}>{item.title}</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">{item.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 8. FAQ ══════════════════════════════════════════════════════════ */}
      <section className="py-24 bg-slate-50 dark:bg-slate-900/30 border-y border-slate-100 dark:border-slate-800/50">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <SectionHeading title="پرسش‌ها و تردیدهای رایج" />
          <FAQAccordion items={faqItems} />
          <p className="mt-8 text-center text-xs text-slate-500 dark:text-slate-500 leading-relaxed">
            اطلاعات LiteLLM بر اساس مستندات عمومی آن در زمان انتشار این صفحه است و ممکن است تغییر کرده باشد. پیش از تصمیم نهایی، نسخه فعلی هر دو محصول را در محیط خود ارزیابی کنید.
          </p>
        </div>
      </section>

      {/* ═══ 9. Final CTA ════════════════════════════════════════════════════ */}
      <section className="py-24 relative overflow-hidden">
        <div className="absolute inset-0 bg-indigo-600 dark:bg-indigo-900/40" />
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 relative z-10 text-center">
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4" style={{ fontFamily: 'var(--font-heading)' }}>
            نمونه‌سازی با LiteLLM؛ محیط عملیاتی با گدارAI
          </h2>
          <p className="text-lg text-indigo-200 mb-10 leading-relaxed">
            ۱۵ روز گدارAI را رایگان و کنار درگاه فعلی خود آزمایش کنید؛ با دسترسی به بیش از ۱۲۰۰ مدل، بدون پرداخت ارزی و با پشتیبانی فارسی.
          </p>
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
            <Link
              href="/onboarding"
              className="w-full sm:w-auto px-8 py-4 bg-white text-indigo-700 hover:bg-indigo-50 rounded-xl font-bold text-lg shadow-[0_0_30px_rgba(0,0,0,0.2)] transition-all active:scale-95 flex items-center justify-center gap-2 group"
            >
              کنار LiteLLM اجرا کنید
              <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
            </Link>
            <Link
              href="/demo"
              className="w-full sm:w-auto px-8 py-4 bg-white/10 hover:bg-white/20 text-white border border-white/20 hover:border-white/40 rounded-xl font-bold text-lg transition-all active:scale-95 flex items-center justify-center"
            >
              گفت‌وگو با کارشناس
            </Link>
          </div>
          <div className="mt-8 flex flex-wrap justify-center gap-x-6 gap-y-2 text-sm text-indigo-100">
            {['بدون نیاز به کارت اعتباری خارجی', 'صورتحساب ریالی', 'استقرار On-Premise در پلن سازمانی'].map((item) => (
              <span key={item} className="flex items-center gap-1.5">
                <CheckCircle className="w-4 h-4" />
                {item}
              </span>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
