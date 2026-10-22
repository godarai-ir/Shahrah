import { ClosingBand } from './patterns';

export function ClosingCta() {
  return (
    <ClosingBand
      title="بی‌گدار به آب نزنید."
      copy="پانزده روز، رایگان و با اعتبار اولیه. اول امتحان کنید، بعد تصمیم بگیرید."
      primary={{ href: '/onboarding', label: '۱۵ روز رایگان شروع کنید' }}
      secondary={{ href: '/demo', label: 'درخواست دمو' }}
    />
  );
}
