import type { Metadata } from "next";
import { Terms } from '@/app/pages/Terms';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "شرایط استفاده",
  description:
    "شرایط استفاده، سیاست‌های مصرف مجاز و توافق‌های حقوقی گدارAI را مرور کنید.",
  path: "/terms",
});

export default function Page() {
  return <Terms />;
}
