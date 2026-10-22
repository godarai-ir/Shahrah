import { isReady } from '@/app/lib/navigation';
import { ControlChecklist, Section } from './patterns';
import type { ControlGroup, ControlItem } from './patterns';
import { sectionCopy, sectionHeading } from './sectionType';

// Links only to pages that exist; the rest of the claim still stands as plain text.
const item = (label: string, href?: string): ControlItem => ({ label, href: href && isReady(href) ? href : undefined });

// Every item is a confirmed capability (PRODUCT.md and docs/site-architecture/04-open-questions.md).
// SSO and user sync carry the On-Premise qualifier until user sync is confirmed for the cloud too.
const groups: ControlGroup[] = [
  {
    title: 'امنیت',
    surface: 'guard',
    items: [
      item('گاردریل معنایی روی ورودی، خروجی و فایل‌ها', '/features/guardrails'),
      item('گدارAI با داده‌ی شما مدل آموزش نمی‌دهد', '/features/security'),
    ],
  },
  {
    title: 'مدیریت',
    surface: 'govern',
    items: [
      item('نقش و دسترسی هر تیم به مدل، دستیار و ابزار', '/features/governance'),
      item('ثبت هر درخواست و هر تغییر تنظیمات'),
      item('مدت نگهداری گزارش‌ها به انتخاب شما', '/features/observability'),
    ],
  },
  {
    title: 'هزینه',
    surface: 'cost',
    items: [
      item('سقف مصرف برای هر تیم و هر کلید', '/features/cost-control'),
      item('پرداخت تومانی و فاکتور رسمی', '/pricing'),
    ],
  },
  {
    title: 'استقرار',
    surface: 'panel',
    items: [
      item('روی ابر گدارAI، بدون نصب'),
      item('روی سرورهای خودتان، با قرارداد سالانه', '/on-premise'),
      item('ورود سازمانی و همگام‌سازی کاربران در استقرار داخل سازمان', '/on-premise'),
    ],
  },
];

export function Enterprise() {
  return (
    <Section id="enterprise" layout="flex flex-col gap-14">
      <div className="flex flex-col gap-6">
        <div>
          <h2 className={`max-w-3xl ${sectionHeading}`}>داده و دسترسی‌ها دست خودتان است</h2>
        </div>
        <div>
          <p className={`max-w-3xl ${sectionCopy}`}>
            دستیاری که به CRM و اسناد شما سر می‌زند باید قاعده داشته باشد. قاعده را شما می‌نویسید؛ اجرایش با ماست.
          </p>
        </div>
      </div>

      <ControlChecklist groups={groups} />
    </Section>
  );
}
