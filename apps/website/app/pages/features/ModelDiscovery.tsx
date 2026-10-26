import { Calendar, Database, Filter, Scale, Search } from 'lucide-react';
import { ClosingBand, ComparisonTable, FeatureRow, MockRows, PageHero, ProductStage, ProductWindow, Section } from '@/app/components/patterns';
import { sectionCopy, sectionHeading } from '@/app/components/sectionType';

// Introduces the model catalog (/catalog). Prices live only on /pricing and in the catalog itself; the old plan
// table with invented prices is gone.
export function ModelDiscovery() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="مدل مناسب را پیدا کنید. پیش از اینکه هزینه‌اش را بدهید."
        lead="بیش از ۱۲۰۰ مدل از ۲۱ ارائه‌دهنده را بر اساس کار، قابلیت و قیمت جست‌وجو و مقایسه کنید."
        primary={{ href: '/catalog', label: 'باز کردن کاتالوگ' }}
        secondary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        visual={
          <ProductStage layout="size-full p-6 sm:p-10">
            <ProductWindow title="کاتالوگ مدل‌ها" meta="نمونه" label="نمونه‌ای از جست‌وجو در کاتالوگ مدل‌ها با فیلتر قابلیت.">
              <MockRows
                rows={[
                  { icon: Search, title: 'جست‌وجو: خلاصه‌سازی فارسی', text: 'فیلتر: چت، پنجره‌ی متن بلند' },
                  { icon: Database, title: 'مدل‌های پیدا‌شده', text: 'از چند ارائه‌دهنده', badge: { label: 'فعال', tone: 'success' } },
                  { icon: Scale, title: 'مقایسه‌ی سه مدل', text: 'کنار هم' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="search"
        tone="panel"
        reverse
        title="جست‌وجو و فیلتر. نه حدس و گمان."
        copy="مدل‌ها را بر اساس ورودی و خروجی، قابلیت، پنجره‌ی متن و قیمت فیلتر کنید."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="فیلترها" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از فیلترهای کاتالوگ.">
              <MockRows
                rows={[
                  { icon: Filter, title: 'قابلیت', text: 'چت، تصویر، embedding' },
                  { icon: Filter, title: 'ارائه‌دهنده', text: '۲۱ ارائه‌دهنده' },
                  { icon: Filter, title: 'وضعیت', text: 'فعال یا منسوخ' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="compare" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>سه مدل، یک سناریو، کنار هم.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>نمونه‌ی نمای مقایسه برای سناریوی پشتیبانی مشتری. عددها نمونه‌اند؛ داده‌ی واقعی در کاتالوگ است.</p>
        </div>
        <ComparisonTable
          caption="نمونه‌ی مقایسه‌ی سه مدل برای سناریوی پشتیبانی مشتری"
          columns={['مدل الف', 'مدل ب', 'مدل ج']}
          rows={[
            { label: 'ورودی و خروجی', values: ['متن', 'متن و تصویر', 'متن'] },
            { label: 'پنجره‌ی متن', values: ['بلند', 'خیلی بلند', 'متوسط'] },
            { label: 'قیمت نسبی', values: ['کم', 'زیاد', 'متوسط'], note: 'قیمت دقیق در کاتالوگ' },
          ]}
        />
      </Section>

      <FeatureRow
        id="trust"
        tone="panel"
        title="داده‌ای که می‌شود به آن تکیه کرد."
        copy="منبع قیمت، تاریخ اثر، وضعیت مدل و ارائه‌دهنده و پیوند مستندات رسمی، کنار هر مدل."
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="جزئیات مدل" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از جزئیات یک مدل در کاتالوگ: منبع قیمت، تاریخ اثر و وضعیت.">
              <MockRows
                rows={[
                  { icon: Database, title: 'منبع قیمت', text: 'صفحه‌ی رسمی ارائه‌دهنده' },
                  { icon: Calendar, title: 'تاریخ اثر', text: 'آخرین به‌روزرسانی' },
                  { icon: Database, title: 'وضعیت', text: 'فعال', badge: { label: 'فعال', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <ClosingBand
        title="انتخاب در کاتالوگ."
        titleMuted="مصرف از درگاه."
        copy="مدل را در کاتالوگ پیدا کنید و از درگاه یا یک مدل مجازی به کار ببرید."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/features/ai-gateway', label: 'درگاه مدل‌های زبانی' }}
      />
    </main>
  );
}
