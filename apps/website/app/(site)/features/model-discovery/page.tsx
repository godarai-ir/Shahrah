import type { Metadata } from "next";
import { ModelDiscovery } from "@/app/pages/features/ModelDiscovery";
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "کشف و مقایسه مدل‌ها",
  description:
    "با Model Discovery گدارAI، مدل‌ها و پروایدرها را جست‌وجو، فیلتر و مقایسه کنید و مدل مناسب را برای Gateway و محصول خود انتخاب کنید.",
  path: "/features/model-discovery",
});

export default function Page() {
  return <ModelDiscovery />;
}
