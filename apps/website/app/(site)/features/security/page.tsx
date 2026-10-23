import type { Metadata } from "next";
import { SecurityCompliance } from '@/app/pages/features/SecurityCompliance';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "امنیت و حریم داده در گدارAI",
  description:
    "مسیر داده، محل اجرا، کنترل دسترسی و گاردریل معنایی؛ هر چه مسئول امنیت پیش از تأیید باید بداند.",
  path: "/features/security",
});

export default function Page() {
  return <SecurityCompliance />;
}
