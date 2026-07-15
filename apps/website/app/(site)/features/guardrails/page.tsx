import type { Metadata } from "next";
import { Guardrails } from '@/app/pages/features/Guardrails';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "گاردریل‌ها",
  description:
    "روی ورودی، خروجی و داده حساس در مسیر مدل، کنترل‌های ایمنی و سیاستی گدارAI را اعمال کنید.",
  path: "/features/guardrails",
});

export default function Page() {
  return <Guardrails />;
}
