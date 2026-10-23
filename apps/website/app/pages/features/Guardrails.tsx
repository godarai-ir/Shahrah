import { FileSearch, Filter, ListChecks, ScrollText, ShieldCheck, Tags } from 'lucide-react';
import {
  ClosingBand,
  FAQ,
  FeatureRow,
  MockRows,
  PageHero,
  PillarCrossLinks,
  ProductStage,
  ProductWindow,
  Section,
  TileGrid,
} from '@/app/components/patterns';
import type { Tile } from '@/app/components/patterns';
import { sectionCopy, sectionHeading } from '@/app/components/sectionType';

// The three things the semantic guardrail does today (PRODUCT.md), shown as tiles that do not link anywhere.
const abilities: Tile[] = [
  { key: 'confidential', title: 'تشخیص داده‌ی محرمانه', text: 'قرارداد، اطلاعات مالی و اسناد داخلی را از محتوا می‌شناسد، نه از کلیدواژه.', icon: ShieldCheck },
  { key: 'topics', title: 'سیاست موضوعی هر تیم', text: 'هر تیم مشخص می‌کند درباره‌ی چه چیزهایی می‌شود با مدل حرف زد؛ ورودی و خروجی هر دو.', icon: Tags },
  { key: 'files', title: 'بررسی فایل‌ها', text: 'پیوست‌ها و فایل‌های بارگذاری‌شده هم از همان گاردریل رد می‌شوند.', icon: FileSearch },
];

export function Guardrails() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="داده‌ی محرمانه را از معنایش می‌شناسد. نه از کلمه‌هایش."
        lead="گاردریل معنایی پیش از رسیدن درخواست به مدل، و پیش از رسیدن پاسخ به کاربر، محتوا را بررسی می‌کند. هر تیم سیاست خودش را دارد."
        primary={{ href: '/demo', label: 'درخواست دمو' }}
        secondary={{ href: '/onboarding', label: 'شروع ۱۵ روز رایگان' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <ProductWindow title="گزارش رخداد" meta="نمونه" label="نمونه‌ای از گزارش رخداد: یک درخواست به‌دلیل اطلاعات محرمانه‌ی مالی رد شده و دو درخواست عبور کرده‌اند.">
              <MockRows
                rows={[
                  { icon: ScrollText, title: 'خلاصه‌ی قرارداد تأمین‌کننده', text: 'تیم حقوقی · ورودی', badge: { label: 'رد شد', tone: 'danger' } },
                  { icon: ScrollText, title: 'پیش‌نویس پاسخ تیکت', text: 'تیم پشتیبانی · خروجی', badge: { label: 'عبور', tone: 'success' } },
                  { icon: ScrollText, title: 'گزارش فروش هفته', text: 'تیم فروش · ورودی', badge: { label: 'عبور', tone: 'success' } },
                ]}
              />
              <p className="m-0 border-t border-site-line px-5 py-3 text-caption text-site-ink-soft">علت رد: اطلاعات محرمانه‌ی مالی</p>
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="meaning"
        tone="panel"
        reverse
        title="فیلتر کلیدواژه می‌خوابد. گاردریل معنایی نه."
        copy="جمله‌ای که هیچ واژه‌ی ممنوعی ندارد ولی درباره‌ی قرارداد محرمانه است، از فیلتر کلیدواژه رد می‌شود. گاردریل معنایی آن را از معنایش می‌گیرد."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="یک درخواست، دو بررسی" meta="نمونه" size="m-auto max-w-md" label="نمونه: یک درخواست که فیلتر کلیدواژه آن را عبور می‌دهد و گاردریل معنایی آن را می‌گیرد.">
              <p className="m-0 border-b border-site-line px-5 py-4 text-label text-site-ink">«شرایط پرداخت تأمین‌کننده‌ی جدید را با مبلغ هر قسط خلاصه کن.»</p>
              <MockRows
                rows={[
                  { icon: Filter, title: 'فیلتر کلیدواژه', text: 'هیچ واژه‌ی ممنوعی پیدا نشد', badge: { label: 'عبور', tone: 'neutral' } },
                  { icon: ShieldCheck, title: 'گاردریل معنایی', text: 'اطلاعات محرمانه‌ی قراردادی', badge: { label: 'گرفته شد', tone: 'danger' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="abilities" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>سه کار، هر روز.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>آنچه گاردریل معنایی همین امروز انجام می‌دهد.</p>
        </div>
        <TileGrid items={abilities} columns={3} />
      </Section>

      <FeatureRow
        id="rules"
        tone="panel"
        title="قاعده‌ی ساده؟ قاعده‌ی ساده."
        copy="برای قواعد قابل پیش‌بینی، عبارت مسدود، Regex و محدودیت ساختار خروجی هم کنار گاردریل معنایی هست."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="گاردریل‌های فضای کاری" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از فهرست گاردریل‌ها: معنایی، عبارت مسدود، Regex و ساختار خروجی.">
              <MockRows
                rows={[
                  { icon: ShieldCheck, title: 'معنایی: اطلاعات محرمانه', text: 'ورودی و خروجی' },
                  { icon: ListChecks, title: 'عبارت مسدود', text: 'ورودی' },
                  { icon: ListChecks, title: 'Regex: شماره‌ی قرارداد', text: 'خروجی' },
                  { icon: ListChecks, title: 'ساختار خروجی JSON', text: 'خروجی' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="logged"
        reverse
        title="هر تصمیم گاردریل ثبت می‌شود."
        copy="نتیجه‌ی هر بررسی در گزارش رخداد و رد درخواست دیده می‌شود: چه چیزی، کجا و چرا."
        link={{ href: '/features/observability', label: 'مشاهده‌پذیری' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="رد درخواست" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از رد یک درخواست: بررسی ورودی، فراخوانی مدل و بررسی خروجی.">
              <MockRows
                rows={[
                  { title: 'گاردریل ورودی', text: 'معنایی · بدون مورد', badge: { label: 'عبور', tone: 'success' } },
                  { title: 'فراخوانی مدل', text: 'مدل مجاز تیم' },
                  { title: 'گاردریل خروجی', text: 'ساختار JSON', badge: { label: 'عبور', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <PillarCrossLinks current="guard" />

      <FAQ
        title="پیش از تأیید، بپرسید."
        items={[
          { question: 'با متن فارسی کار می‌کند؟', answer: 'بله. گاردریل معنایی متن فارسی را از معنایش بررسی می‌کند.' },
          { question: 'سیاست را چه کسی تعریف می‌کند؟', answer: 'مدیر فضای کاری، برای هر تیم جدا: چه موضوع‌هایی مجاز است و چه داده‌ای نباید به مدل برسد.' },
          { question: 'اگر گاردریل اشتباه کند چه؟', answer: 'نتیجه‌ی هر بررسی در گزارش رخداد دیده می‌شود، پس می‌بینید کجا سخت‌گیر یا آسان‌گیر بوده و سیاست را تنظیم می‌کنید.' },
          { question: 'آیا با داده‌ی ما مدل آموزش می‌بینید؟', answer: 'نه. گدارAI با داده‌ی شما مدل آموزش نمی‌دهد.' },
        ]}
      />

      <ClosingBand
        title="سیاست تیم‌ها را امتحان کنید."
        titleMuted="روی داده‌ی خودتان."
        copy="در دمو، گاردریل معنایی را روی نمونه‌ی داده‌ی سازمانتان می‌بینید."
        primary={{ href: '/demo', label: 'درخواست دمو' }}
        secondary={{ href: '/onboarding', label: 'شروع ۱۵ روز رایگان' }}
      />
    </main>
  );
}
