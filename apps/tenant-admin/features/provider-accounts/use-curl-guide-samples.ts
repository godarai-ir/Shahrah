"use client";

import { useTranslations } from "next-intl";
import { useMemo } from "react";

import type { CurlGuideSamples } from "@/features/provider-accounts/utils";

/** Localized sample request bodies for the copy-ready API guides. */
export function useCurlGuideSamples(): CurlGuideSamples {
  const t = useTranslations("providerAccounts");

  return useMemo<CurlGuideSamples>(
    () => ({
      textPrompt: t("curlGuideSamples.textPrompt"),
      embeddingInput: t("curlGuideSamples.embeddingInput"),
      moderationInput: t("curlGuideSamples.moderationInput"),
      imagePrompt: t("curlGuideSamples.imagePrompt"),
      speechInput: t("curlGuideSamples.speechInput"),
      transcriptionText: t("curlGuideSamples.transcriptionText"),
    }),
    [t],
  );
}
