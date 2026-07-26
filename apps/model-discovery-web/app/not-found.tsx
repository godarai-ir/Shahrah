import { getTranslations } from "next-intl/server";

import { NotFoundPage } from "@/components/not-found-page";

export default async function NotFound() {
  const t = await getTranslations("NotFound");

  return (
    <NotFoundPage
      copy={{
        eyebrow: t("eyebrow"),
        title: t("title"),
        description: t("description"),
        hint: t("hint"),
        homeAction: t("homeAction"),
        providersAction: t("providersAction"),
        brand: t("brand"),
        brandAria: t("brandAria"),
      }}
    />
  );
}
