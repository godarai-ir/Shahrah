import type { Metadata } from 'next';

import { buildMetadata } from '@/app/lib/seo';
import { GodarAIvsLiteLLM } from '@/app/pages/compare/GodarAIvsLiteLLM';

export const metadata: Metadata = buildMetadata({
  title: 'مقایسه گدارAI و LiteLLM',
  description:
    'تفاوت گدارAI و LiteLLM در معماری درگاه، مسیریابی، گاردریل، درگاه MCP، کنترل هزینه، مشاهده‌پذیری، استقرار و پشتیبانی؛ و اینکه چه زمانی مهاجرت ارزش دارد.',
  path: '/vs/litellm',
  keywords: ['LiteLLM', 'جایگزین LiteLLM', 'AI Gateway', 'درگاه مدل زبانی', 'گدارAI'],
});

export default function Page() {
  return <GodarAIvsLiteLLM />;
}
