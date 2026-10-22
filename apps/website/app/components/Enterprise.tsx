import { ArrowLeft } from 'lucide-react';
import Link from 'next/link';
import { WaterEdge } from './FordArt';
import { ControlChecklist, Section } from './patterns';
import type { ControlGroup } from './patterns';
import { sectionCopy, sectionHeading } from './sectionType';

const groups: ControlGroup[] = [
  {
    title: 'امنیت و حریم خصوصی',
    items: [
      { label: 'گاردریل‌ها داده‌ی حساس را قبل از رسیدن به مدل می‌گیرند' },
      { label: 'هر دستیار فقط به ابزارهای مجاز خودش دسترسی دارد' },
      { label: 'هر درخواست در گزارش رخدادها ثبت می‌شود' },
    ],
    surface: 'panel',
  },
  {
    title: 'مدیریت و مصرف',
    items: [
      { label: 'مشخص می‌کنید چه کسی از کدام دستیار استفاده کند' },
      { label: 'برای هر تیم سقف مصرف می‌گذارید' },
      { label: 'کیف پول تومانی و گزارش مالی روشن' },
    ],
    surface: 'panel-deep',
  },
  {
    title: 'استقرار و پشتیبانی',
    items: [
      { label: 'روی ابر گدارAI، همین امروز و بدون نصب' },
      { label: 'داخل سرور خودتان، برای داده‌ی محرمانه' },
      { label: 'رابط راست‌به‌چپ و پشتیبانی فارسی' },
    ],
    surface: 'panel',
  },
];

export function Enterprise() {
  return (
    <Section id="terms" layout="flex flex-col gap-14">
      <WaterEdge className="text-site-line" />
      <div className="flex flex-col gap-6">
        <div>
          <h2 className={`max-w-3xl ${sectionHeading}`}>داده و دسترسی‌ها همیشه دست خودتان است</h2>
        </div>
        <div>
          <p className={`max-w-3xl ${sectionCopy}`}>
            دستیاری که به CRM و اسناد شما سر می‌زند باید قاعده داشته باشد. قاعده را شما می‌نویسید؛ اجرایش با ماست.
          </p>
        </div>
      </div>

      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap items-baseline justify-between gap-3 px-1">
          <span className="text-sm text-site-ink-soft">کجا اجرا شود</span>
          <Link
            href="/on-premise"
            className="group inline-flex items-center gap-2 rounded-full text-sm font-medium text-site-ink transition-colors hover:text-site-accent focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-site-accent"
          >
            درخواست استقرار درون‌سازمانی
            <ArrowLeft className="size-4 transition-transform duration-500 group-hover:-translate-x-0.5" />
          </Link>
        </div>

        <ControlChecklist groups={groups} />
      </div>
    </Section>
  );
}
