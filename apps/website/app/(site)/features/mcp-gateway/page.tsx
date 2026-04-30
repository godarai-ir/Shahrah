import type { Metadata } from "next";
import { MCPGateway } from '@/app/pages/features/MCPGateway';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "MCP Gateway Feature",
  description:
    "Manage and scale MCP servers in production with secure and observable gateway controls.",
  path: "/features/mcp-gateway",
});

export default function Page() {
  return <MCPGateway />;
}
