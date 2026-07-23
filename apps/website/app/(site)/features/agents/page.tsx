import type { Metadata } from "next";
import { Agents } from '@/app/pages/features/Agents';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "اجرای ایجنت‌های هوش مصنوعی",
  description:
    "ایجنت‌های هوش مصنوعی را از مسیر گدارAI اجرا کنید و رد درخواست، گاردریل، مسیریابی و هزینه هر اجرا را بررسی کنید.",
  path: "/features/agents",
});

export default function Page() {
  return <Agents />;
}
