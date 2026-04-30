import type { Metadata } from "next";
import { About } from '@/app/pages/About';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "About Shahrah",
  description:
    "Learn about Shahrah, our mission, and how we build AI infrastructure for Iranian developers and businesses.",
  path: "/about",
});

export default function Page() {
  return <About />;
}
