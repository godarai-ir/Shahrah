import type { Metadata } from "next";
import { About } from '@/app/pages/About';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "درباره گدارAI",
  description:
    "با ماموریت گدارAI برای ساخت درگاه مدل‌های زبانی، کنترل هزینه و پایش مصرف برای تیم‌های ایرانی آشنا شوید.",
  path: "/about",
});

export default function Page() {
  return <About />;
}
