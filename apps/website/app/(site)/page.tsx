import type { Metadata } from "next";
import { Home } from '@/app/pages/Home';
import { buildMetadata } from "@/app/lib/seo";
import { SOCIAL_LINK_URLS } from "@/app/lib/social-links";

export const metadata: Metadata = buildMetadata({
  title: "هوش مصنوعی امن برای کار روزانه‌ی تیم‌ها",
  description:
    "دستیار هوش مصنوعی برای هر تیم، با گاردریل معنایی، سقف هزینه و گزارش کامل. پرداخت تومانی، روی ابر یا سرور خودتان.",
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
      "فضای کاری هوش مصنوعی و درگاه مدل‌های زبانی برای سازمان‌های ایرانی، با گاردریل معنایی، کنترل دسترسی و سقف هزینه.",
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
