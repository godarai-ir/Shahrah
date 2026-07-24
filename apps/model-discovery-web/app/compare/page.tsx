"use client";

import { Fragment, Suspense, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { CircleAlert, GitCompareArrows, ListFilter, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { useCompareModels } from "@/components/compare/compare-provider";
import { CompareCell, compareListValues, compareScalarValues } from "@/components/compare/compare-cell";
import { Badge } from "@/components/ui/badge";
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
    { key: "modelName", label: tFields("modelName"), section: "identity", kind: "scalar", getValue: (model) => model.displayName || EMPTY },
    { key: "provider", label: tFields("provider"), section: "identity", kind: "scalar", getValue: (model) => model.provider?.name || model.provider?.slug || EMPTY },
    { key: "series", label: tFields("series"), section: "identity", kind: "scalar", getValue: (model) => model.modelSeries?.name || model.modelSeries?.slug || EMPTY },
    { key: "status", label: tFields("status"), section: "identity", kind: "scalar", getValue: (model) => formatStatus(model.status) },
    { key: "releaseDate", label: tFields("releaseDate"), section: "identity", kind: "scalar", getValue: (model) => formatDate(model.releaseDate) },
    { key: "inputPrice", label: tFields("inputPrice"), section: "pricing", kind: "scalar", getValue: (model) => formatPrice(model.pricing?.[0]?.inputPricePerMillionTokens) },
    { key: "outputPrice", label: tFields("outputPrice"), section: "pricing", kind: "scalar", getValue: (model) => formatPrice(model.pricing?.[0]?.outputPricePerMillionTokens) },
    { key: "cachedInput", label: tFields("cachedInput"), section: "pricing", kind: "scalar", getValue: (model) => formatPrice(model.pricing?.[0]?.cachedInputPricePerMillionTokens) },
    {
      key: "inputModalities",
      label: tFields("inputModalities"),
      section: "capabilities",
      kind: "list",
      getValue: (model) => normalizeNameList((model.inputModalities ?? []).map((item) => item.modality ?? null)).join("، ") || EMPTY,
      getList: (model) => normalizeNameList((model.inputModalities ?? []).map((item) => item.modality ?? null)),
    },
    {
      key: "outputModalities",
      label: tFields("outputModalities"),
      section: "capabilities",
      kind: "list",
      getValue: (model) => normalizeNameList((model.outputModalities ?? []).map((item) => item.modality ?? null)).join("، ") || EMPTY,
      getList: (model) => normalizeNameList((model.outputModalities ?? []).map((item) => item.modality ?? null)),
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
    {
      key: "benchmarksSummary",
      label: tFields("benchmarksSummary"),
      section: "regionsBenchmarks",
      kind: "list",
      getValue: (model) => (model.benchmarks ?? []).map((benchmark) => {
        const name = benchmark.benchmark?.name || benchmark.benchmark?.slug;
        if (!name) return null;
        const score = benchmark.score ?? benchmark.percentile;
        return score === null || score === undefined ? name : `${name}: ${numberFormatter.format(score)}`;
      }).filter((value): value is string => Boolean(value)).join("، ") || EMPTY,
      getList: (model) => (model.benchmarks ?? []).map((benchmark) => {
        const name = benchmark.benchmark?.name || benchmark.benchmark?.slug;
        if (!name) return null;
        const score = benchmark.score ?? benchmark.percentile;
        return score === null || score === undefined ? name : `${name}:${score}`;
      }).filter((value): value is string => Boolean(value)),
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
          <Card className="p-0">
            <CardHeader className="sm:flex-row sm:items-center sm:justify-between sm:gap-4">
              <div className="flex min-w-0 flex-col gap-1">
                <h2 className="text-base font-semibold leading-7 sm:text-lg">{t("controlsTitle")}</h2>
                <CardDescription>{t("controlsDescription")}</CardDescription>
              </div>
              <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row">
                <Button
                  type="button"
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
                  variant="outline"
                  onClick={clear}
                  aria-label={t("clearSelection")}
                >
                  <Trash2 data-icon="inline-start" aria-hidden="true" />
                  {t("clearAll")}
                </Button>
              </div>
            </CardHeader>
          </Card>

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
            <CardHeader className="border-b border-border sm:flex-row sm:items-end sm:justify-between sm:gap-4">
              <div className="flex min-w-0 flex-col gap-1">
                <h2 className="text-base font-semibold sm:text-lg">{t("tableTitle")}</h2>
                <CardDescription>{t("tableDescription")}</CardDescription>
              </div>
              <p className="text-xs leading-5 text-muted-foreground lg:hidden">{t("scrollHint")}</p>
            </CardHeader>
            <CardContent className="p-0 sm:p-0">
              {compareQuery.isLoading ? (
                <CompareTableSkeleton label={t("loading")} />
              ) : null}

              {!compareQuery.isLoading && !compareQuery.isError ? (
                <div className="overflow-x-auto" role="region" aria-label={t("tableTitle")} tabIndex={0}>
                  <table
                    className="w-full table-fixed border-collapse text-sm"
                    style={{ minWidth: `${11 + selectedCount * 15}rem` }}
                  >
                    <caption className="sr-only">{t("tableCaption")}</caption>
                    <thead className="bg-muted/70">
                      <tr>
                        <th scope="col" className="sticky start-0 z-10 w-44 bg-muted p-3 text-start align-top">
                          {tCommon("field")}
                        </th>
                        {selectedIds.map((id) => (
                          <th key={id} scope="col" className="w-60 p-3 text-start align-top">
                            <div className="flex min-w-0 flex-col items-start gap-2">
                              <div className="min-w-0">
                                <p dir="auto" className="line-clamp-2 font-semibold leading-6">
                                  {modelById.get(id)?.displayName || id}
                                </p>
                                <p className="truncate text-xs font-normal text-muted-foreground" dir="ltr" data-technical title={id}>
                                  {id}
                                </p>
                              </div>
                              <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                onClick={() => remove(id)}
                                aria-label={t("removeAria", { id })}
                              >
                                <Trash2 data-icon="inline-start" aria-hidden="true" />
                                {t("removeButton")}
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
                                      value={model ? field.getValue(model) : EMPTY}
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
    <Suspense fallback={<CompareTableSkeleton label={t("pageLoading")} />}>
      <ComparePageContent />
    </Suspense>
  );
}
