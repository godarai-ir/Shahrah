import type { Metadata } from "next";
import { MCPGateway } from '@/app/pages/features/MCPGateway';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "درگاه MCP",
  description:
    "دسترسی ایجنت‌ها به سرورها و ابزارهای MCP را با سیاست‌های متمرکز و اجرای قابل پیگیری مدیریت کنید.",
  path: "/features/mcp-gateway",
});

export default function Page() {
  return <MCPGateway />;
}
