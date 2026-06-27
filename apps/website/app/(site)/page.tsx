import type { Metadata } from "next";
import { Home } from '@/app/pages/Home';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "AI Gateway for Iranian Teams",
  description:
    "GodarAI helps Iranian teams connect to top AI models with reliable routing, monitoring, and cost control.",
  path: "/",
  keywords: ["AI Gateway", "LLM Gateway", "GodarAI", "Iran AI Infrastructure"],
});

export default function Page() {
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "GodarAI",
    url: "https://godarai.ir",
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
