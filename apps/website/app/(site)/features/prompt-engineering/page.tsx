import type { Metadata } from "next";
import { redirect } from 'next/navigation';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Prompt Engineering",
  description:
    "Prompt engineering workflows are available within GodarAI AI Gateway features.",
  path: "/features/prompt-engineering",
  noIndex: true,
});

export default function Page() {
  redirect('/features/ai-gateway');
}
