import { Banknote, ShieldCheck, Users } from 'lucide-react';
import { ClosingBand, ComparisonTable, FAQ, FeatureRow, MockBars, MockRows, PageHero, ProductStage, ProductWindow, Section, StepList } from '@/app/components/patterns';
import { sectionCopy, sectionHeading } from '@/app/components/sectionType';

// Fair by design: ChatGPT is a good tool; the problem is scattered personal use inside an organisation.
// Every statement about ChatGPT here is about that scattered use, never about OpenAI's own enterprise product,
// and carries no statistic. No OpenAI logos.
export function GodarAIvsChatGPT() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="همکارانتان همین حالا از هوش مصنوعی استفاده می‌کنند. سؤال این است: با چه کنترلی؟"
        lead="ChatGPT ابزار خوبی است. وقتی هر کس با حساب و کارت خودش از آن استفاده کند، سازمان نمی‌داند چه داده‌ای کجا رفته و چقدر خرج شده."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
      />

      <FeatureRow
        id="leak"
        tone="panel"
        reverse
        title="ریسک اول: نشت داده. پاسخ: گاردریل معنایی."
        copy="در استفاده‌ی پراکنده، هیچ چیز جلوی چسباندن یک قرارداد محرمانه را نمی‌گیرد. در گدارAI، گاردریل معنایی داده‌ی محرمانه را پیش از رسیدن به مدل می‌گیرد."
        link={{ href: '/features/guardrails', label: 'گاردریل معنایی' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="گزارش رخداد" meta="نمونه" size="m-auto max-w-md" label="نمونه: درخواستی با اطلاعات محرمانه‌ی مالی که گاردریل آن را رد کرده است.">
              <MockRows rows={[{ icon: ShieldCheck, title: 'خلاصه‌ی قرارداد تأمین‌کننده', text: 'اطلاعات محرمانه‌ی مالی', badge: { label: 'رد شد', tone: 'danger' } }]} />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="access"
        title="ریسک دوم: دسترسی بی‌قاعده. پاسخ: دسترسی تیمی."
        copy="حساب شخصی یعنی هر کس هر چیزی را به هر مدلی می‌دهد. در گدارAI هر تیم فقط به مدل‌ها، دستیارها و ابزارهای خودش می‌رسد."
        link={{ href: '/features/governance', label: 'حاکمیت' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="دسترسی تیم‌ها" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از دسترسی دو تیم به ابزارها.">
              <MockRows
                rows={[
                  { icon: Users, title: 'تیم فروش', text: 'CRM و دستیار فروش', badge: { label: 'مجاز', tone: 'success' } },
                  { icon: Users, title: 'تیم پشتیبانی', text: 'اسناد حقوقی', badge: { label: 'بسته', tone: 'neutral' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="visibility"
        tone="panel"
        reverse
        title="ریسک سوم: هیچ‌کس نمی‌بیند. پاسخ: گزارش هر تیم."
        copy="استفاده‌ی پراکنده در هیچ گزارشی نیست. در گدارAI مصرف و هزینه‌ی هر تیم و رد هر درخواست را می‌بینید."
        link={{ href: '/features/observability', label: 'مشاهده‌پذیری' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="مصرف این ماه" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از مصرف دو تیم نسبت به سقف ماهانه.">
              <MockBars
                bars={[
                  { label: 'تیم پشتیبانی', percent: 72, caption: '۷۲٪ سقف ماهانه' },
                  { label: 'تیم فروش', percent: 41, caption: '۴۱٪ سقف ماهانه' },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <FeatureRow
        id="cost"
        title="ریسک چهارم: هزینه‌ی پراکنده و ارزی. پاسخ: سقف و تومان."
        copy="ده اشتراک شخصی با ده کارت ارزی، بدون سقف. در گدارAI برای هر تیم بودجه می‌گذارید و به تومان می‌پردازید."
        link={{ href: '/features/cost-control', label: 'مدیریت هزینه' }}
        visual={
          <ProductStage layout="w-full max-w-lg p-8 sm:p-12">
            <ProductWindow title="کیف پول" meta="نمونه" size="m-auto max-w-md" label="نمونه‌ای از کیف پول تومانی و بودجه‌ی تیم‌ها.">
              <MockRows
                rows={[
                  { icon: Banknote, title: 'کیف پول تومانی', text: 'کردیت پیش‌پرداخت' },
                  { icon: Banknote, title: 'بودجه‌ی تیم فروش', text: 'سقف ماهانه', badge: { label: 'فعال', tone: 'success' } },
                ]}
              />
            </ProductWindow>
          </ProductStage>
        }
      />

      <Section id="compare" tone="panel" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>کنار هم، منصفانه.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>ستون اول استفاده‌ی شخصی و پراکنده از ChatGPT در سازمان است، نه محصول سازمانی OpenAI.</p>
        </div>
        <ComparisonTable
          caption="مقایسه‌ی استفاده‌ی شخصی و پراکنده از ChatGPT با گدارAI در سازمان"
          columns={['استفاده‌ی شخصی و پراکنده از ChatGPT', 'گدارAI در سازمان']}
          highlight={1}
          rows={[
            { label: 'پرداخت', values: ['هر نفر با کارت ارزی خودش', 'کیف پول تومانی سازمان، با فاکتور رسمی'] },
            { label: 'دسترسی', values: ['حساب شخصی هر نفر', 'نقش و دسترسی هر تیم'] },
            { label: 'داده‌ی محرمانه', values: ['به عهده‌ی هر نفر', 'گاردریل معنایی روی ورودی، خروجی و فایل‌ها'] },
            { label: 'گزارش مصرف', values: ['در دسترس سازمان نیست', 'مصرف و هزینه‌ی هر تیم و رد هر درخواست'] },
            { label: 'مدل‌ها', values: ['مدل‌های OpenAI', 'مدل‌های OpenAI و بیش از ۱۲۰۰ مدل از ۲۱ ارائه‌دهنده'] },
            { label: 'استقرار', values: ['سرویس خارجی', 'ابر گدارAI یا سرورهای خود سازمان'] },
          ]}
        />
      </Section>

      <Section id="switch" layout="flex flex-col gap-8">
        <div className="flex flex-col gap-4">
          <h2 className={`max-w-3xl ${sectionHeading}`}>جابه‌جایی آرام. بدون شیرجه.</h2>
          <p className={`max-w-2xl ${sectionCopy}`}>همان مدل‌های OpenAI، در کنار مدل‌های ۲۰ ارائه‌دهنده‌ی دیگر، با کنترل سازمانی.</p>
        </div>
        <StepList
          steps={[
            { title: 'فضای کاری بسازید', text: 'همان لحظه آماده است.' },
            { title: 'تیم‌ها را دعوت کنید', text: 'هر تیم با دسترسی خودش.' },
            { title: 'سقف و سیاست بگذارید', text: 'بودجه و گاردریل از روز اول.' },
          ]}
        />
      </Section>

      <FAQ
        title="سؤال‌هایی که قبل از جابه‌جایی می‌پرسند."
        items={[
          { question: 'باید ChatGPT را ممنوع کنیم؟', answer: 'لازم نیست. یک جایگزین کنترل‌شده بدهید؛ همان مدل‌ها، با قاعده‌ی سازمان.' },
          { question: 'همکاران باید چیز تازه‌ای یاد بگیرند؟', answer: 'نه چندان: همان گفتگو با مدل‌هاست، این بار در فضای کاری سازمان و با دستیارهای تیم.' },
          { question: 'تاریخچه‌ی گفتگوهای قبلی منتقل می‌شود؟', answer: 'بله، تاریخچه‌ی گفتگوهای ChatGPT را می‌شود به گدارAI آورد.', link: { href: '/onboarding', label: 'شروع با اعتبار هدیه' } },
        ]}
      />

      <ClosingBand
        title="جایگزین سازمانی را امتحان کنید."
        titleMuted="با اعتبار هدیه."
        copy="اعتبار هدیه تاریخ انقضا ندارد؛ اول ببینید، بعد تصمیم بگیرید."
        primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
        secondary={{ href: '/demo', label: 'درخواست دمو' }}
      />
    </main>
  );
}
