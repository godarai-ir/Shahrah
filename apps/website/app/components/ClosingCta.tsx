import Link from 'next/link';
import { Reveal } from './Reveal';

const stats = [
  { value: '۱۲۰۰+', label: 'مدل آماده‌ی استفاده', color: 'text-blue-400' },
  { value: '۲۱', label: 'ارائه‌دهنده‌ی مدل', color: 'text-amber-400' },
  { value: '۱۵ روز', label: 'آزمایش رایگان', color: 'text-green-400' },
];

export function ClosingCta() {
  return (
    <section className="relative w-full overflow-hidden bg-slate-950 py-32 text-white">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-px bg-linear-to-r from-transparent via-white/20 to-transparent" />
      <div className="mx-auto flex w-full max-w-6xl flex-col items-center gap-10 px-6 text-center">
        <Reveal>
          <span className="inline-flex h-7 w-fit items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 text-xs text-white/70 backdrop-blur-sm">
            <span className="h-1.5 w-1.5 rounded-full bg-blue-400" />
            بی‌گدار به آب نزنید
          </span>
        </Reveal>
        <Reveal>
          <h2 className="m-0 max-w-3xl text-balance text-center text-4xl font-semibold leading-tight text-white md:text-5xl xl:text-6xl xl:leading-tight">
            اول یک دستیار بسازید،{' '}
            <span className="font-normal text-blue-300">بعد ببینید کار چقدر سبک می‌شود.</span>
          </h2>
        </Reveal>
        <Reveal className="flex max-w-2xl flex-col gap-6 text-white/70">
          <p className="m-0 text-base leading-8">
            به یکی از ابزارهایتان وصلش کنید، به یک همکار بدهید و ببینید چطور کار می‌کند. به کارتان نیامد؟ چیزی از دست نداده‌اید؛ ۱۵ روز اول رایگان است.
          </p>
        </Reveal>
        <Reveal className="mt-2 flex flex-col items-center gap-5">
          <Link
            href="/onboarding"
            className="inline-flex h-12 items-center justify-center rounded-2xl bg-blue-500 px-7 text-base font-medium text-white transition-colors hover:bg-blue-400"
          >
            ۱۵ روز رایگان امتحان کنید
          </Link>
          <Link href="/demo" className="text-sm text-white/70 transition-colors hover:text-white">
            یا با ما حرف بزنید ←
          </Link>
        </Reveal>

        <dl className="mt-12 grid w-full max-w-3xl grid-cols-1 gap-8 sm:grid-cols-3 sm:gap-0 sm:divide-x sm:divide-x-reverse sm:divide-white/10">
          {stats.map((stat) => (
            <div key={stat.label} className="flex flex-col items-center gap-3 px-6">
              <dt className={`text-3xl font-semibold leading-none md:text-4xl ${stat.color}`}>{stat.value}</dt>
              <dd className="m-0 text-xs leading-5 text-white/65">{stat.label}</dd>
            </div>
          ))}
        </dl>
      </div>
    </section>
  );
}
