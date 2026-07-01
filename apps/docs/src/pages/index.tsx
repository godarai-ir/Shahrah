import type {ReactNode, SVGProps} from 'react';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './index.module.css';

type CatalogItem = {
  title: string;
  description: string;
  to: string;
  meta: string;
  icon: (props: SVGProps<SVGSVGElement>) => ReactNode;
};

function BookIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" />
      <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2Z" />
    </svg>
  );
}

function CompassIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <circle cx="12" cy="12" r="9" />
      <path d="m16.2 7.8-2.8 8.4-8.4 2.8 2.8-8.4 8.4-2.8Z" />
    </svg>
  );
}

function LayersIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="m12 3 9 4.5-9 4.5-9-4.5L12 3Z" />
      <path d="m3 12 9 4.5 9-4.5" />
      <path d="m3 16.5 9 4.5 9-4.5" />
    </svg>
  );
}

function KeyIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <circle cx="7.5" cy="15.5" r="3.5" />
      <path d="M11 13h10" />
      <path d="M18 10v6" />
      <path d="M15 12v4" />
    </svg>
  );
}

function ShieldIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M12 3 5 6v6c0 5 3.4 8.4 7 9 3.6-.6 7-4 7-9V6l-7-3Z" />
      <path d="m9.5 12 1.7 1.7 3.3-3.4" />
    </svg>
  );
}

function ChartIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" {...props}>
      <path d="M3 3v18h18" />
      <path d="m7 14 3-3 3 2 4-5" />
      <circle cx="7" cy="14" r="1" />
      <circle cx="10" cy="11" r="1" />
      <circle cx="13" cy="13" r="1" />
      <circle cx="17" cy="8" r="1" />
    </svg>
  );
}

const catalogItems: CatalogItem[] = [
  {
    title: 'شروع کار',
    description:
      'اگر تازه وارد GodarAI شده‌اید، از مسیر نصب، احراز هویت و اولین درخواست شروع کنید.',
    to: '/quick-start',
    meta: 'Quick start, auth, first request',
    icon: BookIcon,
  },
  {
    title: 'مفاهیم و معماری',
    description:
      'مدل ذهنی گیت‌وی، مدل مجازی، رهگیری هزینه و مشاهده‌پذیری را یکجا ببینید.',
    to: '/feature-overview',
    meta: 'Gateway, virtual model, cost, observability',
    icon: CompassIcon,
  },
  {
    title: 'APIهای پشتیبانی‌شده',
    description:
      'از Chat Completions تا Embeddings، Image، Audio و Moderation را بر اساس use case پیدا کنید.',
    to: '/supported-apis',
    meta: 'Text, image, audio, retrieval',
    icon: LayersIcon,
  },
  {
    title: 'مدیریت دسترسی و هویت',
    description:
      'PAT، VAT، مدیریت کاربران، تیم‌ها، نقش‌ها و چرخش امن کلیدها را مرور کنید.',
    to: '/guides/api-keys',
    meta: 'PAT, VAT, users, teams, roles',
    icon: KeyIcon,
  },
  {
    title: 'سیاست‌ها و عملیات',
    description:
      'Guardrails، caching، rate limiting، budget limiting و کنترل داده حساس را فعال کنید.',
    to: '/guardrails',
    meta: 'Policies, safety, cost control',
    icon: ShieldIcon,
  },
  {
    title: 'مشاهده‌پذیری و تحلیل',
    description:
      'لاگ‌ها، export، OTel و تحلیل مصرف را برای debugging و تصمیم‌گیری عملیاتی دنبال کنید.',
    to: '/observability-logs',
    meta: 'Logs, exports, OTel, analytics',
    icon: ChartIcon,
  },
];

const readingPaths = [
  {
    title: 'برای اولین integration',
    steps: 'شروع سریع ← احراز هویت ← اولین درخواست ← Chat Completions',
    to: '/quick-start',
  },
  {
    title: 'برای طراحی معماری',
    steps: 'نمای کلی قابلیت‌ها ← معرفی GodarAI ← مدل مجازی ← رهگیری هزینه',
    to: '/feature-overview',
  },
  {
    title: 'برای production و governance',
    steps: 'API Keyها ← چرخش کلید ← Guardrails ← Logs و OTel',
    to: '/guides/api-keys',
  },
];

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();

  return (
    <Layout title={siteConfig.title} description="کاتالوگ مستندات پلتفرم گدارAI">
      <main className={styles.page} dir="rtl">
        <section className={styles.hero}>
          <div className={styles.heroBackdrop} />
          <div className={styles.heroContent}>
            <div className={styles.heroBadge}>کاتالوگ مستندات GodarAI</div>
            <div className={styles.heroGrid}>
              <div>
                <h1 className={styles.heroTitle}>
                  سریع‌تر شاخه درست را پیدا کنید و با اطمینان ادامه بدهید
                </h1>
                <p className={styles.heroDescription}>
                  این صفحه، نقشه ورود به مستندات گدارAI است. به‌جای جست‌وجوی
                  پراکنده، از اینجا می‌توانید مسیر مناسب خود را بر اساس هدف
                  مطالعه، نوع integration و سطح درگیری با محصول انتخاب کنید.
                </p>
                <div className={styles.heroActions}>
                  <Link className="button button--primary button--lg" to="/intro">
                    ورود به مستندات
                  </Link>
                  <Link className={styles.secondaryCta} to="/supported-apis">
                    دیدن شاخه اصلی APIها
                  </Link>
                </div>
              </div>

              <aside className={styles.highlightPanel}>
                <h2>از کجا شروع کنم؟</h2>
                <ul className={styles.highlightList}>
                  <li>اگر تازه‌کار هستید: از «شروع کار» شروع کنید.</li>
                  <li>اگر دنبال endpoint مناسب هستید: وارد «APIهای پشتیبانی‌شده» شوید.</li>
                  <li>اگر روی production کار می‌کنید: «هویت»، «سیاست‌ها» و «Observability» را ببینید.</li>
                </ul>
                <Link className={styles.panelLink} to="/quick-start">
                  شروع مطالعه پیشنهادی
                </Link>
              </aside>
            </div>
          </div>
        </section>

        <section className={styles.catalogSection}>
          <div className={styles.sectionHeader}>
            <p className={styles.sectionEyebrow}>شاخه‌های اصلی</p>
            <h2>محتوا بر اساس نیاز واقعی شما دسته‌بندی شده است</h2>
            <p>
              هر کارت شما را به یکی از شاخه‌های اصلی docs می‌برد تا سریع‌تر به
              مستند مرتبط برسید.
            </p>
          </div>

          <div className={styles.catalogGrid}>
            {catalogItems.map(item => {
              const Icon = item.icon;
              return (
                <Link key={item.title} className={styles.catalogCard} to={item.to}>
                  <span className={styles.catalogIcon}>
                    <Icon width={24} height={24} />
                  </span>
                  <span className={styles.catalogMeta}>{item.meta}</span>
                  <h3>{item.title}</h3>
                  <p>{item.description}</p>
                  <span className={styles.catalogLink}>ورود به این شاخه</span>
                </Link>
              );
            })}
          </div>
        </section>

        <section className={styles.readingSection}>
          <div className={styles.sectionHeader}>
            <p className={styles.sectionEyebrow}>مسیرهای پیشنهادی</p>
            <h2>اگر هدفتان مشخص است، این مسیرها زمان شما را کم می‌کنند</h2>
          </div>

          <div className={styles.pathGrid}>
            {readingPaths.map(path => (
              <div key={path.title} className={styles.pathCard}>
                <h3>{path.title}</h3>
                <p>{path.steps}</p>
                <Link className={styles.inlineAction} to={path.to}>
                  شروع این مسیر
                </Link>
              </div>
            ))}
          </div>
        </section>

        <section className={styles.ctaSection}>
          <div className={styles.ctaCard}>
            <div>
              <p className={styles.sectionEyebrow}>گام بعدی</p>
              <h2>اگر فقط یک CTA بخواهیم، همین است: وارد شاخه اصلی شوید</h2>
              <p>
                برای بیشتر کاربران، بهترین نقطه ورود یکی از این دو مسیر است:
                «شروع کار» برای اولین integration یا «APIهای پشتیبانی‌شده» برای
                پیدا کردن endpoint مناسب.
              </p>
            </div>
            <div className={styles.ctaActions}>
              <Link className="button button--primary button--lg" to="/quick-start">
                شروع کار
              </Link>
              <Link className="button button--secondary button--lg" to="/supported-apis">
                رفتن به APIها
              </Link>
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
