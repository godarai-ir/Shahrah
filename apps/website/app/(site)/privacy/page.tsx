import type { Metadata } from "next";
import { Privacy } from '@/app/pages/Privacy';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "حریم خصوصی",
  description:
    "سیاست حریم خصوصی گدارAI و شیوه رسیدگی، محافظت و پردازش داده‌های کاربران را بخوانید.",
  path: "/privacy",
});

export default function Page() {
  return <Privacy />;
}
