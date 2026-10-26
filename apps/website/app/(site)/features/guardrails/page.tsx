import type { Metadata } from "next";
import { Guardrails } from '@/app/pages/features/Guardrails';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "گاردریل معنایی: جلوگیری از نشت داده به مدل‌ها",
  description:
    "داده‌ی محرمانه‌ی سازمان را بر اساس معنا تشخیص دهید و برای هر تیم سیاست جدا بگذارید؛ روی ورودی، خروجی و فایل‌ها.",
  path: "/features/guardrails",
});

export default function Page() {
  return <Guardrails />;
}
