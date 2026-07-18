import type {ReactNode} from 'react';
import type {LucideIcon} from 'lucide-react';
import {
  Activity,
  BookOpen,
  CheckCircle2,
  Code2,
  Compass,
  CreditCard,
  Gauge,
  KeyRound,
  Layers,
  MessageSquare,
  Route,
  Search,
  ShieldCheck,
  Users,
} from 'lucide-react';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import styles from './index.module.css';

type EntryCard = {
  eyebrow: string;
  title: string;
  description: string;
  to: string;
  action: string;
  icon: LucideIcon;
};

type LinkItem = {
  label: string;
  to: string;
};

type DocGroup = {
  title: string;
  description: string;
  icon: LucideIcon;
  links: LinkItem[];
};

type HelpItem = {
  title: string;
  description: string;
  icon: LucideIcon;
  links: LinkItem[];
};

const entryCards: EntryCard[] = [
  {
    eyebrow: 'مسیر پیشنهادی',
    title: 'اولین درخواست را بفرستید',
    description:
      'فضای کاری، حساب ارائه‌دهنده، توکن دسترسی و شناسه مدل را آماده کنید و یک درخواست Chat Completions از مسیر گدارAI بفرستید.',
    to: '/quick-start',
    action: 'شروع سریع',
    icon: BookOpen,
  },
  {
    eyebrow: 'انتخاب قرارداد',
    title: 'API مناسب را پیدا کنید',
    description:
      'قراردادهای متن، داده، تصویر و صوت را بر اساس کاری که اپلیکیشن شما انجام می‌دهد بررسی کنید.',
    to: '/supported-apis',
    action: 'دیدن APIها',
    icon: Layers,
  },
  {
    eyebrow: 'مدل ذهنی',
    title: 'معماری درگاه را بفهمید',
    description:
      'جایگاه گدارAI میان اپلیکیشن، ارائه‌دهنده مدل، مدل مجازی، کنترل هزینه و داده‌های پایش را ببینید.',
    to: '/intro-to-gateway',
    action: 'خواندن معماری',
    icon: Compass,
  },
];

const requestEssentials = [
  {
    title: 'نشانی پایه درگاه',
    description:
      'نشانی فضای کاری که درخواست‌های سازگار با OpenAI را دریافت می‌کند. مقدار معمولاً با `/v1` پایان می‌یابد.',
    value: 'https://YOUR_WORKSPACE.godarai.ir/v1',
    to: '/making-llm-requests-via-gateway',
  },
  {
    title: 'توکن دسترسی',
    description:
      'برای تست انسانی از PAT استفاده کنید. برای سرویس‌ها و محیط عملیاتی، VAT انتخاب مناسب‌تری است.',
    value: 'YOUR_GODARAI_TOKEN',
    to: '/guides/api-keys',
  },
  {
    title: 'شناسه مدل',
    description:
      'شناسه را از مدل‌های مجاز فضای کاری یا مدل مجازی بردارید. شناسه مدل را از روی نام عمومی مدل حدس نزنید.',
    value: 'YOUR_MODEL_ID',
    to: '/virtual-model',
  },
];

const docGroups: DocGroup[] = [
  {
    title: 'شروع و نمای کلی',
    description: 'شناخت محصول، مسیر شروع و انتخاب صفحه بعدی.',
    icon: BookOpen,
    links: [
      {label: 'معرفی مستندات', to: '/intro'},
      {label: 'نمای کلی قابلیت‌ها', to: '/feature-overview'},
      {label: 'شروع سریع', to: '/quick-start'},
    ],
  },
  {
    title: 'درگاه، مدل‌ها و اتصال',
    description: 'اتصال حساب ارائه‌دهنده، مدل مجازی، احراز هویت و ارسال درخواست.',
    icon: Route,
    links: [
      {label: 'معرفی درگاه', to: '/intro-to-gateway'},
      {label: 'مدل مجازی', to: '/virtual-model'},
      {label: 'اولین درخواست به درگاه', to: '/making-llm-requests-via-gateway'},
      {label: 'احراز هویت', to: '/authentication'},
    ],
  },
  {
    title: 'APIهای پشتیبانی‌شده',
    description: 'قراردادهای سازگار با OpenAI برای متن، داده، تصویر و صوت.',
    icon: Code2,
    links: [
      {label: 'فهرست APIها', to: '/supported-apis'},
      {label: 'Chat Completions', to: '/supported-apis/chat-completions/overview'},
      {label: 'Embeddings', to: '/supported-apis/embeddings'},
      {label: 'تبدیل گفتار به متن', to: '/supported-apis/speech-to-text'},
    ],
  },
  {
    title: 'رجیستری پرامپت',
    description: 'ساخت، نسخه‌بندی و اجرای پرامپت‌های قابل استفاده مجدد.',
    icon: MessageSquare,
    links: [
      {label: 'نمای کلی رجیستری پرامپت', to: '/prompt-registry'},
      {label: 'ساخت پرامپت', to: '/prompt-registry-create'},
      {label: 'متغیرهای ورودی', to: '/prompt-registry-variables'},
      {label: 'اجرای نسخه ذخیره‌شده', to: '/prompt-registry-run'},
    ],
  },
  {
    title: 'کاربران و دسترسی',
    description: 'کاربران، تیم‌ها، نقش‌ها، PAT، VAT و چرخه عمر کلیدها.',
    icon: Users,
    links: [
      {label: 'نمای کلی دسترسی', to: '/guides/users-access-overview'},
      {label: 'توکن‌ها و کلیدها', to: '/guides/api-keys'},
      {label: 'کاربران مجازی', to: '/guides/virtual-account-management'},
      {label: 'نقش‌ها و مجوزها', to: '/guides/roles-and-permissions'},
    ],
  },
  {
    title: 'کنترل مصرف',
    description: 'حافظه نهان، محدودیت نرخ درخواست و محدودیت بودجه.',
    icon: CreditCard,
    links: [
      {label: 'حافظه نهان', to: '/guides/caching'},
      {label: 'محدودیت نرخ درخواست', to: '/guides/rate-limiting'},
      {label: 'محدودیت بودجه', to: '/guides/budget-limiting'},
    ],
  },
  {
    title: 'مالی',
    description: 'کیف پول، BYOK، قیمت‌گذاری سفارشی، تسویه، فاکتور و رهگیری هزینه.',
    icon: CreditCard,
    links: [
      {label: 'مالی، کیف پول و صورتحساب', to: '/financial/overview'},
      {label: 'رهگیری هزینه', to: '/cost-tracking'},
      {label: 'سیاست BYOK', to: '/guides/byok'},
      {label: 'مقایسه مدل‌ها و هزینه', to: '/provider-models-compare'},
    ],
  },
  {
    title: 'گاردریل‌ها',
    description: 'کنترل محتوای ورودی و خروجی، داده حساس و اعتبارسنجی قالب.',
    icon: ShieldCheck,
    links: [
      {label: 'نمای کلی گاردریل‌ها', to: '/guardrails'},
      {label: 'قابلیت‌های گاردریل', to: '/guardrails-capabilities'},
      {label: 'تشخیص داده حساس', to: '/guardrails-secrets-detection'},
      {label: 'اعتبارسنجی JSON Schema', to: '/guardrails-json-schema'},
    ],
  },
  {
    title: 'مشاهده‌پذیری و پایش',
    description: 'گزارش رخداد، رد درخواست، سنجه‌های مدل و OpenTelemetry.',
    icon: Activity,
    links: [
      {label: 'نمای کلی پایش', to: '/observability-overview'},
      {label: 'لاگ درخواست‌ها', to: '/observability-logs'},
      {label: 'ردیابی درخواست‌ها', to: '/observability-traces'},
      {label: 'متریک مدل‌ها', to: '/observability-model-metrics'},
    ],
  },
];

const helpItems: HelpItem[] = [
  {
    title: 'درخواست، خطای احراز هویت دارد',
    description:
      'توکن باید در خود گدارAI ساخته شده باشد. کلید خام ارائه‌دهنده مدل را به‌عنوان توکن گدارAI نفرستید.',
    icon: KeyRound,
    links: [
      {label: 'احراز هویت', to: '/authentication'},
      {label: 'توکن‌های دسترسی', to: '/guides/api-keys'},
    ],
  },
  {
    title: 'شناسه مدل یا مسیر API روشن نیست',
    description:
      'ابتدا نوع API را انتخاب کنید، سپس شناسه مدل را از مدل‌های مجاز یا مدل مجازی بردارید.',
    icon: Gauge,
    links: [
      {label: 'فهرست APIها', to: '/supported-apis'},
      {label: 'مدل مجازی', to: '/virtual-model'},
    ],
  },
  {
    title: 'می‌خواهید نتیجه را بررسی کنید',
    description:
      'بعد از ارسال درخواست، گزارش رخداد، رد درخواست و سنجه‌ها نشان می‌دهند چه اتفاقی افتاده است.',
    icon: CheckCircle2,
    links: [
      {label: 'لاگ درخواست‌ها', to: '/observability-logs'},
      {label: 'ردیابی درخواست‌ها', to: '/observability-traces'},
    ],
  },
];

export default function Home(): ReactNode {
  return (
    <Layout
      title="مستندات گدارAI"
      description="نقطه شروع مستندات گدارAI برای شناخت درگاه، ارسال اولین درخواست، انتخاب API و بررسی مسیرهای راهنما."
    >
      <main className={styles.page} dir="rtl">
        <section className={styles.hero} aria-labelledby="home-title">
          <div className={styles.heroText}>
            <p className={styles.eyebrow}>مستندات گدارAI</p>
            <div id="home-title" className={styles.heroIntro}>
              <p>
                <strong>گٌدارAI</strong>{' '}
                <span className={styles.inlineLatin} dir="ltr">
                  (GodarAI)
                </span>{' '}
                یک درگاه مدل‌های زبانی{' '}
                <span className={styles.inlineLatin} dir="ltr">
                  (LLM Gateway)
                </span>{' '}
                است که به‌عنوان لایه میانی بین اپلیکیشن‌های شما و
                ارائه‌دهندگان مدل‌های هوش مصنوعی عمل می‌کند.
              </p>
              <p>
                با استفاده از گدارAI می‌توانید تنها از طریق یک API سازگار با
                استاندارد OpenAI به صدها مدل و ارائه‌دهنده مختلف مانند OpenAI،
                Anthropic، Google Gemini، Azure OpenAI و AWS Bedrock متصل شوید،
                بدون آنکه وابستگی مستقیمی به هر ارائه‌دهنده داشته باشید.
              </p>
              <p>
                گدارAI امکاناتی مانند مدیریت کلیدهای API، مسیریابی هوشمند
                درخواست‌ها، پایش مصرف، کنترل هزینه‌ها، حافظه نهان پاسخ‌ها،
                سیاست‌های امنیتی و حاکمیتی، و مدیریت یکپارچه مدل‌ها را در
                اختیار تیم‌های توسعه و سازمان‌ها قرار می‌دهد.
              </p>
            </div>
            <div className={styles.heroActions}>
              <Link className="button button--primary button--lg" to="/quick-start">
                شروع سریع
              </Link>
              <Link className="button button--secondary button--lg" to="/supported-apis">
                فهرست APIها
              </Link>
              <Link className={styles.searchAction} to="/search">
                <Search aria-hidden="true" size={18} strokeWidth={2} />
                جستجو در مستندات
              </Link>
            </div>
          </div>

          <aside className={styles.startPanel} aria-label="مسیر کوتاه اولین درخواست">
            <div className={styles.panelHeader}>
              <span>مسیر کوتاه</span>
              <strong>اولین درخواست موفق</strong>
            </div>
            <ol className={styles.startList}>
              <li>یک حساب ارائه‌دهنده فعال یا مدل مجاز داشته باشید.</li>
              <li>برای تست PAT و برای سرویس‌ها VAT بسازید.</li>
              <li>شناسه مدل یا مدل مجازی را انتخاب کنید.</li>
              <li>درخواست را بفرستید و نتیجه را در داده‌های پایش بررسی کنید.</li>
            </ol>
            <Link className={styles.panelAction} to="/making-llm-requests-via-gateway">
              جزئیات ارسال درخواست
            </Link>
          </aside>
        </section>

        <section className={styles.entrySection} aria-labelledby="entry-title">
          <div className={styles.sectionHeader}>
            <p className={styles.eyebrow}>شروع بر اساس هدف</p>
            <h2 id="entry-title">سه مسیر پرتکرار برای ورود به مستندات</h2>
            <p>
              اگر دقیق نمی‌دانید کدام صفحه را باید باز کنید، یکی از این مسیرها
              شما را به نزدیک‌ترین راهنمای عملی می‌رساند.
            </p>
          </div>

          <div className={styles.entryGrid}>
            {entryCards.map(item => {
              const Icon = item.icon;

              return (
                <Link key={item.title} className={styles.entryCard} to={item.to}>
                  <span className={styles.cardTop}>
                    <span className={styles.iconBox}>
                      <Icon aria-hidden="true" size={22} strokeWidth={2} />
                    </span>
                    <span className={styles.cardEyebrow}>{item.eyebrow}</span>
                  </span>
                  <h3>{item.title}</h3>
                  <p>{item.description}</p>
                  <span className={styles.cardAction}>{item.action}</span>
                </Link>
              );
            })}
          </div>
        </section>

        <section className={styles.essentialsSection} aria-labelledby="essentials-title">
          <div className={styles.sectionHeader}>
            <p className={styles.eyebrow}>قبل از نمونه‌کد</p>
            <h2 id="essentials-title">سه مقدار لازم برای اولین درخواست</h2>
            <p>
              بیشتر خطاهای شروع از همین سه مقدار می‌آیند. آن‌ها را از فضای کاری
              خودتان بردارید و در نمونه‌کد جایگزین کنید.
            </p>
          </div>

          <div className={styles.essentialsGrid}>
            {requestEssentials.map((item, index) => (
              <article key={item.title} className={styles.essentialItem}>
                <span className={styles.stepNumber}>{index + 1}</span>
                <div>
                  <h3>{item.title}</h3>
                  <p>{item.description}</p>
                  <code dir="ltr">{item.value}</code>
                  <Link className={styles.inlineLink} to={item.to}>
                    راهنمای مرتبط
                  </Link>
                </div>
              </article>
            ))}
          </div>
        </section>

        <section className={styles.mapSection} aria-labelledby="map-title">
          <div className={styles.sectionHeader}>
            <p className={styles.eyebrow}>نقشه مستندات</p>
            <h2 id="map-title">دسته‌بندی موضوعی</h2>
            <p>
              برای تسلط هرچه سریع‌تر بر گدارAI این دسته از موضوعات اهمیت
              ویژه‌ای دارند.
            </p>
          </div>

          <div className={styles.docMapGrid}>
            {docGroups.map(group => {
              const Icon = group.icon;

              return (
                <article key={group.title} className={styles.docGroup}>
                  <div className={styles.docGroupHeader}>
                    <span className={styles.iconBox}>
                      <Icon aria-hidden="true" size={21} strokeWidth={2} />
                    </span>
                    <div>
                      <h3>{group.title}</h3>
                      <p>{group.description}</p>
                    </div>
                  </div>
                  <ul className={styles.linkList}>
                    {group.links.map(link => (
                      <li key={link.to}>
                        <Link to={link.to}>{link.label}</Link>
                      </li>
                    ))}
                  </ul>
                </article>
              );
            })}
          </div>
        </section>

        <section className={styles.helpSection} aria-labelledby="help-title">
          <div className={styles.sectionHeader}>
            <p className={styles.eyebrow}>کمک سریع</p>
            <h2 id="help-title">به مشکلی خورده‌اید؟ احتمالاً مشکل شما یکی از این‌هاست.</h2>
          </div>

          <div className={styles.helpGrid}>
            {helpItems.map(item => {
              const Icon = item.icon;

              return (
                <article key={item.title} className={styles.helpItem}>
                  <span className={styles.iconBox}>
                    <Icon aria-hidden="true" size={21} strokeWidth={2} />
                  </span>
                  <h3>{item.title}</h3>
                  <p>{item.description}</p>
                  <div className={styles.helpLinks}>
                    {item.links.map(link => (
                      <Link key={link.to} to={link.to}>
                        {link.label}
                      </Link>
                    ))}
                  </div>
                </article>
              );
            })}
          </div>
        </section>
      </main>
    </Layout>
  );
}
