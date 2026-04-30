import type { Metadata } from "next";
import { Guardrails } from '@/app/pages/features/Guardrails';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Guardrails Feature",
  description:
    "Apply safety and policy controls on prompts and model outputs with Shahrah guardrails.",
  path: "/features/guardrails",
});

export default function Page() {
  return <Guardrails />;
}
