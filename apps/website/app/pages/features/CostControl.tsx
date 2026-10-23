import { Banknote, Database, Route, Wallet } from 'lucide-react';
import {
  ClosingBand,
  FAQ,
  FeatureRow,
  MockBars,
  MockRows,
  PageHero,
  PillarCrossLinks,
  ProductStage,
  ProductWindow,
  Section,
  TileGrid,
} from '@/app/components/patterns';
import { sectionCopy, sectionHeading } from '@/app/components/sectionType';

// Blocking at a cap is configurable per budget rule (BlockOnBudgetExceed), so the copy offers it rather than promising it.
// Never state a credit-to-dollar ratio or a savings figure here (PRODUCT.md).
export function CostControl() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="هزینه‌ی هوش مصنوعی، زیر سقفی که خودتان گذاشته‌اید."
        lead="برای هر تیم و هر کلید بودجه بگذارید. کار ساده به مدل ارزان‌تر می‌رود و پرسش تکراری دوباره هزینه نمی‌دهد."
        primary={{ href: '/demo', label: 'درخواست دمو' }}
        secondary={{ href: '/pricing', label: 'دیدن قیمت‌ها' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <ProductWindow title="بودجه‌ی تیم‌ها" meta="نمونه" label="نمونه‌ای از بودجه‌ی سه تیم؛ تیم فنی به سقف رسیده و مصرفش متوقف شده است.">
              <MockBars
                bars={[
                  { label: 'تیم پشتیبانی', percent: 64, caption: '۶۴٪ بودجه' },
                  { label: 'تیم فروش', percent: 38, caption: '۳۸٪ بودجه' },
                  { label: 'تیم فنی', percent: 100, caption: 'به سقف رسید، متوقف شد', over: true },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="two-jobs" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>دو کار: سقف بگذارید. کمتر خرج کنید.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>کنترل جلوی هزینه‌ی ناگهانی را می‌گیرد؛ کاهش، همان کار را ارزان‌تر انجام می‌دهد.</p>
        </div>
        <TileGrid
          columns={2}
          items={[
            { key: 'control', icon: Wallet, title: 'کنترل', text: 'بودجه و سقف برای هر تیم و هر کلید، با محدودیت نرخ درخواست.' },
            { key: 'reduce', icon: Route, title: 'کاهش', text: 'مسیریابی کار ساده به مدل سبک‌تر و پاسخ پرسش تکراری از حافظه‌ی نهان.' },
          ]}
        />
      </Section>

      <FeatureRow
        id="caps"
        reverse
        title="بودجه بگذارید. غافلگیری نه."
        copy="بودجه‌ی هر تیم و هر کلید را برای هر دوره تعیین کنید. می‌توانید بگویید با رسیدن به سقف، مصرف متوقف شود تا هزینه‌ی ناگهانی پیش نیاید."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="قاعده‌ی بودجه" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از یک قاعده‌ی بودجه برای تیم فروش، با توقف مصرف در سقف.">
              <MockRows
                rows={[
                  { icon: Wallet, title: 'دامنه', text: 'تیم فروش' },
                  { icon: Wallet, title: 'دوره', text: 'ماهانه' },
                  { icon: Wallet, title: 'با رسیدن به سقف', text: 'مصرف متوقف شود', badge: { label: 'روشن', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="routing"
        tone="panel"
        title="کار ساده، مدل سبک. کار سخت، مدل قوی."
        copy="با مدل مجازی، هر کار به مدل مناسبش می‌رود. برنامه‌ها فقط یک نام پایدار می‌شناسند و شما پشت آن مدل را عوض می‌کنید."
        link={{ href: '/features/ai-gateway', label: 'درگاه مدل‌های زبانی' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="مدل مجازی: support-reply" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از یک مدل مجازی با دو مقصد و مسیر جایگزین.">
              <MockRows
                rows={[
                  { icon: Route, title: 'مقصد اول', text: 'مدل سبک', badge: { label: 'اولویت ۱', tone: 'neutral' } },
                  { icon: Route, title: 'مسیر جایگزین', text: 'مدل قوی، هنگام خطا', badge: { label: 'اولویت ۲', tone: 'neutral' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="cache"
        reverse
        title="یک بار پرسیدند. دوباره پول ندهید."
        copy="پاسخ پرسش‌های تکراری و هم‌معنا از حافظه‌ی نهان می‌آید و دوباره به مدل نمی‌رود. حافظه‌ی نهان هم‌معنا در پلن‌هایی است که در صفحه‌ی قیمت آمده."
        link={{ href: '/pricing', label: 'قیمت‌گذاری' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="حافظه‌ی نهان" meta="نمونه" size="m-auto max-w-md" label="نمونه: دو پرسش هم‌معنا که دومی از حافظه‌ی نهان پاسخ گرفته است.">
              <MockRows
                rows={[
                  { icon: Database, title: 'ساعت کاری پشتیبانی چیست؟', text: 'به مدل رفت' },
                  { icon: Database, title: 'پشتیبانی تا چه ساعتی جواب می‌دهد؟', text: 'هم‌معنای پرسش قبلی', badge: { label: 'از حافظه‌ی نهان', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="toman"
        tone="panel"
        title="به تومان بپردازید. بی‌دغدغه‌ی کارت ارزی."
        copy="کیف پول تومانی و کردیت پیش‌پرداخت؛ بدون کارت بین‌المللی. هزینه‌ی هر تیم و هر مدل را به تومان می‌بینید و فاکتور رسمی می‌گیرید."
        link={{ href: '/features/observability', label: 'گزارش هزینه' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="کیف پول" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از کیف پول تومانی با شارژ، مصرف و فاکتور رسمی.">
              <MockRows
                rows={[
                  { icon: Banknote, title: 'شارژ کیف پول', text: 'به تومان' },
                  { icon: Banknote, title: 'مصرف کردیت', text: 'به تفکیک تیم و مدل' },
                  { icon: Banknote, title: 'فاکتور رسمی', text: 'برای هر پرداخت', badge: { label: 'صادر شد', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <PillarCrossLinks current="cost" />

      <FAQ
        title="سؤال‌هایی که مدیر مالی می‌پرسد."
        items={[
          { question: 'اگر تیمی به سقف برسد چه می‌شود؟', answer: 'بسته به تنظیم شما: مصرف متوقف می‌شود، یا ادامه می‌یابد و در گزارش دیده می‌شود.' },
          { question: 'فاکتور رسمی می‌دهید؟', answer: 'بله.' },
          { question: 'قیمت مدل‌ها را از کجا ببینم؟', answer: 'قیمت هر مدل در کاتالوگ مدل‌ها آمده است.', link: { href: '/features/model-discovery', label: 'کشف و مقایسه‌ی مدل‌ها' } },
          { question: 'امتحان کردن هزینه دارد؟', answer: 'نه. با ثبت‌نام اعتبار هدیه می‌گیرید و این اعتبار تاریخ انقضا ندارد.' },
        ]}
      />

      <ClosingBand
        title="اول بودجه را بگذارید."
        titleMuted="بعد خیالتان راحت."
        copy="با اعتبار هدیه شروع کنید؛ سقف را از روز اول خودتان تعیین می‌کنید."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
      />
    </main>
  );
}
