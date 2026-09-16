import type { Metadata } from 'next';

import { buildMetadata } from '@/app/lib/seo';
import { GodarAIvsPortkey } from '@/app/pages/compare/GodarAIvsPortkey';

export const metadata: Metadata = buildMetadata({
  title: 'مقایسه گدارAI و Portkey',
  description:
    'تفاوت گدارAI و Portkey در استقلال و نقشه راه، استقرار و حاکمیت داده، درگاه MCP، گاردریل، کنترل هزینه، پرداخت و پشتیبانی برای تیم‌های ایرانی.',
  path: '/vs/portkey',
  keywords: ['Portkey', 'جایگزین Portkey', 'AI Gateway', 'درگاه مدل زبانی', 'گدارAI'],
});

export default function Page() {
  return <GodarAIvsPortkey />;
}
