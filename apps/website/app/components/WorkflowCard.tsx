import { Check, Database, FileText, PenLine, UserCheck } from 'lucide-react';

const steps = [
  { icon: Database, text: 'سفارش ۱۴۸۲ را در CRM پیدا کرد', done: true },
  { icon: FileText, text: 'فاکتور و سابقه‌ی پرداخت را کنار هم گذاشت', done: true },
  { icon: PenLine, text: 'پیش‌نویس پاسخ را نوشت', done: true },
  { icon: UserCheck, text: 'منتظر تأیید همکار پشتیبانی', done: false },
];

export function WorkflowCard() {
  return (
    <div
      role="img"
      aria-label="نمونه‌ای از کار یک دستیار: پیام مشتری درباره‌ی فاکتور اشتباه را می‌خواند، سفارش و فاکتور را در سیستم‌ها پیدا می‌کند، پیش‌نویس پاسخ می‌نویسد و برای تأیید به همکار پشتیبانی می‌سپارد."
      className="m-auto w-full max-w-md overflow-hidden rounded-2xl bg-white shadow-sm dark:bg-slate-900"
    >
      <div className="flex items-center justify-between border-b border-gray-100 px-5 py-3 dark:border-gray-800">
        <span className="text-sm font-semibold text-slate-950 dark:text-white">نمونه: پشتیبانی مشتری</span>
        <span className="rounded-full bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700 dark:bg-amber-500/10 dark:text-amber-300">
          منتظر تأیید شما
        </span>
      </div>

      <div className="space-y-5 p-5">
        <div className="rounded-xl bg-slate-50 p-4 dark:bg-slate-800/60">
          <p className="mb-1 text-xs text-slate-500 dark:text-gray-400">پیام مشتری</p>
          <p className="text-sm leading-7 text-slate-800 dark:text-gray-200">
            سلام، مبلغ فاکتور سفارش ۱۴۸۲ اشتباه است. لطفاً بررسی کنید.
          </p>
        </div>

        <ol className="space-y-3">
          {steps.map((step) => (
            <li key={step.text} className="flex items-center gap-3 text-sm">
              <span
                className={
                  step.done
                    ? 'flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-50 text-blue-600 dark:bg-blue-500/10 dark:text-blue-400'
                    : 'flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-amber-50 text-amber-600 dark:bg-amber-500/10 dark:text-amber-300'
                }
              >
                <step.icon className="h-4 w-4" />
              </span>
              <span className={step.done ? 'text-slate-700 dark:text-gray-300' : 'font-medium text-slate-950 dark:text-white'}>
                {step.text}
              </span>
              {step.done ? <Check className="ms-auto h-4 w-4 shrink-0 text-green-500" /> : null}
            </li>
          ))}
        </ol>

        <div className="rounded-xl border border-dashed border-blue-200 p-4 dark:border-blue-500/30">
          <p className="mb-1 text-xs text-slate-500 dark:text-gray-400">پیش‌نویس پاسخ</p>
          <p className="text-sm leading-7 text-slate-800 dark:text-gray-200">
            سلام، حق با شماست. مبلغ ردیف دوم دو بار محاسبه شده بود. فاکتور اصلاحی را همین امروز برایتان می‌فرستیم.
          </p>
        </div>
      </div>
    </div>
  );
}
