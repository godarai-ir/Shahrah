import type { Metadata } from "next";
import { SecurityCompliance } from '@/app/pages/features/SecurityCompliance';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Security and Compliance Feature",
  description:
    "Enterprise-grade security controls and compliance-ready architecture for AI operations.",
  path: "/features/security",
});

export default function Page() {
  return <SecurityCompliance />;
}
