import type { Metadata } from "next";
import { CommercialPricing } from '@/app/pages/CommercialPricing';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "قیمت‌گذاری",
  description:
    "اشتراک ماهانه‌ی تومانی برای هر همکار، با اعتبار ماهانه‌ی خودش؛ طرح کسب‌وکارها تا ۱۰۰ نفر و طرح سازمان‌ها با قیمت سفارشی.",
  path: "/pricing",
});

export default function Page() {
  return <CommercialPricing />;
}
