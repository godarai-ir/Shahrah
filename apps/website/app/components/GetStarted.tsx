import { ArrowLeft, Check } from 'lucide-react';
import Link from 'next/link';
import { Section, StepList } from './patterns';
import { sectionHeading, textLink } from './sectionType';

const steps = [
  { title: 'فضای کاری بسازید', text: 'ثبت‌نام می‌کنید و فضای کاری سازمانتان همان لحظه آماده است.' },
  { title: 'یک دستیار و یک اتصال تعریف کنید', text: 'مثلاً دستیار پشتیبانی که به CRM شما وصل است. برای شروع همین کافی است.' },
  { title: 'تیم را دعوت کنید', text: 'سقف مصرف را بگذارید و بگذارید همکارها امتحان کنند.' },
];

const developerPoints = [
  { title: 'یک کلید، همه‌ی مدل‌ها', text: 'به‌جای چند حساب و چند کلید، از یک جا به همه‌ی مدل‌ها وصل می‌شوید.' },
  { title: 'بدون بازنویسی کد', text: 'اگر امروز از OpenAI استفاده می‌کنید، فقط نشانی و کلید عوض می‌شود.' },
  { title: 'همان سقف، همان گزارش', text: 'مصرف برنامه‌تان هم سقف و گزارش دارد، درست مثل مصرف تیم.' },
];

export function GetStarted() {
  return (
    <Section id="start" tone="panel">
      <div>
        <h2 className={`mb-8 max-w-3xl ${sectionHeading}`}>سه قدم تا اولین دستیار. بدون شیرجه.</h2>
      </div>

      <div className="mb-6">
        <StepList steps={steps} />
      </div>
      <p className="mb-10 text-label text-site-ink-soft lg:mb-16">
        اگر وسط راه گیر کردید، به ما بگویید.{' '}
        <Link href="/contact" className={textLink}>
          تماس با ما
        </Link>
      </p>

      <div className="grid gap-8 rounded-card border border-site-line bg-site-surface p-6 sm:p-10 lg:grid-cols-2 lg:gap-16">
        <div className="flex flex-col gap-4">
          <h3 className="m-0 text-h3 font-semibold text-site-ink">محصول خودتان را می‌سازید؟</h3>
          <p className="m-0 text-body text-site-ink-soft">
            همان مدل‌ها و همان کنترل‌ها را در برنامه‌ی خودتان هم دارید. لازم نیست چیزی را از اول بسازید.
          </p>
          <div className="flex flex-wrap gap-x-8 gap-y-3">
            <Link href="/features/ai-gateway" className={textLink}>
              درگاه مدل‌های زبانی
              <ArrowLeft aria-hidden="true" className="size-4" />
            </Link>
            <a href="/docs" className={textLink}>
              مستندات
              <ArrowLeft aria-hidden="true" className="size-4" />
            </a>
          </div>
        </div>
        <ul className="m-0 flex list-none flex-col divide-y divide-site-line p-0">
          {developerPoints.map((point) => (
            <li key={point.title} className="flex items-start gap-3 py-4 first:pt-0 last:pb-0">
              <Check aria-hidden="true" className="mt-1 size-5 shrink-0 text-site-ink" />
              <div className="flex flex-col gap-1">
                <span className="text-body font-semibold text-site-ink">{point.title}</span>
                <span className="text-label text-site-ink-soft">{point.text}</span>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </Section>
  );
}
