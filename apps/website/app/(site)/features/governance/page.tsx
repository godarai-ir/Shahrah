import type { Metadata } from 'next';
import { Governance } from '@/app/pages/features/Governance';
import { buildMetadata } from '@/app/lib/seo';

export const metadata: Metadata = buildMetadata({
  title: 'حاکمیت هوش مصنوعی در سازمان',
  description:
    'نقش و دسترسی هر تیم به مدل‌ها، دستیارها و ابزارها، ثبت هر درخواست و تغییر تنظیمات و ورود سازمانی؛ هوش مصنوعی را با قاعده به سازمان بیاورید.',
  path: '/features/governance',
});

export default function Page() {
  return <Governance />;
}
