import type { Metadata } from "next";
import { Agents } from '@/app/pages/features/Agents';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "AI Agents Feature",
  description:
    "Deploy and operate AI agents in production with Shahrah's reliable and secure orchestration layer.",
  path: "/features/agents",
});

export default function Page() {
  return <Agents />;
}
