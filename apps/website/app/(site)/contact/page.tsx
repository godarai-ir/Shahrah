import type { Metadata } from "next";
import { Contact } from '@/app/pages/Contact';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "Contact",
  description:
    "Contact GodarAI for sales, partnerships, or product questions about our AI gateway platform.",
  path: "/contact",
});

export default function Page() {
  return <Contact />;
}
