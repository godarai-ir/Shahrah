import type { Metadata } from "next";
import { SecurityCompliance } from '@/app/pages/features/SecurityCompliance';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "امنیت و تطابق",
  description:
    "توکن‌ها، دسترسی‌ها، سیاست‌ها و داده حساس را در مسیر مصرف مدل‌های زبانی کنترل کنید.",
  path: "/features/security",
});

export default function Page() {
  return <SecurityCompliance />;
}
