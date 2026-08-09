"use client";

import { useTranslations } from "next-intl";

import { Badge } from "@/components/ui/badge";
import type { EffectiveModelView } from "./types";
export function EffectiveModelsTable(props: { models: EffectiveModelView[] }) {
  const ux = useTranslations("providerAccounts.ux");
  if (props.models.length === 0) {
    return <p className="text-sm text-muted-foreground">{ux("effectiveModels.empty")}</p>;
  }
  return (
    <div className="overflow-hidden rounded-lg border">
      <div className="grid grid-cols-[1.2fr_1fr_120px_120px] border-b bg-muted/30 px-3 py-2 text-xs font-semibold text-muted-foreground">
        <div>{ux("effectiveModels.columns.model")}</div>
        <div>{ux("effectiveModels.columns.capabilities")}</div>
        <div>{ux("effectiveModels.columns.pricing")}</div>
        <div>{ux("effectiveModels.columns.status")}</div>
      </div>
      {props.models.map((item) => (
        <div key={item.model} className="grid grid-cols-[1.2fr_1fr_120px_120px] items-center border-b px-3 py-2 text-sm last:border-b-0">
          <div className="truncate font-medium">{item.model}</div>
          <div className="flex flex-wrap gap-1">
            {(item.types ?? []).map((type) => (
              <Badge key={`${item.model}-${type}`} variant="secondary">{type}</Badge>
            ))}
          </div>
          <div className="flex items-center gap-1">
            <span>{item.pricing_mode}</span>
            <Badge variant={item.pricing_resolved ? "default" : "outline"}>
              {item.pricing_resolved ? ux("effectiveModels.resolved") : ux("effectiveModels.unresolved")}
            </Badge>
          </div>
          <div>
            <Badge variant={item.allowed ? "default" : "destructive"}>
              {item.allowed ? ux("effectiveModels.allowed") : ux("effectiveModels.denied")}
            </Badge>
            {!item.allowed && item.denied_reason ? (
              <p className="mt-1 truncate text-xs text-muted-foreground">{item.denied_reason}</p>
            ) : null}
          </div>
        </div>
      ))}
    </div>
  );
}


