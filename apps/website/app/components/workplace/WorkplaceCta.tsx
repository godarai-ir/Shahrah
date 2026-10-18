import { Languages, Server, Wallet } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import Link from 'next/link';
import { Reveal } from '../Reveal';
import { pillOutline, pillPrimary, sectionPad, TwoTone } from './parts';

const badges: { icon: LucideIcon; title: string; text: string }[] = [
  { icon: Wallet, title: 'پرداخت تومانی', text: 'بدون کارت بین‌المللی' },
  { icon: Server, title: 'استقرار هرجا', text: 'ابر یا سرور خودتان' },
  { icon: Languages, title: 'فارسی و راست‌به‌چپ', text: 'از رابط تا پشتیبانی' },
];

export function WorkplaceCta() {
  return (
    <section className={`${sectionPad} pb-20 md:pb-28`}>
      <div className="mx-auto w-full max-w-7xl">
        <Reveal className="mb-8 max-w-3xl">
          <TwoTone lead="گدارAI را ۱۵ روز رایگان امتحان کنید." muted="اگر به کارتان نیامد، چیزی از دست نداده‌اید." />
        </Reveal>

        <Reveal className="mb-14 flex flex-wrap items-center gap-3">
          <Link href="/onboarding" className={pillPrimary}>
            ۱۵ روز رایگان امتحان کنید
          </Link>
          <Link href="/demo" className={pillOutline}>
            با ما حرف بزنید
          </Link>
        </Reveal>

        <Reveal>
          <ul className="grid grid-cols-1 gap-6 border-t border-zinc-900/10 pt-8 dark:border-white/15 sm:grid-cols-3">
            {badges.map((badge) => (
              <li key={badge.title} className="flex items-center gap-3">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-zinc-900/5 text-zinc-900 dark:bg-white/10 dark:text-white">
                  <badge.icon className="size-5" />
                </span>
                <span className="flex flex-col text-sm">
                  <span className="font-medium text-zinc-900 dark:text-white">{badge.title}</span>
                  <span className="font-light text-zinc-900/55 dark:text-white/55">{badge.text}</span>
                </span>
              </li>
            ))}
          </ul>
        </Reveal>
      </div>
    </section>
  );
}
