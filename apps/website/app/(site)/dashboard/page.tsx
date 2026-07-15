import type { Metadata } from "next";
import { Dashboard } from '@/app/pages/Dashboard';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "نمای کلی داشبورد",
  description:
    "قابلیت‌های داشبورد گدارAI برای پایش درخواست‌ها، مصرف، هزینه و عملکرد مدل‌ها را ببینید.",
  path: "/dashboard",
});

export default function Page() {
  return <Dashboard />;
}
