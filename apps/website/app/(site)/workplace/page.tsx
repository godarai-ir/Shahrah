import type { Metadata } from "next";
import { Workplace } from "@/app/pages/Workplace";
import { buildMetadata } from "@/app/lib/seo";

export const metadata: Metadata = buildMetadata({
  title: "هوش مصنوعی در کار روزانه‌ی سازمان",
  description:
    "گدارAI هوش مصنوعی را به دست همه‌ی تیم‌ها می‌رساند؛ امن، قابل کنترل و متصل به ابزارهای کار روزانه، با پرداخت تومانی.",
  path: "/workplace",
  keywords: ["پذیرش هوش مصنوعی در سازمان", "هوش مصنوعی در محیط کار", "دستیار هوش مصنوعی سازمانی", "گدارAI"],
  noIndex: true,
});

export default function Page() {
  return <Workplace />;
}
