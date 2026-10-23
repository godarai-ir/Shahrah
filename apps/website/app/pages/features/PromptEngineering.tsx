import { GitCommit, Play, TerminalSquare, Variable } from 'lucide-react';
import { ClosingBand, FAQ, FeatureRow, MockRows, PageHero, ProductStage, ProductWindow } from '@/app/components/patterns';

export function PromptEngineering() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="پرامپت‌ها را مثل کد نگه دارید. نه در یادداشت‌های پراکنده."
        lead="پرامپت را بسازید، نسخه بزنید و با یک شناسه‌ی پایدار از درگاه اجرا کنید."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/docs', label: 'مستندات' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <ProductWindow title="پرامپت: support-reply" meta="نمونه" label="نمونه‌ای از تاریخچه‌ی نسخه‌های یک پرامپت.">
              <MockRows
                rows={[
                  { icon: GitCommit, title: 'نسخه‌ی ۳', text: 'لحن کوتاه‌تر', badge: { label: 'فعال', tone: 'success' } },
                  { icon: GitCommit, title: 'نسخه‌ی ۲', text: 'افزودن متغیر نام مشتری' },
                  { icon: GitCommit, title: 'نسخه‌ی ۱', text: 'نسخه‌ی اول' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="build"
        tone="panel"
        reverse
        title="یک قالب، چند متغیر."
        copy="پرامپت را با متغیرها بسازید تا یک قالب برای هزار درخواست کار کند."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="متغیرها" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از متغیرهای یک قالب پرامپت.">
              <MockRows
                rows={[
                  { icon: Variable, title: 'customer_name', text: 'نام مشتری' },
                  { icon: Variable, title: 'order_id', text: 'شماره‌ی سفارش' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="versions"
        title="هر تغییر، یک نسخه. هر نسخه، قابل برگشت."
        copy="تغییرهای پرامپت نسخه‌بندی می‌شوند؛ می‌بینید چه کسی چه چیزی را عوض کرد و نسخه‌ی قبلی همیشه در دسترس است."
        link={{ href: '/features/governance', label: 'حاکمیت' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="تاریخچه" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از تاریخچه‌ی تغییر یک پرامپت.">
              <MockRows
                rows={[
                  { icon: GitCommit, title: 'نسخه‌ی ۳ منتشر شد', text: 'سرپرست تیم پشتیبانی' },
                  { icon: GitCommit, title: 'نسخه‌ی ۲ منتشر شد', text: 'سرپرست تیم پشتیبانی' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="run"
        tone="panel"
        reverse
        title="اجرا از همان درگاه."
        copy="پرامپت را با شناسه‌اش از درگاه مدل‌های زبانی اجرا کنید؛ مسیریابی، حافظه‌ی نهان و گزارش همان‌جاست."
        link={{ href: '/features/ai-gateway', label: 'درگاه مدل‌های زبانی' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="اجرای پرامپت" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از اجرای یک پرامپت از درگاه.">
              <MockRows
                rows={[
                  { icon: TerminalSquare, title: 'support-reply', text: 'نسخه‌ی فعال' },
                  { icon: Play, title: 'اجرا', text: 'از درگاه', badge: { label: '۲۰۰', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FAQ
        title="سؤال‌های تیم محصول."
        items={[
          { question: 'رجیستری پرامپت در کدام پلن‌هاست؟', answer: 'در همه‌ی پلن‌ها، با تعداد قالب متفاوت؛ جزئیات در صفحه‌ی قیمت.', link: { href: '/pricing', label: 'قیمت‌گذاری' } },
          { question: 'نسخه‌های قبلی کجا می‌مانند؟', answer: 'در تاریخچه‌ی همان پرامپت، همیشه در دسترس.' },
        ]}
      />

      <ClosingBand
        title="پرامپت را از کد بیرون بکشید."
        titleMuted="و نسخه بزنید."
        copy="با اعتبار هدیه وارد پنل شوید؛ تاریخ انقضا ندارد."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/docs', label: 'مستندات' }}
      />
    </main>
  );
}
