import type { Metadata } from "next";
import { Blog } from '@/app/pages/Blog';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "وبلاگ",
  description:
    "یادداشت‌های گدارAI درباره درگاه مدل‌های زبانی، کنترل هزینه، امنیت و تجربه ساخت محصول در ایران.",
  path: "/blog",
});

export default function Page() {
  return <Blog />;
}
