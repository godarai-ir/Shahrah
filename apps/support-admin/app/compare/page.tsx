"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Fragment } from "react";
import { Suspense } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { useCompareModels } from "@/components/compare/compare-provider";
import { CompareCell, compareListValues, compareScalarValues } from "@/components/compare/compare-cell";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { apiGet } from "@/lib/api";
import { type CompareModel, MAX_COMPARE_MODELS, normalizeNameList, parseCompareIds, serializeCompareIds } from "@/lib/compare-models";
import { formatDisplayDate, formatDisplayPrice } from "@/lib/display-settings";
import { useDisplaySettings } from "@/components/display-settings-provider";

type FieldKind = "scalar" | "list";

type CompareField = {
  label: string;
  section: "Identity" | "Pricing" | "Capabilities" | "Limits" | "Governance" | "Regions & Benchmarks";
  kind: FieldKind;
  getValue: (model: CompareModel) => string;
  getList?: (model: CompareModel) => string[];
};

const EMPTY = "-";

function yesNo(value: boolean | null | undefined): string {
  if (value === null || value === undefined) return EMPTY;
  return value ? "Yes" : "No";
}

function toNumber(value: number | null | undefined): string {
  if (value === null || value === undefined) return EMPTY;
  return new Intl.NumberFormat("en-US").format(value);
}

function ComparePageContent() {
  const router = useRouter();
  const pathname = usePathname();
  const params = useSearchParams();
  const urlIds = parseCompareIds(params.get("ids"));
  const { selectedIds, selectedCount, setFromIds, remove, clear } = useCompareModels();
  const { settings } = useDisplaySettings();
  const [onlyDiffs, setOnlyDiffs] = useState(false);
  const hydratedRef = useRef(false);

  useEffect(() => {
    if (hydratedRef.current) return;
    if (urlIds.length > 0) {
      setFromIds(urlIds);
    }
    hydratedRef.current = true;
  }, [setFromIds, urlIds]);

  useEffect(() => {
    if (!hydratedRef.current) return;
    const currentIdsParam = params.get("ids");
    const nextIdsParam = selectedIds.length > 0 ? serializeCompareIds(selectedIds) : null;
    if ((currentIdsParam ?? null) === (nextIdsParam ?? null)) return;
    const next = new URLSearchParams(params.toString());
    if (nextIdsParam) {
      next.set("ids", nextIdsParam);
    } else {
      next.delete("ids");
    }
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

  const fields: CompareField[] = useMemo(
    () => [
      { label: "Model name", section: "Identity", kind: "scalar", getValue: (m) => m.displayName || EMPTY },
      { label: "Provider", section: "Identity", kind: "scalar", getValue: (m) => m.provider?.name || m.provider?.slug || EMPTY },
      { label: "Series", section: "Identity", kind: "scalar", getValue: (m) => m.modelSeries?.name || m.modelSeries?.slug || EMPTY },
      { label: "Status", section: "Identity", kind: "scalar", getValue: (m) => m.status || EMPTY },
      { label: "Release date", section: "Identity", kind: "scalar", getValue: (m) => formatDisplayDate(m.releaseDate, settings) },

      { label: "Input price", section: "Pricing", kind: "scalar", getValue: (m) => formatDisplayPrice(m.pricing?.[0]?.inputPricePerMillionTokens, settings) },
      { label: "Output price", section: "Pricing", kind: "scalar", getValue: (m) => formatDisplayPrice(m.pricing?.[0]?.outputPricePerMillionTokens, settings) },
      { label: "Cached input", section: "Pricing", kind: "scalar", getValue: (m) => formatDisplayPrice(m.pricing?.[0]?.cachedInputPricePerMillionTokens, settings) },

      {
        label: "Input modalities",
        section: "Capabilities",
        kind: "list",
        getValue: (m) => normalizeNameList((m.inputModalities ?? []).map((x) => x.modality ?? null)).join(", ") || EMPTY,
        getList: (m) => normalizeNameList((m.inputModalities ?? []).map((x) => x.modality ?? null)),
      },
      {
        label: "Output modalities",
        section: "Capabilities",
        kind: "list",
        getValue: (m) => normalizeNameList((m.outputModalities ?? []).map((x) => x.modality ?? null)).join(", ") || EMPTY,
        getList: (m) => normalizeNameList((m.outputModalities ?? []).map((x) => x.modality ?? null)),
      },
      {
        label: "Capabilities",
        section: "Capabilities",
        kind: "list",
        getValue: (m) => normalizeNameList((m.capabilities ?? []).map((x) => x.capability ?? null)).join(", ") || EMPTY,
        getList: (m) => normalizeNameList((m.capabilities ?? []).map((x) => x.capability ?? null)),
      },
      {
        label: "Supported params",
        section: "Capabilities",
        kind: "list",
        getValue: (m) => normalizeNameList((m.supportedParams ?? []).map((x) => x.supportedParameter ?? null)).join(", ") || EMPTY,
        getList: (m) => normalizeNameList((m.supportedParams ?? []).map((x) => x.supportedParameter ?? null)),
      },

      { label: "Context window", section: "Limits", kind: "scalar", getValue: (m) => toNumber(m.contextWindow) },
      { label: "Max output tokens", section: "Limits", kind: "scalar", getValue: (m) => toNumber(m.maxOutputTokens) },

      { label: "Open-source", section: "Governance", kind: "scalar", getValue: (m) => yesNo(m.openSource) },
      { label: "Free", section: "Governance", kind: "scalar", getValue: (m) => yesNo(m.free) },
      { label: "License", section: "Governance", kind: "scalar", getValue: (m) => m.license || EMPTY },
      { label: "Docs available", section: "Governance", kind: "scalar", getValue: (m) => yesNo(Boolean(m.officialDocsUrl)) },

      {
        label: "Regions",
        section: "Regions & Benchmarks",
        kind: "list",
        getValue: (m) => normalizeNameList((m.regions ?? []).map((x) => x.region ?? null)).join(", ") || EMPTY,
        getList: (m) => normalizeNameList((m.regions ?? []).map((x) => x.region ?? null)),
      },
      {
        label: "Benchmarks summary",
        section: "Regions & Benchmarks",
        kind: "list",
        getValue: (m) =>
          (m.benchmarks ?? [])
            .map((b) => {
              const name = b.benchmark?.name || b.benchmark?.slug;
              if (!name) return null;
              const score = b.score ?? b.percentile;
              return score === null || score === undefined ? name : `${name}: ${score}`;
            })
            .filter((value): value is string => Boolean(value))
            .join(", ") || EMPTY,
        getList: (m) =>
          (m.benchmarks ?? [])
            .map((b) => {
              const name = b.benchmark?.name || b.benchmark?.slug;
              if (!name) return null;
              const score = b.score ?? b.percentile;
              return score === null || score === undefined ? name : `${name}:${score}`;
            })
            .filter((value): value is string => Boolean(value)),
      },
    ],
    [settings],
  );

  const sections = ["Identity", "Pricing", "Capabilities", "Limits", "Governance", "Regions & Benchmarks"] as const;

  if (selectedCount === 0) {
    return (
      <div className="space-y-3">
        <h1 className="text-2xl font-semibold">Compare Models</h1>
        <p className="text-zinc-400">No models selected. Add models from catalog or model detail pages to start comparing.</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-2">
        <h1 className="text-2xl font-semibold">Compare Models</h1>
        <Badge>{selectedCount}/{MAX_COMPARE_MODELS}</Badge>
        <label className="ml-auto flex items-center gap-2 text-sm text-zinc-300">
          <input
            type="checkbox"
            checked={onlyDiffs}
            onChange={(e) => setOnlyDiffs(e.target.checked)}
            aria-label="Show only different fields"
          />
          Only differences
        </label>
        <Button type="button" className="h-8 bg-zinc-800 hover:bg-zinc-700" onClick={clear} aria-label="Clear all selected compare models">Clear all</Button>
      </div>

      {selectedCount === 1 ? <p className="text-sm text-amber-300">Add at least one more model to see meaningful differences.</p> : null}
      {missingIds.length > 0 ? <p className="text-sm text-amber-300">Some models were not found or are no longer available: {missingIds.join(", ")}</p> : null}
      {compareQuery.isError ? <p className="text-sm text-red-400">Failed to load compare data.</p> : null}
      {compareQuery.isLoading ? <p className="text-sm text-zinc-400">Loading compare data...</p> : null}

      <div className="overflow-x-auto rounded-xl border border-zinc-800">
        <table className="min-w-[900px] text-sm">
          <thead className="bg-zinc-900">
            <tr>
              <th className="min-w-48 p-2 text-left">Field</th>
              {selectedIds.map((id) => (
                <th key={id} className="min-w-64 p-2 text-left align-top">
                  <div className="space-y-2">
                    <p className="text-zinc-100">{modelById.get(id)?.displayName || id}</p>
                    <p className="text-xs text-zinc-500">{id}</p>
                    <Button type="button" className="h-7 bg-zinc-800 px-2 text-xs hover:bg-zinc-700" onClick={() => remove(id)} aria-label={`Remove ${id} from compare`}>Remove</Button>
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {sections.map((section) => {
              const sectionFields = fields.filter((f) => f.section === section);
              const renderedFields = sectionFields.filter((field) => {
                if (!onlyDiffs) return true;
                if (field.kind === "list") {
                  const values = rows.map((model) => (model ? (field.getList?.(model) ?? []) : []));
                  return compareListValues(values);
                }
                return compareScalarValues(rows.map((model) => (model ? field.getValue(model) : EMPTY)));
              });

              if (renderedFields.length === 0) return null;

              return (
                <Fragment key={section}>
                  <tr key={`${section}-header`} className="border-t border-zinc-800 bg-zinc-900/60">
                    <td className="p-2 font-semibold text-zinc-200" colSpan={selectedCount + 1}>{section}</td>
                  </tr>
                  {renderedFields.map((field) => {
                    const scalarValues = rows.map((model) => (model ? field.getValue(model) : EMPTY));
                    const listValues = rows.map((model) => (model ? (field.getList?.(model) ?? []) : []));
                    const isDifferent = field.kind === "list" ? compareListValues(listValues) : compareScalarValues(scalarValues);
                    return (
                      <tr key={`${section}-${field.label}`} className="border-t border-zinc-800">
                        <td className="p-2 font-medium text-zinc-300">{field.label}</td>
                        {rows.map((model, index) => (
                          <td key={`${field.label}-${selectedIds[index]}`} className="p-2 align-top">
                            <CompareCell value={model ? field.getValue(model) : EMPTY} isDifferent={isDifferent} />
                          </td>
                        ))}
                      </tr>
                    );
                  })}
                </Fragment>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default function ComparePage() {
  return (
    <Suspense fallback={<p className="text-sm text-zinc-400">Loading compare page...</p>}>
      <ComparePageContent />
    </Suspense>
  );
}
