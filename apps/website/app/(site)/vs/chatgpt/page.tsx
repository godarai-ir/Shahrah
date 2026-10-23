import type { Metadata } from 'next';
import { GodarAIvsChatGPT } from '@/app/pages/compare/GodarAIvsChatGPT';
import { buildMetadata } from '@/app/lib/seo';

export const metadata: Metadata = buildMetadata({
  title: 'جایگزین سازمانی ChatGPT',
  description:
    'استفاده‌ی پراکنده از ChatGPT را با فضای کاری سازمانی جایگزین کنید: گاردریل معنایی، دسترسی تیمی، گزارش مصرف و پرداخت تومانی.',
  path: '/vs/chatgpt',
  keywords: ['جایگزین ChatGPT', 'ChatGPT سازمانی', 'هوش مصنوعی سازمانی', 'گدارAI'],
});

export default function Page() {
  return <GodarAIvsChatGPT />;
}
