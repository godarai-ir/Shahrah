import type { Metadata } from "next";
import { ModelDiscovery } from "@/app/pages/features/ModelDiscovery";
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "کشف و مقایسه مدل‌ها",
  description:
    "مدل‌ها و ارائه‌دهنده‌ها را جست‌وجو، فیلتر و مقایسه کنید و پیش از اتصال به درگاه، هزینه و قابلیت‌های هر مدل را بسنجید.",
  path: "/features/model-discovery",
});

export default function Page() {
  return <ModelDiscovery />;
}
