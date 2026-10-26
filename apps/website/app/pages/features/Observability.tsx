import { Activity, Gauge, ScrollText, Timer } from 'lucide-react';
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
} from '@/app/components/patterns';

// Aimed at the technology lead and the finance lead first; the per-request detail for engineers sits lower down.
export function Observability() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="مصرف هر تیم را ببینید. و پشت هر عدد را."
        lead="داشبورد نشان می‌دهد کدام تیم با کدام مدل چقدر خرج کرده. از هر عدد می‌شود تا تک‌تک درخواست‌ها رفت."
        primary={{ href: '/demo', label: 'درخواست دمو' }}
        secondary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <ProductWindow title="مصرف این ماه" meta="نمونه" label="نمونه‌ای از داشبورد مصرف به تفکیک تیم: پشتیبانی، فروش، فنی و حقوقی، هر کدام نسبت به سقف خودش.">
              <MockBars
                bars={[
                  { label: 'تیم پشتیبانی', percent: 72, caption: '۷۲٪ سقف ماهانه' },
                  { label: 'تیم فروش', percent: 41, caption: '۴۱٪ سقف ماهانه' },
                  { label: 'تیم فنی', percent: 88, caption: '۸۸٪ سقف ماهانه' },
                  { label: 'تیم حقوقی', percent: 17, caption: '۱۷٪ سقف ماهانه' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="by-team"
        tone="panel"
        reverse
        title="گزارشی که مدیر مالی می‌خواهد. به تومان."
        copy="هزینه به تفکیک تیم و مدل، به تومان؛ بدون جمع زدن صورتحساب چند ارائه‌دهنده."
        link={{ href: '/features/cost-control', label: 'مدیریت هزینه' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="هزینه به تفکیک مدل" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از تقسیم هزینه‌ی یک تیم بین سه مدل.">
              <MockBars
                bars={[
                  { label: 'مدل سبک', percent: 55, caption: '۵۵٪ هزینه‌ی تیم' },
                  { label: 'مدل قوی', percent: 35, caption: '۳۵٪ هزینه‌ی تیم' },
                  { label: 'مدل تصویر', percent: 10, caption: '۱۰٪ هزینه‌ی تیم' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="logs"
        title="هر درخواست یک ردپا دارد."
        copy="هر درخواست با زمان، مدل، هزینه و نتیجه‌ی گاردریل در گزارش رخداد ثبت می‌شود."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="گزارش رخداد" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از گزارش رخداد با سه درخواست، مدل و نتیجه‌ی گاردریل هر کدام.">
              <MockRows
                rows={[
                  { icon: ScrollText, title: 'پیش‌نویس پاسخ تیکت', text: 'تیم پشتیبانی · مدل سبک', badge: { label: 'عبور', tone: 'success' } },
                  { icon: ScrollText, title: 'خلاصه‌ی قرارداد', text: 'تیم حقوقی · مدل قوی', badge: { label: 'رد شد', tone: 'danger' } },
                  { icon: ScrollText, title: 'گزارش فروش هفته', text: 'تیم فروش · مدل سبک', badge: { label: 'عبور', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="trace"
        tone="panel"
        reverse
        title="پاسخ عجیب بود؟ مسیرش را دنبال کنید."
        copy="رد درخواست نشان می‌دهد کدام مدل، کدام ابزار و کجا کند شد؛ نتیجه‌ی گاردریل هم در همان مسیر است."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="رد درخواست" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از رد یک درخواست: گاردریل ورودی، فراخوانی ابزار، فراخوانی مدل و گاردریل خروجی.">
              <MockRows
                rows={[
                  { icon: Activity, title: 'گاردریل ورودی', text: 'معنایی', badge: { label: 'عبور', tone: 'success' } },
                  { icon: Activity, title: 'فراخوانی ابزار', text: 'CRM از راه MCP', badge: { label: 'کند', tone: 'neutral' } },
                  { icon: Activity, title: 'فراخوانی مدل', text: 'مدل مجاز تیم' },
                  { icon: Activity, title: 'گاردریل خروجی', text: 'ساختار JSON', badge: { label: 'عبور', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="metrics"
        title="برای تیم فنی: سنجه‌ها."
        copy="تأخیر، خطا و مصرف توکن هر مدل و هر کلید، برای وقتی که باید بدانید مشکل از کجاست."
        link={{ href: '/dashboard', label: 'داشبورد' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="سنجه‌های مدل" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از سنجه‌های یک مدل: تأخیر، خطا و مصرف توکن.">
              <MockRows
                rows={[
                  { icon: Timer, title: 'تأخیر', text: 'میانه و صدک ۹۵' },
                  { icon: Gauge, title: 'خطا', text: 'به تفکیک کد پاسخ' },
                  { icon: Activity, title: 'مصرف توکن', text: 'ورودی و خروجی' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <PillarCrossLinks current="observe" />

      <FAQ
        title="پیش از تأیید، بپرسید."
        items={[
          { question: 'گزارش‌ها چه مدت نگه داشته می‌شوند؟', answer: 'مدتش را خودتان تنظیم می‌کنید.' },
          { question: 'به سامانه‌ی پایش خودمان وصل می‌شود؟', answer: 'با سرآیند x-godarai-trace-id هر درخواست را با رد درخواست در سامانه‌ی پایش خودتان هم‌بسته می‌کنید.' },
        ]}
      />

      <ClosingBand
        title="هزینه را قبل از صورتحساب ببینید."
        titleMuted="نه بعدش."
        copy="در دمو، داشبورد مصرف و گزارش رخداد را روی محیط واقعی می‌بینید."
        primary={{ href: '/demo', label: 'درخواست دمو' }}
        secondary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
      />
    </main>
  );
}
