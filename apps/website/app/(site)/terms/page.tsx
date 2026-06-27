import type { Metadata } from "next";
import { Terms } from '@/app/pages/Terms';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Terms of Service",
  description:
    "Review GodarAI terms of service, acceptable use policies, and legal agreements.",
  path: "/terms",
});

export default function Page() {
  return <Terms />;
}
