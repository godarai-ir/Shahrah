import type { Metadata } from 'next';
import { Patterns } from '@/app/pages/Patterns';
import { buildMetadata } from '@/app/lib/seo';

// Internal playground for the shared section patterns. The `%5F` folder name serves it at /_patterns;
// it stays out of the sitemap and the navigation, and is not indexed.
export const metadata: Metadata = buildMetadata({
  title: 'الگوهای وب‌سایت',
  description: 'نمونه‌ی الگوهای مشترک وب‌سایت گدارAI در چهار رنگ تمایز.',
  path: '/_patterns',
  noIndex: true,
});

export default function Page() {
  return <Patterns />;
}
