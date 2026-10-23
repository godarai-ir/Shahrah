import type { Metadata } from "next";
import { Agents } from '@/app/pages/features/Agents';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "دستیار هوش مصنوعی برای هر تیم",
  description:
    "دستیار آماده یا ساخت بی‌کد با دانش و ابزارهای تیم، با دسترسی و سقف مصرف کنترل‌شده.",
  path: "/features/agents",
});

export default function Page() {
  return <Agents />;
}
