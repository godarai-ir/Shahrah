import { ClosingBand } from './patterns';

export function ClosingCta() {
  return (
    <ClosingBand
      title="اول امتحان کنید."
      titleMuted="بعد تصمیم بگیرید."
      copy="پانزده روز رایگان، با اعتبار اولیه. به کارتان نیامد؟ چیزی از دست نداده‌اید."
      primary={{ href: '/onboarding', label: 'شروع ۱۵ روز رایگان' }}
      secondary={{ href: '/demo', label: 'درخواست دمو' }}
    />
  );
}
