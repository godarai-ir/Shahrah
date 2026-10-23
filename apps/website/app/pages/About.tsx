import { Heart, Lightbulb, Target, Users } from 'lucide-react';
import { ClosingBand, PageHero, Section, TileGrid } from '@/app/components/patterns';
import { sectionCopy, sectionHeading } from '@/app/components/sectionType';

const values = [
  { key: 'goal', icon: Target, title: 'هدف‌محور', text: 'تمرکز ما روی مشکل‌های واقعی تیم‌هاست.' },
  { key: 'user', icon: Users, title: 'کاربرمحور', text: 'تجربه‌ی ساده و روان، پیش از هر چیز.' },
  { key: 'precise', icon: Lightbulb, title: 'دقیق', text: 'ادعا، هزینه و محدودیت را شفاف می‌گوییم.' },
  { key: 'responsible', icon: Heart, title: 'مسئولیت‌پذیر', text: 'سرویسی پایدار، قابل بررسی و قابل پشتیبانی.' },
];

export function About() {
  return (
    <main className="bg-site-ground">
      <PageHero
        title="درباره‌ی گدارAI"
        lead="ما مسیر استفاده‌ی سازمان‌های ایرانی از مدل‌های زبانی را ساده‌تر، قابل‌پیگیری‌تر و قابل‌کنترل‌تر می‌کنیم."
        primary={{ href: '/demo', label: 'رزرو دمو' }}
        secondary={{ href: '/contact', label: 'تماس با ما' }}
      />

      <Section id="mission" tone="panel" layout="grid gap-12 md:grid-cols-2 lg:gap-16">
        <div className="flex flex-col gap-4">
          <h2 className={sectionHeading}>ماموریت ما</h2>
          <p className={`m-0 ${sectionCopy}`}>
            گدارAI برای حل یکی از چالش‌های جدی تیم‌های ایرانی شکل گرفت: اتصال پایدارتر به مدل‌های زبانی، پرداخت تومانی، کنترل هزینه و پایش مصرف در یک مسیر.
          </p>
          <p className={`m-0 ${sectionCopy}`}>
            نمی‌خواهیم محدودیت پرداخت، پراکندگی کلیدها و نبود مشاهده‌پذیری مانع استفاده از هوش مصنوعی شود.
          </p>
        </div>
        <div className="flex flex-col gap-4">
          <h2 className={sectionHeading}>چشم‌انداز ما</h2>
          <p className={`m-0 ${sectionCopy}`}>
            می‌خواهیم هر تیم بتواند مدل مناسبش را انتخاب کند، مصرف را ببیند، هزینه را کنترل کند و درباره‌ی مسیر هر درخواست پاسخ روشن داشته باشد.
          </p>
          <p className={`m-0 ${sectionCopy}`}>گدارAI قرار است لایه‌ای قابل اتکا برای استفاده از مدل‌های زبانی در بازار ایران باشد.</p>
        </div>
      </Section>

      <Section id="name" width="narrow" layout="flex flex-col gap-4">
        <h2 className={sectionHeading}>چرا «گدار»؟</h2>
        <p className={`m-0 ${sectionCopy}`}>گدار جای امن عبور از رودخانه است؛ بی‌گدار به آب نزنید.</p>
      </Section>

      <Section id="values" tone="panel" layout="flex flex-col gap-8">
        <h2 className={`max-w-3xl ${sectionHeading}`}>چیزهایی که کوتاه نمی‌آییم.</h2>
        <TileGrid items={values} size="compact" />
      </Section>

      <ClosingBand
        title="گدارAI را از نزدیک ببینید."
        titleMuted="در یک جلسه‌ی کوتاه."
        copy="جلسه‌ای برای دیدن محصول روی سناریوهای سازمان شما."
        primary={{ href: '/demo', label: 'درخواست دمو' }}
        secondary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
      />
    </main>
  );
}
