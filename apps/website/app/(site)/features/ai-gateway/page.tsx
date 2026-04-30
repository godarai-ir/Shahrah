import type { Metadata } from "next";
import { AIGateway } from '@/app/pages/features/AIGateway';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "AI Gateway Feature",
  description:
    "Route traffic to multiple AI providers with failover, model routing, and cost-aware optimization.",
  path: "/features/ai-gateway",
});

export default function Page() {
  return <AIGateway />;
}
