import { Languages, Server, Wallet } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

const items: { icon: LucideIcon; title: string; text: string }[] = [
  { icon: Wallet, title: 'پرداخت تومانی', text: 'بدون کارت بین‌المللی' },
  { icon: Server, title: 'ابر یا سرور خودتان', text: 'روی ابر گدارAI یا داخل سازمان' },
  { icon: Languages, title: 'فارسی', text: 'از رابط تا پشتیبانی' },
];

// The three local advantages a foreign product cannot claim (PRODUCT.md, Positioning).
export function TrustStrip() {
  return (
    <ul className="m-0 grid list-none grid-cols-1 gap-6 p-0 sm:grid-cols-3">
      {items.map((item) => (
        <li key={item.title} className="flex items-center gap-3">
          <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-site-panel-deep text-site-accent">
            <item.icon className="size-5" />
          </span>
          <span className="flex flex-col text-sm">
            <span className="font-medium text-site-ink">{item.title}</span>
            <span className="text-site-ink-soft">{item.text}</span>
          </span>
        </li>
      ))}
    </ul>
  );
}
