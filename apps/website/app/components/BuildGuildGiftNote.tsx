import { Gift } from 'lucide-react';

export function BuildGuildGiftNote({ className = '' }: { className?: string }) {
  return (
    <aside
      className={`flex gap-3 rounded-card border border-site-line bg-site-sand px-4 py-3.5 text-sm leading-7 text-site-ink ${className}`}
    >
      <span className="mt-0.5 grid size-8 shrink-0 place-items-center rounded-full bg-site-surface text-site-accent-text">
        <Gift aria-hidden="true" className="size-4" />
      </span>
      <p>
        به پاس حضور شما در ایونت BuildGuild، <strong className="font-semibold text-site-accent-text">یک میلیون تومان اعتبار هدیه</strong> برای
        استفاده از API مدل‌های هوش مصنوعی دریافت می‌کنید. این اعتبار پس از ساخت فضای کاری در گدار به اکانت شما اضافه خواهد شد.
        <span className="mt-0.5 block font-medium text-site-ink">برای ساخت فضای کاری در گدارAI همین الان اقدام کنید.</span>
      </p>
    </aside>
  );
}
