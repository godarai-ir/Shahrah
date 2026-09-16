import { Gift } from 'lucide-react';

export function BuildGuildGiftNote({ className = '' }: { className?: string }) {
  return (
    <aside
      className={`flex gap-3 rounded-2xl border border-violet-200/80 bg-violet-50/70 px-4 py-3.5 text-sm leading-7 text-slate-700 dark:border-violet-500/20 dark:bg-violet-500/10 dark:text-slate-300 ${className}`}
    >
      <span className="mt-0.5 grid size-8 shrink-0 place-items-center rounded-full bg-white text-violet-600 shadow-sm dark:bg-violet-500/15 dark:text-violet-300">
        <Gift aria-hidden="true" className="size-4" />
      </span>
      <p>
        به پاس حضور شما در ایونت BuildGuild، <strong className="font-semibold text-violet-700 dark:text-violet-300">یک میلیون تومان اعتبار هدیه</strong> برای
        استفاده از API مدل‌های هوش مصنوعی دریافت می‌کنید. این اعتبار پس از ساخت فضای کاری در گدار به اکانت شما اضافه خواهد شد.
        <span className="mt-0.5 block font-medium text-slate-900 dark:text-white">برای ساخت فضای کاری در گدارAI همین الان اقدام کنید.</span>
      </p>
    </aside>
  );
}
