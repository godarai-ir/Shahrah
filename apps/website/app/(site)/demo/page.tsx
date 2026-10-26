import type { Metadata } from 'next';

import { DemoBooking } from '@/app/components/DemoBooking';
import { buildMetadata } from '@/app/lib/seo';

export const metadata: Metadata = buildMetadata({
  title: 'درخواست دمو',
  description: 'یک جلسه‌ی ۶۰ دقیقه‌ای برای دیدن فضای کاری، گاردریل معنایی و گزارش‌های گدارAI روی نیاز سازمان شما.',
  path: '/demo',
});

export default function Page() {
  return <DemoBooking />;
}
