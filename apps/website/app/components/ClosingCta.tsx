import { Pebbles, WaterLines } from './FordArt';
import { ClosingBand } from './patterns';

export function ClosingCta() {
  return (
    <ClosingBand
      title="اول یک دستیار بسازید،"
      titleMuted="بعد ببینید کار چقدر سبک می‌شود."
      copy="به یکی از ابزارهایتان وصلش کنید، به یک همکار بدهید و ببینید چطور کار می‌کند. به کارتان نیامد؟ چیزی از دست نداده‌اید؛ ۱۵ روز اول رایگان است."
      primary={{ href: '/onboarding', label: '۱۵ روز رایگان امتحان کنید' }}
      secondary={{ href: '/demo', label: 'یا با ما حرف بزنید' }}
      backdrop={
        <>
          <WaterLines className="text-site-night-ink/10" />
          <Pebbles className="absolute bottom-6 start-8 w-56 opacity-25 sm:w-72" />
        </>
      }
    />
  );
}
