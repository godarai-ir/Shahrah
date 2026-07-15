import type { Metadata } from "next";
import { Documentation } from '@/app/pages/Documentation';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "مستندات",
  description:
    "با مستندات API، راهنمای اتصال و مرجع پیاده‌سازی گدارAI شروع کنید.",
  path: "/docs",
});

export default function Page() {
  return <Documentation />;
}
