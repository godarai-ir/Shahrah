"use client";

import { Fragment, Suspense, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { CircleAlert, GitCompareArrows, ListFilter, Trash2, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { DisplayPrice } from "@/components/catalog/display-price";
import { ModalityFlow } from "@/components/catalog/modality-icon-list";
import { useCompareModels } from "@/components/compare/compare-provider";
import { CompareCell, compareListValues, compareScalarValues } from "@/components/compare/compare-cell";
import { Badge } from "@/components/ui/badge";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { apiGet } from "@/lib/api";
import { type CompareModel, MAX_COMPARE_MODELS, normalizeNameList, parseCompareIds, serializeCompareIds } from "@/lib/compare-models";
import { cn } from "@/lib/utils";
import { useDisplayFormatters } from "@/lib/use-display-formatters";
import { PageHero } from "@/components/page-hero";

type FieldKind = "scalar" | "list";
type SectionKey = "identity" | "pricing" | "capabilities" | "limits" | "governance" | "regionsBenchmarks";

type CompareField = {
  key: string;
  label: string;
  section: SectionKey;
  kind: FieldKind;
  getValue: (model: CompareModel) => string;
  renderValue?: (model: CompareModel) => React.ReactNode;
  getList?: (model: CompareModel) => string[];
};

type StatusKey = "active" | "inactive" | "disabled" | "maintenance" | "deprecated" | "preview" | "available" | "unavailable";

const EMPTY = "-";
const sectionKeys: SectionKey[] = ["identity", "pricing", "capabilities", "limits", "governance", "regionsBenchmarks"];

function getStatusKey(status: string): StatusKey | null {
  const normalized = status.toLowerCase().replaceAll("_", "-");
  const keys: StatusKey[] = ["active", "inactive", "disabled", "maintenance", "deprecated", "preview", "available", "unavailable"];
  return keys.includes(normalized as StatusKey) ? (normalized as StatusKey) : null;
}

function ComparePageContent() {
  const t = useTranslations("Compare");
  const tFields = useTranslations("Compare.fields");
  const tSections = useTranslations("Compare.sections");
  const tCommon = useTranslations("Common");
  const tStatus = useTranslations("Statuses");
  const { formatDate, formatPrice } = useDisplayFormatters();
  const numberFormatter = useMemo(() => new Intl.NumberFormat("fa-IR"), []);
  const router = useRouter();
  const pathname = usePathname();
  const params = useSearchParams();
  const urlIds = parseCompareIds(params.get("ids"));
  const { selectedIds, selectedCount, setFromIds, remove, clear } = useCompareModels();
  const [onlyDiffs, setOnlyDiffs] = useState(false);
  const hydratedRef = useRef(false);
  const initialUrlIdsRef = useRef(urlIds);
  const urlSyncReadyRef = useRef(urlIds.length === 0);

  useEffect(() => {
    if (hydratedRef.current) return;
    if (initialUrlIdsRef.current.length > 0) setFromIds(initialUrlIdsRef.current);
    hydratedRef.current = true;
  }, [setFromIds]);

  useEffect(() => {
    if (!hydratedRef.current) return;
    if (!urlSyncReadyRef.current) {
      const initialIdsParam = serializeCompareIds(initialUrlIdsRef.current);
      if (serializeCompareIds(selectedIds) !== initialIdsParam) return;
      urlSyncReadyRef.current = true;
    }

    const currentIdsParam = params.get("ids");
    const nextIdsParam = selectedIds.length > 0 ? serializeCompareIds(selectedIds) : null;
    if ((currentIdsParam ?? null) === (nextIdsParam ?? null)) return;
    const next = new URLSearchParams(params.toString());
    if (nextIdsParam) next.set("ids", nextIdsParam);
    else next.delete("ids");
    const nextUrl = next.toString() ? `${pathname}?${next.toString()}` : pathname;
    router.replace(nextUrl);
  }, [params, pathname, router, selectedIds]);

  const ids = selectedIds.join(",");
  const compareQuery = useQuery({
    queryKey: ["models-compare", ids],
    queryFn: () => apiGet<CompareModel[]>(`/models/compare?ids=${encodeURIComponent(ids)}`),
    enabled: selectedIds.length > 0,
  });

  const modelById = useMemo(() => {
    const map = new Map<string, CompareModel>();
    for (const row of compareQuery.data ?? []) map.set(row.canonicalId, row);
    return map;
  }, [compareQuery.data]);

  const rows = useMemo(() => selectedIds.map((id) => modelById.get(id) ?? null), [modelById, selectedIds]);
  const missingIds = useMemo(() => selectedIds.filter((id) => !modelById.has(id)), [modelById, selectedIds]);
  const yesNo = (value: boolean | null | undefined) => value === null || value === undefined ? EMPTY : value ? tCommon("yes") : tCommon("no");
  const toNumber = (value: number | null | undefined) => value === null || value === undefined ? EMPTY : numberFormatter.format(value);
  const formatStatus = (status: string | null | undefined) => {
    if (!status) return EMPTY;
    const key = getStatusKey(status);
    return key ? tStatus(key) : status;
  };

  const fields: CompareField[] = [
    {
      key: "provider",
      label: tFields("provider"),
      section: "identity",
      kind: "scalar",
      getValue: (model) => model.provider?.name || model.provider?.slug || EMPTY,
      renderValue: (model) => model.provider?.slug ? (
        <Link
          href={`/providers/${model.provider.slug}`}
          className="rounded-sm font-medium text-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <bdi dir="auto">{model.provider.name || model.provider.slug}</bdi>
        </Link>
      ) : EMPTY,
    },
    { key: "status", label: tFields("status"), section: "identity", kind: "scalar", getValue: (model) => formatStatus(model.status) },
    { key: "releaseDate", label: tFields("releaseDate"), section: "identity", kind: "scalar", getValue: (model) => formatDate(model.releaseDate) },
    {
      key: "inputPrice",
      label: tFields("inputPrice"),
      section: "pricing",
      kind: "scalar",
      getValue: (model) => formatPrice(model.pricing?.[0]?.inputPricePerMillionTokens),
      renderValue: (model) => <DisplayPrice value={model.pricing?.[0]?.inputPricePerMillionTokens} />,
    },
    {
      key: "outputPrice",
      label: tFields("outputPrice"),
      section: "pricing",
      kind: "scalar",
      getValue: (model) => formatPrice(model.pricing?.[0]?.outputPricePerMillionTokens),
      renderValue: (model) => <DisplayPrice value={model.pricing?.[0]?.outputPricePerMillionTokens} />,
    },
    {
      key: "cachedInput",
      label: tFields("cachedInput"),
      section: "pricing",
      kind: "scalar",
      getValue: (model) => formatPrice(model.pricing?.[0]?.cachedInputPricePerMillionTokens),
      renderValue: (model) => <DisplayPrice value={model.pricing?.[0]?.cachedInputPricePerMillionTokens} />,
    },
    {
      key: "modalities",
      label: tFields("modalities"),
      section: "capabilities",
      kind: "list",
      getValue: (model) => {
        const inputValues = normalizeNameList((model.inputModalities ?? []).map((item) => item.modality ?? null));
        const outputValues = normalizeNameList((model.outputModalities ?? []).map((item) => item.modality ?? null));
        return [...inputValues.map((value) => `input:${value}`), ...outputValues.map((value) => `output:${value}`)].join("، ") || EMPTY;
      },
      getList: (model) => {
        const inputValues = normalizeNameList((model.inputModalities ?? []).map((item) => item.modality ?? null));
        const outputValues = normalizeNameList((model.outputModalities ?? []).map((item) => item.modality ?? null));
        return [...inputValues.map((value) => `input:${value}`), ...outputValues.map((value) => `output:${value}`)];
      },
      renderValue: (model) => {
        const inputValues = normalizeNameList((model.inputModalities ?? []).map((item) => item.modality ?? null));
        const outputValues = normalizeNameList((model.outputModalities ?? []).map((item) => item.modality ?? null));
        return (
          <ModalityFlow
            inputValues={inputValues}
            outputValues={outputValues}
            inputLabel={tFields("inputModalities")}
            outputLabel={tFields("outputModalities")}
            emptyLabel={EMPTY}
          />
        );
      },
    },
    {
      key: "capabilities",
      label: tFields("capabilities"),
      section: "capabilities",
      kind: "list",
      getValue: (model) => normalizeNameList((model.capabilities ?? []).map((item) => item.capability ?? null)).join("، ") || EMPTY,
      getList: (model) => normalizeNameList((model.capabilities ?? []).map((item) => item.capability ?? null)),
    },
    {
      key: "supportedParams",
      label: tFields("supportedParams"),
      section: "capabilities",
      kind: "list",
      getValue: (model) => normalizeNameList((model.supportedParams ?? []).map((item) => item.supportedParameter ?? null)).join("، ") || EMPTY,
      getList: (model) => normalizeNameList((model.supportedParams ?? []).map((item) => item.supportedParameter ?? null)),
    },
    { key: "contextWindow", label: tFields("contextWindow"), section: "limits", kind: "scalar", getValue: (model) => toNumber(model.contextWindow) },
    { key: "maxOutputTokens", label: tFields("maxOutputTokens"), section: "limits", kind: "scalar", getValue: (model) => toNumber(model.maxOutputTokens) },
    { key: "openSource", label: tFields("openSource"), section: "governance", kind: "scalar", getValue: (model) => yesNo(model.openSource) },
    { key: "free", label: tFields("free"), section: "governance", kind: "scalar", getValue: (model) => yesNo(model.free) },
    { key: "license", label: tFields("license"), section: "governance", kind: "scalar", getValue: (model) => model.license || EMPTY },
    { key: "docsAvailable", label: tFields("docsAvailable"), section: "governance", kind: "scalar", getValue: (model) => yesNo(Boolean(model.officialDocsUrl)) },
    {
      key: "regions",
      label: tFields("regions"),
      section: "regionsBenchmarks",
      kind: "list",
      getValue: (model) => normalizeNameList((model.regions ?? []).map((item) => item.region ?? null)).join("، ") || EMPTY,
      getList: (model) => normalizeNameList((model.regions ?? []).map((item) => item.region ?? null)),
    },
  ];

  const visibleSections = sectionKeys.map((section) => {
    const sectionFields = fields.filter((field) => field.section === section);
    const visibleFields = sectionFields.filter((field) => {
      if (!onlyDiffs) return true;
      if (field.kind === "list") {
        return compareListValues(rows.map((model) => model ? field.getList?.(model) ?? [] : []));
      }
      return compareScalarValues(rows.map((model) => model ? field.getValue(model) : EMPTY));
    });
    return { section, fields: visibleFields };
  }).filter(({ fields: visibleFields }) => visibleFields.length > 0);

  return (
    <div className="flex min-w-0 flex-col gap-6 sm:gap-8">
      <PageHero
        eyebrow={t("eyebrow")}
        title={t("title")}
        description={t("subtitle")}
        icon={GitCompareArrows}
        afterTitle={selectedCount > 0 ? (
          <Badge aria-live="polite">
            {t("selectionSummary", {
              count: numberFormatter.format(selectedCount),
              max: numberFormatter.format(MAX_COMPARE_MODELS),
            })}
          </Badge>
        ) : null}
      />

      <Breadcrumb dir="rtl" className="px-1">
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink asChild>
              <Link href="/providers?section=providers">{t("breadcrumbProviders")}</Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbLink asChild>
              <Link href="/models">{t("breadcrumbModels")}</Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{t("title")}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      {selectedCount === 0 ? (
        <Card role="status" className="p-0">
          <CardHeader className="items-start sm:flex-row sm:items-center sm:gap-4">
            <span className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-muted text-muted-foreground">
              <GitCompareArrows className="size-5" aria-hidden="true" />
            </span>
            <div className="flex min-w-0 flex-col gap-1">
              <h2 className="text-base font-semibold leading-7">{t("noSelectionTitle")}</h2>
              <CardDescription>{t("noSelection")}</CardDescription>
            </div>
            <Link
              href="/models"
              className="inline-flex shrink-0 items-center justify-center rounded-md bg-primary px-3 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring sm:ms-auto"
            >
              {t("browseModels")}
            </Link>
          </CardHeader>
        </Card>
      ) : (
        <>
          <div className="flex flex-col gap-3">
            {selectedCount === 1 ? (
              <CompareNotice tone="warning">{t("oneModelHint")}</CompareNotice>
            ) : null}
            {!compareQuery.isLoading && !compareQuery.isError && missingIds.length > 0 ? (
              <CompareNotice tone="warning">
                {t("missingModels")} {" "}
                <bdi dir="ltr" data-technical>{missingIds.join(", ")}</bdi>
              </CompareNotice>
            ) : null}
            {compareQuery.isError ? (
              <CompareNotice tone="error">{t("loadError")}</CompareNotice>
            ) : null}
          </div>

          <Card className="overflow-hidden p-0">
            <CardContent className="p-0 sm:p-0">
              {compareQuery.isLoading ? (
                <CompareTableSkeleton label={t("loading")} />
              ) : null}

              {!compareQuery.isLoading && !compareQuery.isError ? (
                <div className="max-h-[70dvh] overflow-auto" role="region" aria-label={t("tableTitle")} tabIndex={0}>
                  <table
                    className="w-full table-fixed border-collapse text-sm"
                    style={{ minWidth: `${11 + selectedCount * 15}rem` }}
                  >
                    <caption className="sr-only">{t("tableCaption")}</caption>
                    <thead className="sticky top-0 z-20 bg-muted shadow-sm">
                      <tr>
                        <th scope="col" className="sticky start-0 z-30 w-44 bg-muted p-3 text-start align-top">
                          <span className="sr-only">{tCommon("field")}</span>
                          <div className="flex items-center justify-center gap-1">
                            <Button
                              type="button"
                              size="sm"
                              className="px-2"
                              aria-pressed={onlyDiffs}
                              aria-label={t("onlyDifferencesAria")}
                              variant={onlyDiffs ? "secondary" : "outline"}
                              onClick={() => setOnlyDiffs((current) => !current)}
                            >
                              <ListFilter data-icon="inline-start" aria-hidden="true" />
                              {t("onlyDifferences")}
                            </Button>
                            <Button
                              type="button"
                              size="sm"
                              className="px-2"
                              variant="outline"
                              onClick={clear}
                              aria-label={t("clearSelection")}
                            >
                              <Trash2 data-icon="inline-start" aria-hidden="true" />
                              {t("clearAll")}
                            </Button>
                          </div>
                        </th>
                        {selectedIds.map((id) => (
                          <th key={id} scope="col" className="w-60 bg-muted p-3 text-start align-top">
                            <div className="flex min-w-0 items-center gap-2">
                              <Link
                                href={`/models/${id}`}
                                dir="auto"
                                className="line-clamp-2 min-w-0 flex-1 rounded-sm font-semibold leading-6 hover:text-primary hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                              >
                                {modelById.get(id)?.displayName || id}
                              </Link>
                              <Button
                                type="button"
                                variant="ghost"
                                size="icon-sm"
                                className="shrink-0"
                                onClick={() => remove(id)}
                                aria-label={t("removeAria", { id })}
                                title={t("removeAria", { id })}
                              >
                                <X aria-hidden="true" />
                              </Button>
                            </div>
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {visibleSections.length === 0 ? (
                        <tr className="border-t border-border">
                          <td colSpan={selectedCount + 1} className="p-6 text-center text-sm text-muted-foreground">
                            {t("noDifferences")}
                          </td>
                        </tr>
                      ) : null}
                      {visibleSections.map(({ section, fields: visibleFields }) => (
                        <Fragment key={section}>
                          <tr className="border-t border-border bg-muted/50">
                            <th scope="rowgroup" className="p-3 text-start font-semibold" colSpan={selectedCount + 1}>
                              {tSections(section)}
                            </th>
                          </tr>
                          {visibleFields.map((field) => {
                            const scalarValues = rows.map((model) => model ? field.getValue(model) : EMPTY);
                            const listValues = rows.map((model) => model ? field.getList?.(model) ?? [] : []);
                            const isDifferent = field.kind === "list"
                              ? compareListValues(listValues)
                              : compareScalarValues(scalarValues);

                            return (
                              <tr key={`${section}-${field.key}`} className="group border-t border-border hover:bg-muted/30">
                                <th
                                  scope="row"
                                  className="sticky start-0 z-10 bg-card p-3 text-start font-medium text-muted-foreground group-hover:bg-muted/30"
                                >
                                  {field.label}
                                </th>
                                {rows.map((model, index) => (
                                  <td key={`${field.key}-${selectedIds[index]}`} className="p-2 align-top sm:p-3">
                                    <CompareCell
                                      value={model ? field.renderValue?.(model) ?? field.getValue(model) : EMPTY}
                                      isDifferent={isDifferent}
                                    />
                                  </td>
                                ))}
                              </tr>
                            );
                          })}
                        </Fragment>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : null}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}

function CompareNotice({
  tone,
  children,
}: {
  tone: "warning" | "error";
  children: React.ReactNode;
}) {
  return (
    <Card role={tone === "error" ? "alert" : "status"} className="p-0">
      <CardHeader className="flex-row items-center gap-3 py-3 sm:py-3">
        <span className={cn(
          "flex size-9 shrink-0 items-center justify-center rounded-lg",
          tone === "warning" && "bg-warning-soft text-warning-soft-foreground",
          tone === "error" && "bg-destructive-soft text-destructive-soft-foreground",
        )}>
          <CircleAlert className="size-4" aria-hidden="true" />
        </span>
        <p className="text-sm font-medium leading-6">{children}</p>
      </CardHeader>
    </Card>
  );
}

function CompareTableSkeleton({ label }: { label: string }) {
  return (
    <div className="flex flex-col gap-3 p-4 sm:p-5" role="status" aria-label={label}>
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {Array.from({ length: 4 }, (_, index) => <Skeleton key={index} className="h-16" />)}
      </div>
      {Array.from({ length: 6 }, (_, index) => <Skeleton key={index} className="h-10" />)}
    </div>
  );
}

export default function ComparePage() {
  const t = useTranslations("Compare");
  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:py-10">
      <Suspense fallback={<CompareTableSkeleton label={t("pageLoading")} />}>
        <ComparePageContent />
      </Suspense>
    </div>
  );
}
