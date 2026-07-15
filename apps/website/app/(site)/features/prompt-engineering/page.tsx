import type { Metadata } from "next";
import { PromptEngineering } from '@/app/pages/features/PromptEngineering';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "رجیستری پرامپت",
  description:
    "پرامپت‌ها را با نسخه، متغیر، FQN و اجرای قابل‌پیگیری در گدارAI مدیریت کنید.",
  path: "/features/prompt-engineering",
});

export default function Page() {
  return <PromptEngineering />;
}
