import { Check, Database, FileText, PenLine, UserCheck } from 'lucide-react';
import { ProductWindow } from './patterns';

const steps = [
  { icon: Database, text: 'سفارش ۱۴۸۲ را در CRM پیدا کرد', done: true },
  { icon: FileText, text: 'فاکتور و سابقه‌ی پرداخت را کنار هم گذاشت', done: true },
  { icon: PenLine, text: 'پیش‌نویس پاسخ را نوشت', done: true },
  { icon: UserCheck, text: 'منتظر تأیید همکار پشتیبانی', done: false },
];

export function WorkflowCard() {
  return (
    <ProductWindow
      title="نمونه: پشتیبانی مشتری"
      meta={<span className="rounded-full bg-site-panel-deep px-3 py-1 text-xs font-medium text-site-ink-soft">منتظر تأیید شما</span>}
      label="نمونه‌ای از کار یک دستیار: پیام مشتری درباره‌ی فاکتور اشتباه را می‌خواند، سفارش و فاکتور را در سیستم‌ها پیدا می‌کند، پیش‌نویس پاسخ می‌نویسد و برای تأیید به همکار پشتیبانی می‌سپارد."
      size="m-auto max-w-md"
    >
      <div className="space-y-5 p-5">
        <div className="rounded-2xl bg-site-panel p-4">
          <p className="mb-1 text-xs text-site-ink-soft">پیام مشتری</p>
          <p className="text-sm leading-7 text-site-ink">سلام، مبلغ فاکتور سفارش ۱۴۸۲ اشتباه است. لطفاً بررسی کنید.</p>
        </div>

        <ol className="space-y-3">
          {steps.map((step) => (
            <li key={step.text} className="flex items-center gap-3 text-sm">
              <span
                className={`flex size-8 shrink-0 items-center justify-center rounded-full ${
                  step.done ? 'bg-site-panel-deep text-site-accent' : 'border border-dashed border-site-accent text-site-accent'
                }`}
              >
                <step.icon className="size-4" />
              </span>
              <span className={step.done ? 'text-site-ink-soft' : 'font-medium text-site-ink'}>{step.text}</span>
              {step.done ? <Check className="ms-auto size-4 shrink-0 text-site-accent" /> : null}
            </li>
          ))}
        </ol>

        <div className="rounded-2xl border border-dashed border-site-line p-4">
          <p className="mb-1 text-xs text-site-ink-soft">پیش‌نویس پاسخ</p>
          <p className="text-sm leading-7 text-site-ink">
            سلام، حق با شماست. مبلغ ردیف دوم دو بار محاسبه شده بود. فاکتور اصلاحی را همین امروز برایتان می‌فرستیم.
          </p>
        </div>
      </div>
    </ProductWindow>
  );
}
