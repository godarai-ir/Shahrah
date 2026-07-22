import type { Metadata } from "next";
import { getTranslations } from "next-intl/server";

export async function generateMetadata(): Promise<Metadata> {
  const t = await getTranslations("Compare");

  return {
    title: t("metadataTitle"),
    description: t("metadataDescription"),
  };
}

export default function CompareLayout({ children }: { children: React.ReactNode }) {
  return children;
}
