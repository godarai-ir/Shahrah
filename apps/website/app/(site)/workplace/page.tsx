import type { Metadata } from "next";
import { Workplace } from "@/app/pages/Workplace";
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "فضای کاری هوش مصنوعی برای سازمان‌ها",
  description:
    "گفتگو با هر مدل، دستیار برای هر تیم، اتصال به ابزارهای سازمان و کنترل کامل دسترسی و هزینه.",
  path: "/workplace",
  keywords: ["پذیرش هوش مصنوعی در سازمان", "هوش مصنوعی در محیط کار", "دستیار هوش مصنوعی سازمانی", "گدارAI"],
  noIndex: true,
});

export default function Page() {
  return <Workplace />;
}
