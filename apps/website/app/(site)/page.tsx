import type { Metadata } from "next";
import { Home } from '@/app/pages/Home';
import { buildMetadata } from "@/app/lib/seo";
import { SOCIAL_LINK_URLS } from "@/app/lib/social-links";

export const metadata: Metadata = buildMetadata({
  title: "دستیار هوش مصنوعی برای سازمان‌های ایرانی",
  description:
    "دستیارهای گدارAI به CRM، پیام‌رسان‌ها، Jira، Notion و Git وصل می‌شوند و کار تیم‌ها را جلو می‌برند؛ با کنترل دسترسی، سقف مصرف و پرداخت به تومان.",
  path: "/",
  keywords: ["دستیار هوش مصنوعی سازمانی", "هوش مصنوعی برای شرکت", "MCP", "گدارAI", "درگاه مدل‌های زبانی"],
});

export default function Page() {
  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "گدارAI",
    url: "https://godarai.ir",
    description:
      "دستیارهای هوش مصنوعی متصل به ابزارهای سازمان، با کنترل دسترسی و مصرف، برای کسب‌وکارهای ایرانی.",
    areaServed: "IR",
    sameAs: SOCIAL_LINK_URLS,
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
