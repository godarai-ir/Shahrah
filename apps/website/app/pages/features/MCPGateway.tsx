import { Activity, KeyRound, List, Lock, Server, ShieldCheck, Wrench } from 'lucide-react';
import { ClosingBand, FeatureRow, MockRows, PageHero, ProductStage, ProductWindow, Section, TileGrid } from '@/app/components/patterns';
import { sectionHeading } from '@/app/components/sectionType';

// The technical layer behind workplace integrations. Human approval of tool actions is not claimed (PRODUCT.md).
export function MCPGateway() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="یک مسیر برای همه‌ی سرورهای MCP. نه ده کلید پراکنده."
        lead="سرورها و ابزارها را یک‌جا ثبت کنید، دسترسی را تا سطح ابزار بدهید و هر فراخوانی را ببینید."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/integrations', label: 'اتصال‌ها' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <ProductWindow title="سرورهای MCP" meta="نمونه" label="نمونه‌ای از فهرست مرکزی سرورهای MCP و ابزارهایشان.">
              <MockRows
                rows={[
                  { icon: Server, title: 'CRM', text: 'سه ابزار', badge: { label: 'وصل', tone: 'success' } },
                  { icon: Server, title: 'Jira', text: 'چهار ابزار', badge: { label: 'وصل', tone: 'success' } },
                  { icon: Server, title: 'سامانه‌ی انبار (داخلی)', text: 'دو ابزار', badge: { label: 'وصل', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="controls" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>چهار کنترل پایه در مسیر MCP.</h2>
        </div>
        <TileGrid
          columns={4}
          size="compact"
          items={[
            { key: 'auth', icon: KeyRound, title: 'احراز هویت', text: 'متناسب با هر سرور.' },
            { key: 'registry', icon: List, title: 'فهرست مرکزی', text: 'همه‌ی سرورها و ابزارها یک‌جا.' },
            { key: 'access', icon: Lock, title: 'دسترسی', text: 'در سطح سرور و ابزار.' },
            { key: 'observe', icon: Activity, title: 'پایش', text: 'هر فراخوانی ابزار ثبت می‌شود.' },
          ]}
        />
      </Section>

      <FeatureRow
        id="credentials"
        reverse
        title="ابزار را به اشتراک بگذارید. کلیدش را نه."
        copy="اعتبارنامه‌ی هر سرور در گدارAI می‌ماند و به کاربر یا ایجنت داده نمی‌شود."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="اعتبارنامه‌ی CRM" meta="نمونه" size="m-auto max-w-md" label="نمونه: اعتبارنامه‌ی یک سرور MCP فقط در گدارAI نگه داشته می‌شود.">
              <MockRows rows={[{ icon: KeyRound, title: 'کلید CRM', text: 'فقط در گدارAI', badge: { label: 'پنهان', tone: 'neutral' } }]} />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="policy"
        tone="panel"
        title="سیاست، پیش از فراخوانی."
        copy="پیش از هر فراخوانی ابزار، دسترسی تیم و ابزار بررسی می‌شود و نتیجه در گزارش می‌ماند."
        link={{ href: '/features/governance', label: 'حاکمیت' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="فراخوانی‌ها" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از دو فراخوانی ابزار، یکی مجاز و یکی بسته.">
              <MockRows
                rows={[
                  { icon: Wrench, title: 'CRM: خواندن مشتری', text: 'تیم پشتیبانی', badge: { label: 'مجاز', tone: 'success' } },
                  { icon: ShieldCheck, title: 'CRM: ثبت کار بعدی', text: 'تیم پشتیبانی', badge: { label: 'بسته', tone: 'danger' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="usage"
        reverse
        title="کدام ابزار، چقدر، با چه تأخیری."
        copy="مصرف و عملکرد هر سرور و هر ابزار را ببینید تا بدانید کجا کند است و کجا پرکار."
        link={{ href: '/features/observability', label: 'مشاهده‌پذیری' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="سنجه‌های MCP" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از سنجه‌های سرورهای MCP.">
              <MockRows
                rows={[
                  { icon: Activity, title: 'CRM', text: 'پرکارترین سرور' },
                  { icon: Activity, title: 'سامانه‌ی انبار', text: 'کندترین پاسخ', badge: { label: 'کند', tone: 'neutral' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <ClosingBand
        title="سرورهای MCP را یک‌جا کنید."
        titleMuted="و آرام بخوابید."
        copy="با اعتبار هدیه وارد پنل شوید؛ تاریخ انقضا ندارد."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/integrations', label: 'اتصال‌ها' }}
      />
    </main>
  );
}
