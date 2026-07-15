import type { Metadata } from "next";
import { AIGateway } from '@/app/pages/features/AIGateway';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "درگاه مدل‌های زبانی",
  description:
    "درخواست‌ها را با مدل مجازی، مسیر جایگزین، مسیریابی و کنترل هزینه بین حساب‌های ارائه‌دهنده مدیریت کنید.",
  path: "/features/ai-gateway",
});

export default function Page() {
  return <AIGateway />;
}
