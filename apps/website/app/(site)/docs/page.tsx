import type { Metadata } from "next";
import { Documentation } from '@/app/pages/Documentation';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Documentation",
  description:
    "Get started with GodarAI API documentation, integration guides, and implementation references.",
  path: "/docs",
});

export default function Page() {
  return <Documentation />;
}
