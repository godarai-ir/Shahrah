import { Globe, Puzzle, Server } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { iconWell } from '../sectionType';

const items: { icon: LucideIcon; title: string; text: string }[] = [
  { icon: Globe, title: 'مدل‌های روز دنیا، بدون تحریم', text: 'دسترسی پایدار، پرداخت تومانی' },
  { icon: Server, title: 'روی سرورهای خودتان', text: 'یا ابر گدارAI، به انتخاب شما' },
  { icon: Puzzle, title: 'سازگار با نرم‌افزارهای ایرانی', text: 'حسابداری‌ و فروش، مسیریاب‌، پیام‌رسان‌ و ...' },
];

// The three local advantages a foreign product cannot claim (PRODUCT.md, Positioning).
// `night` sets it on the navy footer.
export function TrustStrip({ tone = 'ground' }: { tone?: 'ground' | 'night' }) {
  const night = tone === 'night';

  return (
    <ul className="m-0 grid list-none grid-cols-1 gap-6 p-0 sm:grid-cols-3">
      {items.map((item) => (
        <li key={item.title} className="flex items-center gap-3">
          <span className={night ? 'flex size-10 shrink-0 items-center justify-center rounded-card bg-site-night-raised text-site-night-ink' : `size-10 ${iconWell}`}>
            <item.icon aria-hidden="true" className="size-5" />
          </span>
          <span className="flex flex-col text-label">
            <span className={`font-medium ${night ? 'text-site-night-ink' : 'text-site-ink'}`}>{item.title}</span>
            <span className={night ? 'text-site-night-soft' : 'text-site-ink-soft'}>{item.text}</span>
          </span>
        </li>
      ))}
    </ul>
  );
}
