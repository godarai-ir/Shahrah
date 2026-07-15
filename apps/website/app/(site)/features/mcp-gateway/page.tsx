import type { Metadata } from "next";
import { MCPGateway } from '@/app/pages/features/MCPGateway';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "درگاه MCP",
  description:
    "سرورهای MCP را در محیط عملیاتی با کنترل دسترسی، مشاهده‌پذیری و سیاست‌های قابل پیگیری مدیریت کنید.",
  path: "/features/mcp-gateway",
});

export default function Page() {
  return <MCPGateway />;
}
