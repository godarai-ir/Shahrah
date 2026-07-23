import type { Metadata } from "next";
import { Guardrails } from '@/app/pages/features/Guardrails';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "گاردریل‌های ورودی و خروجی",
  description:
    "متن ورودی و خروجی مدل را پیش از ادامه مسیر بررسی، نشانه‌گذاری، متوقف یا اصلاح کنید.",
  path: "/features/guardrails",
});

export default function Page() {
  return <Guardrails />;
}
