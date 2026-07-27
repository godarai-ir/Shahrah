"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

type Blocker = { code: string; message: string; region?: string };
type PricingProfile = {
  region: string;
  currency: string;
  effectiveFrom?: string | null;
  effectiveTo?: string | null;
  sourceUrl?: string | null;
  rates: Array<Record<string, unknown>>;
};
type ModelRow = {
  canonicalId: string;
  displayName: string;
  mode: string[];
  enabled: boolean;
  lifecycleStatus: string;
  pricingStatus: string;
  pricingAuthority: string;
  provider: { slug: string; name: string };
  pricingProfiles: PricingProfile[];
  publication: {
    state: string;
    eligible: boolean;
    blockers: Blocker[];
  };
};
type ModelsResponse = {
  data: ModelRow[];
  pagination: { page: number; total: number; totalPages: number };
  summary: Record<string, number>;
  providerSummary: Record<string, number>;
};

const publicationOptions = [
  ["needs_pricing", "Needs pricing"],
  ["ready_to_publish", "Ready to publish"],
  ["published", "Published"],
  ["retired", "Retired"],
] as const;

export default function AdminModelsPage() {
  const [payload, setPayload] = useState<ModelsResponse>({
    data: [],
    pagination: { page: 1, total: 0, totalPages: 1 },
    summary: {},
    providerSummary: {},
  });
  const [query, setQuery] = useState("");
  const [readiness, setReadiness] = useState("needs_pricing");
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<ModelRow | null>(null);
  const [pricingJson, setPricingJson] = useState("");
  const [notice, setNotice] = useState("");
  const [pending, setPending] = useState(false);

  const load = useCallback(async () => {
    const params = new URLSearchParams({
      page: String(page),
      limit: "50",
      mode: "image",
    });
    if (query.trim()) params.set("q", query.trim());
    if (readiness) params.set("readiness", readiness);
    const response = await fetch(`/api/admin/models?${params}`, {
      cache: "no-store",
    });
    if (!response.ok) {
      setNotice("Failed to load model publication queue.");
      return;
    }
    setPayload((await response.json()) as ModelsResponse);
  }, [page, query, readiness]);

  useEffect(() => {
    void load();
  }, [load]);

  const providerCounts = useMemo(
    () =>
      Object.entries(payload.providerSummary ?? {}).sort((a, b) => b[1] - a[1]),
    [payload.providerSummary],
  );

  function choose(row: ModelRow) {
    setSelected(row);
    setPricingJson(
      JSON.stringify(
        {
          pricingStatus: row.pricingStatus,
          pricingProfiles: row.pricingProfiles,
        },
        null,
        2,
      ),
    );
    setNotice("");
  }

  async function action(kind: "publish" | "unpublish") {
    if (!selected) return;
    if (
      kind === "publish" &&
      !window.confirm(`Publish ${selected.displayName}?`)
    ) {
      return;
    }
    setPending(true);
    const response = await fetch(
      `/api/admin/models/${encodeURIComponent(selected.canonicalId)}/${kind}`,
      { method: "POST" },
    );
    const result = await response.json().catch(() => ({}));
    setPending(false);
    if (!response.ok) {
      setNotice(
        result?.message?.blockers?.map((item: Blocker) => item.message).join(" ") ||
          result?.message ||
          `Unable to ${kind} model.`,
      );
      return;
    }
    setSelected(result as ModelRow);
    const reconcile = await fetch(
      "/api/admin/control-plane/provider-catalog/reconcile",
      { method: "POST" },
    );
    setNotice(
      reconcile.ok
        ? kind === "publish"
          ? "Model published and runtime catalog reconciled."
          : "Model unpublished and runtime catalog reconciled."
        : `${kind === "publish" ? "Model published" : "Model unpublished"}; runtime reconciliation is pending.`,
    );
    await load();
  }

  async function savePricing() {
    if (!selected) return;
    let body: unknown;
    try {
      body = JSON.parse(pricingJson);
    } catch {
      setNotice("Pricing JSON is invalid.");
      return;
    }
    setPending(true);
    const response = await fetch(
      `/api/admin/models/${encodeURIComponent(selected.canonicalId)}/pricing`,
      {
        method: "PUT",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(body),
      },
    );
    const result = await response.json().catch(() => ({}));
    setPending(false);
    if (!response.ok) {
      setNotice(result?.message || "Unable to save pricing.");
      return;
    }
    choose(result as ModelRow);
    setNotice(
      result.publication?.eligible
        ? "Pricing saved. Model is ready to publish."
        : "Draft saved. Pricing still needs completion.",
    );
    await load();
  }

  async function releasePricingAuthority() {
    if (!selected || !window.confirm("Return pricing control to the YAML importer? The model will remain unpublished until the next sync.")) return;
    setPending(true);
    const response = await fetch(
      `/api/admin/models/${encodeURIComponent(selected.canonicalId)}/pricing-authority`,
      { method: "POST" },
    );
    const result = await response.json().catch(() => ({}));
    setPending(false);
    if (!response.ok) {
      setNotice(result?.message || "Unable to release pricing authority.");
      return;
    }
    choose(result as ModelRow);
    setNotice("Pricing control returned to the importer; model remains unpublished.");
    await load();
  }

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-sm text-zinc-400">Image model publication</p>
          <h1 className="text-2xl font-semibold">Model readiness queue</h1>
        </div>
        <div className="flex gap-2">
          <Link
            href="/admin/models/new"
            className="rounded-md bg-cyan-500 px-3 py-2 text-sm font-medium text-black"
          >
            Create model
          </Link>
          <Button onClick={() => void load()}>Refresh</Button>
        </div>
      </div>

      {notice ? (
        <div className="rounded-md border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm">
          {notice}
        </div>
      ) : null}

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {publicationOptions.map(([key, label]) => (
          <button key={key} type="button" onClick={() => { setPage(1); setReadiness(key); }}>
            <Card className={readiness === key ? "border-cyan-500" : undefined}>
              <div className="text-sm text-zinc-400">{label}</div>
              <div className="mt-1 text-2xl font-semibold">
                {payload.summary[key] ?? 0}
              </div>
            </Card>
          </button>
        ))}
      </div>

      <Card>
        <div className="grid gap-3 lg:grid-cols-[minmax(0,1fr)_220px]">
          <Input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search model or canonical ID"
          />
          <div className="text-xs text-zinc-400">
            {providerCounts.map(([provider, count]) => `${provider}: ${count}`).join(" · ")}
          </div>
        </div>
      </Card>

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_460px]">
        <Card className="overflow-hidden p-0">
          <div className="divide-y divide-zinc-800">
            {payload.data.map((row) => (
              <button
                key={row.canonicalId}
                type="button"
                onClick={() => choose(row)}
                className="grid w-full gap-2 px-4 py-3 text-left hover:bg-zinc-800/60 md:grid-cols-[minmax(0,1fr)_150px_180px]"
              >
                <span>
                  <span className="block font-medium">{row.displayName}</span>
                  <span className="block text-xs text-zinc-500">{row.canonicalId}</span>
                </span>
                <span className="text-sm text-zinc-300">{row.provider.name}</span>
                <span className="text-sm text-amber-300">
                  {row.publication.blockers[0]?.message ?? row.publication.state}
                </span>
              </button>
            ))}
            {payload.data.length === 0 ? (
              <p className="px-4 py-10 text-center text-sm text-zinc-400">
                No models match this queue.
              </p>
            ) : null}
          </div>
          <div className="flex items-center justify-between border-t border-zinc-800 px-4 py-3 text-sm">
            <span>{payload.pagination.total} models</span>
            <div className="flex gap-2">
              <Button disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>Previous</Button>
              <Button disabled={page >= payload.pagination.totalPages} onClick={() => setPage((value) => value + 1)}>Next</Button>
            </div>
          </div>
        </Card>

        <Card>
          {!selected ? (
            <p className="text-sm text-zinc-400">Select a model to review pricing.</p>
          ) : (
            <div className="space-y-4">
              <div>
                <h2 className="font-semibold">{selected.displayName}</h2>
                <p className="text-xs text-zinc-500">{selected.canonicalId}</p>
                <p className="mt-2 text-xs text-zinc-400">
                  {selected.lifecycleStatus} · {selected.pricingStatus} · authority: {selected.pricingAuthority}
                </p>
              </div>
              {selected.publication.blockers.length > 0 ? (
                <ul className="space-y-1 rounded-md border border-amber-900/50 bg-amber-950/20 p-3 text-sm text-amber-200">
                  {selected.publication.blockers.map((blocker, index) => (
                    <li key={`${blocker.code}-${index}`}>
                      {blocker.region ? `${blocker.region}: ` : ""}{blocker.message}
                    </li>
                  ))}
                </ul>
              ) : null}
              <label className="block space-y-2 text-sm">
                <span>Pricing profiles</span>
                <textarea
                  className="min-h-80 w-full rounded-md border border-zinc-700 bg-zinc-950 p-3 font-mono text-xs"
                  value={pricingJson}
                  onChange={(event) => setPricingJson(event.target.value)}
                  disabled={selected.lifecycleStatus === "published"}
                />
              </label>
              <div className="flex flex-wrap gap-2">
                {selected.lifecycleStatus === "published" ? (
                  <Button disabled={pending} onClick={() => void action("unpublish")}>Unpublish to edit</Button>
                ) : (
                  <>
                    <Button disabled={pending} onClick={() => void savePricing()}>Save draft</Button>
                    <Button
                      disabled={pending || !selected.publication.eligible}
                      onClick={() => void action("publish")}
                    >
                      Publish
                    </Button>
                    {selected.pricingAuthority === "admin" ? (
                      <Button disabled={pending} onClick={() => void releasePricingAuthority()}>
                        Use imported pricing
                      </Button>
                    ) : null}
                  </>
                )}
              </div>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}
