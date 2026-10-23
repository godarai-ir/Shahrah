import type { Metadata } from 'next';
import { Integrations } from '@/app/pages/Integrations';
import { buildMetadata } from '@/app/lib/seo';

export const metadata: Metadata = buildMetadata({
  title: 'اتصال هوش مصنوعی به ابزارهای سازمان',
  description:
    'دستیارهای گدارAI به Jira، Notion، Git، CRM، پیام‌رسان‌ها و سرورهای MCP سازمان وصل می‌شوند، با دسترسی کنترل‌شده.',
  path: '/integrations',
});

export default function Page() {
  return <Integrations />;
}
