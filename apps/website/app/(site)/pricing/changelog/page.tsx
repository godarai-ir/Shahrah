import type { Metadata } from "next";
import { PricingChangelog, type PublicPricingChangelog } from "@/app/pages/PricingChangelog";
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "تغییرات قیمت و نرخ‌نامه",
  description:
    "تاریخچه‌ی عمومی قیمت هر کردیت و نسخه‌های نرخ‌نامه‌ی مدل‌ها در گدارAI، همراه با تاریخ اعمال هر تغییر.",
  path: "/pricing/changelog",
});

export const dynamic = "force-dynamic";

export default async function Page() {
  return <PricingChangelog changelog={await loadChangelog()} />;
}

async function loadChangelog(): Promise<PublicPricingChangelog | null> {
  const baseURL = process.env.SHAHRAH_CONTROL_PLANE_URL ?? "http://127.0.0.1:4000";
  try {
    const url = new URL("/v1/pricing/changelog", baseURL.endsWith("/") ? baseURL : `${baseURL}/`);
    const response = await fetch(url, { next: { revalidate: 300 } });
    if (!response.ok) return null;
    return (await response.json()) as PublicPricingChangelog;
  } catch {
    return null;
  }
}
