import type { Metadata } from "next";
import { Home } from '@/app/pages/Home';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "AI Gateway for Iranian Teams",
  description:
    "Shahrah helps Iranian teams connect to top AI models with reliable routing, monitoring, and cost control.",
  path: "/",
  keywords: ["AI Gateway", "LLM Gateway", "Shahrah", "Iran AI Infrastructure"],
});

export default function Page() {
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "Shahrah",
    url: "https://shahrah.ir",
    description:
      "AI gateway infrastructure for Iranian developers and businesses.",
    areaServed: "IR",
    sameAs: [],
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <Home />
    </>
  );
}
