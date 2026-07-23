import type { Metadata } from "next";
import { SecurityCompliance } from '@/app/pages/features/SecurityCompliance';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "امنیت و کنترل دسترسی",
  description:
    "توکن‌ها، نقش‌ها، مجوزها، سقف مصرف و داده حساس را در مسیر استفاده از مدل‌های زبانی کنترل کنید.",
  path: "/features/security",
});

export default function Page() {
  return <SecurityCompliance />;
}
