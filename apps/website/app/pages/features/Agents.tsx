import { Bot, Brain, Cpu, FileText, FileUp, Lock, ScrollText, Share2, ShieldCheck, Wallet, Wrench } from 'lucide-react';
import {
  ClosingBand,
  FeatureRow,
  MockRows,
  PageHero,
  ProductStage,
  ProductWindow,
  Section,
  StepList,
  TileGrid,
} from '@/app/components/patterns';
import { sectionCopy, sectionHeading, textLink } from '@/app/components/sectionType';
import Link from 'next/link';

// Workplace assistants and knowledge. The former developer-agent content now lives in the #agents section of
// /features/ai-gateway. Automatic sync of sources such as Drive is not claimed (PRODUCT.md).
export function Agents() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="برای هر تیم، دستیار خودش. بدون یک خط کد."
        lead="دستورالعمل، مدل، دانش و ابزارها را انتخاب کنید. دستیار برای همه‌ی تیم آماده است."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <ProductWindow title="ساخت دستیار" meta="نمونه" label="نمونه‌ای از صفحه‌ی ساخت دستیار با دستورالعمل، مدل، دانش و ابزارها.">
              <MockRows
                rows={[
                  { icon: FileText, title: 'دستورالعمل', text: 'به سؤال‌های مشتری مؤدب و کوتاه جواب بده.' },
                  { icon: Cpu, title: 'مدل', text: 'یکی از مدل‌های مجاز تیم' },
                  { icon: Brain, title: 'دانش', text: 'راهنمای محصول، سیاست بازگشت وجه' },
                  { icon: Wrench, title: 'ابزارها', text: 'CRM از راه MCP' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="ready-made"
        tone="panel"
        reverse
        title="از دستیار آماده شروع کنید. بعد مال خودتان کنید."
        copy="با دستیارهای آماده شروع کنید و به‌مرور دستورالعمل، دانش و ابزارهایشان را برای کار خودتان تغییر دهید."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="دستیارهای آماده" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از فهرست دستیارهای آماده برای پشتیبانی، فروش و منابع انسانی.">
              <MockRows
                rows={[
                  { icon: Bot, title: 'پاسخ به تیکت', text: 'برای تیم پشتیبانی' },
                  { icon: Bot, title: 'خلاصه‌ی پیش از جلسه', text: 'برای تیم فروش' },
                  { icon: Bot, title: 'پرسش‌های آیین‌نامه', text: 'برای منابع انسانی' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="build" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>چهار قدم تا دستیار خودتان.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>بدون کد و بدون تیم فنی.</p>
        </div>
        <StepList
          steps={[
            { title: 'دستورالعمل', text: 'بگویید دستیار چه کاری بکند و چطور حرف بزند.' },
            { title: 'مدل', text: 'از مدل‌های مجاز تیم یکی را انتخاب کنید.' },
            { title: 'دانش', text: 'فایل‌ها و پایگاه دانش تیم را به آن بدهید.' },
            { title: 'ابزار', text: 'به ابزارهای تیم از راه MCP وصلش کنید.' },
          ]}
        />
      </Section>

      <FeatureRow
        id="knowledge"
        tone="panel"
        title="دستیار از روی اسناد خودتان جواب می‌دهد."
        copy="فایل‌ها و پایگاه دانش تیم را به دستیار بدهید تا پاسخ‌هایش از روی اسناد سازمان باشد، نه از حدس."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="پایگاه دانش تیم پشتیبانی" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از پایگاه دانش یک تیم با سه سند بارگذاری‌شده.">
              <MockRows
                rows={[
                  { icon: FileUp, title: 'راهنمای محصول', text: 'سند بارگذاری‌شده' },
                  { icon: FileUp, title: 'سیاست بازگشت وجه', text: 'سند بارگذاری‌شده' },
                  { icon: FileUp, title: 'پرسش‌های پرتکرار', text: 'سند بارگذاری‌شده' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="tools"
        reverse
        title="دستیار فقط حرف نمی‌زند. کار هم می‌کند."
        copy="از راه MCP به Jira، Notion، Git، CRM و ابزارهای خودتان وصل می‌شود و در پیام‌رسان‌ها مانند تلگرام، واتساپ، بله و … پیام می‌خواند و می‌فرستد."
        link={{ href: '/integrations', label: 'اتصال‌ها' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="ابزارهای دستیار" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از ابزارهای وصل‌شده به یک دستیار.">
              <MockRows
                rows={[
                  { icon: Wrench, title: 'CRM', text: 'از راه MCP', badge: { label: 'وصل', tone: 'success' } },
                  { icon: Wrench, title: 'Jira', text: 'از راه MCP', badge: { label: 'وصل', tone: 'success' } },
                  { icon: Wrench, title: 'پیام‌رسان‌ها', text: 'تلگرام، واتساپ، بله و …', badge: { label: 'وصل', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="sharing"
        tone="panel"
        title="با تیم به اشتراک بگذارید. نه با همه."
        copy="فقط کسانی که دسترسی دارند از دستیار استفاده می‌کنند، و ساخت دستیار را هم می‌توانید به تیم یا نقش خاصی محدود کنید."
        link={{ href: '/features/governance', label: 'حاکمیت' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="اشتراک دستیار پشتیبانی" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از تنظیم اشتراک یک دستیار با یک تیم.">
              <MockRows
                rows={[
                  { icon: Share2, title: 'تیم پشتیبانی', text: 'استفاده', badge: { label: 'مجاز', tone: 'success' } },
                  { icon: Lock, title: 'سایر تیم‌ها', text: 'استفاده', badge: { label: 'بسته', tone: 'neutral' } },
                  { icon: Lock, title: 'ویرایش دستیار', text: 'فقط سرپرست تیم', badge: { label: 'محدود', tone: 'neutral' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="controls" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>هر دستیار با کنترلش می‌آید.</h2>
        </div>
        <TileGrid
          columns={3}
          size="compact"
          items={[
            { key: 'guard', icon: ShieldCheck, title: 'گاردریل روی گفتگوها', text: 'داده‌ی محرمانه پیش از رسیدن به مدل گرفته می‌شود.', href: '/features/guardrails' },
            { key: 'cost', icon: Wallet, title: 'سقف مصرف تیم', text: 'هزینه‌ی دستیار زیر سقف تیم می‌ماند.', href: '/features/cost-control' },
            { key: 'log', icon: ScrollText, title: 'ثبت هر درخواست', text: 'هر گفتگو در گزارش رخداد دیده می‌شود.', href: '/features/observability' },
          ]}
        />
        <p className="m-0 text-label text-site-ink-soft">
          ایجنت خودتان را می‌سازید؟{' '}
          <Link href="/features/ai-gateway#agents" className={textLink}>
            از درگاه مدل‌های زبانی استفاده کنید
          </Link>
        </p>
      </Section>

      <ClosingBand
        title="اولین دستیار تیمتان را بسازید."
        titleMuted="امروز، نه فصل بعد."
        copy="با اعتبار هدیه وارد پنل شوید؛ تاریخ انقضا ندارد."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
      />
    </main>
  );
}
