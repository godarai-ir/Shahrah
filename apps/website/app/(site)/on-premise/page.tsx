import type { Metadata } from 'next';

import { buildMetadata } from '@/app/lib/seo';
import { OnPremiseRequest } from '@/app/pages/OnPremiseRequest';

export const metadata: Metadata = buildMetadata({
  title: 'استقرار گدارAI در سازمان (On-Premise)',
  description:
    'با تکمیل یک فرم کوتاه مرحله‌ای، وضعیت زیرساخت سازمان خود را معرفی کنید تا استقرار On-Premise گدارAI را برنامه‌ریزی کنیم.',
  path: '/on-premise',
});

export default function Page() {
  return <OnPremiseRequest />;
}
