"use client";

import { useTranslations } from "next-intl";

import { AppCard } from "@/components/app-ui/AppCard";
import { CopyableValue } from "@/components/app-ui/CopyableValue";

export type IntegrationDetailsProps = {
  /** The tenant's issuer URL, which is also the gateway origin. */
  issuerURL: string;
  onCopyError: () => void;
};

/** The values an administrator copies into a relying party such as LibreChat. */
export function IntegrationDetails({ issuerURL, onCopyError }: IntegrationDetailsProps) {
  const t = useTranslations("oidcClients");
  const values = [
    { key: "issuer", label: t("integration.issuer"), value: issuerURL },
    { key: "discovery", label: t("integration.discovery"), value: `${issuerURL}/.well-known/openid-configuration` },
    { key: "gateway", label: t("integration.gateway"), value: `${issuerURL}/v1` },
  ];

  return (
    <AppCard title={t("integration.title")} description={t("integration.description")}>
      <dl className="grid gap-4 lg:grid-cols-3">
        {values.map((item) => (
          <div key={item.key} className="flex min-w-0 flex-col gap-1.5">
            <dt className="text-xs text-muted-foreground">{item.label}</dt>
            <dd className="min-w-0">
              <CopyableValue
                value={item.value}
                labels={{
                  valueAriaLabel: item.label,
                  copy: t("actions.copy"),
                  copying: t("actions.copying"),
                  copied: t("actions.copied"),
                }}
                onCopyError={onCopyError}
              />
            </dd>
          </div>
        ))}
      </dl>
    </AppCard>
  );
}
