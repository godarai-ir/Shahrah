import type { Metadata } from "next";
import { Observability } from '@/app/pages/features/Observability';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "پایش و مشاهده‌پذیری",
  description:
    "لاگ، رد درخواست و سنجه‌های مصرف مدل‌ها را بررسی کنید و علت خطا، تأخیر یا افزایش هزینه را پیدا کنید.",
  path: "/features/observability",
});

export default function Page() {
  return <Observability />;
}
