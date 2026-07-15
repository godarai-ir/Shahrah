import type { Metadata } from "next";
import { Home } from '@/app/pages/Home';
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "درگاه مدل‌های زبانی برای تیم‌های ایرانی",
  description:
    "گدارAI به تیم‌های ایرانی کمک می‌کند اتصال به مدل‌های زبانی، مسیریابی، پایش و کنترل هزینه را از یک مسیر واحد مدیریت کنند.",
  path: "/",
  keywords: ["درگاه مدل‌های زبانی", "LLM Gateway", "گدارAI", "زیرساخت هوش مصنوعی ایران"],
});

export default function Page() {
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "گدارAI",
    url: "https://godarai.ir",
    description:
      "درگاه مدل‌های زبانی برای توسعه‌دهندگان و کسب‌وکارهای ایرانی.",
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
