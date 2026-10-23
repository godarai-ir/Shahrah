import type { Metadata } from 'next';
import { CostControl } from '@/app/pages/features/CostControl';
import { buildMetadata } from '@/app/lib/seo';

export const metadata: Metadata = buildMetadata({
  title: 'کنترل و کاهش هزینه‌ی هوش مصنوعی',
  description: 'سقف بودجه برای هر تیم، مسیریابی به مدل ارزان‌تر، حافظه‌ی نهان و پرداخت تومانی.',
  path: '/features/cost-control',
});

export default function Page() {
  return <CostControl />;
}
