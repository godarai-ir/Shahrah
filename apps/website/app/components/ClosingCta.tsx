import { ClosingBand } from './patterns';

export function ClosingCta() {
  return (
    <ClosingBand
      title="اول امتحان کنید."
      titleMuted="بعد تصمیم بگیرید."
      copy="با اعتبار هدیه وارد پنل شوید؛ تاریخ انقضا ندارد. به کارتان نیامد؟ چیزی از دست نداده‌اید."
      primary={{ href: '/onboarding', label: 'شروع با اعتبار هدیه' }}
      secondary={{ href: '/demo', label: 'درخواست دمو' }}
    />
  );
}
