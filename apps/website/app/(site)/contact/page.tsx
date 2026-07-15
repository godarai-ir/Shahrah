import type { Metadata } from "next";
import { Contact } from '@/app/pages/Contact';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "تماس با ما",
  description:
    "برای پرسش‌های فروش، همکاری یا راه‌اندازی درگاه مدل‌های زبانی با تیم گدارAI در ارتباط باشید.",
  path: "/contact",
});

export default function Page() {
  return <Contact />;
}
