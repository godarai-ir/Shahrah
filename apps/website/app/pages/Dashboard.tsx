import { Activity, Gauge, KeyRound, Timer } from 'lucide-react';
import { ClosingBand, FeatureRow, MockBars, MockRows, PageHero, ProductStage, ProductWindow } from '@/app/components/patterns';

// Developer-facing overview of the gateway dashboard; the manager's view of usage and cost is /features/observability.
export function Dashboard() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="سنجه‌های هر مدل، یک‌جا. نه در پنج داشبورد."
        lead="تأخیر، خطا و مصرف توکن هر مدل و هر کلید را ببینید."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/features/observability', label: 'مشاهده‌پذیری' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <ProductWindow title="سنجه‌ها" meta="نمونه" label="نمونه‌ای از سنجه‌های سه مدل نسبت به یکدیگر.">
              <MockBars
                bars={[
                  { label: 'مدل سبک', percent: 30, caption: 'تأخیر کم' },
                  { label: 'مدل قوی', percent: 70, caption: 'تأخیر بیشتر' },
                  { label: 'مدل تصویر', percent: 55, caption: 'تأخیر متوسط' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="latency"
        tone="panel"
        reverse
        title="کجا کند شد؟ عدد دارد."
        copy="تأخیر هر مدل و هر کلید، با میانه و صدک‌های بالا."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="تأخیر" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از سنجه‌ی تأخیر.">
              <MockRows rows={[{ icon: Timer, title: 'میانه', text: 'هر مدل' }, { icon: Timer, title: 'صدک ۹۵', text: 'هر مدل' }]} />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="errors"
        title="خطاها، به تفکیک."
        copy="خطاها به تفکیک کد پاسخ، مدل و کلید؛ تا بدانید مشکل از ارائه‌دهنده است یا از برنامه."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="خطاها" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از خطاها به تفکیک کد پاسخ.">
              <MockRows
                rows={[
                  { icon: Gauge, title: '۴۲۹', text: 'محدودیت نرخ', badge: { label: 'خطا', tone: 'danger' } },
                  { icon: Gauge, title: '۵۰۳', text: 'ارائه‌دهنده در دسترس نبود', badge: { label: 'خطا', tone: 'danger' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="tokens"
        tone="panel"
        reverse
        title="توکن‌ها، به تفکیک کلید."
        copy="مصرف توکن ورودی و خروجی هر کلید، برای وقتی که باید بدانید کدام سرویس پرمصرف است."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="مصرف توکن" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از مصرف توکن دو کلید.">
              <MockRows
                rows={[
                  { icon: KeyRound, title: 'سرویس پشتیبانی', text: 'ورودی و خروجی' },
                  { icon: Activity, title: 'اپ موبایل', text: 'ورودی و خروجی' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <ClosingBand
        title="داشبورد را روشن کنید."
        titleMuted="حدس را خاموش."
        copy="با اعتبار هدیه وارد پنل شوید؛ تاریخ انقضا ندارد."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/features/ai-gateway', label: 'درگاه مدل‌های زبانی' }}
      />
    </main>
  );
}
