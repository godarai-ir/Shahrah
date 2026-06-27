import type { Metadata } from "next";
import { Observability } from '@/app/pages/features/Observability';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Observability Feature",
  description:
    "Track logs, traces, and key metrics across your AI stack with GodarAI observability features.",
  path: "/features/observability",
});

export default function Page() {
  return <Observability />;
}
