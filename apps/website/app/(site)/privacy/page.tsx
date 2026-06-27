import type { Metadata } from "next";
import { Privacy } from '@/app/pages/Privacy';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Privacy Policy",
  description:
    "Read GodarAI privacy policy and how we handle, protect, and process user data responsibly.",
  path: "/privacy",
});

export default function Page() {
  return <Privacy />;
}
