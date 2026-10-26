import { Bot, Cpu, KeyRound, ScrollText, Server, UserMinus, Users, Wrench } from 'lucide-react';
import {
  ClosingBand,
  ControlChecklist,
  FAQ,
  FeatureRow,
  MockRows,
  PageHero,
  PillarCrossLinks,
  ProductStage,
  ProductWindow,
  Section,
} from '@/app/components/patterns';
import { sectionCopy, sectionHeading } from '@/app/components/sectionType';

// Every claim here is confirmed in docs/site-architecture/04-open-questions.md; SSO and user sync carry the
// On-Premise qualifier until user sync is confirmed for the cloud.
export function Governance() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="حاکمیت از روز اول. نه وصله‌ی بعدی."
        lead="برای هر تیم مشخص کنید به کدام مدل، دستیار و ابزار برسد. هر درخواست و هر تغییر تنظیمات ثبت می‌شود."
        primary={{ href: '/demo', label: 'درخواست دمو' }}
        secondary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <ProductWindow title="دسترسی تیم‌ها" meta="نمونه" label="نمونه‌ای از دسترسی تیم‌ها: تیم فروش به CRM و دستیار فروش، تیم حقوقی به اسناد حقوقی، تیم فنی به درگاه API.">
              <MockRows
                rows={[
                  { icon: Users, title: 'تیم فروش', text: 'دستیار فروش · CRM', badge: { label: 'مجاز', tone: 'success' } },
                  { icon: Users, title: 'تیم حقوقی', text: 'دستیار حقوقی · اسناد حقوقی', badge: { label: 'مجاز', tone: 'success' } },
                  { icon: Users, title: 'تیم پشتیبانی', text: 'اسناد حقوقی', badge: { label: 'بسته', tone: 'neutral' } },
                  { icon: Users, title: 'تیم فنی', text: 'درگاه API · سرورهای MCP', badge: { label: 'مجاز', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="access"
        tone="panel"
        reverse
        title="دسترسی تیم به تیم. نه نفر به نفر."
        copy="تیم فروش به CRM و دستیار فروش می‌رسد، تیم حقوقی به اسناد خودش. ساخت دستیار را هم می‌توانید به تیم یا نقش خاصی محدود کنید."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="تیم فروش" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از تنظیم دسترسی یک تیم به مدل، دستیار، ابزار و ساخت دستیار.">
              <MockRows
                rows={[
                  { icon: Cpu, title: 'مدل‌ها', text: 'سه مدل مجاز', badge: { label: 'مجاز', tone: 'success' } },
                  { icon: Bot, title: 'دستیارها', text: 'دستیار فروش', badge: { label: 'مجاز', tone: 'success' } },
                  { icon: Wrench, title: 'ابزارها', text: 'CRM از راه MCP', badge: { label: 'مجاز', tone: 'success' } },
                  { icon: Bot, title: 'ساخت دستیار', text: 'فقط سرپرست تیم', badge: { label: 'محدود', tone: 'neutral' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="audit"
        title="چه کسی، چه چیزی، کی. همه ثبت می‌شود."
        copy="هر درخواست و پاسخ با کاربر، تیم، مدل و نتیجه‌ی گاردریل ثبت می‌شود. تغییر تنظیمات هم ثبت می‌شود، و مدت نگهداری گزارش‌ها را خودتان تعیین می‌کنید."
        link={{ href: '/features/observability', label: 'مشاهده‌پذیری' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="گزارش تغییرات" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از گزارش تغییر تنظیمات: چه کسی چه تنظیمی را کی عوض کرد.">
              <MockRows
                rows={[
                  { icon: ScrollText, title: 'سقف مصرف تیم فروش تغییر کرد', text: 'مدیر فضای کاری · دیروز' },
                  { icon: ScrollText, title: 'دستیار حقوقی به تیم حقوقی داده شد', text: 'سرپرست تیم حقوقی · دیروز' },
                  { icon: ScrollText, title: 'یک مدل از فهرست مجاز حذف شد', text: 'مدیر فضای کاری · هفته‌ی پیش' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="directory"
        tone="panel"
        reverse
        title="کسی که می‌رود، دسترسی‌اش هم می‌رود."
        copy="در استقرار داخل سازمان، کاربران و تیم‌ها از فهرست سازمانتان همگام می‌شوند و کارکنان با حساب سازمانی وارد می‌شوند (SSO)."
        link={{ href: '/on-premise', label: 'استقرار در سازمان' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="همگام‌سازی کاربران" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از همگام‌سازی کاربران با فهرست سازمان: یک کاربر جدید اضافه و یک کاربر غیرفعال شده است.">
              <MockRows
                rows={[
                  { icon: Users, title: 'یک کاربر به تیم فروش اضافه شد', text: 'از فهرست سازمان', badge: { label: 'همگام', tone: 'success' } },
                  { icon: UserMinus, title: 'یک کاربر غیرفعال شد', text: 'دسترسی‌ها بسته شد', badge: { label: 'همگام', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="keys"
        title="کلید آدم با کلید سرویس فرق دارد."
        copy="توکن دسترسی شخصی برای افراد، کلید دسترسی مجازی برای سرویس‌ها؛ هر کدام با دسترسی و سقف خودش."
        link={{ href: '/features/security', label: 'امنیت و اعتماد' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="کلیدها" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از دو نوع کلید: توکن دسترسی شخصی و کلید دسترسی مجازی.">
              <MockRows
                rows={[
                  { icon: KeyRound, title: 'توکن دسترسی شخصی', text: 'برای آزمایش و کار شخصی', badge: { label: 'PAT', tone: 'neutral' } },
                  { icon: Server, title: 'کلید دسترسی مجازی', text: 'برای سرویس پشتیبانی', badge: { label: 'VAT', tone: 'neutral' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="checklist" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>فهرست حاکمیت، یک‌جا.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>همه‌ی آنچه پیش از تأیید باید بدانید، در یک نگاه.</p>
        </div>
        <ControlChecklist
          groups={[
            {
              title: 'دسترسی',
              surface: 'surface',
              items: [{ label: 'نقش و دسترسی هر تیم' }, { label: 'دسترسی به مدل، دستیار و ابزار' }, { label: 'محدود کردن ساخت دستیار به تیم یا نقش' }],
            },
            {
              title: 'ثبت',
              surface: 'sand',
              items: [{ label: 'ثبت هر درخواست و پاسخ' }, { label: 'ثبت هر تغییر تنظیمات' }, { label: 'مدت نگهداری قابل تنظیم' }],
            },
            {
              title: 'هویت',
              surface: 'surface',
              items: [{ label: 'کلید جدا برای انسان و سرویس' }, { label: 'ورود سازمانی و همگام‌سازی کاربران در استقرار داخل سازمان', href: '/on-premise' }],
            },
            {
              title: 'هزینه',
              surface: 'sand',
              items: [{ label: 'سقف مصرف برای هر تیم و هر کلید', href: '/features/cost-control' }],
            },
          ]}
        />
      </Section>

      <PillarCrossLinks current="govern" />

      <FAQ
        title="سؤال‌هایی که مدیر فناوری می‌پرسد."
        items={[
          { question: 'می‌شود ساخت دستیار را به تیم خاصی محدود کرد؟', answer: 'بله، به تیم یا نقش خاص.' },
          { question: 'کارمندی که می‌رود چه می‌شود؟', answer: 'در استقرار داخل سازمان، با همگام‌سازی فهرست کاربران دسترسی‌اش بسته می‌شود؛ توکن‌هایش را هم می‌شود باطل کرد.' },
          { question: 'گزارش‌ها چه مدت می‌مانند؟', answer: 'مدت نگهداری گزارش‌ها را خودتان تنظیم می‌کنید.' },
        ]}
      />

      <ClosingBand
        title="قاعده را شما بنویسید."
        titleMuted="اجرایش با ما."
        copy="در دمو، دسترسی تیم‌ها و گزارش‌ها را روی محیط واقعی می‌بینید."
        primary={{ href: '/demo', label: 'درخواست دمو' }}
        secondary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
      />
    </main>
  );
}
