import type { Metadata } from "next";
import { AIGateway } from '@/app/pages/features/AIGateway';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "درگاه مدل‌های زبانی سازگار با OpenAI",
  description:
    "یک API برای ۱۲۰۰+ مدل از ۲۱ ارائه‌دهنده، با مدل مجازی، حافظه‌ی نهان، کلید جدا و گزارش کامل؛ پرداخت تومانی.",
  path: "/features/ai-gateway",
});

export default function Page() {
  return <AIGateway />;
}
