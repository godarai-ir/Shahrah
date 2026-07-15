import type { Metadata } from "next";
import { Agents } from '@/app/pages/features/Agents';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "ایجنت‌های هوش مصنوعی",
  description:
    "ایجنت‌های هوش مصنوعی را در محیط عملیاتی با رد درخواست، گاردریل، مسیریابی و کنترل هزینه اجرا کنید.",
  path: "/features/agents",
});

export default function Page() {
  return <Agents />;
}
