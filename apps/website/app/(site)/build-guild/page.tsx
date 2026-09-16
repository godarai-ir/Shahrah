import type { Metadata } from 'next';

import { buildMetadata } from '@/app/lib/seo';
import { BuildGuildSurvey } from '@/app/pages/BuildGuildSurvey';

export const metadata: Metadata = buildMetadata({
  title: 'هدیهٔ گدارAI به شرکت‌کنندگان BuildGuild',
  description:
    'شرکت‌کنندگان ایونت BuildGuild یک میلیون تومان اعتبار رایگان برای استفاده از API مدل‌های هوش مصنوعی در گدارAI دریافت می‌کنند.',
  path: '/build-guild',
});

export default function Page() {
  return <BuildGuildSurvey />;
}
