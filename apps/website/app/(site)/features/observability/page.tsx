import type { Metadata } from "next";
import { Observability } from '@/app/pages/features/Observability';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "پایش و مشاهده‌پذیری",
  description:
    "گزارش رخدادها، رد درخواست‌ها و سنجه‌های کلیدی مصرف مدل‌ها را در گدارAI دنبال کنید.",
  path: "/features/observability",
});

export default function Page() {
  return <Observability />;
}
