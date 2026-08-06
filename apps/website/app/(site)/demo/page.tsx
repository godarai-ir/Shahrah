import type { Metadata } from 'next';

import { DemoBooking } from '@/app/components/DemoBooking';
import { buildMetadata } from '@/app/lib/seo';

export const metadata: Metadata = buildMetadata({
  title: 'رزرو دموی خصوصی گدارAI',
  description: 'یک جلسه خصوصی ۶۰ دقیقه‌ای برای آشنایی با گدارAI و بررسی نیازهای فنی تیم خود رزرو کنید.',
  path: '/demo',
});

export default function Page() {
  return <DemoBooking />;
}
