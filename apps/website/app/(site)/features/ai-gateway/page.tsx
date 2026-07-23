import type { Metadata } from "next";
import { AIGateway } from '@/app/pages/features/AIGateway';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "درگاه مدل‌های زبانی",
  description:
    "اتصال به مدل‌ها، مسیریابی، مسیر جایگزین، کنترل مصرف و پایش درخواست‌ها را در یک درگاه مدیریت کنید.",
  path: "/features/ai-gateway",
});

export default function Page() {
  return <AIGateway />;
}
